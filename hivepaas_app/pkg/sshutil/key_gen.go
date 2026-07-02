package sshutil

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"

	"golang.org/x/crypto/ssh"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func GenerateKey(keyType base.PrivateKeyType, passphrase string) (privKeyStr string, pubKeyStr string, err error) {
	var privKey crypto.Signer
	//nolint:mnd
	switch keyType {
	case base.PrivateKeyTypeEd25519:
		_, privKey, err = ed25519.GenerateKey(rand.Reader)
	case base.PrivateKeyTypeECP256:
		privKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case base.PrivateKeyTypeECP384:
		privKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case base.PrivateKeyTypeECP521:
		privKey, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	case base.PrivateKeyTypeRSA2048:
		privKey, err = rsa.GenerateKey(rand.Reader, 2048)
	case base.PrivateKeyTypeRSA3072:
		privKey, err = rsa.GenerateKey(rand.Reader, 3072)
	case base.PrivateKeyTypeRSA4096:
		privKey, err = rsa.GenerateKey(rand.Reader, 4096)
	case base.PrivateKeyTypeRSA8192:
		privKey, err = rsa.GenerateKey(rand.Reader, 8192)
	default:
		return "", "", apperrors.New(apperrors.ErrPrivateKeyTypeUnsupported).
			WithParam("Type", keyType)
	}
	if err != nil {
		return "", "", apperrors.New(err)
	}

	// 1. Generate Private Key String
	var pemBlock *pem.Block
	if passphrase != "" {
		pemBlock, err = ssh.MarshalPrivateKeyWithPassphrase(privKey, "", []byte(passphrase))
	} else {
		pemBlock, err = ssh.MarshalPrivateKey(privKey, "")
	}
	if err != nil {
		return "", "", apperrors.New(err)
	}
	privKeyStr = string(pem.EncodeToMemory(pemBlock))

	// 2. Generate Public Key String
	pub, err := ssh.NewPublicKey(privKey.Public())
	if err != nil {
		return "", "", apperrors.New(err)
	}
	pubKeyStr = string(ssh.MarshalAuthorizedKey(pub))

	return privKeyStr, pubKeyStr, nil
}

func GeneratePublicKey(privKey, passphrase string) (base.PrivateKeyType, string, error) {
	var signer ssh.Signer
	var err error

	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(privKey), []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey([]byte(privKey))
	}
	if err != nil {
		return "", "", apperrors.New(err)
	}

	pub := signer.PublicKey()
	cryptoPub, ok := pub.(ssh.CryptoPublicKey)
	if !ok {
		return "", "", apperrors.NewUnsupportedNT("Public key").WithExtraDetail("cannot extract crypto public key")
	}
	pubKeyStr := string(ssh.MarshalAuthorizedKey(pub))

	var pType base.PrivateKeyType
	switch rawPub := cryptoPub.CryptoPublicKey().(type) {
	case ed25519.PublicKey:
		pType = base.PrivateKeyTypeEd25519
	case *ecdsa.PublicKey:
		curveName := rawPub.Curve.Params().Name
		switch curveName {
		case "P-256":
			pType = base.PrivateKeyTypeECP256
		case "P-384":
			pType = base.PrivateKeyTypeECP384
		case "P-521":
			pType = base.PrivateKeyTypeECP521
		default:
			return "", pubKeyStr, nil
		}
	case *rsa.PublicKey:
		//nolint:mnd
		switch rawPub.N.BitLen() {
		case 2048:
			pType = base.PrivateKeyTypeRSA2048
		case 3072:
			pType = base.PrivateKeyTypeRSA3072
		case 4096:
			pType = base.PrivateKeyTypeRSA4096
		case 8192:
			pType = base.PrivateKeyTypeRSA8192
		default:
			return "", pubKeyStr, nil
		}
	default:
		return "", pubKeyStr, nil
	}

	return pType, pubKeyStr, nil
}
