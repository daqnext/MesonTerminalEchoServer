package MesonTerminalEchoServer

import (
	"bufio"
	"context"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	echoTool "github.com/universe-30/EchoMiddleware/tool"
)

type HttpServer struct {
	*echo.Echo
	PauseMoment int64
}

func New() (hs *HttpServer) {
	hs = &HttpServer{echo.New(), 0}
	return hs
}

func (hs *HttpServer) UseJsoniter() {
	hs.JSONSerializer = echoTool.NewJsoniter()
}

func (hs *HttpServer) SetPauseSeconds(secs int64) {
	hs.PauseMoment = time.Now().Unix() + secs
}

func (hs *HttpServer) GetPauseMoment() int64 {
	return hs.PauseMoment
}

func FileWithPause(hs *HttpServer, c echo.Context, filePath string, header map[string][]string, ignoreHeaderMap map[string]struct{}) (err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return echo.NotFoundHandler(c)
	}
	defer f.Close()
	fi, _ := f.Stat()

	for headerKey, headerValue := range header {
		_, exist := ignoreHeaderMap[headerKey]
		if exist {
			continue
		}
		for _, v := range headerValue {
			c.Response().Header().Add(headerKey, v)
		}
	}
	ServeContent(hs, c.Response(), c.Request(), fi.Name(), fi.ModTime(), f)
	return
}

//func FileWithPause(hs *HttpServer, c echo.Context, filePath string, needSavedHeader bool, ignoreHeaderMap map[string]struct{}) (err error) {
//	f, err := os.Open(filePath)
//	if err != nil {
//		return echo.NotFoundHandler(c)
//	}
//	defer f.Close()
//
//	fi, _ := f.Stat()
//	if fi.IsDir() {
//		filePath = filepath.Join(filePath, "index.html")
//		f, err = os.Open(filePath)
//		if err != nil {
//			return echo.NotFoundHandler(c)
//		}
//		defer f.Close()
//		if fi, err = f.Stat(); err != nil {
//			return
//		}
//	}
//
//	if needSavedHeader {
//		AddHeader(c, filePath+".header", ignoreHeaderMap)
//	}
//
//	ServeContent(hs, c.Response(), c.Request(), fi.Name(), fi.ModTime(), f)
//	return
//}

func (hs *HttpServer) CloseServer() {
	hs.Close()
}

func (hs *HttpServer) WaitForServerStart(isTLS bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30000*time.Millisecond)
	defer cancel()

	time.Sleep(2 * time.Second)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			var addr net.Addr
			if isTLS {
				addr = hs.Echo.TLSListenerAddr()
			} else {
				addr = hs.Echo.ListenerAddr()
			}
			if addr != nil && strings.Contains(addr.String(), ":") {
				return nil // was started
			}
		}
	}
}

func AddHeader(c echo.Context, filePath string, ignoreHeaderMap map[string]struct{}) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	for {
		//key
		key, _, err := buf.ReadLine()
		if err != nil {
			if err == io.EOF { //read end
				return nil
			}
			return err
		}
		_, exist := ignoreHeaderMap[string(key)]
		if exist {
			continue
		}
		//value count
		countStr, _, err := buf.ReadLine()
		count, err := strconv.Atoi(string(countStr))
		for i := 0; i < count; i++ {
			value, _, _ := buf.ReadLine()
			c.Response().Header().Add(string(key), string(value))
		}
	}
}
