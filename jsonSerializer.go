package MesonTerminalEchoServer

import (
	"encoding/json"
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
)

type JsoniterHandler struct {
	json jsoniter.API
}

func NewJsoniter() *JsoniterHandler {
	return &JsoniterHandler{
		jsoniter.ConfigCompatibleWithStandardLibrary,
	}
}

func (j *JsoniterHandler) Serialize(c echo.Context, i interface{}, indent string) error {
	enc := j.json.NewEncoder(c.Response())
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(i)
}

func (j *JsoniterHandler) Deserialize(c echo.Context, i interface{}) error {
	err := j.json.NewDecoder(c.Request().Body).Decode(i)
	if ute, ok := err.(*json.UnmarshalTypeError); ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, field=%v, offset=%v", ute.Type, ute.Value, ute.Field, ute.Offset)).SetInternal(err)
	} else if se, ok := err.(*json.SyntaxError); ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Syntax error: offset=%v, error=%v", se.Offset, se.Error())).SetInternal(err)
	}
	return err
}
