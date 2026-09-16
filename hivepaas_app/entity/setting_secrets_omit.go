package entity

import (
	"reflect"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// OmitSecrets clears every encrypted value a setting holds, returning how many
// it cleared.
//
// It clears rather than leaves alone, which is the whole point. An
// EncryptedField marshals as the ciphertext it was loaded with, so a setting
// written out untouched carries its secrets in a form that works on exactly one
// installation - the one that holds the data key. Anywhere else the value is
// present, looks like a secret, and is silently useless. Removing it says
// plainly that the secret was not exported, and lets an import ask for one.
//
// The field is zeroed rather than dropped from the document, so the key stays
// visible and the reader can see there is a secret to supply.
func OmitSecrets(data SettingData) (int, error) {
	if data == nil {
		return 0, nil
	}
	count, err := omitSecretsIn(reflect.ValueOf(data))
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	return count, nil
}

// omitSecretsIn walks the same shape reencryptValue does, and for the same
// reason: only typed values reach exactly the fields that really are encrypted.
// A textual pass would also hit HashField, which shares the stored prefix but
// is a hash rather than a secret.
func omitSecretsIn(value reflect.Value) (int, error) {
	switch value.Kind() { //nolint:exhaustive
	case reflect.Pointer, reflect.Interface:
		if value.IsNil() {
			return 0, nil
		}
		return omitSecretsIn(value.Elem())

	case reflect.Struct:
		if value.Type() == encryptedFieldType {
			return clearEncryptedField(value)
		}
		count := 0
		for i := range value.NumField() {
			if !value.Type().Field(i).IsExported() {
				continue
			}
			n, err := omitSecretsIn(value.Field(i))
			if err != nil {
				return 0, err
			}
			count += n
		}
		return count, nil

	case reflect.Slice, reflect.Array:
		count := 0
		for i := range value.Len() {
			n, err := omitSecretsIn(value.Index(i))
			if err != nil {
				return 0, err
			}
			count += n
		}
		return count, nil

	case reflect.Map:
		return omitSecretsInMap(value)

	default:
		return 0, nil
	}
}

func omitSecretsInMap(value reflect.Value) (int, error) {
	if value.IsNil() || !typeHoldsEncryptedField(value.Type().Elem()) {
		return 0, nil
	}
	count := 0
	for _, key := range value.MapKeys() {
		entry := reflect.New(value.Type().Elem()).Elem()
		entry.Set(value.MapIndex(key))

		n, err := omitSecretsIn(entry)
		if err != nil {
			return 0, err
		}
		if n > 0 {
			value.SetMapIndex(key, entry)
			count += n
		}
	}
	return count, nil
}

func clearEncryptedField(value reflect.Value) (int, error) {
	if !value.CanSet() {
		return 0, hperrors.Wrap(hperrors.ErrInternal).
			WithMsgLog("encrypted field is not settable, cannot omit it")
	}
	field, _ := value.Addr().Interface().(*EncryptedField)
	if field.IsEmpty() {
		return 0, nil
	}
	value.Set(reflect.Zero(encryptedFieldType))
	return 1, nil
}

// SecretPlaintexts maps each of a setting's stored ciphertexts to its plaintext.
//
// It exists because a secret cannot be marshaled in the clear, by design.
// EncryptedField.MarshalJSON always emits ciphertext: it returns the value it
// was loaded with, and seals the plaintext when there is none. Calling Decrypt
// first does not change that - Decrypt fills the plaintext beside the
// ciphertext rather than replacing it, which is exactly right for re-saving an
// unrelated setting and exactly wrong for writing a readable export.
//
// So an export that must show secrets marshals as usual and then substitutes,
// using this map. Ciphertexts are sealed with a random nonce each time, so two
// fields never collide even when their plaintext is identical.
func SecretPlaintexts(data SettingData) (map[string]string, error) {
	if data == nil {
		return nil, nil
	}
	out := map[string]string{}
	if err := collectSecretPlaintexts(reflect.ValueOf(data), out); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return out, nil
}

func collectSecretPlaintexts(value reflect.Value, out map[string]string) error {
	switch value.Kind() { //nolint:exhaustive
	case reflect.Pointer, reflect.Interface:
		if value.IsNil() {
			return nil
		}
		return collectSecretPlaintexts(value.Elem(), out)

	case reflect.Struct:
		if value.Type() == encryptedFieldType {
			return collectOneSecret(value, out)
		}
		for i := range value.NumField() {
			if !value.Type().Field(i).IsExported() {
				continue
			}
			if err := collectSecretPlaintexts(value.Field(i), out); err != nil {
				return err
			}
		}
		return nil

	case reflect.Slice, reflect.Array:
		for i := range value.Len() {
			if err := collectSecretPlaintexts(value.Index(i), out); err != nil {
				return err
			}
		}
		return nil

	case reflect.Map:
		if value.IsNil() || !typeHoldsEncryptedField(value.Type().Elem()) {
			return nil
		}
		for _, key := range value.MapKeys() {
			entry := reflect.New(value.Type().Elem()).Elem()
			entry.Set(value.MapIndex(key))
			if err := collectSecretPlaintexts(entry, out); err != nil {
				return err
			}
		}
		return nil

	default:
		return nil
	}
}

func collectOneSecret(value reflect.Value, out map[string]string) error {
	if !value.CanAddr() {
		return nil
	}
	field, _ := value.Addr().Interface().(*EncryptedField)
	if field.IsEmpty() {
		return nil
	}
	ciphertext, err := field.GetEncrypted()
	if err != nil {
		return hperrors.Wrap(err)
	}
	plaintext, err := field.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if ciphertext != "" {
		out[ciphertext] = plaintext
	}
	return nil
}
