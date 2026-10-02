package mcp

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// registryAuthRoutes answers an env's credentials: one with a username and a
// password, one for Amazon ECR - and, as no API answer does, a token beside
// it, which the tool must not pass on.
func registryAuthRoutes(api *gin.RouterGroup) {
	api.GET("/projects/p1/prod/registry-auth", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"meta": gin.H{}, "data": []gin.H{
			{"id": "ra1", "name": "hub", "kind": "", "address": "docker.io", "username": "bot",
				"password": "********", "secretMasked": true},
			{"id": "ra2", "name": "ecr", "kind": "aws-ecr", "address": "123456789012.dkr.ecr.eu-west-1.amazonaws.com",
				"username": "AWS", "password": "", "secretMasked": true, "token": "tok-secret-value",
				"ecr": gin.H{"region": "eu-west-1", "keyAuth": gin.H{"id": "ka1", "name": "aws-pull"},
					"tokenExpiresAt": "2026-10-03T00:00:00Z",
					"token":          "tok-secret-value"}},
		}})
	})
}

// list_registry_auths shows each credential's kind and, for ECR, its region
// and when its token expires - and only what the API's answer type holds.
func TestRegistryAuthsShowTheirKindAndNeverAToken(t *testing.T) {
	w := newMCPWorld(t, registryAuthRoutes)
	session, err := w.connect(t, "key1")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	text, isErr := callTool(t, session, "list_registry_auths", map[string]any{"project": "shop", "env": "prod"})
	assert.False(t, isErr, text)
	assert.Contains(t, text, `"kind":"aws-ecr"`)
	assert.Contains(t, text, `"region":"eu-west-1"`)
	assert.Contains(t, text, `"tokenExpiresAt":"2026-10-03T00:00:00Z"`)
	assert.Contains(t, text, `"name":"aws-pull"`)
	assert.NotContains(t, text, "tok-secret-value")
	assert.NotContains(t, text, `"token"`)
}
