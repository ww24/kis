package api

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ww24/kis/storage"
)

var (
	// use LevelDB instead of FileSystem
	store = storage.NewStorage(storage.LevelDB)
)

// API structure
type API struct {
	router *gin.RouterGroup
}

// NewAPI constructor
func NewAPI(router *gin.RouterGroup) (api *API) {
	api = &API{
		router: router,
	}

	// API end point
	api.router.GET("/", func(ctx *gin.Context) {
		defer internalServerError(ctx)

		ctx.JSON(200, gin.H{
			"status":  "ok",
			"version": "0.3.0",
		})
	})

	// download image file
	api.router.GET("/:idext", func(ctx *gin.Context) {
		defer internalServerError(ctx)

		idext := ctx.Param("idext")
		ext := filepath.Ext(idext)
		id := strings.TrimSuffix(idext, ext)

		if ext == ".json" {
			width, height, err := store.ReadMetaData(id)
			if err != nil {
				panic(err)
			}
			if width == 0 && height == 0 {
				panic(404)
			}
			ctx.JSON(200, gin.H{
				"status": "ok",
				"width":  width,
				"height": height,
			})
			return
		}

		buff, mimeType, err := store.Fetch(id, ext)
		if err == storage.ErrUnsupportedFileExtension {
			ctx.JSON(400, gin.H{
				"status": "ng",
				"error":  err.Error(),
			})
			return
		}
		if err != nil {
			panic(err)
		}
		if buff.Len() == 0 {
			panic(404)
		}

		ctx.Data(200, mimeType, buff.Bytes())
	})

	// upload image file
	api.router.POST("/", func(ctx *gin.Context) {
		defer internalServerError(ctx)

		contentType := strings.Split(ctx.Request.Header.Get("Content-Type"), ";")[0]
		fmt.Println(contentType)

		id, err := store.GenerateID()
		if err != nil {
			panic(err)
		}

		// save image into store
		switch contentType {
		case "multipart/form-data":
			// receive image file
			file, _, err := ctx.Request.FormFile("image")
			if err != nil {
				ctx.JSON(400, gin.H{
					"status": "ng",
					"error":  `require "image" key for multipart/form-data`,
				})
				return
			}
			defer file.Close()
			err = store.Save(id, file)
			if err != nil {
				panic(err)
			}
		default:
			err := store.Save(id, ctx.Request.Body)
			if err == storage.ErrUnsupportedMIMEType {
				ctx.JSON(400, gin.H{
					"status": "ng",
					"error":  err.Error(),
				})
				return
			}
			if err != nil {
				panic(err)
			}
		}

		ctx.JSON(200, gin.H{
			"status": "ok",
			"id":     id,
		})
	})

	return
}
