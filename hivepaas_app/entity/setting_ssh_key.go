package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentSSHKeyVersion = 1
)

var _ = registerSettingParser(base.SettingTypeSSHKey, &sshKeyParser{})

type sshKeyParser struct {
}

func (s *sshKeyParser) New() SettingData {
	return &SSHKey{}
}

type SSHKey struct {
	KeyType    base.PrivateKeyType `json:"keyType"`
	PublicKey  string              `json:"publicKey,omitempty"`
	PrivateKey EncryptedField      `json:"privateKey"`
	Passphrase EncryptedField      `json:"passphrase,omitzero"`
}

func (s *SSHKey) GetType() base.SettingType {
	return base.SettingTypeSSHKey
}

func (s *SSHKey) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *SSHKey) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *SSHKey) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSSHKeyVersion {
		return false, nil
	}
	if setting.Version > CurrentSSHKeyVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSSHKeyVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *SSHKey) Decrypt() error {
	_, err := s.PrivateKey.GetPlain()
	if err != nil {
		return apperrors.Wrap(err)
	}
	_, err = s.Passphrase.GetPlain()
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *Setting) AsSSHKey() (*SSHKey, error) {
	return parseSettingAs[*SSHKey](s)
}

func (s *Setting) MustAsSSHKey() *SSHKey {
	return gofn.Must(s.AsSSHKey())
}
