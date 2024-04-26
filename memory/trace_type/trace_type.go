package trace_type

import "fmt"

type Type string

const (
	Slice Type = "Slice"

	StackHeader Type = "StackHeader"
	StackNode   Type = "StackNode"

	MapHeader Type = "Map"
	MapTable  Type = "MapTable"

	StringFactory Type = "StringFactory"

	Shared Type = "Shared"
)

func StringFactoryHolds(s string) Type {
	return Type(fmt.Sprintf("%s[%5s...]", StringFactory, s))
}

func SkipTrace(_type Type) bool {
	switch _type {
	case StackNode, MapTable:
		return true
	default:
		return false
	}
}
