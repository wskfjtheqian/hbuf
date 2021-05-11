package golang

type EData struct {
	age int
}

func ExampleData_test() {
	bb := []interface{}{int32(100), int32(200)}

	aa := ListToBytes(0xffEEFFEE, &bb)
	println(len(aa))
	// output:
}
