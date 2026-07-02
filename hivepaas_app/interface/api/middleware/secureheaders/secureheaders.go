package secureheaders

import "github.com/gin-gonic/gin"

// SecureHeaders secures request with headers
// Ref: https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html
func SecureHeaders(c *gin.Context) {
	header := c.Writer.Header()
	header.Add("X-Frame-Options", "DENY")
	// header.Add("Cache-Control", "no-store")
	header.Add("X-Content-Type-Options", "nosniff")
	header.Add("X-XSS-Protection", "1; mode=block")

	c.Next()
}
