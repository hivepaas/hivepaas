package entity

// TaskFunctionAutoscaleOutput is what a run of the function autoscale job
// changed; a run that changed nothing saves no task.
type TaskFunctionAutoscaleOutput struct {
	Scaled []*FunctionAutoscaleChange `json:"scaled"`
}

// FunctionAutoscaleChange is one function scaled, and what it was scaled from:
// its calls in flight over the window, its calls and those turned away, its
// Concurrency and target.
type FunctionAutoscaleChange struct {
	App         string  `json:"app"`
	Name        string  `json:"name"`
	From        int     `json:"from"`
	To          int     `json:"to"`
	InFlight    float64 `json:"inFlight"`
	Calls       int64   `json:"calls"`
	Throttled   int64   `json:"throttled"`
	Concurrency int     `json:"concurrency"`
	Target      int     `json:"target"`
	Reason      string  `json:"reason"`
}
