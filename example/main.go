package main

import (
	"bufio"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
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

func readHeader(filePath string) (map[string][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	buf := bufio.NewReader(f)
	header := map[string][]string{}
	for {
		//key
		key, _, err := buf.ReadLine()
		if err != nil {
			if err == io.EOF { //read end
				return header, nil
			}
			return nil, err
		}
		//value count
		countStr, _, err := buf.ReadLine()
		count, err := strconv.Atoi(string(countStr))
		if count <= 0 {
			continue
		}
		header[string(key)] = []string{}
		for i := 0; i < count; i++ {
			value, _, _ := buf.ReadLine()
			header[string(key)] = append(header[string(key)], string(value))
		}
	}
}

func main() {
	logger, _ := LogrusULog.New("./logs", 2, 20, 30)
	logger.SetLevel(ULog.InfoLevel)

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
	//log request info in Debug level
	hs.Use(EchoMiddleware.LoggerWithConfig(EchoMiddleware.LoggerConfig{
		Logger:            logger,
		RecordFailRequest: true, // log failed request with Error
	}))
	//use recover to handle panic
	hs.Use(EchoMiddleware.RecoverWithConfig(EchoMiddleware.RecoverConfig{
		OnPanic: func(panic_err interface{}) {
			logger.Errorln(panic_err)
		},
	}))

	///////////// api //////////
	hs.GET("/redirect", func(c echo.Context) error {

		return c.Redirect(http.StatusTemporaryRedirect, "http://127.0.0.1:8080/somefile/some/aaa.jpg#randkey=123456")
	})

	hs.GET("/test1", func(c echo.Context) error {
		sign := c.Request().Header.Get("Signature")
		log.Println(sign)

		var content struct {
			Response  string    `json:"response"`
			Timestamp time.Time `json:"timestamp"`
			Random    int       `json:"random"`
		}
		content.Response = "Response"
		content.Timestamp = time.Now().UTC()
		content.Random = rand.Intn(1000)
		return c.JSON(http.StatusOK, &content)
	})

	//example panic in handler
	hs.GET("/test2", func(c echo.Context) error {
		var content struct {
			Response  string    `json:"response"`
			Timestamp time.Time `json:"timestamp"`
			Random    int       `json:"random"`
		}
		content.Response = "Response"
		content.Timestamp = time.Now().UTC()
		content.Random = rand.Intn(1000)

		// panic
		a := 1
		_ = 3 / (a - 1)

		return c.JSON(http.StatusOK, &content)
	})

	//example request a file in server
	hs.GET("/sendfiletest/:filename", func(c echo.Context) error {
		name := c.Param("filename")
		header, err := readHeader("assets/" + name + ".header")
		if err != nil {
			log.Println("readHeader error", err)
		}
		err = EchoServer.FileWithPause(hs, c, "assets/"+name, header, IgnoreHeader)
		if err != nil {
			log.Println("file not found")
		}
		return err
	})

	//example a cdn file request http://127.0.0.1:8080/api/cdn/somefile/path/filename.jpg?randkey=123456
	hs.GET("/api/cdn/*", func(c echo.Context) error {
		//get bindname
		rPath := c.Param("*")
		log.Println(rPath)
		s := strings.SplitN(rPath, "/", 2)
		for i, v := range s {
			log.Println(i, v)
		}
		if len(s) < 1 {
			return c.String(200, "url Error:"+c.Request().RequestURI)
		}
		bindName := s[0]
		log.Println("bindname", bindName)
		if bindName == "" {
			return c.String(200, "url Error:"+c.Request().RequestURI)
		}
		fileName := "index.html"
		if len(s) > 1 && s[1] != "" {
			fileName = s[1]
		}
		log.Println("fileName", fileName)

		return c.String(200, c.Request().RequestURI)
	})

	//example request http://127.0.0.1:8080/somefile/path/filename.jpg?randkey=123456
	hs.GET("*", func(c echo.Context) error {

		//uri
		uri := c.Request().RequestURI
		if uri == "/" {
			uri = "/index.html"
		}
		log.Println("uri=", uri)
		// output: uri= /somefile/path/filename.jpg?randkey=123456

		//

		//host
		log.Println("host=", c.Request().Host)

		//path
		log.Println("path=", c.Param("*"))
		// output: path= somefile/path/filename.jpg

		//param
		log.Println("QueryParam randkey=", c.QueryParam("b"))
		// output: QueryParam randkey= 123456

		return c.String(200, uri)
	})

	hs.GET("/test/:bindname/*", func(c echo.Context) error {
		bindName := c.Param("bindname")
		log.Println(bindName)
		uri := c.Request().RequestURI
		filePath := strings.Trim(uri, "/test/"+c.Param("bindname"))
		log.Println(filePath)

		return c.String(200, uri)
	})

	//post test
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

	//start echo server
	errCh := make(chan error)
	go func() {
		errCh <- hs.Start(":8080")
	}()

	//check start
	err := hs.WaitForServerStart(false)
	if err != nil {
		logger.Debugln(err)
	}
	logger.Infoln("echo server started")

	time.Sleep(1 * time.Hour)
}
