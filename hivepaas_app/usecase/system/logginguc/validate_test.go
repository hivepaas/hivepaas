package logginguc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/logging"
)

func validEnabled() *entity.Logging {
	return &entity.Logging{
		Enabled: true,
		Sources: entity.LoggingSources{Apps: true},
		Collector: entity.LoggingCollector{
			Type: entity.LoggingCollectorTypeVlagent, Managed: true,
		},
		Backend: entity.LoggingBackend{
			Type: entity.LoggingBackendTypeVictoriaLogs, Managed: true,
			VictoriaLogs: &entity.LoggingVictoriaLogs{NodeID: "n1", VolumeID: "v1"},
		},
	}
}

func TestValidateAcceptsAWorkingConfiguration(t *testing.T) {
	assert.NoError(t, validateSettings(validEnabled()))
}

// Disabled is the default and must stay valid however empty it is, or an
// operator could not turn logging off without filling in a form first.
func TestValidateAcceptsDisabledAndEmpty(t *testing.T) {
	assert.NoError(t, validateSettings(&entity.Logging{}))
	assert.NoError(t, validateSettings(nil))
}

func TestValidateRejectsEnabledWithNoSources(t *testing.T) {
	cfg := validEnabled()
	cfg.Sources = entity.LoggingSources{}

	assert.ErrorIs(t, validateSettings(cfg), logging.ErrNoSources)
}

func TestValidateRejectsAManagedBackendWithNoVolume(t *testing.T) {
	cfg := validEnabled()
	cfg.Backend.VictoriaLogs.VolumeID = ""

	assert.ErrorIs(t, validateSettings(cfg), loggingservice.ErrVolumeMissing)
}

func TestValidateRejectsAManagedBackendWithNoNode(t *testing.T) {
	cfg := validEnabled()
	cfg.Backend.VictoriaLogs.NodeID = ""

	assert.ErrorIs(t, validateSettings(cfg), loggingservice.ErrBackendNodeMissing)
}

func TestValidateRejectsAnUnmanagedBackendWithNoIngestURL(t *testing.T) {
	cfg := validEnabled()
	cfg.Backend.Managed = false
	cfg.Backend.VictoriaLogs = nil

	assert.ErrorIs(t, validateSettings(cfg), logging.ErrIngestEndpointRequired)
}

// An unmanaged backend is legitimate: the user runs their own VictoriaLogs and
// HivePaaS only ships to it. Type and Managed are separate questions.
func TestValidateAcceptsAnUnmanagedBackend(t *testing.T) {
	cfg := validEnabled()
	cfg.Backend.Managed = false
	cfg.Backend.VictoriaLogs = nil
	cfg.Backend.Ingest = &entity.LoggingEndpoint{URL: "https://logs.example/insert"}

	assert.NoError(t, validateSettings(cfg))
}

func TestValidateRejectsAForwardWithNoURL(t *testing.T) {
	cfg := validEnabled()
	cfg.Forwards = []entity.LoggingForward{{Name: "siem"}}

	assert.ErrorIs(t, validateSettings(cfg), logging.ErrIngestEndpointRequired)
}

func TestValidateRejectsAForwardWithNoName(t *testing.T) {
	cfg := validEnabled()
	cfg.Forwards = []entity.LoggingForward{{
		Endpoint: entity.LoggingEndpoint{URL: "https://siem.example/ingest"},
	}}

	assert.Error(t, validateSettings(cfg), "a forward with no name cannot be reported on or removed")
}

// Two forwards with one name would be indistinguishable in status and in any
// request to remove one of them.
func TestValidateRejectsDuplicateForwardNames(t *testing.T) {
	cfg := validEnabled()
	ep := entity.LoggingEndpoint{URL: "https://siem.example/ingest"}
	cfg.Forwards = []entity.LoggingForward{{Name: "siem", Endpoint: ep}, {Name: "siem", Endpoint: ep}}

	assert.Error(t, validateSettings(cfg))
}
