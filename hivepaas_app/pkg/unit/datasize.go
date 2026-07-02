// NOTE: source copied from https://github.com/c2h5oh/datasize with modification

package unit

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tracerr"
)

type DataSize int64

const (
	B  DataSize = 1
	KB          = B << 10
	MB          = KB << 10
	GB          = MB << 10
	TB          = GB << 10
	PB          = TB << 10
	EB          = PB << 10

	fnUnmarshalText string = "UnmarshalText"
	maxInt64        int64  = math.MaxInt64
	cutoff          int64  = maxInt64 / 10
)

func (b DataSize) Bytes() int64 {
	return int64(b)
}

func (b DataSize) KBytes() float64 {
	v := b / KB
	r := b % KB
	return float64(v) + float64(r)/float64(KB)
}

func (b DataSize) MBytes() float64 {
	v := b / MB
	r := b % MB
	return float64(v) + float64(r)/float64(MB)
}

func (b DataSize) GBytes() float64 {
	v := b / GB
	r := b % GB
	return float64(v) + float64(r)/float64(GB)
}

func (b DataSize) TBytes() float64 {
	v := b / TB
	r := b % TB
	return float64(v) + float64(r)/float64(TB)
}

func (b DataSize) PBytes() float64 {
	v := b / PB
	r := b % PB
	return float64(v) + float64(r)/float64(PB)
}

func (b DataSize) EBytes() float64 {
	v := b / EB
	r := b % EB
	return float64(v) + float64(r)/float64(EB)
}

func (b DataSize) Truncate(sz DataSize) DataSize {
	if sz == 0 {
		return b
	}
	if sz < 0 {
		sz = -sz
	}
	if b < 0 { // NOTE: we don't handle the case b == MinInt64
		return -((-b / sz) * sz)
	}
	return (b / sz) * sz
}

func (b DataSize) String() string {
	switch {
	case b == 0:
		return "0"
	case b%EB == 0:
		return fmt.Sprintf("%deb", b/EB)
	case b%PB == 0:
		return fmt.Sprintf("%dpb", b/PB)
	case b%TB == 0:
		return fmt.Sprintf("%dtb", b/TB)
	case b%GB == 0:
		return fmt.Sprintf("%dgb", b/GB)
	case b%MB == 0:
		return fmt.Sprintf("%dmb", b/MB)
	case b%KB == 0:
		return fmt.Sprintf("%dkb", b/KB)
	default:
		return fmt.Sprintf("%db", b)
	}
}

func (b DataSize) HR() string {
	return b.HumanReadable()
}

func (b DataSize) HumanReadable() string {
	if b >= GB {
		switch {
		case b >= EB:
			return fmt.Sprintf("%.1f EB", b.EBytes())
		case b >= PB:
			return fmt.Sprintf("%.1f PB", b.PBytes())
		case b >= TB:
			return fmt.Sprintf("%.1f TB", b.TBytes())
		default:
			return fmt.Sprintf("%.1f GB", b.GBytes())
		}
	} else {
		switch {
		case b >= MB:
			return fmt.Sprintf("%.1f MB", b.MBytes())
		case b >= KB:
			return fmt.Sprintf("%.1f KB", b.KBytes())
		case b >= B:
			return fmt.Sprintf("%d B", b)
		default:
			return "0"
		}
	}
}

func (b DataSize) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

//nolint:gocognit
func (b *DataSize) UnmarshalText(t []byte) error {
	var val int64
	var unit string

	// copy for error message
	t0 := t

	var c byte
	var i int

ParseLoop:
	for i < len(t) {
		c = t[i]
		switch {
		case '0' <= c && c <= '9':
			if val > cutoff {
				goto Overflow
			}

			c -= '0'
			val *= 10

			if val > val+int64(c) {
				// val+v overflows
				goto Overflow
			}
			val += int64(c)
			i++

		default:
			if i == 0 {
				goto SyntaxError
			}
			break ParseLoop
		}
	}

	unit = strings.TrimSpace(string(t[i:]))
	unit = strings.ToLower(unit)
	switch unit {
	case "", "b":
		// do nothing - already in bytes

	case "kb":
		if val > maxInt64/int64(KB) {
			goto Overflow
		}
		val *= int64(KB)

	case "mb":
		if val > maxInt64/int64(MB) {
			goto Overflow
		}
		val *= int64(MB)

	case "gb":
		if val > maxInt64/int64(GB) {
			goto Overflow
		}
		val *= int64(GB)

	case "tb":
		if val > maxInt64/int64(TB) {
			goto Overflow
		}
		val *= int64(TB)

	case "pb":
		if val > maxInt64/int64(PB) {
			goto Overflow
		}
		val *= int64(PB)

	case "eb":
		if val > maxInt64/int64(EB) {
			goto Overflow
		}
		val *= int64(EB)

	default:
		goto SyntaxError
	}

	*b = DataSize(val)
	return nil

Overflow:
	*b = DataSize(maxInt64)
	return &strconv.NumError{Func: fnUnmarshalText, Num: string(t0), Err: strconv.ErrRange}

SyntaxError:
	*b = 0
	return &strconv.NumError{Func: fnUnmarshalText, Num: string(t0), Err: strconv.ErrSyntax}
}

func (b DataSize) MarshalJSON() ([]byte, error) {
	return []byte(`"` + b.String() + `"`), nil
}

func (b *DataSize) UnmarshalJSON(in []byte) error {
	if bytes.Equal(in, []byte("null")) {
		*b = 0
		return nil
	}

	// Remove double quotes covering the str
	if len(in) > 1 && in[0] == '"' {
		in = in[1 : len(in)-1]
		d, err := ParseDataSize(in)
		if err != nil {
			return tracerr.Wrap(err)
		}
		*b = d
		return nil
	}

	// Parse unit as integer number
	v, err := strconv.ParseInt(reflectutil.UnsafeBytesToStr(in), 10, 64)
	if err != nil {
		return tracerr.Wrap(err)
	}
	*b = DataSize(v)
	return nil
}

func ParseDataSize(t []byte) (DataSize, error) {
	var v DataSize
	err := v.UnmarshalText(t)
	return v, err
}

func MustParseDataSize(t []byte) DataSize {
	v, err := ParseDataSize(t)
	if err != nil {
		panic(err)
	}
	return v
}

func ParseDataSizeString(s string) (DataSize, error) {
	return ParseDataSize([]byte(s))
}

func MustParseDataSizeString(s string) DataSize {
	return MustParseDataSize([]byte(s))
}
