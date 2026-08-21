package gittool

import (
	"context"
	"encoding/pem"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/services/git/github"
)

const (
	sshKeyFileMode = 0600
)

type AuthMethod interface {
	Name() string
}

type authBasic struct {
	Username string
	Password string
}

func (b *authBasic) Name() string {
	return "http-basic-auth"
}

type authSSHKey struct {
	PEMBytes []byte
}

func (a *authSSHKey) Name() string {
	return "ssh-key"
}

func calcGitAuthMethod(
	ctx context.Context,
	gitCreds *entity.Setting,
) (auth AuthMethod, err error) {
	if gitCreds == nil {
		return auth, nil
	}
	switch gitCreds.Type { //nolint:exhaustive
	case base.SettingTypeGithubApp:
		client, err := github.NewFromSetting(gitCreds)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		token, err := client.CreateAppToken(ctx)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		auth = &authBasic{
			Username: "default", // this can be anything except an empty string
			Password: token,
		}

	case base.SettingTypeAccessToken:
		token, err := gitCreds.MustAsAccessToken().Token.GetPlain()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		auth = &authBasic{
			Username: "default", // this can be anything except an empty string
			Password: token,
		}

	case base.SettingTypeSSHKey:
		sshKey := gitCreds.MustAsSSHKey()
		privateKey, err := sshKey.PrivateKey.GetPlain()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		passphrase, err := sshKey.Passphrase.GetPlain()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		pemBytes := reflectutil.UnsafeStrToBytes(privateKey)
		if passphrase != "" {
			rawKey, err := ssh.ParseRawPrivateKeyWithPassphrase(pemBytes, reflectutil.UnsafeStrToBytes(passphrase))
			if err != nil {
				return nil, apperrors.Wrap(err)
			}
			pemBlock, err := ssh.MarshalPrivateKey(rawKey, "")
			if err != nil {
				return nil, apperrors.Wrap(err)
			}
			pemBytes = pem.EncodeToMemory(pemBlock)
		}

		auth = &authSSHKey{
			PEMBytes: pemBytes,
		}
	}
	return auth, nil
}

func writeSshKeyFile(baseDir string, pemBytes []byte) (sshKeyFile string, err error) {
	fh, err := os.CreateTemp(baseDir, "git-ssh-*")
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	defer fh.Close()

	// NOTE: file will be removed along with the whole temp dir by the caller
	sshKeyFile = fh.Name()

	if err := os.Chmod(sshKeyFile, sshKeyFileMode); err != nil {
		return "", apperrors.Wrap(err)
	}

	if _, err := fh.Write(pemBytes); err != nil {
		return "", apperrors.Wrap(err)
	}

	if len(pemBytes) > 0 && pemBytes[len(pemBytes)-1] != '\n' {
		if _, err := fh.Write([]byte("\n")); err != nil {
			return "", apperrors.Wrap(err)
		}
	}

	return sshKeyFile, nil
}
