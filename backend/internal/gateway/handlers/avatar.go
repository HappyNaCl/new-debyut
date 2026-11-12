package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const maxAvatarBytes = 2 << 20

func readAvatar(c *gin.Context) ([]byte, bool) {
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"avatar": "is required"}})
		return nil, false
	}
	if file.Size > maxAvatarBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"avatar": "must be 2 MB or smaller"}})
		return nil, false
	}

	opened, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"avatar": "could not be read"}})
		return nil, false
	}
	defer opened.Close()

	content, err := io.ReadAll(io.LimitReader(opened, maxAvatarBytes+1))
	if err != nil || len(content) > maxAvatarBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"avatar": "must be 2 MB or smaller"}})
		return nil, false
	}
	return content, true
}
