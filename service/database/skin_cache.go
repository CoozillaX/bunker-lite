package database

import (
	"fmt"

	"go.etcd.io/bbolt"
)

// SetUserSkinCache ..
func SetUserSkinCache(eulogistUniqueID string, skinDownloadURL string, useLock bool) error {
	if useLock {
		mu.Lock()
		defer mu.Unlock()
	}

	err := serverDatabase.Update(func(tx *bbolt.Tx) error {
		return tx.
			Bucket([]byte(DATABASE_KEY_SKIN_CACHE)).
			Put([]byte(eulogistUniqueID), []byte(skinDownloadURL))
	})
	if err != nil {
		return fmt.Errorf("SetUserSkinCache: %v", err)
	}

	return nil
}

// GetAndDeleteUserSkinCache ..
func GetAndDeleteUserSkinCache(
	eulogistUniqueID string,
	useLock bool,
) (haveData bool, skinDownloadURL string, err error) {
	if useLock {
		mu.Lock()
		defer mu.Unlock()
	}

	err = serverDatabase.Update(func(tx *bbolt.Tx) error {
		payload := tx.Bucket([]byte(DATABASE_KEY_SKIN_CACHE)).Get([]byte(eulogistUniqueID))
		if len(payload) == 0 {
			return nil
		}

		haveData = true
		skinDownloadURL = string(payload)

		return tx.
			Bucket([]byte(DATABASE_KEY_SKIN_CACHE)).
			Delete([]byte(eulogistUniqueID))
	})
	if err != nil {
		return false, "", fmt.Errorf("SetUserSkinCache: %v", err)
	}

	return haveData, skinDownloadURL, nil
}
