package jsonl

import "time"

type Metadata struct {
	Name      string    `json:"name,omitempty"`
	Type      string    `json:"type,omitempty"`
	Version   string    `json:"version,omitempty"`
	Timestamp time.Time `json:"timestamp,omitzero"`
	Note      string    `json:"note,omitempty"`
}

type Chunk[T any] struct {
	Type    string `json:"type,omitempty"`
	Version string `json:"version,omitempty"`
	Note    string `json:"note,omitempty"`
	Data    T      `json:"data"`
}

func NewChunk[T any](typ string, data T) *Chunk[T] {
	return &Chunk[T]{
		Type: typ,
		Data: data,
	}
}
