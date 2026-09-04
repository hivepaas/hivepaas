package envutil

import (
	"slices"
	"testing"
)

func TestSafeEnviron(t *testing.T) {
	t.Setenv("HP_APP_SECRET", "must-not-leak")
	t.Setenv("HP_DB_PASSWORD", "must-not-leak")
	t.Setenv("PATH_FOR_TEST", "kept")

	safe := SafeEnviron()

	if slices.Contains(safe, "HP_APP_SECRET=must-not-leak") {
		t.Error("the app secret reached a child process environment")
	}
	if slices.Contains(safe, "HP_DB_PASSWORD=must-not-leak") {
		t.Error("a HP_ variable reached a child process environment")
	}
	if !slices.Contains(safe, "PATH_FOR_TEST=kept") {
		t.Error("an unrelated variable was dropped")
	}
	for _, entry := range safe {
		if len(entry) >= 3 && entry[:3] == "HP_" {
			t.Errorf("HP_ variable survived: %s", entry)
		}
	}
}
