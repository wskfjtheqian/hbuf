package golang

import "encoding/json"

type DData struct {
	Name string
	Info map[string]string
}

func (r *DData) ToBytes() []byte {
	//return JoinBytes(Data, 0x1111, func() ([]byte, int) {
	//	temp := []byte{}
	//	temp = append(temp, ToBytes(Int64, 1020, r.Name)...)
	//	temp = append(temp, ToBytes(Map, 2020, r.Info)...)
	//	return temp, len(temp)
	//})
	return nil
}

type EData struct {
	DData
	Age int
}

func (r *EData) ToBytes() []byte {
	//return JoinBytes(Data, 0x2222, func() ([]byte, int) {
	//	temp := []byte{}
	//	temp = append(temp, ToBytes(Int64, 10, r.Age)...)
	//	temp = append(temp, r.DData.ToBytes()...)
	//	return temp, len(temp)
	//})
	return nil
}

func ExampleData_idToBytes() {
	for i := 0; i <= 0xfffffff; i++ {
		ee, _ := idToBytes(uint32(i))
		aa, _ := idFromBytes(ee)
		if aa != uint32(i) {
			println(i)
			println(aa)
		}
	}

	// output:
}
func ExampleData_test() {
	var ed = EData{
		Age: 8522585452,
		DData: DData{
			Name: "heqian",
			Info: map[string]string{
				"Headimage": "Https://www.baidu.com",
				"Type":      "Manager",
			},
		},
	}

	aa := ed.ToBytes()
	println(len(aa))

	//ee := toBytes(nil, nil, &ed)
	//println(len(ee))

	ff, _ := json.Marshal(ed)
	println(len(ff))
	// output:
}
