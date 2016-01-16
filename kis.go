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

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/ww24/kis/api"
)

var (
	port int
	mode string
)

func init() {
	flag.IntVar(&port, "port", 3000, "Set port.")
	modes := []string{"debug", "release"}
	flag.StringVar(&mode, "mode", "debug", "Set debug mode for Web Framework. ["+strings.Join(modes, " or ")+"]")
	flag.Parse()
}

func main() {
	listener, errch := Serve(":"+strconv.Itoa(port), func() http.Handler {
		router := echo.New()
		router.SetDebug(mode == "debug")

		router.Use(middleware.Logger())
		router.Use(api.ErrorMiddleware())

		router.Get("/", func(ctx *echo.Context) (err error) {
			err = ctx.String(200, "KIS server works.\n")
			return
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
