package middleware

import (
	"youfun/shipyard/internal/database"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTSecret is the secret key used to sign JWT tokens
// MUST be set via JWT_SECRET environment variable to persist across restarts
var JWTSecret []byte

// InitJWTSecret initializes the JWT secret from environment variable
// This should be called after loading .env file
func InitJWTSecret() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Generate a temporary random secret if not provided
		secret = uuid.New().String()
		log.Println("⚠️  WARNING: JWT_SECRET environment variable is not set!")
		log.Println("⚠️  Generated a temporary secret for this session.")
		log.Println("⚠️  All user login sessions will be INVALIDATED after server restart.")
		log.Println("⚠️  Please set JWT_SECRET in your .env file to persist login sessions.")
		log.Printf("⚠️  Example: echo 'JWT_SECRET=%s' >> .env\n", secret)
	} else {
		log.Println("✅ JWT Secret loaded successfully from environment variable")
	}
	JWTSecret = []byte(secret)
}

// Claims represents the JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	DeviceID string `json:"device_id,omitempty"` // Optional: for device tracking
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID, username string, deviceID string) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour) // 7 days

	claims := &Claims{
		UserID:   userID,
		Username: username,
		DeviceID: deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "deployer",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

// GenerateTemp2FAToken creates a temporary token for 2FA verification
func GenerateTemp2FAToken(userID, username string) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute) // 5 minutes

	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "deployer-2fa",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

// ValidateToken parses and validates a JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// AuthMiddleware is a Gin middleware for JWT authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// First, try to get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Check for Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// If no token in header, try query parameter (for WebSocket connections)
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}

		claims, err := ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("device_id", claims.DeviceID)

		// Update device last used time if device_id is present
		if claims.DeviceID != "" {
			deviceID, err := uuid.Parse(claims.DeviceID)
			if err == nil {
				_ = database.UpdateAuthDeviceLastUsed(deviceID)
			}
		}

		c.Next()
	}
}

// GetUserIDFromContext extracts the user ID from the Gin context
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("user_id not found in context")
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}
