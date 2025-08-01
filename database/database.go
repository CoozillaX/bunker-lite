package database

import (
	"sync"

	"go.etcd.io/bbolt"
)

const DatabseFileName = "eulogist-bunker-lite.db"

const (
	DATABASE_KEY_EULOGIST_USER = "EULOGIST_USER" // map[UserUniqueID]define.EulogistUser
	DATABASE_KEY_AUTH_HELPER   = "AUTH_HELPER"   // map[HelperUniqueID]define.AuthServerHelper
)

const (
	DATABASE_KEY_NTEU_MAPPING = "NAME_TO_EULOGIST_USER"  // map[EulogistUserName]UserUniqueID
	DATABSE_KEY_TTEU_MAPPING  = "TOKEN_TO_EULOGIST_USER" // map[EulogistToken]UserUniqueID
	DATABASE_KEY_TTAH_MAPPING = "TOEKN_TO_AUTH_HELPER"   // map[AuthServerHelperToken]HelperUniqueID
)

const (
	DATABASE_KEY_ALLOW_LIST_CONFIG = "ALLOW_LIST_CONFIG" // map[RentalServerNumber][]AllowListConfig
	DATABASE_KEY_GAME_SAVES_KEYS   = "GAME_SAVES_KEYS"   // map[UserUniqueID + RentalServerNumver]GameSavesKey
)

var (
	mu             *sync.RWMutex
	serverDatabase *bbolt.DB
)

func init() {
	var err error
	options := bbolt.Options{
		FreelistType: bbolt.FreelistMapType,
	}

	serverDatabase, err = bbolt.Open(DatabseFileName, 0600, &options)
	if err != nil {
		panic(err)
	}

	err = serverDatabase.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(DATABASE_KEY_EULOGIST_USER)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(DATABASE_KEY_AUTH_HELPER)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(DATABASE_KEY_NTEU_MAPPING)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(DATABSE_KEY_TTEU_MAPPING)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(DATABASE_KEY_TTAH_MAPPING)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(DATABASE_KEY_ALLOW_LIST_CONFIG)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(DATABASE_KEY_GAME_SAVES_KEYS)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	mu = new(sync.RWMutex)
}
