package apphelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectAppIcon(t *testing.T) {
	testCases := []struct {
		name      string
		appName   string
		imageName string
		expected  string
	}{
		{
			name:      "Standard image with tag",
			appName:   "my-db",
			imageName: "postgres:15-alpine",
			expected:  "postgresql",
		},
		{
			name:      "Registry domain and organization prefix",
			appName:   "db",
			imageName: "ghcr.io/bitnami/postgresql:15",
			expected:  "postgresql",
		},
		{
			name:      "Full registry URL with Grafana",
			appName:   "dashboard",
			imageName: "docker.io/grafana/grafana-enterprise:latest",
			expected:  "grafana",
		},
		{
			name:      "MySQL database image",
			appName:   "app-mysql",
			imageName: "mysql:8.0",
			expected:  "mysql",
		},
		{
			name:      "Redis cache image",
			appName:   "cache",
			imageName: "redis:7.0-alpine",
			expected:  "redis",
		},
		{
			name:      "Alias resolution - mongo to mongodb",
			appName:   "my-mongo",
			imageName: "mongo:latest",
			expected:  "mongodb",
		},
		{
			name:      "Alias resolution - k8s to kubernetes",
			appName:   "cluster-k8s",
			imageName: "k8s-tool:v1",
			expected:  "kubernetes",
		},
		{
			name:      "Fallback to appName when imageName is custom",
			appName:   "my-redis-cache",
			imageName: "myregistry.com/myteam/custom-app:v1",
			expected:  "redis",
		},
		{
			name:      "Image name priority over appName",
			appName:   "my-redis-app",
			imageName: "postgres:15",
			expected:  "postgresql",
		},
		{
			name:      "Unknown image and appName returns empty string",
			appName:   "unknown-service",
			imageName: "myregistry.com/myteam/custom-app:v1",
			expected:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := DetectAppIcon(tc.appName, tc.imageName)
			assert.Equal(t, tc.expected, result)
		})
	}
}
