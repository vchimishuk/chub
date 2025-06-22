package db

import "github.com/akrylysov/pogreb"

var db *pogreb.DB

func Open(file string) error {
	// TODO: pogreb.SetLogger()

	var err error
	db, err = pogreb.Open(file, nil)

	return err
}

func Close() error {
	return db.Close()
}

func Put(path string, val []byte) error {
	return db.Put([]byte(path), val)
}

func Get(path string) ([]byte, error) {
	return db.Get([]byte(path))
}
