package cacheentity

type ConsoleTicket struct {
	AppID    string `json:"appId"`
	TargetID string `json:"targetId,omitempty"`
}
