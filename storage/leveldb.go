package storage

import (
	"bytes"
	"image"

	"github.com/chai2010/webp"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// LevelDBStore structure
type LevelDBStore struct {
	dir string
}

// NewLevelDBStore constructor
func NewLevelDBStore() (ld *LevelDBStore) {
	ld = &LevelDBStore{
		dir: "store/leveldb",
	}
	return
}

func (ld *LevelDBStore) open(options *opt.Options) (db *leveldb.DB, err error) {
	db, err = leveldb.OpenFile(ld.dir, options)
	return
}

// Save method
func (ld *LevelDBStore) Save(id string, img image.Image) (err error) {
	var db *leveldb.DB
	db, err = ld.open(nil)
	if err != nil {
		return
	}
	defer db.Close()
	// encode and save image to LevelDB

	buff := bytes.Buffer{}
	err = webp.Encode(&buff, img, &webp.Options{
		Lossless: true,
		Quality:  100,
	})

	err = db.Put([]byte(id), buff.Bytes(), nil)

	return
}

// Fetch method
func (ld *LevelDBStore) Fetch(id string) (img image.Image, err error) {
	var db *leveldb.DB
	db, err = ld.open(nil)
	if err != nil {
		return
	}
	defer db.Close()

	var data []byte
	data, err = db.Get([]byte(id), nil)
	if err == leveldb.ErrNotFound {
		err = nil
		return
	}
	if err != nil {
		return
	}

	// decode and read image from LevelDB
	buff := bytes.NewBuffer(data)
	img, err = webp.Decode(buff)
	return
}

// Exists method
func (ld *LevelDBStore) Exists(id string) (exists bool, err error) {
	var db *leveldb.DB
	db, err = ld.open(nil)
	if err != nil {
		return
	}
	defer db.Close()

	exists, err = db.Has([]byte(id), nil)
	return
}

// Remove method
func (ld *LevelDBStore) Remove(id string) (err error) {
	var db *leveldb.DB
	db, err = ld.open(nil)
	if err != nil {
		return
	}
	defer db.Close()

	err = db.Delete([]byte(id), nil)
	return
}
