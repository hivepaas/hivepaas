package obi

import (
	"bufio"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
)

// sample is one line of Prometheus' text format: a series and its value.
type sample struct {
	Name   string
	Labels map[string]string
	Value  float64
}

var errBadLine = errors.New("not a Prometheus sample")

// parseText reads Prometheus' text exposition format, keeping the samples
// whose name keep says to. Comments are skipped, and so is a line it cannot
// read: one bad series does not lose the scrape.
func parseText(r io.Reader, keep func(name string) bool) ([]sample, error) {
	var out []sample
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64<<10), 1<<20) //nolint:mnd // long label sets
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		end := strings.IndexAny(line, "{ ")
		if end <= 0 || !keep(line[:end]) {
			continue
		}
		s, err := parseSample(line)
		if err != nil {
			continue
		}
		out = append(out, s)
	}
	return out, scanner.Err() //nolint:wrapcheck // the reader's
}

// parseSample reads `name{label="value",...} value [timestamp]`.
func parseSample(line string) (sample, error) {
	s := sample{Labels: map[string]string{}}
	i := strings.IndexAny(line, "{ ")
	s.Name = line[:i]
	rest := line[i:]
	if rest[0] == '{' {
		n, err := parseLabels(rest[1:], s.Labels)
		if err != nil {
			return s, err
		}
		rest = rest[1+n:]
	}
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return s, errBadLine
	}
	v, err := parseValue(fields[0])
	if err != nil {
		return s, errBadLine
	}
	s.Value = v
	return s, nil
}

// parseLabels reads `a="x",b="y"}` into labels, and says how many bytes it
// read, the closing brace included.
func parseLabels(in string, labels map[string]string) (int, error) {
	i := 0
	for i < len(in) {
		switch in[i] {
		case '}':
			return i + 1, nil
		case ',', ' ':
			i++
			continue
		}
		eq := strings.IndexByte(in[i:], '=')
		if eq <= 0 || i+eq+1 >= len(in) || in[i+eq+1] != '"' {
			return 0, errBadLine
		}
		name := in[i : i+eq]
		i += eq + 2 //nolint:mnd // past `="`
		var b strings.Builder
		for ; i < len(in) && in[i] != '"'; i++ {
			if in[i] == '\\' && i+1 < len(in) {
				i++
				switch in[i] {
				case 'n':
					b.WriteByte('\n')
				default: // \\ and \"
					b.WriteByte(in[i])
				}
				continue
			}
			b.WriteByte(in[i])
		}
		if i >= len(in) {
			return 0, errBadLine
		}
		i++ // the closing quote
		labels[name] = b.String()
	}
	return 0, errBadLine
}

// parseValue reads a sample's value, Prometheus' spellings of infinity and
// NaN included.
func parseValue(s string) (float64, error) {
	switch s {
	case "+Inf":
		return math.Inf(1), nil
	case "-Inf":
		return math.Inf(-1), nil
	case "NaN":
		return math.NaN(), nil
	}
	return strconv.ParseFloat(s, 64) //nolint:wrapcheck // the caller's to judge
}
