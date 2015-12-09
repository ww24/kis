package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/ww24/kis/api"
)

var (
	port int
	mode string
)

func init() {
	flag.IntVar(&port, "port", 3000, "Set port.")
	modes := []string{gin.DebugMode, gin.ReleaseMode, gin.TestMode}
	flag.StringVar(&mode, "mode", gin.DebugMode, "Set Gin Web Framework mode. ["+strings.Join(modes, " or ")+"]")
	flag.Parse()
}

func main() {
	gin.SetMode(mode)
	listener, errch := Serve(":"+strconv.Itoa(port), func() http.Handler {
		router := gin.Default()

		router.GET("/", func(ctx *gin.Context) {
			ctx.String(200, "KIS server works.\n")
		})

		api.NewAPI(router.Group("/api"))

		return router
	}())

	go func() {
		sig := make(chan os.Signal)
		signal.Notify(sig, syscall.SIGINT)
		for {
			select {
			case s := <-sig:
				log.Println(s)
				listener.Close()
			}
		}
	}()

	log.Println("KIS server started at", listener.Addr())
	log.Println("error:", <-errch)
}

// Serve is server bootstrap
func Serve(addr string, router http.Handler) (listener net.Listener, errch <-chan error) {
	ch := make(chan error)
	errch = ch

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		ch <- err
	}

	go func() {
		ch <- http.Serve(listener, router)
	}()

	return
}
