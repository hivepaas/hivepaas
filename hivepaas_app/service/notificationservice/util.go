package notificationservice

import "time"

func formatDateTime(d time.Time) string {
	if d.IsZero() {
		return ""
	}
	return d.UTC().Format("Jan 02, 2006, 15:04:05 UTC")
}
