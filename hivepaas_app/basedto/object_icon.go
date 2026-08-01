package basedto

import (
	"fmt"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
)

func TransformObjectIcon(iconID string) string {
	if iconID == "" {
		return ""
	}
	if strings.HasPrefix(iconID, "/") {
		return iconID
	}
	return fmt.Sprintf("%v/images/%v", config.Current.HTTPServer.BasePath, iconID)
}
