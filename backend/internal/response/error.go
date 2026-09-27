package response

import "github.com/gin-gonic/gin"

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func AbortError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, struct {
		Error Error `json:"error"`
	}{Error: Error{Code: code, Message: message}})
}
