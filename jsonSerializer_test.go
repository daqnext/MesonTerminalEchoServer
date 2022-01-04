package MesonTerminalEchoServer

import (
	"encoding/json"
	"fmt"
	"testing"

	jsoniter "github.com/json-iterator/go"
)

type testStruct struct {
	Somestring string  `json:"somestring"`
	Someint    int     `json:"someint"`
	Somefloat  float64 `json:"somefloat"`
}

var originS = `{
    "somestring":"this is a string",
    "someint":1,
    "somefloat":1.5
}`

func TestJsoniter_Deserialize(t *testing.T) {
	var jsoni = jsoniter.ConfigCompatibleWithStandardLibrary

	var data testStruct

	err := jsoni.Unmarshal([]byte(originS), &data)
	if err != nil {
		t.Log(err)
	}

	if ute, ok := err.(*json.UnmarshalTypeError); ok {
		t.Log(fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, field=%v, offset=%v", ute.Type, ute.Value, ute.Field, ute.Offset))
	} else if se, ok := err.(*json.SyntaxError); ok {
		t.Log(fmt.Sprintf("Syntax error: offset=%v, error=%v", se.Offset, se.Error()))
	}

	t.Log(data)
}
