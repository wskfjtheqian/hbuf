package golang

type DData struct {
	Name string
}

func (r *DData) ToBytes() []byte {
	return JoinBytes(Data, 0x1111, func() ([]byte, int) {
		temp := []byte{}
		temp = append(temp, ToBytes(Int64, 10, r.Name)...)
		return temp, len(temp)
	})
}

type EData struct {
	DData
	Age int
}

func (r *EData) ToBytes() []byte {
	return JoinBytes(Data, 0x2222, func() ([]byte, int) {
		temp := []byte{}
		temp = append(temp, ToBytes(Int64, 10, r.Age)...)
		temp = append(temp, r.DData.ToBytes()...)
		return temp, len(temp)
	})
}

func ExampleData_test() {
	bb := ToBytes(Bool, 0x0808, true)
	FromBytes(bb, func(typ *Type, id *int, data *interface{}) {

	})
	// output:
}
