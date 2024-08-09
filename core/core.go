package core

import (
	"context"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/zeim839/mailbox/data"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Timeout is the time to wait before canceling a database transaction.
var Timeout = 2 * time.Second

// BasicAuthMw returns a Gin middleware that implements the basic
// authorization scheme. username and password define the expected
// credentials.
func BasicAuthMw(username, password string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: need username and password",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Basic" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: scheme must be Basic",
			})
			return
		}

		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: could not decode base64",
			})
			return
		}

		credentials := strings.SplitN(string(decoded), ":", 2)
		if len(credentials) != 2 || credentials[0] != username || credentials[1] != password {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: wrong credentials",
			})
			return
		}

		c.Next()
	}
}

// Create returns a gin middleware that creates a new mailbox entry.
func Create(db data.Data) gin.HandlerFunc {
	return func(c *gin.Context) {
		var form data.Form
		if err := c.ShouldBindJSON(&form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if err := form.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := db.Create(ctx, form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.String(http.StatusOK, "")
	}
}

// CreateWithCaptcha returns a gin middleware that creates a new mailbox
// entry, but only if the associated captcha token is valid.
func CreateWithCaptcha(db data.Data, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var form data.FormWithCaptcha
		if err := c.ShouldBindJSON(&form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if err := form.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		if !validateCaptcha(secret, form.Captcha, form.RemoteIP) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "bad captcha",
			})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		defer cancel()
		if _, err := db.Create(ctx, form.Form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.String(http.StatusOK, "")
	}
}

// ReadAll returns a Gin middleware that fetches paginated batches of
// mailbox entries.
func ReadAll(db data.Data) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "0")
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid page number",
			})
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		defer cancel()
		forms, err := db.ReadAll(ctx, 20, int64(page))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"page":        page,
			"page_count":  int64(db.Count(ctx) / 20),
			"entry_count": len(forms),
			"entries":     forms,
		})
	}
}

// Read returns a Gin middleware that fetches a mailbox entry by its ID.
func Read(db data.Data) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		defer cancel()
		form, err := db.Read(ctx, id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, form)
	}
}

// Delete returns a Gin middleware that deletes a mailbox entry by its ID.
func Delete(db data.Data) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		defer cancel()
		if err := db.Delete(ctx, id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.String(http.StatusOK, "")
	}
}
