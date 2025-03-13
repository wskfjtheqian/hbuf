package parser

import (
	"encoding/json"
)

type Test struct {
	Id int32 `json:"id,omitempty"` //
}

func (g *Test) ToData() ([]byte, error) {
	return json.Marshal(g)
}

func (g *Test) FormData(data []byte) error {
	return json.Unmarshal(data, g)
}

func (g *Test) GetId() int32 {
	return g.Id
}

func (g *Test) SetId(val int32) {
	g.Id = val
}
