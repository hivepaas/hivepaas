package entity

// IDEntity base interface for every entity kind which has `ID`
type IDEntity interface {
	GetID() string
}

// NamedEntity base interface for every entity kind which has `name`
type NamedEntity interface {
	GetName() string
}

type ObjectID struct {
	ID string `json:"id"`
}

type ObjectIDSlice []*ObjectID

// Unique is the slice without its empty ids and with each id once, where it
// first is.
func (o ObjectIDSlice) Unique() ObjectIDSlice {
	res := make(ObjectIDSlice, 0, len(o))
	seen := make(map[string]struct{}, len(o))
	for _, obj := range o {
		if obj == nil || obj.ID == "" {
			continue
		}
		if _, ok := seen[obj.ID]; ok {
			continue
		}
		seen[obj.ID] = struct{}{}
		res = append(res, obj)
	}
	return res
}

func (o ObjectIDSlice) ToIDStringSlice() []string {
	res := make([]string, 0, len(o))
	for _, obj := range o {
		res = append(res, obj.ID)
	}
	return res
}

type ObjectValue struct {
	ID    string `json:"id,omitempty"`
	Value string `json:"value,omitempty"`
}

func (o ObjectValue) IsValid() bool {
	return o.ID != "" || o.Value != ""
}
