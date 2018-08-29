package mods

import (
	"github.com/labstack/echo"
)

func InitMods(e *echo.Echo) {
	e.GET("/test", test)
}

func test(c echo.Context) error {
	return c.HTML(200, `
<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="utf-8">
	<script src="http://code.jquery.com/jquery-1.11.3.min.js"></script>
	<script src="/js/api.js"></script>
</head>
<body>
</body>`)
}
