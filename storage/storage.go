package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/chai2010/webp"
	"github.com/satori/go.uuid"
)

// Storage structure
type Storage struct {
	store store
}

type store interface {
	Save(string, *Item) error
	Keys(prefixes ...string) ([]string, error)
	Fetch(string) (*Item, error)
	Exists(string) (bool, error)
	Remove(string) error
}

// Item structure for storage
type Item struct {
	Width     int       `codec:"width" json:"width"`
	Height    int       `codec:"height" json:"height"`
	CreatedAt time.Time `codec:"created_at" json:"created_at"`
	UpdatedAt time.Time `codec:"updated_at" json:"updated_at"`
	// 登録者情報
	IP string `codec:"ip" json:"-"`
	UA string `codec:"ua" json:"-"`
	// 画像データ
	Webp []byte `codec:"file" json:"-"`
	// 汎用メタデータ (JSON) for MessagePack
	MsgpData []byte `codec:"data" json:"-"`
	// 汎用メタデータ (JSON) for API
	JSONData map[string]interface{} `codec:"-" json:"data"`
}

const (
	// FileSystem constant value
	FileSystem = 0
	// LevelDB constant value
	LevelDB = 1
)

var (
	// ErrUnsupportedMIMEType error value
	ErrUnsupportedMIMEType = errors.New("unsupported MIME type")
	// ErrUnsupportedFileExtension error value
	ErrUnsupportedFileExtension = errors.New("unsupported file extension")
	// ErrInvalidJSON error value
	ErrInvalidJSON = errors.New("invalid json")
)

// NewStorage constructor
func NewStorage(storeType int) (storage *Storage) {
	var storeImplementation store
	switch storeType {
	case FileSystem:
		panic(errors.New("Use LevelDBStore instead of FileSystemStore."))
		// storeImplementation = NewFileSystemStore()
	case LevelDB:
		storeImplementation = NewLevelDBStore()
	default:
		panic(errors.New("invalid store type"))
	}

	storage = &Storage{
		store: storeImplementation,
	}
	return storage
}

// GenerateID will generate unique file ID
func (storage *Storage) GenerateID() (id string, err error) {
	// (生成順が辞書順に並ぶ為 UUIDv1 を採用する)
	idv1 := uuid.NewV1()
	id = idv1.String()

	exists, err := storage.Exists(id)
	if err != nil {
		panic(err)
	}
	if exists {
		id, err = storage.GenerateID()
	}
	return
}

// Save method
func (storage *Storage) Save(id string, reader io.Reader, metadata ...map[string]interface{}) (err error) {
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
		err = ErrUnsupportedMIMEType
	}
	if err != nil {
		return
	}

	rect := img.Bounds()
	item := &Item{
		Width:     rect.Max.X - rect.Min.X,
		Height:    rect.Max.Y - rect.Min.Y,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// check metadata existance and encode to JSON
	if len(metadata) > 0 && metadata[0] != nil {
		item.MsgpData, err = json.Marshal(metadata[0])
		if err != nil {
			return
		}
	}

	buff := &bytes.Buffer{}
	err = webp.Encode(buff, img, &webp.Options{
		Lossless: true,
		Quality:  100,
	})
	if err != nil {
		return
	}
	item.Webp = buff.Bytes()

	err = storage.store.Save(id, item)
	return
}

// ReadMetaData method
func (storage *Storage) ReadMetaData(id string) (item *Item, err error) {
	item, err = storage.store.Fetch(id)
	return
}

// Keys method
func (storage *Storage) Keys(prefixes ...string) (list []string, err error) {
	list, err = storage.store.Keys(prefixes...)
	return
}

// Fetch method
func (storage *Storage) Fetch(id string, extension string) (buff bytes.Buffer, mimeType string, err error) {
	var item *Item
	item, err = storage.store.Fetch(id)
	if err != nil || item == nil {
		return
	}

	var img image.Image
	img, err = webp.Decode(bytes.NewBuffer(item.Webp))
	if err != nil {
		return
	}

	buff = bytes.Buffer{}
	switch extension {
	case ".gif":
		// 	lossy compression
		err = gif.Encode(&buff, img, &gif.Options{
			NumColors: 256,
		})
		mimeType = "image/gif"
	case ".png":
		// lossless compression
		err = png.Encode(&buff, img)
		mimeType = "image/png"
	case "":
		fallthrough
	case ".jpg":
		// 	lossy compression
		err = jpeg.Encode(&buff, img, &jpeg.Options{
			Quality: 99,
		})
		mimeType = "image/jpeg"
	case ".webp":
		// lossless compression
		buff = *bytes.NewBuffer(item.Webp)
		mimeType = "image/webp"
	default:
		err = ErrUnsupportedFileExtension
	}
	if err != nil {
		return
	}

	return
}

// Exists method
func (storage *Storage) Exists(id string) (bool, error) {
	return storage.store.Exists(id)
}

// Remove method
func (storage *Storage) Remove(id string) error {
	return storage.store.Remove(id)
}
