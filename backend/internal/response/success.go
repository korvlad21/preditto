package response

import "github.com/gin-gonic/gin"

func Success(c *gin.Context, status int, data any) {
	c.JSON(status, struct {
		Success bool `json:"success"`
		Data    any  `json:"data,omitempty"`
	}{Success: true, Data: data})
}
