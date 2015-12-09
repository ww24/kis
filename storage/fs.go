package storage

import (
	"image"
	"os"
	"path/filepath"
	"sync"

	"github.com/chai2010/webp"
)

var (
	mutex = new(sync.Mutex)
)

// FileSystemStore structure
type FileSystemStore struct {
	dir string
}

// NewFileSystemStore constructor
func NewFileSystemStore() (fs *FileSystemStore) {
	fs = &FileSystemStore{
		dir: "store",
	}
	return
}

// Save method
func (fs *FileSystemStore) Save(id string, img image.Image) (err error) {
	mutex.Lock()
	defer mutex.Unlock()

	var file *os.File
	pathstr := filepath.Join(fs.dir, id+".webp")
	file, err = os.Create(pathstr)
	if err != nil {
		return
	}
	defer file.Close()
	// encode and save image to file system
	err = webp.Encode(file, img, &webp.Options{
		Lossless: true,
		Quality:  100,
	})
	return
}

// Fetch method
func (fs *FileSystemStore) Fetch(id string) (img image.Image, err error) {
	var file *os.File
	pathstr := filepath.Join(fs.dir, id+".webp")
	file, err = os.Open(pathstr)
	if err != nil {
		return
	}
	defer file.Close()
	// decode and read image from file system
	img, err = webp.Decode(file)
	return
}

// Exists method
func (fs *FileSystemStore) Exists(id string) (exists bool) {
	mutex.Lock()
	defer mutex.Unlock()

	pathstr := filepath.Join(fs.dir, id+".webp")
	_, err := os.Stat(pathstr)
	exists = err == nil
	return
}

// Remove method
func (fs *FileSystemStore) Remove(id string) (err error) {
	mutex.Lock()
	defer mutex.Unlock()

	pathstr := filepath.Join(fs.dir, id+".webp")
	err = os.Remove(pathstr)
	return
}
