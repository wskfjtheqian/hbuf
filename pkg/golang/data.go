package golang

type Type int

const (
	Server  Type = iota // | type[1] | id | len[uint] | data |
	Data                // | type[1] | id | len[uint] | data |
	Int8                // | type[1] | id | data[1] |
	Int16               // | type[1] | id | data[2] |
	Int32               // | type[1] | id | data[4] |
	Int64               // | type[1] | id | data[8] |
	UInt8               // | type[1] | id | data[1] |
	UInt16              // | type[1] | id | data[2] |
	UInt32              // | type[1] | id | data[4] |
	UInt64              // | type[1] | id | data[8] |
	Float               // | type[1] | id | data[4] |
	Double              // | type[1] | id | data[8] |
	StrUTF8             // | type[1] | id | len[uint] | data |
	List                // | type[1] | id | len[uint] | val[1] | data |
	Map                 // | type[1] | id | len[uint] | key[1] | val[1] | data |
)
