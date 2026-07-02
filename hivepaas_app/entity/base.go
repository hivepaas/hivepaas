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

func (o ObjectIDSlice) ToIDStringSlice() []string {
	res := make([]string, 0, len(o))
	for _, obj := range o {
		res = append(res, obj.ID)
	}
	return res
}
