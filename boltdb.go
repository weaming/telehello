package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
	"strings"
	"time"
)

type BoltConnection struct {
	db *bolt.DB
}

func NewDB(path string) *BoltConnection {
	if ok, _ := ExistsFile(path); !ok {
		fd, err := os.Create(path)
		fatalErr(err)
		fd.Close()
		fmt.Println("created database file", path)
	}
	db, err := bolt.Open(path, 0666, &bolt.Options{Timeout: 2 * time.Second})
	fatalErr(err)
	return &BoltConnection{db: db}
}

func (p *BoltConnection) CreateBucketIfNotExists(bucket string) error {
	return p.db.Batch(func(tx *bolt.Tx) error {
		// create bucket
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	})
}

func (p *BoltConnection) Set(bucket, key, value string) error {
	return p.db.Batch(func(tx *bolt.Tx) error {
		// create bucket
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}

		// set key
		if err := b.Put([]byte(key), []byte(value)); err != nil {
			return err
		}
		return nil
	})
}

func (p *BoltConnection) Get(bucket, key string) ([]byte, error) {
	var value []byte
	err := p.db.View(func(tx *bolt.Tx) error {
		value = tx.Bucket([]byte(bucket)).Get([]byte(key))
		return nil
	})
	if err != nil {
		return value, err
	}
	return value, nil
}

func (p *BoltConnection) Close() error {
	return p.db.Close()
}

func (p *BoltConnection) Clear() error {
	fp := p.db.Path()
	err := p.Close()
	if err == nil {
		return os.Remove(fp)
		log.Println("deleted database", fp)
	}
	return err
}

func (db *BoltConnection) GetFieldsInDB(bucket, key string) ([]string, error) {
	db.CreateBucketIfNotExists(bucket)
	old, err := db.Get(bucket, key)
	if err != nil {
		return []string{}, err
	}
	return strings.Fields(string(old)), nil
}

func (db *BoltConnection) AddFieldInDB(bucket, key, value string) ([]string, error) {
	fields, err1 := db.GetFieldsInDB(bucket, key)
	if err1 != nil {
		return fields, err1
	}

	tmp := append(fields, value)
	var newFields []string

outer:
	for x, a := range tmp {
		for y, b := range newFields {
			if y != x && b == a {
				continue outer
			}
		}
		newFields = append(newFields, a)
	}

	err2 := db.Set(bucket, key, strings.Join(newFields, " "))
	if err2 != nil {
		return fields, err2
	}
	return newFields, nil
}
