package bunex

import "strings"

func MakeLikeOpStr(keyword string, quoteByPercent bool) string {
	keyword = strings.NewReplacer("%", "\\%", "*", "%").Replace(keyword)
	if quoteByPercent && !strings.HasPrefix(keyword, "%") {
		keyword = "%" + keyword
	}
	if quoteByPercent && !strings.HasSuffix(keyword, "%") {
		keyword += "%"
	}
	return strings.ToLower(keyword)
}
