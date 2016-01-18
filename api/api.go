package api

import (
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/labstack/echo"
	"github.com/ww24/kis/storage"
)

var (
	// use LevelDB instead of FileSystem
	store = storage.NewStorage(storage.LevelDB)
)

// JSON type
type JSON map[string]interface{}

// API structure
type API struct {
	router *echo.Group
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
			"version": "1.0.0",
		})

		return
	})

	api.router.Get("/list", func(ctx *echo.Context) (err error) {
		if isAdmin(ctx) == false {
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

	// authKey is storage key prefix
	api.router.Get("/list/:authKey", func(ctx *echo.Context) (err error) {
		if isAdmin(ctx) == false {
			panic(403)
		}

		var list []string

		authKey := ctx.Param("authKey")
		if authKey != "" {
			list, err = store.Keys(authKey + ":")
		} else {
			list, err = store.Keys()
		}
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
			var item *storage.Item
			item, err = store.ReadMetaData(id)
			if err != nil {
				panic(err)
			}
			err = ctx.JSON(200, JSON{
				"status": "ok",
				"item":   item,
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

	// upload image file (and metadata JSON)
	api.router.Post("/", func(ctx *echo.Context) (err error) {
		contentType := strings.Split(ctx.Request().Header.Get("Content-Type"), ";")[0]

		id, err := store.GenerateID()
		if err != nil {
			panic(err)
		}

		request := ctx.Request()

		authKey := request.Header.Get("Authorization")
		if authKey != "" {
			id = authKey + ":" + id
		}

		// save image into store
		switch contentType {
		case "multipart/form-data":
			// receive image file
			var file multipart.File
			file, _, err = request.FormFile("image")
			if err != nil {
				ctx.JSON(400, JSON{
					"status": "ng",
					"error":  `require "image" key for multipart/form-data`,
				})
				return
			}
			defer file.Close()

			metadata := ctx.Form("metadata")
			err = store.Save(id, file, metadata)
			switch err {
			case storage.ErrUnsupportedMIMEType:
			case storage.ErrInvalidJSON:
				err = ctx.JSON(400, JSON{
					"status": "ng",
					"error":  err.Error(),
				})
				return
			}
			if err != nil {
				panic(err)
			}
		default:
			// metadata not supported
			err = store.Save(id, request.Body)
			switch err {
			case storage.ErrUnsupportedMIMEType:
			case storage.ErrInvalidJSON:
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
