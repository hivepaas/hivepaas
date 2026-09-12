package loggingdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func storedWithSecrets() *entity.Logging {
	return &entity.Logging{
		Backend: entity.LoggingBackend{
			Ingest: &entity.LoggingEndpoint{URL: "https://in", BearerToken: entity.NewEncryptedField("ingest-token")},
		},
		Forwards: []entity.LoggingForward{
			{Name: "siem", Endpoint: entity.LoggingEndpoint{
				URL: "https://siem", BearerToken: entity.NewEncryptedField("siem-token"),
			}},
		},
	}
}

func plain(t *testing.T, f entity.EncryptedField) string {
	t.Helper()
	v, err := f.GetPlain()
	if err != nil {
		t.Fatalf("GetPlain: %v", err)
	}
	return v
}

// The response must never carry a credential in the clear.
func TestFromEntityMasksEveryCredential(t *testing.T) {
	data, masked := FromEntity(storedWithSecrets())

	assert.True(t, masked)
	assert.Equal(t, basedto.MaskedSecret, data.Backend.Ingest.BearerToken)
	assert.Equal(t, basedto.MaskedSecret, data.Forwards[0].Endpoint.BearerToken)
	assert.Empty(t, data.Backend.Ingest.Password, "an unset credential is not masked, it is empty")
}

func TestFromEntityIsNotMaskedWithoutCredentials(t *testing.T) {
	_, masked := FromEntity(&entity.Logging{Enabled: true})

	assert.False(t, masked)
}

// Save what GET returned, unchanged: every stored credential must survive.
func TestRoundTripThroughTheMaskKeepsStoredSecrets(t *testing.T) {
	current := storedWithSecrets()
	data, _ := FromEntity(current)

	out, err := ToEntity(data, current)
	if err != nil {
		t.Fatalf("ToEntity: %v", err)
	}

	assert.Equal(t, "ingest-token", plain(t, out.Backend.Ingest.BearerToken))
	assert.Equal(t, "siem-token", plain(t, out.Forwards[0].Endpoint.BearerToken))
}

// Forwards are matched by name, not position: reordering them must not hand one
// forward's token to another.
func TestMaskedForwardSecretFollowsItsName(t *testing.T) {
	current := storedWithSecrets()
	current.Forwards = append(current.Forwards, entity.LoggingForward{
		Name: "audit", Endpoint: entity.LoggingEndpoint{
			URL: "https://audit", BearerToken: entity.NewEncryptedField("audit-token"),
		},
	})
	data, _ := FromEntity(current)
	data.Forwards[0], data.Forwards[1] = data.Forwards[1], data.Forwards[0]

	out, err := ToEntity(data, current)
	if err != nil {
		t.Fatalf("ToEntity: %v", err)
	}

	assert.Equal(t, "audit", out.Forwards[0].Name)
	assert.Equal(t, "audit-token", plain(t, out.Forwards[0].Endpoint.BearerToken))
	assert.Equal(t, "siem-token", plain(t, out.Forwards[1].Endpoint.BearerToken))
}

func TestNewPlaintextReplacesTheStoredSecret(t *testing.T) {
	current := storedWithSecrets()
	data, _ := FromEntity(current)
	data.Backend.Ingest.BearerToken = "rotated"

	out, err := ToEntity(data, current)
	if err != nil {
		t.Fatalf("ToEntity: %v", err)
	}

	assert.Equal(t, "rotated", plain(t, out.Backend.Ingest.BearerToken))
}

// A new forward sent with the placeholder has nothing behind it; saving it would
// store "********" as the password.
func TestPlaceholderWithNothingStoredIsRefused(t *testing.T) {
	data, _ := FromEntity(&entity.Logging{})
	data.Forwards = []ForwardData{{
		Name: "new", Endpoint: EndpointData{URL: "https://x", BearerToken: basedto.MaskedSecret},
	}}

	_, err := ToEntity(data, &entity.Logging{})

	assert.Error(t, err)
}

func TestValidateRejectsMissingData(t *testing.T) {
	assert.NotEmpty(t, (&UpdateSettingsReq{}).Validate())
}
