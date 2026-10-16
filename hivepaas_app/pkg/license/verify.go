package license

import (
	"encoding/json"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/releasesig"
)

// Trust is the public keys license keys are verified against.
//
// Dev is the development key set. Its private halves are committed with
// tools/licensesign, so anybody can sign with them, and a key they signed is
// accepted only when AllowDev is set: in development, which the caller decides.
// The set is kept either way, so that a development key pasted into a production
// installation is refused for what it is, not as a key nobody knows.
type Trust struct {
	Prod     releasesig.PublicKeys
	Dev      releasesig.PublicKeys
	AllowDev bool
}

// Verified is a key whose signatures verify and whose payload this binary can
// apply. Whether it is in force here and now is Evaluate's question.
type Verified struct {
	Payload *Payload
	// Development is a key signed by the development keys.
	Development bool
}

// Verify checks a key's text: its signatures, then its payload. Every refusal is
// ErrLicenseInvalid, whose detail says what failed and never what the key holds.
//
// The envelope is opened with the production keys, or with the development keys
// when its signatures name some of those and none of the production ones. A
// refusal then says what is wrong against the set the key was signed by, rather
// than that no production key signed it.
func Verify(trust *Trust, keyText string) (*Verified, error) {
	envelope, err := DecodeKeyText(keyText)
	if err != nil {
		return nil, err
	}

	keys, development := trust.Prod, false
	if signedByDevOnly(trust, envelope) {
		keys, development = trust.Dev, true
	}
	data, err := Purpose.Open(keys, envelope)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if development && !trust.AllowDev {
		return nil, invalid("signed by a development key, accepted only in development")
	}

	var payload Payload
	if err = json.Unmarshal(data, &payload); err != nil {
		// The decoder's message can quote the payload, so it is not passed on.
		return nil, invalid("malformed payload")
	}
	if err = payload.validate(); err != nil {
		return nil, err
	}
	return &Verified{Payload: &payload, Development: development}, nil
}

// signedByDevOnly reports whether the envelope's signatures name a development
// key and no production key. Only key ids are read; whether the envelope holds
// together is for Open to decide.
func signedByDevOnly(trust *Trust, envelope []byte) bool {
	var env releasesig.Envelope
	if json.Unmarshal(envelope, &env) != nil {
		return false
	}
	namesProd, namesDev := false, false
	for _, entry := range env.Signatures {
		_, inProd := trust.Prod[entry.KeyID]
		_, inDev := trust.Dev[entry.KeyID]
		namesProd = namesProd || inProd
		namesDev = namesDev || inDev
	}
	return namesDev && !namesProd
}
