package core

type FieldType int32

const (
	UnknownType FieldType = iota
	StringType
	IntType
	Int64Type
	UintType
	Uint64Type
	Float32Type
	Float64Type
	DurationType
)

type Field struct {
	Key       string
	Value     interface{}
	Type      FieldType
	StringVal string
	Int64Val  int64
}
