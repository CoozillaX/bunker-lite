package database

import (
	"bunker-lite/define"
	"bytes"
	"fmt"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"go.etcd.io/bbolt"
)

// GetOrCreateGameSavesKey ..
func GetOrCreateGameSavesKey(eulogistUniqueID string, rentalServerNumber string, useLock bool) (aesCipher []byte, err error) {
	if useLock {
		mu.Lock()
		defer mu.Unlock()
	}

	key := define.GameSavesKey{
		EulogistUserUniqueID: eulogistUniqueID,
		RentalServerNumber:   rentalServerNumber,
	}

	keyBuf := bytes.NewBuffer(nil)
	writer := protocol.NewWriter(keyBuf, 0)
	key.MarshalKey(writer)

	err = serverDatabase.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(DATABASE_KEY_GAME_SAVES_KEYS))
		payload := bucket.Get(keyBuf.Bytes())

		if len(payload) == 0 {
			uniqueID := uuid.New()
			key.GameSavesAESCipher = uniqueID[:]

			dataBuf := bytes.NewBuffer(nil)
			writer = protocol.NewWriter(dataBuf, 0)
			key.MarshalData(writer)

			return bucket.Put(keyBuf.Bytes(), dataBuf.Bytes())
		}

		dataBuf := bytes.NewBuffer(payload)
		reader := protocol.NewReader(dataBuf, 0, false)
		key.MarshalData(reader)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("GetOrCreateGameSavesKey: 获取存档 AES 密钥时出现问题，原因是 %v", err)
	}

	return key.GameSavesAESCipher, nil
}
