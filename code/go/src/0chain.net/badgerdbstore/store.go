package badgerdbstore

import (
	"bytes"
	"context"
	"encoding/json"

	"0chain.net/datastore"
	"github.com/dgraph-io/badger"
)

var storageAPI *Store

func SetupStorageProvider() {
	storageAPI = &Store{}
	opts := badger.DefaultOptions
	opts.Dir = "data/badgerdb/blobberstate"
	opts.ValueDir = "data/badgerdb/blobberstate"
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	storageAPI.DB = db
}

/*GetStorageProvider - get the storage provider for the memorystore */
func GetStorageProvider() datastore.Store {
	return storageAPI
}

/*Store - just a struct to implement the datastore.Store interface */
type Store struct {
	DB *badger.DB
}

/*Read - read an entity from the store */
func (ps *Store) Read(ctx context.Context, key datastore.Key, entity datastore.Entity) error {
	txn := ps.GetConnection(ctx)
	item, err := txn.Get([]byte(key))
	if err != nil && err == badger.ErrKeyNotFound {
		return datastore.ErrKeyNotFound
	}
	if err != nil {
		return err
	}
	valCopy, err := item.ValueCopy(nil)
	if err != nil {
		return err
	}
	err = json.NewDecoder(bytes.NewReader(valCopy)).Decode(entity)
	if err != nil {
		return err
	}

	return nil
}

/*ReadBytes - reads a key from the store */
func (ps *Store) ReadBytes(ctx context.Context, key datastore.Key) ([]byte, error) {
	resultBytes := make([]byte, 0)
	txn := ps.GetConnection(ctx)
	item, err := txn.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	resultBytes, err = item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}

	return resultBytes, err
}

/*Write - write an entity to the store */
func (ps *Store) Write(ctx context.Context, entity datastore.Entity) error {
	// Start a writable transaction.
	txn := ps.GetConnection(ctx)

	// Use the transaction...
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(entity)
	if err != nil {
		return err
	}

	err = txn.Set([]byte(entity.GetKey()), b.Bytes())
	if err != nil {
		return err
	}

	return nil
}

/*WriteBytes - write bytes to the store */
func (ps *Store) WriteBytes(ctx context.Context, key datastore.Key, value []byte) error {
	// Start a writable transaction.
	txn := ps.GetConnection(ctx)

	err := txn.Set([]byte(key), value)
	if err != nil {
		return err
	}

	return nil
}

/*Delete - Delete an entity from the store */
func (ps *Store) Delete(ctx context.Context, entity datastore.Entity) error {
	// Start a writable transaction.
	txn := ps.GetConnection(ctx)

	err := txn.Delete([]byte(entity.GetKey()))
	if err != nil {
		return err
	}

	return nil
}

/*MultiRead - read multiple entities from the store */
func (ps *Store) MultiRead(ctx context.Context, entityMetadata datastore.EntityMetadata, keys []datastore.Key, entities []datastore.Entity) error {
	return nil
}

/*MultiWrite - Write multiple entities to the store */
func (ps *Store) MultiWrite(ctx context.Context, entityMetadata datastore.EntityMetadata, entities []datastore.Entity) error {
	return nil
}

/*MultiDelete - delete multiple entities from the store */
func (ps *Store) MultiDelete(ctx context.Context, entityMetadata datastore.EntityMetadata, entities []datastore.Entity) error {
	// TODO
	return nil
}

func (ps *Store) Iterate(ctx context.Context, handler datastore.StoreIteratorHandler) error {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchSize = 10
	txn := ps.GetConnection(ctx)
	it := txn.NewIterator(opts)
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
		item := it.Item()
		k := item.Key()
		valueBytes, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		err = handler(ctx, string(k), valueBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ps *Store) IteratePrefix(ctx context.Context, prefix string, handler datastore.StoreIteratorHandler) error {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchSize = 10
	txn := ps.GetConnection(ctx)
	it := txn.NewIterator(opts)
	defer it.Close()
	prefixI := []byte(prefix)
	for it.Seek(prefixI); it.ValidForPrefix(prefixI); it.Next() {
		item := it.Item()
		k := item.Key()
		valueBytes, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		err = handler(ctx, string(k), valueBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

/*WithConnection takes a context and adds a connection value to it */
func (ps *Store) WithConnection(ctx context.Context) context.Context {
	return context.WithValue(ctx, datastore.CONNECTION_CONTEXT_KEY, ps.GetCon())
}

func (ps *Store) GetCon() *badger.Txn {
	return ps.DB.NewTransaction(true)
}

func (ps *Store) GetConnection(ctx context.Context) *badger.Txn {
	conn := ctx.Value(datastore.CONNECTION_CONTEXT_KEY)
	if conn != nil {
		return conn.(*badger.Txn)
	}
	return ps.GetCon()
}

func (ps *Store) Commit(ctx context.Context) error {
	return ps.GetConnection(ctx).Commit()
}

func (ps *Store) Discard(ctx context.Context) {
	ps.GetConnection(ctx).Discard()
}
