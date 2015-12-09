package storage

import (
	"bytes"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/chai2010/webp"
	"github.com/satori/go.uuid"
)

// Storage structure
type Storage struct {
	store store
}

type store interface {
	Save(string, image.Image) error
	Fetch(string) (image.Image, error)
	Exists(string) bool
	Remove(string) error
}

// NewStorage constructor
func NewStorage(st store) (storage *Storage) {
	storage = &Storage{
		store: st,
	}
	return storage
}

// GenerateID will generate unique file ID
func (storage *Storage) GenerateID() (id string) {
	idv4 := uuid.NewV4()
	id = idv4.String()
	if storage.Exists(id) {
		id = storage.GenerateID()
	}
	return
}

// Save method
func (storage *Storage) Save(id string, reader io.Reader) (err error) {
	var data []byte
	data, err = ioutil.ReadAll(reader)
	if err != nil {
		return
	}

	var img image.Image
	mimeType := http.DetectContentType(data)
	switch mimeType {
	case "image/gif":
		img, err = gif.Decode(bytes.NewReader(data))
	case "image/png":
		img, err = png.Decode(bytes.NewReader(data))
	case "image/jpeg":
		img, err = jpeg.Decode(bytes.NewReader(data))
	case "image/webp":
		img, err = webp.Decode(bytes.NewReader(data))
	default:
		err = errors.New("unsupported MIME type")
	}
	if err != nil {
		return
	}

	err = storage.store.Save(id, img)
	return
}

// ReadMetaData method
func (storage *Storage) ReadMetaData(id string) (width, height int, err error) {
	var img image.Image
	img, err = storage.store.Fetch(id)
	if err != nil {
		return
	}

	rect := img.Bounds()
	width = rect.Max.X - rect.Min.X
	height = rect.Max.Y - rect.Min.Y
	return
}

// Fetch method
func (storage *Storage) Fetch(id string, extension string) (buff bytes.Buffer, mimeType string, err error) {
	var img image.Image
	img, err = storage.store.Fetch(id)
	if err != nil {
		return
	}

	switch extension {
	case ".gif":
		err = gif.Encode(&buff, img, &gif.Options{
			NumColors: 256,
		})
		mimeType = "image/gif"
	case "":
		fallthrough
	case ".png":
		err = png.Encode(&buff, img)
		mimeType = "image/png"
	case ".jpg":
		err = jpeg.Encode(&buff, img, &jpeg.Options{
			Quality: 100,
		})
		mimeType = "image/jpeg"
	case ".webp":
		err = webp.Encode(&buff, img, &webp.Options{
			Lossless: true,
			Quality:  100,
		})
		mimeType = "image/webp"
	default:
		err = errors.New("unsupported file extension")
	}
	if err != nil {
		return
	}

	return
}

// Exists method
func (storage *Storage) Exists(id string) bool {
	return storage.store.Exists(id)
}

// Remove method
func (storage *Storage) Remove(id string) error {
	return storage.store.Remove(id)
}
