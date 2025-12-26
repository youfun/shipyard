package handlers

import (
	"crypto/rand"
	"youfun/shipyard/internal/api/middleware"
	"youfun/shipyard/internal/api/response"
	"youfun/shipyard/internal/database"
	"encoding/base32"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Legacy function wrappers for backward compatibility
var default2FARepo = &DefaultRepository{}

// TwoFactorSetupResponse contains the data needed to set up 2FA
type TwoFactorSetupResponse struct {
	Secret        string   `json:"secret"`
	QRCodeURL     string   `json:"qr_code_url"`
	RecoveryCodes []string `json:"recovery_codes"`
}

// TwoFactorEnableRequest represents the request to enable 2FA
type TwoFactorEnableRequest struct {
	Secret string `json:"secret" binding:"required"`
	OTP    string `json:"otp" binding:"required"`
}

// TwoFactorDisableRequest represents the request to disable 2FA
type TwoFactorDisableRequest struct {
	Password string `json:"password" binding:"required"`
	OTP      string `json:"otp" binding:"required"`
}

// Setup2FA generates a new 2FA secret and recovery codes
func Setup2FA(c *gin.Context) {
	h := &Handlers{Repo: default2FARepo}
	h.Setup2FA(c)
}

// Setup2FAHandler generates a new 2FA secret and recovery codes (method on Handlers)
func (h *Handlers) Setup2FA(c *gin.Context) {
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

	if user.TwoFactorEnabled {
		response.BadRequest(c, "2FA is already enabled")
		return
	}

	// Generate secret
	secret := generateTOTPSecret()

	// Generate recovery codes
	recoveryCodes := generateRecoveryCodes(10)

	// Build QR code URL (otpauth format)
	// Format: otpauth://totp/{issuer}:{username}?secret={secret}&issuer={issuer}
	issuer := "Deployer"
	qrCodeURL := "otpauth://totp/" + issuer + ":" + user.Username + "?secret=" + secret + "&issuer=" + issuer

	response.Data(c, TwoFactorSetupResponse{
		Secret:        secret,
		QRCodeURL:     qrCodeURL,
		RecoveryCodes: recoveryCodes,
	})
}

// Enable2FA enables 2FA after verifying the OTP
func Enable2FA(c *gin.Context) {
	h := &Handlers{Repo: default2FARepo}
	h.Enable2FA(c)
}

// Enable2FAHandler enables 2FA after verifying the OTP (method on Handlers)
func (h *Handlers) Enable2FA(c *gin.Context) {
	var req TwoFactorEnableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid user")
		return
	}

	// Verify OTP
	if !verifyTOTP(req.Secret, req.OTP) {
		response.BadRequest(c, "Invalid OTP")
		return
	}

	// Generate and hash recovery codes
	recoveryCodes := generateRecoveryCodes(10)
	hashedCodes := make([]string, len(recoveryCodes))
	for i, code := range recoveryCodes {
		hash, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		hashedCodes[i] = string(hash)
	}

	// Enable 2FA in transaction
	tx, err := database.DB.Beginx()
	if err != nil {
		response.InternalServerError(c, "Database error")
		return
	}

	if err := h.Repo.Enable2FAForUser(tx, userID, req.Secret, hashedCodes); err != nil {
		tx.Rollback()
		response.InternalServerError(c, "Failed to enable 2FA")
		return
	}

	if err := tx.Commit(); err != nil {
		response.InternalServerError(c, "Database error")
		return
	}

	response.Data(c, gin.H{
		"message":        "2FA enabled successfully",
		"recovery_codes": recoveryCodes,
	})
}

// Disable2FA disables 2FA after verifying password and OTP
func Disable2FA(c *gin.Context) {
	h := &Handlers{Repo: default2FARepo}
	h.Disable2FA(c)
}

// Disable2FAHandler disables 2FA after verifying password and OTP (method on Handlers)
func (h *Handlers) Disable2FA(c *gin.Context) {
	var req TwoFactorDisableRequest
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

	// Verify password
	if !user.CheckPassword(req.Password) {
		response.Error(c, http.StatusUnauthorized, "Invalid password")
		return
	}

	// Verify OTP
	if !verifyTOTP(user.TwoFactorSecret, req.OTP) {
		response.BadRequest(c, "Invalid OTP")
		return
	}

	// Disable 2FA in transaction
	tx, err := database.DB.Beginx()
	if err != nil {
		response.InternalServerError(c, "Database error")
		return
	}

	if err := h.Repo.Disable2FAForUser(tx, userID); err != nil {
		tx.Rollback()
		response.InternalServerError(c, "Failed to disable 2FA")
		return
	}

	if err := tx.Commit(); err != nil {
		response.InternalServerError(c, "Database error")
		return
	}

	response.Message(c, "2FA disabled successfully")
}

// generateTOTPSecret generates a new TOTP secret
func generateTOTPSecret() string {
	bytes := make([]byte, 20)
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}

// generateRecoveryCodes generates recovery codes
func generateRecoveryCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		bytes := make([]byte, 5)
		rand.Read(bytes)
		code := base32.StdEncoding.EncodeToString(bytes)[:8]
		codes[i] = code[:4] + "-" + code[4:]
	}
	return codes
}
