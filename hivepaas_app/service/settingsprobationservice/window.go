package settingsprobationservice

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

const (
	// The window is not the budget. Confirmation is refused for the first
	// settleDelay of it, so what the operator actually has is the window minus
	// that - which is why the numbers here are not the round ones somebody
	// reading "you have two minutes" would expect.
	//
	// Erring long is the cheap direction. A window that runs out under a working
	// configuration throws away a change somebody had only to click to keep, and
	// that is the failure that makes people want the whole mechanism turned off. A
	// window that is too long costs the operator a longer wait to get back into a
	// dashboard.
	WindowDefault = 3 * time.Minute
	WindowMax     = 15 * time.Minute

	// answerMargin is the least a caller gets between confirmation becoming
	// possible and the deadline taking the change away.
	//
	// The floor is derived from that rather than written down as a second number,
	// because the two have already drifted apart once: a 60s minimum outlived its
	// settle delay until service settings arrived with a 75s one, at which point
	// the smallest window expired fifteen seconds before anybody was allowed to
	// answer it. Deriving it means a slower settle delay cannot reintroduce that.
	answerMargin = 45 * time.Second
)

// ResolveWindow clamps the requested window against the settle delay the caller
// is going to arm with.
//
// It lives in the interface package rather than behind the Service, because
// callers need the answer while building their request - before there is anything
// to arm - and because it is a pure function of the two arguments.
//
// It takes the delay rather than the setting type because the setting type does
// not decide it: the same HivePaaS service settings endpoint restarts the app or
// leaves it alone depending on what the request carries, and only the caller
// knows which.
//
// There is deliberately no way to ask for zero. A caller that cannot confirm is
// exactly the caller this exists for: a script that applies a change and dies
// leaves the change reverted, which is the outcome we want and not one the script
// gets to opt out of.
func ResolveWindow(requested, settleDelay time.Duration) time.Duration {
	minWindow := max(settleDelay, entity.SettingsProbationSettleDelay) + answerMargin
	if requested <= 0 {
		return max(WindowDefault, minWindow)
	}
	return gofn.Clamp(requested, minWindow, WindowMax)
}
