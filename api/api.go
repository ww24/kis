package api

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ww24/kis/storage"
)

var (
	st = storage.NewStorage(storage.NewFileSystemStore())
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
			"version": "0.1.0",
		})
	})

	// download image file
	api.router.GET("/:idext", func(ctx *gin.Context) {
		defer internalServerError(ctx)

		idext := ctx.Param("idext")
		ext := filepath.Ext(idext)
		id := strings.TrimSuffix(idext, ext)

		if ext == ".json" {
			width, height, err := st.ReadMetaData(id)
			if err != nil {
				if os.IsNotExist(err) {
					ctx.JSON(404, gin.H{
						"status": "ng",
						"error":  err.Error(),
					})
					return
				}

				panic(err)
			}
			ctx.JSON(200, gin.H{
				"status": "ok",
				"width":  width,
				"height": height,
			})
			return
		}

		buff, mimeType, err := st.Fetch(id, ext)
		if err != nil {
			if os.IsNotExist(err) {
				ctx.JSON(404, gin.H{
					"status": "ng",
					"error":  "file not found",
				})
				return
			}

			panic(err)
		}

		ctx.Data(200, mimeType, buff.Bytes())
	})

	// upload image file
	api.router.POST("/", func(ctx *gin.Context) {
		defer internalServerError(ctx)

		// receive image file
		file, _, err := ctx.Request.FormFile("image")
		if err != nil {
			ctx.JSON(400, gin.H{
				"status": "ng",
				"error":  "bad request",
			})
			return
		}
		defer file.Close()

		// save image into store
		id := st.GenerateID()
		err = st.Save(id, file)
		if err != nil {
			panic(err)
		}

		ctx.JSON(200, gin.H{
			"status": "ok",
			"id":     id,
		})
	})

	return
}
