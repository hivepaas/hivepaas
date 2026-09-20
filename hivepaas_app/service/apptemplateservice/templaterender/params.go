package templaterender

import (
	"crypto/rand"
	"fmt"
	"maps"
	"math/big"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const (
	alnumAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	hexAlphabet   = "0123456789abcdef"
)

// Value is one parameter resolved to its canonical, typed value.
type Value struct {
	Param *templatemodel.Parameter
	// Value is a string, int64, bool or unit.DataSize according to Param.Type,
	// or nil for an optional parameter nobody set.
	Value     any
	Generated bool
}

// Text is the value as text: what the app-template setting stores, and what a
// placeholder inside a longer string is replaced with.
func (v *Value) Text() string {
	if v.Value == nil {
		return ""
	}
	return fmt.Sprint(v.Value)
}

// ResolveParams checks input against a template's parameters, fills in defaults
// and generates the secrets nobody gave. A parameter the template does not
// declare is refused rather than ignored: it is a form and a template that
// disagree, and ignoring it would provision without the value somebody meant.
func ResolveParams(defs []*templatemodel.Parameter, input map[string]any) (map[string]*Value, error) {
	declared := make(map[string]bool, len(defs))
	for _, def := range defs {
		declared[def.Name] = true
	}
	for _, name := range slices.Sorted(maps.Keys(input)) {
		if !declared[name] {
			return nil, paramInvalid(name, "the template declares no such parameter")
		}
	}

	out := make(map[string]*Value, len(defs))
	for _, def := range defs {
		value, err := resolveParam(def, input[def.Name])
		if err != nil {
			return nil, err
		}
		out[def.Name] = value
	}
	return out, nil
}

// ValidateDefaults checks every declared default against its own parameter's
// constraints. The linter runs it; a default that fails its own pattern would
// otherwise surface as a form that cannot be submitted unchanged.
func ValidateDefaults(defs []*templatemodel.Parameter) error {
	for _, def := range defs {
		if isBlank(def.Default) {
			continue
		}
		if _, err := convertParam(def, def.Default); err != nil {
			return err
		}
	}
	return nil
}

func resolveParam(def *templatemodel.Parameter, raw any) (*Value, error) {
	if isBlank(raw) {
		raw = def.Default
	}
	if isBlank(raw) {
		switch {
		case def.Type == templatemodel.ParamTypeSecret && def.Generate != nil:
			secret, err := generateSecret(def.Generate)
			if err != nil {
				return nil, err
			}
			return &Value{Param: def, Value: secret, Generated: true}, nil
		case def.Optional:
			return &Value{Param: def}, nil
		default:
			return nil, paramInvalid(def.Name, "a value is required")
		}
	}

	value, err := convertParam(def, raw)
	if err != nil {
		return nil, err
	}
	return &Value{Param: def, Value: value}, nil
}

// domainPattern is the host name an app can be reached at. A wildcard is not one
// of them: it is what a certificate covers, not what a request arrives at.
var domainPattern = regexp.MustCompile(
	`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)+$`)

// appKeyPattern is the shape projecthelper.CalcAppKey produces: slugified, and
// with every dash replaced by an underscore. Whether the app exists is not a
// question rendering can answer - the caller with the database decides that.
var appKeyPattern = regexp.MustCompile(`^[a-z0-9_][a-z0-9_]{0,99}$`)

func isBlank(v any) bool {
	if v == nil {
		return true
	}
	text, isText := v.(string)
	return isText && text == ""
}

func convertParam(def *templatemodel.Parameter, raw any) (any, error) {
	switch def.Type {
	case templatemodel.ParamTypeString, templatemodel.ParamTypeSecret:
		text, ok := raw.(string)
		if !ok {
			return nil, paramInvalid(def.Name, "must be text")
		}
		return text, checkText(def, text)
	case templatemodel.ParamTypeInt:
		number, ok := toInt(raw)
		if !ok {
			return nil, paramInvalid(def.Name, "must be a whole number")
		}
		return number, checkInt(def, number)
	case templatemodel.ParamTypeSize:
		size, ok := toSize(raw)
		if !ok {
			return nil, paramInvalid(def.Name, "must be a size such as 512MB")
		}
		return size, checkSize(def, size)
	case templatemodel.ParamTypeBool:
		flag, ok := toBool(raw)
		if !ok {
			return nil, paramInvalid(def.Name, "must be true or false")
		}
		return flag, nil
	case templatemodel.ParamTypeSelect:
		text, ok := raw.(string)
		if !ok || !slices.ContainsFunc(def.Options, func(o *templatemodel.SelectOption) bool {
			return o.Value == text
		}) {
			return nil, paramInvalid(def.Name, "must be one of the template's options")
		}
		return text, nil
	case templatemodel.ParamTypeVolume:
		text, ok := raw.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return nil, paramInvalid(def.Name, "must name a volume")
		}
		return text, nil
	case templatemodel.ParamTypeApp:
		text, ok := raw.(string)
		if !ok {
			return nil, paramInvalid(def.Name, "must name an app")
		}
		text = strings.TrimSpace(text)
		if !appKeyPattern.MatchString(text) {
			return nil, paramInvalid(def.Name, "must be an app key: lowercase letters, digits and _")
		}
		return text, nil
	case templatemodel.ParamTypeDomain:
		text, ok := raw.(string)
		if !ok {
			return nil, paramInvalid(def.Name, "must be a host name")
		}
		text = strings.TrimSpace(strings.ToLower(text))
		if !domainPattern.MatchString(text) {
			return nil, paramInvalid(def.Name, "must be a host name such as app.example.com")
		}
		return text, nil
	default:
		return nil, paramInvalid(def.Name, fmt.Sprintf("type %q is not supported", def.Type))
	}
}

func checkText(def *templatemodel.Parameter, text string) error {
	length := utf8.RuneCountInString(text)
	if def.MinLength != nil && length < *def.MinLength {
		return paramInvalid(def.Name, fmt.Sprintf("must be at least %d characters", *def.MinLength))
	}
	if def.MaxLength != nil && length > *def.MaxLength {
		return paramInvalid(def.Name, fmt.Sprintf("must be at most %d characters", *def.MaxLength))
	}
	if def.Pattern != "" {
		matched, err := regexp.MatchString(def.Pattern, text)
		if err != nil || !matched {
			return paramInvalid(def.Name, fmt.Sprintf("must match %s", def.Pattern))
		}
	}
	return nil
}

func checkInt(def *templatemodel.Parameter, number int64) error {
	minValue, maxValue := def.IntBounds()
	if minValue != nil && number < *minValue {
		return paramInvalid(def.Name, fmt.Sprintf("must be at least %d", *minValue))
	}
	if maxValue != nil && number > *maxValue {
		return paramInvalid(def.Name, fmt.Sprintf("must be at most %d", *maxValue))
	}
	return nil
}

func checkSize(def *templatemodel.Parameter, size unit.DataSize) error {
	minValue, maxValue := def.SizeBounds()
	if minValue != nil && size < *minValue {
		return paramInvalid(def.Name, fmt.Sprintf("must be at least %s", minValue.HR()))
	}
	if maxValue != nil && size > *maxValue {
		return paramInvalid(def.Name, fmt.Sprintf("must be at most %s", maxValue.HR()))
	}
	return nil
}

func toInt(raw any) (int64, bool) {
	if text, isText := raw.(string); isText {
		number, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
		return number, err == nil
	}
	return templatemodel.ToInt64(raw)
}

func toSize(raw any) (unit.DataSize, bool) {
	text, isText := raw.(string)
	if !isText {
		return 0, false
	}
	size, err := unit.ParseDataSizeString(strings.TrimSpace(text))
	return size, err == nil
}

func toBool(raw any) (bool, bool) {
	switch v := raw.(type) {
	case bool:
		return v, true
	case string:
		flag, err := strconv.ParseBool(strings.TrimSpace(v))
		return flag, err == nil
	default:
		return false, false
	}
}

// generateSecret draws from crypto/rand. rand.Int rejects rather than takes a
// modulus, so every character of the alphabet is equally likely.
func generateSecret(gen *templatemodel.Generate) (string, error) {
	alphabet := alnumAlphabet
	if gen.Charset == templatemodel.CharsetHex {
		alphabet = hexAlphabet
	}
	limit := big.NewInt(int64(len(alphabet)))
	out := make([]byte, gen.Length)
	for i := range out {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", hperrors.Wrap(err)
		}
		out[i] = alphabet[n.Int64()]
	}
	return string(out), nil
}

func paramInvalid(name, reason string) error {
	return hperrors.Wrap(hperrors.ErrAppTemplateParamInvalid).WithParam("Name", name).
		WithExtraDetail("%s: %s", name, reason)
}
