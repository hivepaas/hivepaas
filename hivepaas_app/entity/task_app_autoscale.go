package entity

// TaskAppAutoscaleOutput is what a run of the autoscale job changed; a run
// that changed nothing saves no task.
type TaskAppAutoscaleOutput struct {
	Scaled []*AppAutoscaleChange `json:"scaled"`
}

// AppAutoscaleChange is one app scaled, and what it was scaled from.
//
// A function's: its calls in flight over the window, its calls and those
// turned away, its Concurrency and target. Any other app's: its requests in
// flight and its requests, its CPU as a percent of its limit, and their
// targets - 0 for a signal it does not scale on, or that could not be read.
type AppAutoscaleChange struct {
	App            string  `json:"app"`
	Name           string  `json:"name"`
	From           int     `json:"from"`
	To             int     `json:"to"`
	InFlight       float64 `json:"inFlight"`
	Calls          int64   `json:"calls"`
	Throttled      int64   `json:"throttled"`
	Concurrency    int     `json:"concurrency,omitempty"`
	Target         int     `json:"target,omitempty"`
	Requests       int64   `json:"requests,omitempty"`
	RequestsTarget int     `json:"requestsTarget,omitempty"`
	CPU            float64 `json:"cpu,omitempty"`
	CPUTarget      int     `json:"cpuTarget,omitempty"`
	Reason         string  `json:"reason"`
}
