package repositories

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/ktnyt/labcon/cmd/labcon/app/models"
	"github.com/ktnyt/labcon/cmd/labcon/lib"
	"github.com/vmihailenco/msgpack"
)

type DriverRepositoryImpl struct {
	db *badger.DB
}

func NewDriverRepository(db *badger.DB) DriverRepository {
	return DriverRepositoryImpl{
		db: db,
	}
}

func (repo DriverRepositoryImpl) List() ([]string, error) {
	names := []string{}
	err := repo.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte("driver/")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			name := string(bytes.TrimPrefix(key, prefix))
			names = append(names, name)
		}
		return nil
	})
	return names, err
}

func (repo DriverRepositoryImpl) Key(name string) []byte {
	return []byte(fmt.Sprintf("driver/%s", name))
}

func (repo DriverRepositoryImpl) Create(name string, token string, state interface{}) error {
	return repo.db.Update(func(txn *badger.Txn) error {
		key := repo.Key(name)
		_, err := txn.Get(key)
		if !errors.Is(err, badger.ErrKeyNotFound) {
			if err == nil {
				return lib.ErrAlreadyExists
			}
			return err
		}
		driver := models.NewDriver(name, token, state)
		val, err := msgpack.Marshal(driver)
		if err != nil {
			return err
		}
		return txn.Set(key, val)
	})
}

func (repo DriverRepositoryImpl) Fetch(name string) (models.DriverModel, error) {
	driver := models.DriverModel{Name: name}
	err := repo.db.View(func(txn *badger.Txn) error {
		key := repo.Key(name)
		item, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return lib.ErrNotFound
			}
			return err
		}
		return item.Value(func(val []byte) error {
			return msgpack.Unmarshal(val, &driver)
		})
	})
	return driver, err
}

func (repo DriverRepositoryImpl) Update(driver models.DriverModel) error {
	return repo.db.Update(func(txn *badger.Txn) error {
		val, err := msgpack.Marshal(driver)
		if err != nil {
			return err
		}
		return txn.Set(repo.Key(driver.Name), val)
	})
}

func (repo DriverRepositoryImpl) Delete(name string) error {
	return repo.db.Update(func(txn *badger.Txn) error {
		key := repo.Key(name)
		if _, err := txn.Get(key); err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return lib.ErrNotFound
			}
			return err
		}
		err := txn.Delete(key)
		if errors.Is(err, badger.ErrKeyNotFound) {
			return lib.ErrNotFound
		}
		return err
	})
}
