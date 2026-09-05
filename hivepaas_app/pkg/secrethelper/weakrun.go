package secrethelper

import "unicode"

// keyboardRows are the character rows of a US QWERTY layout. A run along one of them
// ("qwer", "asdf", "rewq") is as easy to type and as easy to guess as "1234", but
// looks random to a check that only knows about code points.
var keyboardRows = []string{
	"`1234567890-=",
	"qwertyuiop[]\\",
	"asdfghjkl;'",
	"zxcvbnm,./",
}

// keyboardIndex maps a character to its column in each row it appears in. A character
// can only appear once per row, so one entry per row is enough.
var keyboardIndex = buildKeyboardIndex()

func buildKeyboardIndex() []map[rune]int {
	index := make([]map[rune]int, len(keyboardRows))
	for i, row := range keyboardRows {
		cols := make(map[rune]int, len(row))
		for col, r := range row {
			cols[r] = col
		}
		index[i] = cols
	}
	return index
}

// longestWeakRun returns the length of the longest stretch of characters that follow
// an obvious pattern: consecutive code points ("1234", "dcba"), the same character
// repeated ("aaaa"), or neighboring keys on one keyboard row ("qwer").
//
// A run of 1 means no pattern at all, so the result is never below 1 for a non-empty
// secret and is 0 only for an empty one.
func longestWeakRun(chars []rune) int {
	best := longestRepeatRun(chars)
	if run := longestStepRun(chars, codePointColumn); run > best {
		best = run
	}
	for i := range keyboardIndex {
		row := keyboardIndex[i]
		column := func(r rune) (int, bool) {
			col, ok := row[unicode.ToLower(r)]
			return col, ok
		}
		if run := longestStepRun(chars, column); run > best {
			best = run
		}
	}
	return best
}

// codePointColumn treats the code point itself as the position, which turns "abcd"
// and "4321" into runs along the same line the keyboard rows are read as.
func codePointColumn(r rune) (int, bool) {
	return int(unicode.ToLower(r)), true
}

// longestStepRun finds the longest stretch where each character sits one position
// away from the previous one, in a consistent direction.
//
// The direction has to stay consistent, otherwise "1213" and "aba" would read as runs
// and the rule would refuse secrets that hold no pattern at all.
func longestStepRun(chars []rune, column func(rune) (int, bool)) int {
	if len(chars) == 0 {
		return 0
	}
	best, run, delta := 1, 1, 0
	for i := 1; i < len(chars); i++ {
		prevCol, prevOK := column(chars[i-1])
		currCol, currOK := column(chars[i])
		step := 0
		if prevOK && currOK {
			step = currCol - prevCol
		}
		switch {
		case step != 1 && step != -1:
			run, delta = 1, 0
		case run == 1 || step == delta:
			run++
			delta = step
		default:
			// A valid step, but the direction turned: the last two characters start a
			// new run rather than extending the one that just ended.
			run, delta = 2, step
		}
		if run > best {
			best = run
		}
	}
	return best
}

// longestRepeatRun finds the longest stretch of one repeated character, compared
// without case so "AAaa" counts.
func longestRepeatRun(chars []rune) int {
	if len(chars) == 0 {
		return 0
	}
	best, run := 1, 1
	for i := 1; i < len(chars); i++ {
		if unicode.ToLower(chars[i]) == unicode.ToLower(chars[i-1]) {
			run++
		} else {
			run = 1
		}
		if run > best {
			best = run
		}
	}
	return best
}

// longestSharedRun returns how much of prev survives into secret, reading prev both
// forwards and backwards: reversing a secret is a common way to answer "pick a
// different one" without picking a different one.
func longestSharedRun(secret, prev []rune) int {
	best := longestCommonRun(secret, prev)
	if run := longestCommonRun(secret, reverseRunes(prev)); run > best {
		best = run
	}
	return best
}

func reverseRunes(chars []rune) []rune {
	out := make([]rune, len(chars))
	for i, r := range chars {
		out[len(chars)-1-i] = r
	}
	return out
}

// longestCommonRun returns the length of the longest substring shared by a and b,
// compared without case so bumping "myp@ssw0rd" to "MyP@ssw0rd" is not treated as a
// fresh secret.
//
// This needs both secrets in plaintext, which is why it can only ever run against a
// secret the caller already holds - see SecretStrengthRequirements.PrevSecrets.
func longestCommonRun(a, b []rune) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	lowerA := toLowerRunes(a)
	lowerB := toLowerRunes(b)

	// Only the previous row of the table is ever read, so two rows are enough.
	prev := make([]int, len(lowerB)+1)
	curr := make([]int, len(lowerB)+1)
	best := 0
	for i := 1; i <= len(lowerA); i++ {
		for j := 1; j <= len(lowerB); j++ {
			if lowerA[i-1] == lowerB[j-1] {
				curr[j] = prev[j-1] + 1
				if curr[j] > best {
					best = curr[j]
				}
			} else {
				curr[j] = 0
			}
		}
		prev, curr = curr, prev
	}
	return best
}

func toLowerRunes(chars []rune) []rune {
	out := make([]rune, len(chars))
	for i, r := range chars {
		out[i] = unicode.ToLower(r)
	}
	return out
}
