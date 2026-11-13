package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const maxAvatarBytes = 2 << 20

func readAvatar(c *gin.Context) ([]byte, bool) {
	return readImage(c, "avatar", maxAvatarBytes)
}

// readImage reads an uploaded file from a multipart form field, writing a 400
// field error and returning false when it is missing or larger than maxBytes.
func readImage(c *gin.Context, field string, maxBytes int64) ([]byte, bool) {
	tooLarge := gin.H{"error": gin.H{field: fmt.Sprintf("must be %d MB or smaller", maxBytes>>20)}}

	file, err := c.FormFile(field)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{field: "is required"}})
		return nil, false
	}
	if file.Size > maxBytes {
		c.JSON(http.StatusBadRequest, tooLarge)
		return nil, false
	}

	opened, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{field: "could not be read"}})
		return nil, false
	}
	defer opened.Close()

	content, err := io.ReadAll(io.LimitReader(opened, maxBytes+1))
	if err != nil || int64(len(content)) > maxBytes {
		c.JSON(http.StatusBadRequest, tooLarge)
		return nil, false
	}
	return content, true
}
