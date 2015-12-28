package api

import (
	"log"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func internalServerError(ctx *gin.Context) {
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
	if os.Getenv("GIN_MODE") != "release" {
		log.Println("Error:", cause)
		debug.PrintStack()
	}
}
