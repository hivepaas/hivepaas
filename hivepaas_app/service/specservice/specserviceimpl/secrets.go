package specserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// secretDecrypter is implemented by every setting type that stores secrets.
//
// It is declared here rather than imported because usecase/settings sits above
// this package in the layering. It is an interface, not logic: Go satisfies it
// structurally, so every type already implementing Decrypt for the reveal path
// is covered here too, including ones added later. That is the property the
// original's comment argues for - "what makes a setting type added later
// covered without anybody remembering to come back here".
type secretDecrypter interface {
	Decrypt() error
}

// revealSettingSecrets decrypts a setting's secrets in place when the mode calls
// for it, so that assembly can substitute them into the document.
//
// The capability check and the audit record are the usecase's, not this
// function's - see specuc.ExportSpec. This only does the decryption, and only
// for settings the scope actually owns.
func revealSettingSecrets(setting *entity.Setting, mode specmodel.SecretsMode) error {
	if !mode.RevealsSecrets() || setting == nil {
		return nil
	}

	// An inherited setting is read through, not owned, by this scope; its
	// secrets belong to whoever defined it. This is the same rule
	// usecase/settings applies, and it is what stops a project-scope export
	// from extracting global secrets.
	if setting.CurrentObjectID != "" && setting.ObjectID != setting.CurrentObjectID {
		return nil
	}

	data, err := setting.Parse()
	if err != nil {
		return hperrors.Wrap(err)
	}
	decrypter, ok := data.(secretDecrypter)
	if !ok {
		return nil // the type holds no secrets, so there is nothing to reveal
	}
	if err = decrypter.Decrypt(); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
