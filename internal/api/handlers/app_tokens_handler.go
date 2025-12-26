package handlers

import (
	"youfun/shipyard/internal/api/response"
	"youfun/shipyard/internal/api/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// Legacy function wrappers for backward compatibility
var defaultTokensRepo = &DefaultRepository{}

// ApplicationTokenResponse represents an application token in API responses
type ApplicationTokenResponse struct {
	UID        string  `json:"uid"`
	Name       string  `json:"name"`
	ExpiresAt  *string `json:"expires_at,omitempty"`
	LastUsedAt *string `json:"last_used_at,omitempty"`
	CreatedAt  string  `json:"created_at,omitempty"`
}

// CreateApplicationTokenRequest represents the request to create a token
type CreateApplicationTokenRequest struct {
	Name      string  `json:"name" binding:"required"`
	ExpiresAt *string `json:"expires_at,omitempty"` // ISO8601 format
}

// CreateApplicationTokenResponse includes the plaintext token (only shown once)
type CreateApplicationTokenResponse struct {
	UID       string `json:"uid"`
	Name      string `json:"name"`
	Token     string `json:"token"` // Only returned once on creation
	CreatedAt string `json:"created_at"`
}

// ListApplicationTokens returns all tokens for an application
func ListApplicationTokens(c *gin.Context) {
	h := &Handlers{Repo: defaultTokensRepo}
	h.ListApplicationTokens(c)
}

// ListApplicationTokensHandler returns all tokens for an application (method on Handlers)
func (h *Handlers) ListApplicationTokens(c *gin.Context) {
	uid := c.Param("uid")
	appID, err := utils.DecodeFriendlyID(utils.PrefixApplication, uid)
	if err != nil {
		response.BadRequest(c, "Invalid application ID")
		return
	}

	tokens, err := h.Repo.GetApplicationTokensByAppID(appID)
	if err != nil {
		response.InternalServerError(c, "Failed to list tokens: "+err.Error())
		return
	}

	responses := make([]ApplicationTokenResponse, len(tokens))
	for i, token := range tokens {
		resp := ApplicationTokenResponse{
			UID:  utils.EncodeFriendlyID(utils.PrefixAppToken, token.ID),
			Name: token.Name,
		}
		if token.ExpiresAt.Time != nil {
			expiresAt := token.ExpiresAt.Time.Format(time.RFC3339)
			resp.ExpiresAt = &expiresAt
		}
		if token.LastUsedAt.Time != nil {
			lastUsedAt := token.LastUsedAt.Time.Format(time.RFC3339)
			resp.LastUsedAt = &lastUsedAt
		}
		if token.CreatedAt.Time != nil {
			resp.CreatedAt = token.CreatedAt.Time.Format(time.RFC3339)
		}
		responses[i] = resp
	}

	response.Data(c, responses)
}

// CreateApplicationToken creates a new token for an application
func CreateApplicationToken(c *gin.Context) {
	h := &Handlers{Repo: defaultTokensRepo}
	h.CreateApplicationToken(c)
}

// CreateApplicationTokenHandler creates a new token for an application (method on Handlers)
func (h *Handlers) CreateApplicationToken(c *gin.Context) {
	uid := c.Param("uid")
	appID, err := utils.DecodeFriendlyID(utils.PrefixApplication, uid)
	if err != nil {
		response.BadRequest(c, "Invalid application ID")
		return
	}

	// Verify application exists
	_, err = h.Repo.GetApplicationByID(appID)
	if err != nil {
		response.NotFound(c, "Application not found")
		return
	}

	var req CreateApplicationTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// Validate token name
	if len(req.Name) == 0 {
		response.BadRequest(c, "Token name is required")
		return
	}
	if len(req.Name) > 100 {
		response.BadRequest(c, "Token name must be 100 characters or less")
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			response.BadRequest(c, "Invalid expires_at format, use ISO8601/RFC3339")
			return
		}
		expiresAt = &t
	}

	token, plaintext, err := h.Repo.CreateApplicationToken(appID, req.Name, expiresAt)
	if err != nil {
		response.InternalServerError(c, "Failed to create token: "+err.Error())
		return
	}

	createdAt := ""
	if token.CreatedAt.Time != nil {
		createdAt = token.CreatedAt.Time.Format(time.RFC3339)
	}

	response.Created(c, CreateApplicationTokenResponse{
		UID:       utils.EncodeFriendlyID(utils.PrefixAppToken, token.ID),
		Name:      token.Name,
		Token:     plaintext,
		CreatedAt: createdAt,
	})
}

// DeleteApplicationToken deletes a token
func DeleteApplicationToken(c *gin.Context) {
	h := &Handlers{Repo: defaultTokensRepo}
	h.DeleteApplicationToken(c)
}

// DeleteApplicationTokenHandler deletes a token (method on Handlers)
func (h *Handlers) DeleteApplicationToken(c *gin.Context) {
	uid := c.Param("uid")
	appID, err := utils.DecodeFriendlyID(utils.PrefixApplication, uid)
	if err != nil {
		response.BadRequest(c, "Invalid application ID")
		return
	}

	tokenID := c.Param("tokenId")
	tokenUUID, err := utils.DecodeFriendlyID(utils.PrefixAppToken, tokenID)
	if err != nil {
		response.BadRequest(c, "Invalid token ID")
		return
	}

	if err := h.Repo.DeleteApplicationToken(tokenUUID, appID); err != nil {
		response.InternalServerError(c, "Failed to delete token: "+err.Error())
		return
	}

	response.Message(c, "Token deleted successfully")
}
