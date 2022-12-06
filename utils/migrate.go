package utils

import (
	"github.com/axelarnetwork/axelar-core/utils/key"
)

func MigrateKeys[T ValidatedProtoMarshaler](store KVStore, originalKey key.Key, defaultValue T, migrateValue func(T) error) error {
	// migrate in batches so memory pressure doesn't become too large
	keysToDelete := make([][]byte, 0, 1000)
	for {
		iter := store.IteratorNew(originalKey)

		if !iter.Valid() {
			return iter.Close()
		}

		value := defaultValue
		for ; iter.Valid() && len(keysToDelete) < 1000; iter.Next() {
			iter.UnmarshalValue(value)
			if err := migrateValue(value); err != nil {
				return err
			}

			keysToDelete = append(keysToDelete, iter.Key())
		}

		if err := iter.Close(); err != nil {
			return err
		}

		for _, poolKey := range keysToDelete {
			store.DeleteRaw(poolKey)
		}

		keysToDelete = keysToDelete[:0]
	}
}
