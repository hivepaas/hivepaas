package entity

type BaseEventNotification struct {
	Success           ObjectID `json:"success,omitzero"`
	SuccessUseDefault bool     `json:"successUseDefault,omitempty"`
	Failure           ObjectID `json:"failure,omitzero"`
	FailureUseDefault bool     `json:"failureUseDefault,omitempty"`
}

func (s *BaseEventNotification) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.Success.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.Success.ID)
	}
	if s.Failure.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.Failure.ID)
	}
	return refIDs
}
