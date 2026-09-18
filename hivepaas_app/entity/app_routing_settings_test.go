package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A domain answers on the app's port unless it names one of its own. The one it
// does not name is stored as a zero, which is not a port: a caller that took it
// for one would ask for port 0 to be reachable, and a resource link would claim
// the app listens there.
func TestActivePortsLeaveOutTheOneADomainDoesNotName(t *testing.T) {
	settings := &AppRoutingSettings{
		Port:           8080,
		ExposePublicly: true,
		Domains: []*AppDomain{
			{Domain: "a.example.com", Enabled: true},
			{Domain: "b.example.com", Enabled: true, ContainerPort: 9000},
			{Domain: "c.example.com", Enabled: true, ContainerPort: 8080},
			{Domain: "d.example.com", Enabled: false, ContainerPort: 7000},
		},
	}

	assert.Equal(t, []int{8080, 9000}, settings.GetActivePorts())
}

func TestActivePortsOfAnAppWithNoDomains(t *testing.T) {
	settings := &AppRoutingSettings{Port: 5432}

	assert.Equal(t, []int{5432}, settings.GetActivePorts())
}
