package http

import (
	"net/http"
	auth2 "open_discord/internal/auth"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Auth  *auth2.Service
	Token *auth2.TokenService
}

type SignInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const signupRoute = "/signup"
const signInRoute = "/signin"

func BindAuthRoutes(router *gin.Engine, authHandler *AuthHandler) {
	router.POST(signInRoute, authHandler.HandleSignIn)
	router.POST(signupRoute, authHandler.HandleSignUp)
}

func (h *AuthHandler) HandleSignIn(c *gin.Context) {
	var req SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	signInResult, err := h.Auth.CheckPassword(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !signInResult {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	mintedToken, err := h.Token.GenerateJWT(req.Username)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mintedToken})
}

func (h *AuthHandler) HandleSignUp(c *gin.Context) {
	var req SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.Auth.Signup(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": "ok"})
}

func AuthMiddleware(t *auth2.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()

		if path == signupRoute || path == signInRoute {
			c.Next()
			return
		}

		var bearerHeader = c.GetHeader("Authorization")
		if bearerHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		}

		var startsWithBearer = strings.HasPrefix(bearerHeader, "Bearer ")
		if !startsWithBearer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		}
		bearerToken := strings.TrimPrefix(bearerHeader, "Bearer ")
		claims, err := t.ValidateJWT(bearerToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		}
		c.Set("username", claims.Username)
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
