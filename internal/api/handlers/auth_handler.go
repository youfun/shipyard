package handlers

import (
	"crypto/rand"
	"youfun/shipyard/internal/api/middleware"
	"youfun/shipyard/internal/api/response"
	"youfun/shipyard/internal/database"
	"encoding/base32"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Legacy function wrappers for backward compatibility
var defaultAuthRepo = &DefaultRepository{}

// DeviceAuthRequest represents a pending device authorization request
type DeviceAuthRequest struct {
	DeviceCode  string    `json:"device_code"`
	UserCode    string    `json:"user_code"`
	SessionID   string    `json:"session_id"`
	OS          string    `json:"os"`
	DeviceName  string    `json:"device_name"`
	PublicIP    string    `json:"public_ip"`
	Status      string    `json:"status"` // pending, authorized, rejected
	AccessToken string    `json:"access_token,omitempty"`
	UserID      uuid.UUID `json:"-"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// In-memory store for device auth requests (in production, use Redis or database)
var (
	deviceAuthRequests = make(map[string]*DeviceAuthRequest) // keyed by session_id
	deviceAuthMutex    sync.RWMutex
)

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	TwoFactorRequired bool   `json:"two_factor_required,omitempty"`
	Temp2FAToken      string `json:"temp_2fa_token,omitempty"`
	AccessToken       string `json:"access_token,omitempty"`
}

// Login2FARequest represents the 2FA verification request
type Login2FARequest struct {
	Temp2FAToken string `json:"temp_2fa_token" binding:"required"`
	OTP          string `json:"otp" binding:"required"`
}

// Login handles user login
func Login(c *gin.Context) {
	h := &Handlers{Repo: defaultAuthRepo}
	h.Login(c)
}

// LoginHandler handles user login (method on Handlers)
func (h *Handlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	user, err := h.Repo.GetUserByUsername(req.Username)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if !user.CheckPassword(req.Password) {
		response.Error(c, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Check if 2FA is enabled
	if user.TwoFactorEnabled {
		// Generate temporary token for 2FA
		tempToken, err := middleware.GenerateTemp2FAToken(user.ID.String(), user.Username)
		if err != nil {
			response.InternalServerError(c, "Failed to generate token")
			return
		}

		response.Data(c, LoginResponse{
			TwoFactorRequired: true,
			Temp2FAToken:      tempToken,
		})
		return
	}

	// Create device record
	tx, err := database.DB.Beginx()
	if err != nil {
		response.InternalServerError(c, "Database error")
		return
	}

	device, err := h.Repo.CreateAuthDevice(tx, user.ID, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		tx.Rollback()
		response.InternalServerError(c, "Failed to create device record")
		return
	}

	if err := tx.Commit(); err != nil {
		response.InternalServerError(c, "Database error")
		return
	}

	// Generate access token
	token, err := middleware.GenerateToken(user.ID.String(), user.Username, device.ID.String())
	if err != nil {
		response.InternalServerError(c, "Failed to generate token")
		return
	}

	response.Data(c, LoginResponse{
		AccessToken: token,
	})
}

// Login2FA handles 2FA verification
func Login2FA(c *gin.Context) {
	h := &Handlers{Repo: defaultAuthRepo}
	h.Login2FA(c)
}

// Login2FAHandler handles 2FA verification (method on Handlers)
func (h *Handlers) Login2FA(c *gin.Context) {
	var req Login2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	// Validate temp token
	claims, err := middleware.ValidateToken(req.Temp2FAToken)
	if err != nil || claims.Issuer != "deployer-2fa" {
		response.Error(c, http.StatusUnauthorized, "Invalid or expired 2FA token")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		response.InternalServerError(c, "Invalid user ID")
		return
	}

	user, err := h.Repo.GetUserByID(userID)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "User not found")
		return
	}

	// Verify OTP (placeholder - implement actual TOTP verification)
	if !verifyTOTP(user.TwoFactorSecret, req.OTP) {
		response.Error(c, http.StatusUnauthorized, "Invalid OTP")
		return
	}

	// Create device record
	tx, err := database.DB.Beginx()
	if err != nil {
		response.InternalServerError(c, "Database error")
		return
	}

	device, err := h.Repo.CreateAuthDevice(tx, user.ID, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		tx.Rollback()
		response.InternalServerError(c, "Failed to create device record")
		return
	}

	if err := tx.Commit(); err != nil {
		response.InternalServerError(c, "Database error")
		return
	}

	// Generate access token
	token, err := middleware.GenerateToken(user.ID.String(), user.Username, device.ID.String())
	if err != nil {
		response.InternalServerError(c, "Failed to generate token")
		return
	}

	response.Data(c, gin.H{
		"access_token": token,
	})
}

// GetCurrentUser returns the current authenticated user
func GetCurrentUser(c *gin.Context) {
	h := &Handlers{Repo: defaultAuthRepo}
	h.GetCurrentUser(c)
}

// GetCurrentUserHandler returns the current authenticated user (method on Handlers)
func (h *Handlers) GetCurrentUser(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid user")
		return
	}

	user, err := h.Repo.GetUserByID(userID)
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	response.Data(c, gin.H{
		"id":                 user.ID,
		"username":           user.Username,
		"two_factor_enabled": user.TwoFactorEnabled,
		"created_at":         user.CreatedAt,
	})
}

// Logout handles user logout
func Logout(c *gin.Context) {
	h := &Handlers{Repo: defaultAuthRepo}
	h.Logout(c)
}

// LogoutHandler handles user logout (method on Handlers)
func (h *Handlers) Logout(c *gin.Context) {
	// Get device ID from context
	deviceIDStr, exists := c.Get("device_id")
	if exists && deviceIDStr != "" {
		deviceID, err := uuid.Parse(deviceIDStr.(string))
		if err == nil {
			userID, _ := middleware.GetUserIDFromContext(c)
			_ = h.Repo.RevokeAuthDevice(userID, deviceID)
		}
	}

	response.Message(c, "Logged out successfully")
}

// ChangePasswordRequest represents the change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword handles password change
func ChangePassword(c *gin.Context) {
	h := &Handlers{Repo: defaultAuthRepo}
	h.ChangePassword(c)
}

// ChangePasswordHandler handles password change (method on Handlers)
func (h *Handlers) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid user")
		return
	}

	user, err := h.Repo.GetUserByID(userID)
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	if !user.CheckPassword(req.CurrentPassword) {
		response.Error(c, http.StatusUnauthorized, "Current password is incorrect")
		return
	}

	if err := h.Repo.UpdateUserPassword(userID, req.NewPassword); err != nil {
		response.InternalServerError(c, "Failed to update password")
		return
	}

	response.Message(c, "Password changed successfully")
}

// DeviceCodeRequest represents the device code request from CLI
type DeviceCodeRequest struct {
	OS         string `json:"os"`
	DeviceName string `json:"device_name"`
}

// DeviceCode handles CLI device authorization request
func DeviceCode(c *gin.Context) {
	var req DeviceCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	// Generate unique codes
	sessionID := uuid.New().String()
	userCode := generateUserCode()

	authReq := &DeviceAuthRequest{
		DeviceCode: uuid.New().String(),
		UserCode:   userCode,
		SessionID:  sessionID,
		OS:         req.OS,
		DeviceName: req.DeviceName,
		PublicIP:   c.ClientIP(),
		Status:     "pending",
		ExpiresAt:  time.Now().Add(10 * time.Minute),
		CreatedAt:  time.Now(),
	}

	deviceAuthMutex.Lock()
	deviceAuthRequests[sessionID] = authReq
	deviceAuthMutex.Unlock()

	response.Data(c, gin.H{
		"session_id":       sessionID,
		"user_code":        userCode,
		"verification_uri": "/cli-device-auth",
		"expires_in":       600,
		"interval":         5,
	})
}

// DeviceToken handles CLI polling for token
func DeviceToken(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		response.BadRequest(c, "session_id required")
		return
	}

	deviceAuthMutex.RLock()
	authReq, exists := deviceAuthRequests[sessionID]
	deviceAuthMutex.RUnlock()

	if !exists {
		response.NotFound(c, "Session not found")
		return
	}

	if time.Now().After(authReq.ExpiresAt) {
		deviceAuthMutex.Lock()
		delete(deviceAuthRequests, sessionID)
		deviceAuthMutex.Unlock()
		response.Error(c, http.StatusGone, "Session expired")
		return
	}

	switch authReq.Status {
	case "pending":
		response.Error(c, http.StatusAccepted, "authorization_pending")
	case "rejected":
		deviceAuthMutex.Lock()
		delete(deviceAuthRequests, sessionID)
		deviceAuthMutex.Unlock()
		response.Error(c, http.StatusForbidden, "Authorization denied")
	case "authorized":
		deviceAuthMutex.Lock()
		delete(deviceAuthRequests, sessionID)
		deviceAuthMutex.Unlock()
		response.Data(c, gin.H{
			"access_token": authReq.AccessToken,
		})
	}
}

// GetDeviceSession returns device session info for the authorization page
func GetDeviceSession(c *gin.Context) {
	sessionID := c.Param("sessionId")

	deviceAuthMutex.RLock()
	authReq, exists := deviceAuthRequests[sessionID]
	deviceAuthMutex.RUnlock()

	if !exists {
		response.NotFound(c, "Session not found")
		return
	}

	if time.Now().After(authReq.ExpiresAt) {
		response.Error(c, http.StatusGone, "Session expired")
		return
	}

	response.Data(c, gin.H{
		"session_id":        authReq.SessionID,
		"os":                authReq.OS,
		"device_name":       authReq.DeviceName,
		"public_ip":         authReq.PublicIP,
		"request_timestamp": authReq.CreatedAt.Unix(),
	})
}

// DeviceConfirmRequest represents the device authorization confirmation
type DeviceConfirmRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Approved  bool   `json:"approved"`
}

// DeviceConfirm handles user confirmation of device authorization
func DeviceConfirm(c *gin.Context) {
	h := &Handlers{Repo: defaultAuthRepo}
	h.DeviceConfirm(c)
}

// DeviceConfirmHandler handles user confirmation of device authorization (method on Handlers)
func (h *Handlers) DeviceConfirm(c *gin.Context) {
	var req DeviceConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	deviceAuthMutex.Lock()
	authReq, exists := deviceAuthRequests[req.SessionID]
	if !exists {
		deviceAuthMutex.Unlock()
		response.NotFound(c, "Session not found")
		return
	}

	if time.Now().After(authReq.ExpiresAt) {
		delete(deviceAuthRequests, req.SessionID)
		deviceAuthMutex.Unlock()
		response.Error(c, http.StatusGone, "Session expired")
		return
	}

	if !req.Approved {
		authReq.Status = "rejected"
		deviceAuthMutex.Unlock()
		response.Message(c, "Authorization denied")
		return
	}

	// Get current user
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		deviceAuthMutex.Unlock()
		response.Error(c, http.StatusUnauthorized, "Invalid user")
		return
	}

	user, err := h.Repo.GetUserByID(userID)
	if err != nil {
		deviceAuthMutex.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Create device record
	tx, err := database.DB.Beginx()
	if err != nil {
		deviceAuthMutex.Unlock()
		response.InternalServerError(c, "Database error")
		return
	}

	device, err := h.Repo.CreateAuthDevice(tx, user.ID, authReq.PublicIP, "CLI: "+authReq.DeviceName)
	if err != nil {
		tx.Rollback()
		deviceAuthMutex.Unlock()
		response.InternalServerError(c, "Failed to create device record")
		return
	}

	if err := tx.Commit(); err != nil {
		deviceAuthMutex.Unlock()
		response.InternalServerError(c, "Database error")
		return
	}

	// Generate token
	token, err := middleware.GenerateToken(user.ID.String(), user.Username, device.ID.String())
	if err != nil {
		deviceAuthMutex.Unlock()
		response.InternalServerError(c, "Failed to generate token")
		return
	}

	authReq.Status = "authorized"
	authReq.AccessToken = token
	authReq.UserID = user.ID
	deviceAuthMutex.Unlock()

	response.Message(c, "Authorization granted")
}

// ListDevices lists all devices for current user
func ListDevices(c *gin.Context) {
	h := &Handlers{Repo: defaultAuthRepo}
	h.ListDevices(c)
}

// ListDevicesHandler lists all devices for current user (method on Handlers)
func (h *Handlers) ListDevices(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid user")
		return
	}

	devices, err := h.Repo.ListAuthDevicesForUser(userID)
	if err != nil {
		response.InternalServerError(c, "Failed to list devices")
		return
	}

	response.Data(c, devices)
}

// RevokeDevice revokes a device session
func RevokeDevice(c *gin.Context) {
	h := &Handlers{Repo: defaultAuthRepo}
	h.RevokeDevice(c)
}

// RevokeDeviceHandler revokes a device session (method on Handlers)
func (h *Handlers) RevokeDevice(c *gin.Context) {
	deviceIDStr := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid device ID")
		return
	}

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid user")
		return
	}

	if err := h.Repo.RevokeAuthDevice(userID, deviceID); err != nil {
		response.InternalServerError(c, "Failed to revoke device")
		return
	}

	response.Message(c, "Device revoked successfully")
}

// generateUserCode generates a short user-friendly code
func generateUserCode() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	code := base32.StdEncoding.EncodeToString(bytes)[:6]
	return code[:3] + "-" + code[3:]
}

// verifyTOTP verifies a TOTP code (placeholder implementation)
func verifyTOTP(secret, code string) bool {
	// TODO: Implement actual TOTP verification using a library like pquerna/otp
	// For now, accept any 6-digit code for development
	return len(code) == 6
}
