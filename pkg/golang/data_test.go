package golang

import "encoding/json"

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
	var ed = EData{
		Age: 16,
		DData: DData{
			Name: "heqian",
		},
	}

	aa := ed.ToBytes()
	println(len(aa))

	ee := toBytes(nil, nil, &ed)
	println(len(ee))

	ff, _ := json.Marshal(ed)
	println(len(ff))
	// output:
}
