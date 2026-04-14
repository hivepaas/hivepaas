package bunex

import "github.com/uptrace/bun"

func List(slice any) any {
	return bun.List(slice)
}

func Safe(column string) bun.Safe {
	return bun.Safe(column)
}
