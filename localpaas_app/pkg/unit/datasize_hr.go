package unit

// DataSizeHR human-readable data size, e.g. "1.5 GB", "500 MB", etc.
// Use this type for output only, it can't parse input.
type DataSizeHR int64

func (b DataSizeHR) String() string {
	return DataSize(b).HumanReadable()
}

func (b DataSizeHR) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b DataSizeHR) MarshalJSON() ([]byte, error) {
	return []byte(`"` + b.String() + `"`), nil
}
