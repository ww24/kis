package api

import (
	"log"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/labstack/echo"
)

// ErrorMiddleware is recover error middleware
func ErrorMiddleware() echo.MiddlewareFunc {
	return func(handler echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx *echo.Context) (err error) {
			defer internalServerError(ctx)
			return handler(ctx)
		}
	}
}

func internalServerError(ctx *echo.Context) {
	cause := recover()
	if cause == nil {
		return
	}

	if code, ok := cause.(int); ok {
		switch code {
		case 404:
			ctx.JSON(404, gin.H{
				"status": "ng",
				"error":  "file not found",
			})
		}
		return
	}

	if err, ok := cause.(error); ok {
		ctx.JSON(500, gin.H{
			"status": "ng",
			"error":  err.Error(),
		})
	} else {
		ctx.JSON(500, gin.H{
			"status": "ng",
			"error":  "unknown error",
		})
	}

	// debug log
	if ctx.Echo().Debug() {
		log.Println("Error:", cause)
		debug.PrintStack()
	}
}
