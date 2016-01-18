package storage

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/ugorji/go/codec"
)

// LevelDBStore structure (implement store interface)
type LevelDBStore struct {
	dir    string
	prefix string `default:"data:"`
	msgh   *codec.MsgpackHandle
}

// NewLevelDBStore constructor
func NewLevelDBStore() (ld *LevelDBStore) {
	ld = &LevelDBStore{
		dir:  "store/leveldb",
		msgh: &codec.MsgpackHandle{RawToString: true},
	}
	return
}

// open leveldb store
func (ld *LevelDBStore) open(options *opt.Options) (db *leveldb.DB, err error) {
	db, err = leveldb.OpenFile(ld.dir, options)
	return
}

// Save method
func (ld *LevelDBStore) Save(id string, item *Item) (err error) {
	var db *leveldb.DB
	db, err = ld.open(nil)
	if err != nil {
		return
	}
	defer db.Close()

	// encode to MessagePack and save into LevelDB
	buff := &bytes.Buffer{}
	err = codec.NewEncoder(buff, ld.msgh).Encode(item)
	if err != nil {
		return
	}
	err = db.Put([]byte(ld.prefix+id), buff.Bytes(), nil)
	return
}

// Keys method
func (ld *LevelDBStore) Keys(prefixes ...string) (list []string, err error) {
	prefix := ""
	if len(prefixes) > 0 {
		prefix = strings.Join(prefixes, ":")
	}

	var db *leveldb.DB
	db, err = ld.open(nil)
	if err != nil {
		return
	}
	defer db.Close()

	list = make([]string, 0, 100)

	iter := db.NewIterator(util.BytesPrefix([]byte(ld.prefix+prefix)), nil)
	defer iter.Release()
	for iter.Next() {
		key := iter.Key()
		list = append(list, string(key))
	}
	err = iter.Error()
	if err != nil {
		return
	}
	return
}

// Fetch method
func (ld *LevelDBStore) Fetch(id string) (item *Item, err error) {
	var db *leveldb.DB
	db, err = ld.open(nil)
	if err != nil {
		return
	}
	defer db.Close()

	var data []byte
	data, err = db.Get([]byte(ld.prefix+id), nil)
	if err == leveldb.ErrNotFound {
		err = nil
		return
	}
	if err != nil {
		return
	}

	// read data from LevelDB and decode MessagePack
	item = &Item{}
	err = codec.NewDecoderBytes(data, ld.msgh).Decode(item)

	// Convert metadata MessagePack -> JSON
	err = json.Unmarshal(item.MsgpData, &item.JSONData)
	if err != nil {
		return
	}

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

	exists, err = db.Has([]byte(ld.prefix+id), nil)
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

	err = db.Delete([]byte(ld.prefix+id), nil)
	return
}
