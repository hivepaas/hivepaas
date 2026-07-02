package base

type BinObjectType string

const (
	BinObjectTypeUserPhoto    BinObjectType = "user-photo"
	BinObjectTypeProjectPhoto BinObjectType = "project-photo"
)

var (
	AllBinObjectTypes = []BinObjectType{BinObjectTypeUserPhoto, BinObjectTypeProjectPhoto}
)

type BinObjectStatus string

const (
	BinObjectStatusActive   BinObjectStatus = "active"
	BinObjectStatusDisabled BinObjectStatus = "disabled"
)

var (
	AllBinObjectStatuses = []BinObjectStatus{BinObjectStatusActive, BinObjectStatusDisabled}
)
