package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"server/settings"

	"github.com/labstack/echo"
)

func initSettings(e *echo.Echo) {
	e.GET("/settings", settingsPage)
	e.POST("/settings/read", settingsRead)
	e.POST("/settings/write", settingsWrite)
}

func settingsPage(c echo.Context) error {
	return c.Render(http.StatusOK, "settingsPage", nil)
}

func settingsRead(c echo.Context) error {
	return c.JSON(http.StatusOK, settings.Get())
}

func settingsWrite(c echo.Context) error {
	err := getJsSettings(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	settings.SaveSettings()
	return c.JSON(http.StatusOK, "Ok")
}

func getJsSettings(c echo.Context) error {
	buf, _ := ioutil.ReadAll(c.Request().Body)
	decoder := json.NewDecoder(bytes.NewBuffer(buf))
	err := decoder.Decode(settings.Get())
	if err != nil {
		if ute, ok := err.(*json.UnmarshalTypeError); ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, offset=%v", ute.Type, ute.Value, ute.Offset))
		} else if se, ok := err.(*json.SyntaxError); ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Syntax error: offset=%v, error=%v", se.Offset, se.Error()))
		} else {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}
	return nil
}
