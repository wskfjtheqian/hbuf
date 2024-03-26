package parser

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"github.com/wskfjtheqian/hbuf_golang/pkg/hbuf"
)

// GetInfoReq 12
type GetInfoReq struct {
	UserId *hbuf.Int64 `json:"user_id"`

	Name *string `json:"name"`

	Age *int32 `json:"age"`
}

func (g *GetInfoReq) ToData() ([]byte, error) {
	return json.Marshal(g)
}

func (g *GetInfoReq) FormData(data []byte) error {
	return json.Unmarshal(data, g)
}

func (g *GetInfoReq) GetUserId() hbuf.Int64 {
	if nil == g.UserId {
		return hbuf.Int64(0)
	}
	return *g.UserId
}

func (g *GetInfoReq) SetUserId(val hbuf.Int64) {
	g.UserId = &val
}

func (g *GetInfoReq) GetName() string {
	if nil == g.Name {
		return ""
	}
	return *g.Name
}

func (g *GetInfoReq) SetName(val string) {
	g.Name = &val
}

func (g *GetInfoReq) GetAge() int32 {
	if nil == g.Age {
		return int32(0)
	}
	return *g.Age
}

func (g *GetInfoReq) SetAge(val int32) {
	g.Age = &val
}

type InfoReq struct {
	UserId *hbuf.Int64 `json:"user_id"`
}

func (g *InfoReq) ToData() ([]byte, error) {
	return json.Marshal(g)
}

func (g *InfoReq) FormData(data []byte) error {
	return json.Unmarshal(data, g)
}

func (g *InfoReq) GetUserId() hbuf.Int64 {
	if nil == g.UserId {
		return hbuf.Int64(0)
	}
	return *g.UserId
}

func (g *InfoReq) SetUserId(val hbuf.Int64) {
	g.UserId = &val
}

type InfoSet struct {
	UserId *hbuf.Int64 `json:"user_id"`

	Name *string `json:"name"`

	Age *int32 `json:"age"`
}

func (g *InfoSet) ToData() ([]byte, error) {
	return json.Marshal(g)
}

func (g *InfoSet) FormData(data []byte) error {
	return json.Unmarshal(data, g)
}

func (g *InfoSet) GetUserId() hbuf.Int64 {
	if nil == g.UserId {
		return hbuf.Int64(0)
	}
	return *g.UserId
}

func (g *InfoSet) SetUserId(val hbuf.Int64) {
	g.UserId = &val
}

func (g *InfoSet) GetName() string {
	if nil == g.Name {
		return ""
	}
	return *g.Name
}

func (g *InfoSet) SetName(val string) {
	g.Name = &val
}

func (g *InfoSet) GetAge() int32 {
	if nil == g.Age {
		return int32(0)
	}
	return *g.Age
}

func (g *InfoSet) SetAge(val int32) {
	g.Age = &val
}

// GetInfoResp 12
type GetInfoResp struct {
	V1 int8 `json:"v1"`

	B1 *int8 `json:"b1"`

	V2 int16 `json:"v2"`

	B2 *int16 `json:"b2"`

	V3 int32 `json:"v3"`

	B3 *int32 `json:"b3"`

	V4 hbuf.Int64 `json:"v4"`

	B4 *hbuf.Int64 `json:"b4"`

	V5 uint8 `json:"v5"`

	B5 *uint8 `json:"b5"`

	V6 uint16 `json:"v6"`

	B6 *uint16 `json:"b6"`

	V7 uint32 `json:"v7"`

	B7 *uint32 `json:"b7"`

	V8 hbuf.Uint64 `json:"v8"`

	B8 *hbuf.Uint64 `json:"b8"`

	V9 bool `json:"v9"`

	B9 *bool `json:"b9"`

	V10 float32 `json:"v10"`

	B10 *float32 `json:"b10"`

	V11 float64 `json:"v11"`

	B11 *float64 `json:"b11"`

	V12 string `json:"v12"`

	B12 *string `json:"b12"`

	V13 hbuf.Time `json:"v13"`

	B13 *hbuf.Time `json:"b13"`

	V14 decimal.Decimal `json:"v14"`

	B14 *decimal.Decimal `json:"b14"`

	V15 Status `json:"v15"`

	B15 *Status `json:"b15"`

	V16 GetInfoReq `json:"v16"`

	B16 *GetInfoReq `json:"b16"`

	V17 []GetInfoReq `json:"v17"`

	B17 []GetInfoReq `json:"b17"`

	V19 []*GetInfoReq `json:"v19"`

	B19 []*GetInfoReq `json:"b19"`

	V18 map[Status]GetInfoReq `json:"v18"`

	B18 map[string]GetInfoReq `json:"b18"`

	V20 map[string]*GetInfoReq `json:"v20"`

	B20 map[string]*GetInfoReq `json:"b20"`
}

func (g *GetInfoResp) ToData() ([]byte, error) {
	return json.Marshal(g)
}

func (g *GetInfoResp) FormData(data []byte) error {
	return json.Unmarshal(data, g)
}

func (g *GetInfoResp) GetV1() int8 {
	return g.V1
}

func (g *GetInfoResp) SetV1(val int8) {
	g.V1 = val
}

func (g *GetInfoResp) GetB1() int8 {
	if nil == g.B1 {
		return int8(0)
	}
	return *g.B1
}

func (g *GetInfoResp) SetB1(val int8) {
	g.B1 = &val
}

func (g *GetInfoResp) GetV2() int16 {
	return g.V2
}

func (g *GetInfoResp) SetV2(val int16) {
	g.V2 = val
}

func (g *GetInfoResp) GetB2() int16 {
	if nil == g.B2 {
		return int16(0)
	}
	return *g.B2
}

func (g *GetInfoResp) SetB2(val int16) {
	g.B2 = &val
}

func (g *GetInfoResp) GetV3() int32 {
	return g.V3
}

func (g *GetInfoResp) SetV3(val int32) {
	g.V3 = val
}

func (g *GetInfoResp) GetB3() int32 {
	if nil == g.B3 {
		return int32(0)
	}
	return *g.B3
}

func (g *GetInfoResp) SetB3(val int32) {
	g.B3 = &val
}

func (g *GetInfoResp) GetV4() hbuf.Int64 {
	return g.V4
}

func (g *GetInfoResp) SetV4(val hbuf.Int64) {
	g.V4 = val
}

func (g *GetInfoResp) GetB4() hbuf.Int64 {
	if nil == g.B4 {
		return hbuf.Int64(0)
	}
	return *g.B4
}

func (g *GetInfoResp) SetB4(val hbuf.Int64) {
	g.B4 = &val
}

func (g *GetInfoResp) GetV5() uint8 {
	return g.V5
}

func (g *GetInfoResp) SetV5(val uint8) {
	g.V5 = val
}

func (g *GetInfoResp) GetB5() uint8 {
	if nil == g.B5 {
		return uint8(0)
	}
	return *g.B5
}

func (g *GetInfoResp) SetB5(val uint8) {
	g.B5 = &val
}

func (g *GetInfoResp) GetV6() uint16 {
	return g.V6
}

func (g *GetInfoResp) SetV6(val uint16) {
	g.V6 = val
}

func (g *GetInfoResp) GetB6() uint16 {
	if nil == g.B6 {
		return uint16(0)
	}
	return *g.B6
}

func (g *GetInfoResp) SetB6(val uint16) {
	g.B6 = &val
}

func (g *GetInfoResp) GetV7() uint32 {
	return g.V7
}

func (g *GetInfoResp) SetV7(val uint32) {
	g.V7 = val
}

func (g *GetInfoResp) GetB7() uint32 {
	if nil == g.B7 {
		return uint32(0)
	}
	return *g.B7
}

func (g *GetInfoResp) SetB7(val uint32) {
	g.B7 = &val
}

func (g *GetInfoResp) GetV8() hbuf.Uint64 {
	return g.V8
}

func (g *GetInfoResp) SetV8(val hbuf.Uint64) {
	g.V8 = val
}

func (g *GetInfoResp) GetB8() hbuf.Uint64 {
	if nil == g.B8 {
		return hbuf.Uint64(0)
	}
	return *g.B8
}

func (g *GetInfoResp) SetB8(val hbuf.Uint64) {
	g.B8 = &val
}

func (g *GetInfoResp) GetV9() bool {
	return g.V9
}

func (g *GetInfoResp) SetV9(val bool) {
	g.V9 = val
}

func (g *GetInfoResp) GetB9() bool {
	if nil == g.B9 {
		return false
	}
	return *g.B9
}

func (g *GetInfoResp) SetB9(val bool) {
	g.B9 = &val
}

func (g *GetInfoResp) GetV10() float32 {
	return g.V10
}

func (g *GetInfoResp) SetV10(val float32) {
	g.V10 = val
}

func (g *GetInfoResp) GetB10() float32 {
	if nil == g.B10 {
		return float32(0)
	}
	return *g.B10
}

func (g *GetInfoResp) SetB10(val float32) {
	g.B10 = &val
}

func (g *GetInfoResp) GetV11() float64 {
	return g.V11
}

func (g *GetInfoResp) SetV11(val float64) {
	g.V11 = val
}

func (g *GetInfoResp) GetB11() float64 {
	if nil == g.B11 {
		return float64(0)
	}
	return *g.B11
}

func (g *GetInfoResp) SetB11(val float64) {
	g.B11 = &val
}

func (g *GetInfoResp) GetV12() string {
	return g.V12
}

func (g *GetInfoResp) SetV12(val string) {
	g.V12 = val
}

func (g *GetInfoResp) GetB12() string {
	if nil == g.B12 {
		return ""
	}
	return *g.B12
}

func (g *GetInfoResp) SetB12(val string) {
	g.B12 = &val
}

func (g *GetInfoResp) GetV13() hbuf.Time {
	return g.V13
}

func (g *GetInfoResp) SetV13(val hbuf.Time) {
	g.V13 = val
}

func (g *GetInfoResp) GetB13() hbuf.Time {
	if nil == g.B13 {
		return hbuf.Time{}
	}
	return *g.B13
}

func (g *GetInfoResp) SetB13(val hbuf.Time) {
	g.B13 = &val
}

func (g *GetInfoResp) GetV14() decimal.Decimal {
	return g.V14
}

func (g *GetInfoResp) SetV14(val decimal.Decimal) {
	g.V14 = val
}

func (g *GetInfoResp) GetB14() decimal.Decimal {
	if nil == g.B14 {
		return decimal.Zero
	}
	return *g.B14
}

func (g *GetInfoResp) SetB14(val decimal.Decimal) {
	g.B14 = &val
}

func (g *GetInfoResp) GetV15() Status {
	return g.V15
}

func (g *GetInfoResp) SetV15(val Status) {
	g.V15 = val
}

func (g *GetInfoResp) GetB15() Status {
	if nil == g.B15 {
		return Status(0)
	}
	return *g.B15
}

func (g *GetInfoResp) SetB15(val Status) {
	g.B15 = &val
}

func (g *GetInfoResp) GetV16() GetInfoReq {
	return g.V16
}

func (g *GetInfoResp) SetV16(val GetInfoReq) {
	g.V16 = val
}

func (g *GetInfoResp) GetB16() GetInfoReq {
	if nil == g.B16 {
		return GetInfoReq{}
	}
	return *g.B16
}

func (g *GetInfoResp) SetB16(val GetInfoReq) {
	g.B16 = &val
}

func (g *GetInfoResp) GetV17() []GetInfoReq {
	return g.V17
}

func (g *GetInfoResp) SetV17(val []GetInfoReq) {
	g.V17 = val
}

func (g *GetInfoResp) GetB17() []GetInfoReq {
	return g.B17
}

func (g *GetInfoResp) SetB17(val []GetInfoReq) {
	g.B17 = val
}

func (g *GetInfoResp) GetV19() []*GetInfoReq {
	return g.V19
}

func (g *GetInfoResp) SetV19(val []*GetInfoReq) {
	g.V19 = val
}

func (g *GetInfoResp) GetB19() []*GetInfoReq {
	return g.B19
}

func (g *GetInfoResp) SetB19(val []*GetInfoReq) {
	g.B19 = val
}

func (g *GetInfoResp) GetV18() map[Status]GetInfoReq {
	return g.V18
}

func (g *GetInfoResp) SetV18(val map[Status]GetInfoReq) {
	g.V18 = val
}

func (g *GetInfoResp) GetB18() map[string]GetInfoReq {
	return g.B18
}

func (g *GetInfoResp) SetB18(val map[string]GetInfoReq) {
	g.B18 = val
}

func (g *GetInfoResp) GetV20() map[string]*GetInfoReq {
	return g.V20
}

func (g *GetInfoResp) SetV20(val map[string]*GetInfoReq) {
	g.V20 = val
}

func (g *GetInfoResp) GetB20() map[string]*GetInfoReq {
	return g.B20
}

func (g *GetInfoResp) SetB20(val map[string]*GetInfoReq) {
	g.B20 = val
}
