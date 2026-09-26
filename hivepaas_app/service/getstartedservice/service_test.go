package getstartedservice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChecklistIsAllDoneOnlyWhenEveryItemIsDone(t *testing.T) {
	done := Item{Status: ItemStatusDone}
	checklist := Checklist{DashboardCert: done, TwoFactor: done, GithubApp: done}
	assert.True(t, checklist.AllDone())

	for _, status := range []ItemStatus{ItemStatusTodo, ItemStatusObtaining, ItemStatusFailed} {
		left := checklist
		left.DashboardCert = Item{Status: status}
		assert.False(t, left.AllDone(), "dashboard certificate %s", status)
	}
	left := checklist
	left.TwoFactor = Item{Status: ItemStatusTodo}
	assert.False(t, left.AllDone(), "two-factor left")
	left = checklist
	left.GithubApp = Item{Status: ItemStatusTodo}
	assert.False(t, left.AllDone(), "GitHub App left")
}
