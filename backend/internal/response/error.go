package response

import "github.com/gin-gonic/gin"

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const failureEnvelopeKey = "response.failure_envelope"

// WithFailureEnvelope must precede middleware that can abort the request.
func WithFailureEnvelope() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(failureEnvelopeKey, true)
		c.Next()
	}
}

func AbortError(c *gin.Context, status int, code, message string) {
	var success *bool
	if c.GetBool(failureEnvelopeKey) {
		success = new(bool)
	}
	c.AbortWithStatusJSON(status, struct {
		Success *bool `json:"success,omitempty"`
		Error   Error `json:"error"`
	}{Success: success, Error: Error{Code: code, Message: message}})
}
