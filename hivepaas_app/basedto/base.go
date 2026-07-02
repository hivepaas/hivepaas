package basedto

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

const (
	PageLimitDefault = 50
	PageLimitMax     = 10000
)

// Paging is used to store pagination request from client side
type Paging struct {
	Sort   Orders
	Offset int
	Limit  int
}

func (p *Paging) Orders() Orders {
	return p.Sort
}

func (p *Paging) OffsetEnd() int {
	return p.Offset + p.Limit - 1
}

func (p *Paging) ToPage() int {
	return p.Offset / gofn.Coalesce(p.Limit, 1)
}

func (p *Paging) ToPageSize() int {
	return p.Limit
}

type Direction string

const (
	DirectionAsc  Direction = "ASC"
	DirectionDesc Direction = "DESC"
)

type Order struct {
	Direction  Direction
	ColumnName string
}

func (o *Order) OrderBy() string {
	return o.ColumnName + " " + string(o.Direction)
}

type Orders []*Order

func (o Orders) Contain(columnName string) bool {
	for i := range o {
		if o[i].ColumnName == columnName {
			return true
		}
	}
	return false
}

// Add merge given orders with skipping duplicated items
func (o *Orders) Add(orders ...*Order) {
	for _, order := range orders {
		if !o.Contain(order.ColumnName) {
			*o = append(*o, order)
		}
	}
}

func (o Orders) ApplyMapping(columnMapping map[string]string) {
	if len(columnMapping) == 0 {
		return
	}
	for _, order := range o {
		if mappedColumn, ok := columnMapping[order.ColumnName]; ok {
			if mappedColumn == "" {
				mappedColumn = order.ColumnName
			}
			order.ColumnName = mappedColumn
		}
	}
}

// ReqModifier any request struct implements this interface will have method `ModifyRequest()`
// to be called when request is parsed.
type ReqModifier interface {
	ModifyRequest() error
}

// ReqValidator any request struct implements this interface will have method `Validate()`
// to be called when request is parsed.
type ReqValidator interface {
	Validate() apperrors.ValidationErrors
}

// ReqParsingErrorHandler any request struct implements this interface will have method `HandleParsingError()`
// to be called when parsing error is raised.
type ReqParsingErrorHandler interface {
	HandleParsingError(err error) error
}

// Meta metadata of single entity response
type Meta struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// ListMeta metadata of list api
type ListMeta struct {
	Meta
	Page *PagingMeta `json:"page"`
}

// PagingMeta metadata of pagination
type PagingMeta struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
}
