package basedto

import (
	"fmt"
	"regexp"

	vld "github.com/tiendc/go-validator"
)

const (
	alpnProtocolsMax   = 10
	alpnProtocolMaxLen = 255
)

// An ALPN protocol ID is up to 255 bytes; the registered ones, and any a client
// is configured with, are printable ASCII without spaces, such as "tds/8.0".
var alpnProtocolRegex = regexp.MustCompile(`^[\x21-\x7e]+$`)

func ValidateALPNProtocols(protocols []string, field string) (res []vld.Validator) {
	res = append(res, ValidateSliceEx(protocols, true, 0, alpnProtocolsMax, nil, field)...)
	for i := range protocols {
		itemField := fmt.Sprintf("%s[%d]", field, i)
		res = append(res,
			vld.StrLen(&protocols[i], 1, alpnProtocolMaxLen).OnError(
				vld.SetField(itemField, nil),
				vld.SetCustomKey("ERR_VLD_FIELD_LENGTH_INVALID"),
			),
			vld.StrByteMatch(&protocols[i], alpnProtocolRegex).OnError(
				vld.SetField(itemField, nil),
				vld.SetCustomKey("ERR_VLD_VALUE_INVALID"),
			),
		)
	}
	return res
}
