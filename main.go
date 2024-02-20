package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sevlyar/go-daemon"
)

const defaultPort = ":80"
const vn_host = "https://vietnamnet.vn"

var port = defaultPort

func init() {
	err := godotenv.Load()
	if err != nil {
		return
	}

	port = os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	} else if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
}

func main() {
	cntxt := &daemon.Context{
		PidFileName: "vn_proxy.pid",
		PidFilePerm: 0644,
		LogFileName: "vn_proxy.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{"[go-daemon vn_proxy]"},
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	log.Print("- - - - - - - - - - - - - - -")
	log.Print("daemon started")

	ginMain()
}

func ginMain() {
	r := gin.Default()

	// r.GET("/test", func(ctx *gin.Context) {
	// 	const testResponse = `Proxy: "This is a test"`
	// 	ctx.String(http.StatusOK, testResponse)
	// })

	r.GET("/*uri", func(ctx *gin.Context) {
		uri := ctx.Param("uri")
		if strings.HasPrefix(uri, "/ori-raw/") {
			url := vn_host + uri[8:]
			oriString, err := getOrigin(url)
			if err != nil {
				ctx.String(http.StatusInternalServerError, "get url error")
				return
			}
			ctx.Data(http.StatusOK, "text/html; chartset=utf-8", oriString)
		}

		// ctx.String(http.StatusOK, "Uri: %s", uri)
		oriString, err := getOrigin(vn_host + uri)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "get url error")
			return
		}

		replacer := strings.NewReplacer(
			"translate=\"no\"", "",
		)
		ctx.Data(http.StatusOK, "text/html; chartset=utf-8", []byte(replacer.Replace(string(oriString))))
	})

	r.Run(port)
}

func getOrigin(urlStr string) (oriString []byte, err error) {
	var resp *http.Response
	log.Println("url:", urlStr)
	resp, err = http.Get(urlStr)
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}

	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
