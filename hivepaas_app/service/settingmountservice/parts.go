// Package settingmountservice mounts parts of settings - a certificate and its
// key, a basic auth pair as htpasswd - as files in an app's containers, which
// follow those settings from then on. What needs no database lives here: the
// parts a source type offers, how a file is named and labeled, and where it may
// go. See docs/superpowers/specs/2026-09-25-setting-mounts-design.md.
package settingmountservice

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"slices"
	"strconv"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/htpasswd"
)

// The names of parts, and of the values they read: a field part is named after
// its field.
const (
	fieldCertificate   = "certificate"
	fieldPrivateKey    = "privateKey"
	fieldCACertificate = "caCertificate"
	fieldPublicKey     = "publicKey"
	fieldUsername      = "username"
	fieldPassword      = "password"
	partHtpasswd       = "htpasswd"
	fieldValue         = "value"
	fieldContent       = "content"
)

// Part is one file a source type offers. A field is the simplest part; a virtual
// one such as htpasswd is computed. The engine does not tell them apart.
type Part struct {
	Name string
	// Required parts must all have their inputs for the source to be mounted at
	// all; an optional part without them is left out alone.
	Required bool
	// Secret parts are stored as Docker secrets; the others as Docker configs.
	Secret bool
	// Gated parts take the Reveal Secrets permission to mount: the app has no
	// other way to read them. A secret's value is not gated - the app reads it
	// through ${secrets.NAME} already.
	Gated bool
	// Version is raised whenever Render's output changes shape for the same
	// inputs, which is what replaces files already mounted.
	Version int
	// Inputs are the source's values the part reads, as Values names them.
	Inputs []string
	// Render turns the inputs' values, in Inputs' order, into the file. Nil
	// writes the one input as it is.
	Render func(values []string) ([]byte, error)
}

type sourceType struct {
	parts  []*Part
	values func(*entity.Setting) (map[string]string, error)
}

var registry = map[base.SettingType]*sourceType{
	base.SettingTypeSecret: {
		parts:  []*Part{{Name: fieldValue, Required: true, Secret: true, Version: 1, Inputs: []string{fieldValue}}},
		values: secretValues,
	},
	base.SettingTypeConfigFile: {
		parts:  []*Part{{Name: fieldContent, Required: true, Version: 1, Inputs: []string{fieldContent}}},
		values: configFileValues,
	},
	base.SettingTypeSSLCert: {
		parts: []*Part{
			{Name: fieldCertificate, Required: true, Version: 1, Inputs: []string{fieldCertificate}},
			{Name: fieldPrivateKey, Required: true, Secret: true, Gated: true, Version: 1, Inputs: []string{fieldPrivateKey}},
			{Name: fieldCACertificate, Version: 1, Inputs: []string{fieldCACertificate}},
		},
		values: sslCertValues,
	},
	base.SettingTypeSSHKey: {
		parts: []*Part{
			{Name: fieldPrivateKey, Required: true, Secret: true, Gated: true, Version: 1, Inputs: []string{fieldPrivateKey}},
			{Name: fieldPublicKey, Version: 1, Inputs: []string{fieldPublicKey}},
		},
		values: sshKeyValues,
	},
	base.SettingTypeBasicAuth: {
		parts: []*Part{
			{Name: fieldUsername, Required: true, Version: 1, Inputs: []string{fieldUsername}},
			{Name: fieldPassword, Required: true, Secret: true, Gated: true, Version: 1, Inputs: []string{fieldPassword}},
			// bcrypt, which Traefik, Apache, Caddy and HivePaaS's registry read,
			// and nginx where the system's crypt is libxcrypt.
			{Name: partHtpasswd, Required: true, Secret: true, Gated: true, Version: 1,
				Inputs: []string{fieldUsername, fieldPassword}, Render: renderHtpasswd},
		},
		values: basicAuthValues,
	},
}

// SourceTypes are the setting types an entry may mount from.
func SourceTypes() []base.SettingType {
	types := make([]base.SettingType, 0, len(registry))
	for typ := range registry {
		types = append(types, typ)
	}
	slices.Sort(types)
	return types
}

func IsSourceType(typ base.SettingType) bool {
	return registry[typ] != nil
}

// PartsOf are the parts a source type offers, nil for a type that is not one.
func PartsOf(typ base.SettingType) []*Part {
	if src := registry[typ]; src != nil {
		return src.parts
	}
	return nil
}

// PartOf is a source type's part by name, nil when it offers none of that name.
func PartOf(typ base.SettingType, name string) *Part {
	for _, part := range PartsOf(typ) {
		if part.Name == name {
			return part
		}
	}
	return nil
}

// Values are what a source's parts read, decrypted, by input name.
func Values(setting *entity.Setting) (map[string]string, error) {
	src := registry[setting.Type]
	if src == nil {
		return nil, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("%s is not a mount source", setting.Type)
	}
	values, err := src.values(setting)
	return values, hperrors.Wrap(err)
}

// Usable reports whether every required part of a source has its inputs: a
// setting that cannot be used is not in the container.
func Usable(typ base.SettingType, values map[string]string) bool {
	if !IsSourceType(typ) {
		return false
	}
	for _, part := range PartsOf(typ) {
		if part.Required && part.Empty(values) {
			return false
		}
	}
	return true
}

// Empty reports whether one of the part's inputs is.
func (p *Part) Empty(values map[string]string) bool {
	for _, input := range p.Inputs {
		if values[input] == "" {
			return true
		}
	}
	return false
}

// RenderFrom is the file's bytes.
func (p *Part) RenderFrom(values map[string]string) ([]byte, error) {
	inputs := p.inputs(values)
	if p.Render == nil {
		return []byte(inputs[0]), nil
	}
	out, err := p.Render(inputs)
	return out, hperrors.Wrap(err)
}

func (p *Part) inputs(values map[string]string) []string {
	inputs := make([]string, len(p.Inputs))
	for i, name := range p.Inputs {
		inputs[i] = values[name]
	}
	return inputs
}

// RotationKey decides whether a part's file is new: it changes with the part's
// inputs and version, and with nothing else. It is keyed so that the name it
// ends up in says nothing of a password to whoever can list secrets.
func RotationKey(key []byte, typ base.SettingType, part *Part, values map[string]string) string {
	mac := hmac.New(sha256.New, key)
	fields := append([]string{string(typ), part.Name, strconv.Itoa(part.Version)}, part.inputs(values)...)
	for _, field := range fields {
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(len(field)))
		mac.Write(size[:])
		mac.Write([]byte(field))
	}
	return hex.EncodeToString(mac.Sum(nil))
}

func renderHtpasswd(values []string) ([]byte, error) {
	hash, err := htpasswd.HashPassword(values[1])
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return []byte(values[0] + ":" + hash + "\n"), nil
}

func sslCertValues(setting *entity.Setting) (map[string]string, error) {
	cert, err := setting.AsSSLCert()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	key, err := cert.PrivateKey.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return map[string]string{fieldCertificate: cert.Certificate, fieldPrivateKey: key,
		fieldCACertificate: cert.CACertificate}, nil
}

func sshKeyValues(setting *entity.Setting) (map[string]string, error) {
	sshKey, err := setting.AsSSHKey()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	key, err := sshKey.PrivateKey.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return map[string]string{fieldPrivateKey: key, fieldPublicKey: sshKey.PublicKey}, nil
}

func basicAuthValues(setting *entity.Setting) (map[string]string, error) {
	auth, err := setting.AsBasicAuth()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	password, err := auth.Password.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return map[string]string{fieldUsername: auth.Username, fieldPassword: password}, nil
}

func secretValues(setting *entity.Setting) (map[string]string, error) {
	secret, err := setting.AsSecret()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	value, err := secret.ValueAsBytes()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return map[string]string{fieldValue: string(value)}, nil
}

// configFileValues decodes a base64 config file itself: ContentAsBytes panics
// on bad base64, and a deployment is no place for that.
func configFileValues(setting *entity.Setting) (map[string]string, error) {
	configFile, err := setting.AsConfigFile()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !configFile.Base64 {
		return map[string]string{fieldContent: configFile.Content}, nil
	}
	content, err := base64.StdEncoding.DecodeString(configFile.Content)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return map[string]string{fieldContent: string(content)}, nil
}
