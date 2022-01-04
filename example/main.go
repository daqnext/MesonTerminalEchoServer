package main

import (
	"math/rand"
	"net/http"
	"strings"
	"time"

	EchoServer "github.com/daqnext/MesonTerminalEchoServer"
	"github.com/labstack/echo/v4"
	"github.com/universe-30/EchoMiddleware"
	"github.com/universe-30/LogrusULog"
	"github.com/universe-30/ULog"
)

type testStruct struct {
	Somestring string  `json:"somestring"`
	Someint    int     `json:"someint"`
	Somefloat  float64 `json:"somefloat"`
}

func main() {
	logger, _ := LogrusULog.New("./logs", 2, 20, 30)
	logger.SetLevel(ULog.DebugLevel)

	var IgnoreHeader = map[string]struct{}{
		"Content-Length":              {},
		"Connection":                  {},
		"Server":                      {},
		"Last-Modified":               {},
		"Expires":                     {},
		"Access-Control-Allow-Origin": {},
		"Allow":                       {},
		"Content-Encoding":            {},
	}

	hs := EchoServer.New()

	//use jsoniter
	hs.UseJsoniter()
	//log request info
	hs.Use(EchoMiddleware.LoggerWithConfig(EchoMiddleware.LoggerConfig{
		Logger: logger,
	}))
	//use recover
	hs.Use(EchoMiddleware.RecoverWithConfig(EchoMiddleware.RecoverConfig{
		Logger: logger,
		OnPanic: func(panic_err interface{}) {

		},
	}))

	//////////start transmiting after 10 seconds////////
	hs.StaticWithPause("/", "assets", IgnoreHeader)
	hs.SetPauseSeconds(10)

	///////////////////  JSONP //////////////////////
	hs.GET("/jsonp", func(c echo.Context) error {
		callback := c.QueryParam("callback")
		var content struct {
			Response  string    `json:"response"`
			Timestamp time.Time `json:"timestamp"`
			Random    int       `json:"random"`
		}
		content.Response = "Sent via JSONP"
		content.Timestamp = time.Now().UTC()
		content.Random = rand.Intn(1000)
		return c.JSONP(http.StatusOK, callback, &content)
	})

	hs.GET("/jsonp2", func(c echo.Context) error {
		callback := c.QueryParam("callback")
		var content struct {
			Response  string    `json:"response"`
			Timestamp time.Time `json:"timestamp"`
			Random    int       `json:"random"`
		}
		content.Response = "Sent via JSONP"
		content.Timestamp = time.Now().UTC()
		content.Random = rand.Intn(1000)
		a := 1
		_ = 1 / (a - 1)
		return c.JSONP(http.StatusOK, callback, &content)
	})

	hs.GET("/sendfiletest/:filename", func(c echo.Context) error {
		name := c.Param("filename")
		needSavedHeader := true
		err := EchoServer.FileWithPause(hs, c, "assets/"+name, needSavedHeader, IgnoreHeader)
		if err != nil {
			logger.Debugln("file not found")
			//file missing
			//notice server
		}
		return err
	})

	hs.GET("/api/cdn/*", func(c echo.Context) error {
		//get bindname
		rPath := c.Param("*")
		logger.Debugln(rPath)
		s := strings.SplitN(rPath, "/", 2)
		for i, v := range s {
			logger.Debugln(i, v)
		}
		if len(s) < 1 {
			return c.String(200, "url Error:"+c.Request().RequestURI)
		}
		bindName := s[0]
		logger.Debugln("bindname", bindName)
		if bindName == "" {
			return c.String(200, "url Error:"+c.Request().RequestURI)
		}
		fileName := "index.html"
		if len(s) > 1 && s[1] != "" {
			fileName = s[1]
		}
		logger.Debugln("fileName", fileName)

		//logger.Debugln("referer",c.Request().Header["Referer"])
		//r,_:=url.Parse(c.Request().Header["Referer"][0])
		//logger.Debugln(r)

		return c.String(200, c.Request().RequestURI)
	})

	hs.GET("*", func(c echo.Context) error {
		uri := c.Request().RequestURI
		if uri == "/" {
			uri = "/index.html"
		}
		logger.Debugln(uri)

		//logger.Debugln("referer",c.Request().Header["Referer"])
		//r,_:=url.Parse(c.Request().Header["Referer"][0])
		//logger.Debugln(r)

		return c.String(200, uri)
	})

	hs.GET("/test/:bindname/*", func(c echo.Context) error {
		bindName := c.Param("bindname")
		logger.Debugln(bindName)
		uri := c.Request().RequestURI
		filePath := strings.Trim(uri, "/test/"+c.Param("bindname"))
		logger.Debugln(filePath)

		//logger.Debugln("referer",c.Request().Header["Referer"])
		//r,_:=url.Parse(c.Request().Header["Referer"][0])
		//logger.Debugln(r)

		return c.String(200, uri)
	})

	hs.GET("/test200301", func(c echo.Context) error {
		c.Response().Header().Add("location", "https://www.baidu.com")
		return c.HTML(302, "")
	})

	hs.POST("/testpost", func(c echo.Context) error {
		var ts testStruct
		if err := c.Bind(&ts); err != nil {
			return err
		}

		ts.Someint++
		ts.Somefloat += 10
		return c.JSON(http.StatusOK, ts)
	})

	hs.Server.SetKeepAlivesEnabled(false)

	errCh := make(chan error)
	go func() {
		errCh <- hs.Start(":8080")
	}()

	err := hs.WaitForServerStart(false)
	if err != nil {
		logger.Debugln(err)
	}
	logger.Infoln("echo server started")

	time.Sleep(1 * time.Hour)

}
