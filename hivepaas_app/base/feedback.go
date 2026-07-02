package base

type FeedbackType string

const (
	FeedbackTypeSecurity  FeedbackType = "security"
	FeedbackTypeIssue     FeedbackType = "issue"
	FeedbackTypeLicensing FeedbackType = "licensing"
	FeedbackTypeGeneral   FeedbackType = "general"
)

var (
	AllFeedbackTypes = []FeedbackType{FeedbackTypeSecurity, FeedbackTypeIssue, FeedbackTypeLicensing,
		FeedbackTypeGeneral}
)
