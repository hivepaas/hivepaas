package cacheentity

type MFAPasscode struct {
	Secret   string `json:"secret"`
	Attempts int    `json:"attempts,omitempty"`
}
