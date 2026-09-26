package mcp

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

// Need is what a tool needs of its caller beyond reading: the key's access
// action that the endpoint it ends up at checks.
type Need int

const (
	NeedRead Need = iota
	// NeedExecute is restarting and redeploying an app.
	NeedExecute
	// NeedWrite is installing an app, changing its configuration, scheduling a job.
	NeedWrite
	// NeedChange is either: apply_plan, which carries out a plan of any kind.
	NeedChange
)

// access is what one request may do: whether the setting allows changes at
// all, and which the key may make. A request is served the tools it can use and
// told why it cannot use the others, so that an assistant asked for a change it
// cannot make says what would let it.
type access struct {
	changesAllowed bool
	execute        bool
	write          bool
}

// accessOf is a request's access. A key with no access actions is not limited.
func accessOf(changesAllowed bool, auth *basedto.Auth) access {
	if !changesAllowed || auth == nil || auth.User == nil || auth.User.AuthClaims == nil {
		return access{changesAllowed: changesAllowed}
	}
	limit := auth.User.AuthClaims.AccessAction
	if limit == nil {
		return access{changesAllowed: true, execute: true, write: true}
	}
	return access{changesAllowed: true, execute: limit.Exec, write: limit.Write}
}

// allAccesses are every access a request can have, one server each.
var allAccesses = []access{
	{},
	{changesAllowed: true},
	{changesAllowed: true, execute: true},
	{changesAllowed: true, write: true},
	{changesAllowed: true, execute: true, write: true},
}

func (a access) serves(n Need) bool {
	switch n {
	case NeedRead:
		return true
	case NeedExecute:
		return a.changesAllowed && a.execute
	case NeedWrite:
		return a.changesAllowed && a.write
	case NeedChange:
		return a.changesAllowed && (a.execute || a.write)
	}
	return false
}

// mayChange is whether any tool that changes things is served.
func (a access) mayChange() bool {
	return a.serves(NeedChange)
}

const (
	whyChangesOff = "Changes are off on this server: an administrator allows them in the dashboard, " +
		"System settings, AI, MCP server."
	whyKeyReads = "Changes are allowed on this server, but this API key may only read. To let an assistant " +
		"restart or redeploy apps, the person connects it with a key that has Execute; to install apps, " +
		"change configuration or schedule jobs, one that has Write. Keys are created in the dashboard, " +
		"Profile, API keys."
	whyNoWrite = "This API key may restart and redeploy apps. Installing apps, changing configuration and " +
		"scheduling jobs need a key that has Write, created in the dashboard, Profile, API keys."
	whyNoExecute = "This API key may install apps, change configuration and schedule jobs. Restarting and " +
		"redeploying need a key that has Execute, created in the dashboard, Profile, API keys."
)

// instructions is what the server tells a model about itself, and about the
// changes this caller cannot make and why.
func (a access) instructions() string {
	switch {
	case !a.changesAllowed:
		return readInstructions + " " + whyChangesOff
	case !a.execute && !a.write:
		return readInstructions + " " + whyKeyReads
	case !a.write:
		return writeInstructions + " " + whyNoWrite
	case !a.execute:
		return writeInstructions + " " + whyNoExecute
	}
	return writeInstructions
}

// refusal is why a change that needs n is refused to this caller, or nil.
func (a access) refusal(n Need) error {
	switch {
	case a.serves(n):
		return nil
	case !a.changesAllowed:
		return errChangesNotAllowed
	case n == NeedExecute && a.write:
		return &InputError{Message: whyNoExecute}
	case n == NeedWrite && a.execute:
		return &InputError{Message: whyNoWrite}
	}
	return &InputError{Message: whyKeyReads}
}
