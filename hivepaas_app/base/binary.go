package base

type BinObjectType string

const (
	BinObjectTypeObjectIcon BinObjectType = "icon"
)

var (
	AllBinObjectTypes = []BinObjectType{BinObjectTypeObjectIcon}
)

type BinObjectStatus string

const (
	BinObjectStatusActive   BinObjectStatus = "active"
	BinObjectStatusDisabled BinObjectStatus = "disabled"
)

var (
	AllBinObjectStatuses = []BinObjectStatus{BinObjectStatusActive, BinObjectStatusDisabled}
)
