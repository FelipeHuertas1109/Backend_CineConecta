package utils

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func SetTokenCookie(c *gin.Context, token string) {
	isProduction := os.Getenv("ENV") == "production"

	cookie := &http.Cookie{
		Name:     "cine_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((24 * time.Hour).Seconds()),
	}

	// Establecer cookie manualmente en el header
	http.SetCookie(c.Writer, cookie)
}
