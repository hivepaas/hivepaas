package gittool

import (
	"context"
	"os"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/services/git/github"
)

const (
	sshKeyFileMode = 0600
)

type authSSHKey struct {
	*ssh.PublicKeys
	PEMBytes []byte
}

func calcGitAuthMethod(
	ctx context.Context,
	gitCreds *entity.Setting,
) (auth transport.AuthMethod, err error) {
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
		auth = &http.BasicAuth{
			Username: "default", // this can be anything except an empty string
			Password: token,
		}

	case base.SettingTypeAccessToken:
		token, err := gitCreds.MustAsAccessToken().Token.GetPlain()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		auth = &http.BasicAuth{
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
		authRaw, err := ssh.NewPublicKeys("git", reflectutil.UnsafeStrToBytes(privateKey), passphrase)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		auth = &authSSHKey{
			PublicKeys: authRaw,
			PEMBytes:   reflectutil.UnsafeStrToBytes(privateKey),
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

	if pemBytes[len(pemBytes)-1] != '\n' {
		if _, err := fh.Write([]byte("\n")); err != nil {
			return "", apperrors.Wrap(err)
		}
	}

	return sshKeyFile, nil
}
