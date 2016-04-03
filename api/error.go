package api

import (
	"log"
	"runtime/debug"

	echo "gopkg.in/labstack/echo.v1"
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
			ctx.JSON(404, JSON{
				"status": "ng",
				"error":  "file not found",
			})
		case 403:
			ctx.JSON(403, JSON{
				"status": "ng",
				"error":  "Forbidden",
			})
		}
		return
	}

	if err, ok := cause.(error); ok {
		ctx.JSON(500, JSON{
			"status": "ng",
			"error":  err.Error(),
		})
	} else {
		ctx.JSON(500, JSON{
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
