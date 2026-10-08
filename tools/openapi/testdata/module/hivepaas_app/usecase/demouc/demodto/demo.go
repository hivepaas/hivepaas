package demodto

import (
	"encoding/json"

	"example.com/m/hivepaas_app/basedto"
)

type Meta struct {
	Total int `json:"total"`
}

// Size writes itself, as a string.
type Size int64

func (s Size) MarshalJSON() ([]byte, error) { return nil, nil }

type Tags []string

type Base struct {
	ID   string  `json:"id"`
	Note *string `json:"note"`
}

type ItemResp struct {
	Base
	Name    string            `json:"name"`
	Nick    string            `json:"nick,omitempty"`
	Meta    *Meta             `json:"meta"`
	Labels  map[string]string `json:"labels"`
	Tags    Tags              `json:"tags"`
	Default any               `json:"default,omitempty"`
	Raw     json.RawMessage   `json:"raw"`
	Params  map[string]any    `json:"params"`
	Size    Size              `json:"size"`
	Typed   any               `json:"typed" swaggertype:"string"`
	Hidden  string            `json:"-"`
	private string
}

type Filter struct {
	Status []string `json:"-" mapstructure:"status"`
}

type ListItemReq struct {
	Filter
	Search    string         `json:"-" mapstructure:"search"`
	ProjectID string         `json:"-"`
	Ignored   string         `json:"-" mapstructure:"-"`
	Paging    basedto.Paging `json:"-"`
}

func NewListItemReq() *ListItemReq { return &ListItemReq{} }

type GetItemReq struct {
	Reveal bool `json:"-" mapstructure:"reveal"`
}

func NewGetItemReq() *GetItemReq { return &GetItemReq{} }
