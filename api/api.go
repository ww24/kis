package api

import (
	"encoding/hex"
	"mime/multipart"
	"path/filepath"
	"strings"

	"crypto/sha256"

	"github.com/labstack/echo"
	"github.com/ww24/kis/storage"
)

var (
	// use LevelDB instead of FileSystem
	store = storage.NewStorage(storage.LevelDB)

	hashedSecret string
)

// JSON type
type JSON map[string]interface{}

// API structure
type API struct {
	router *echo.Group
}

func init() {
	hashedSecret = authSecret("kis.json")["secret"].(string)
	data := sha256.Sum256([]byte(hashedSecret))
	hashedSecret = hex.EncodeToString(data[:])
}

// NewAPI constructor
func NewAPI(router *echo.Group) (api *API) {
	api = &API{
		router: router,
	}

	// API end point
	api.router.Get("/", func(ctx *echo.Context) (err error) {
		err = ctx.JSON(200, JSON{
			"status":  "ok",
			"version": "0.3.0",
		})

		return
	})

	api.router.Get("/list", func(ctx *echo.Context) (err error) {
		if hashCompare(ctx.Query("secret"), hashedSecret) == false {
			panic(403)
		}

		list, err := store.Keys()
		if err != nil {
			panic(err)
		}

		err = ctx.JSON(200, JSON{
			"status": "ok",
			"list":   list,
		})

		return
	})

	// download image file
	api.router.Get("/:idext", func(ctx *echo.Context) (err error) {
		idext := ctx.Param("idext")
		ext := filepath.Ext(idext)
		id := strings.TrimSuffix(idext, ext)

		if ext == ".json" {
			var width, height int
			width, height, err = store.ReadMetaData(id)
			if err != nil {
				panic(err)
			}
			if width == 0 && height == 0 {
				panic(404)
			}
			err = ctx.JSON(200, JSON{
				"status": "ok",
				"width":  width,
				"height": height,
			})

			return
		}

		buff, mimeType, err := store.Fetch(id, ext)
		if err == storage.ErrUnsupportedFileExtension {
			ctx.JSON(400, JSON{
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

		res := ctx.Response()
		res.Header().Set("Content-Type", mimeType)
		res.WriteHeader(200)
		_, err = res.Write(buff.Bytes())

		return
	})

	// upload image file
	api.router.Post("/", func(ctx *echo.Context) (err error) {
		contentType := strings.Split(ctx.Request().Header.Get("Content-Type"), ";")[0]

		id, err := store.GenerateID()
		if err != nil {
			panic(err)
		}

		// save image into store
		switch contentType {
		case "multipart/form-data":
			// receive image file
			var file multipart.File
			file, _, err = ctx.Request().FormFile("image")
			if err != nil {
				ctx.JSON(400, JSON{
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
			err = store.Save(id, ctx.Request().Body)
			if err == storage.ErrUnsupportedMIMEType {
				err = ctx.JSON(400, JSON{
					"status": "ng",
					"error":  err.Error(),
				})
				return
			}
			if err != nil {
				panic(err)
			}
		}

		err = ctx.JSON(200, JSON{
			"status": "ok",
			"id":     id,
		})

		return
	})

	return
}
