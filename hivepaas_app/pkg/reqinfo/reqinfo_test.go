package reqinfo

import (
	"context"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	want := &RequestInfo{RequestID: "req_1", ClientIP: "1.2.3.4", RemoteAddr: "10.0.0.9"}
	got := From(NewContext(context.Background(), want))
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

// Plenty of work runs outside a request - workers, scheduled jobs, the agent -
// and every caller has to survive that rather than assume a request context.
func TestNoInfoIsNotAnError(t *testing.T) {
	if got := From(context.Background()); got != nil {
		t.Errorf("a bare context must carry no info, got %+v", got)
	}
	if got := From(nil); got != nil { //nolint:staticcheck
		t.Errorf("a nil context must carry no info, got %+v", got)
	}
	ctx := context.Background()
	if NewContext(ctx, nil) != ctx {
		t.Error("storing nil must leave the context alone")
	}
}

// The key is a private type, so nothing outside this package can collide with it
// or read the info by guessing a string key.
func TestKeyIsNotReachableByString(t *testing.T) {
	ctx := NewContext(context.Background(), &RequestInfo{RequestID: "req_1"})
	if v := ctx.Value("hp.requestInfo"); v != nil { //nolint:staticcheck
		t.Errorf("the info must not be reachable by a string key, got %v", v)
	}
}
