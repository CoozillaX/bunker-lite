package database

import (
	"bunker-lite/define"
	"bytes"
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"go.etcd.io/bbolt"
)

// GetAllowServerConfig ..
func GetAllowServerConfig(rentalServerNumber string, useLock bool) (result []define.AllowListConfig) {
	if useLock {
		mu.RLock()
		defer mu.RUnlock()
	}

	_ = serverDatabase.View(func(tx *bbolt.Tx) error {
		payload := tx.Bucket([]byte(DATABASE_KEY_ALLOW_LIST_CONFIG)).Get([]byte(rentalServerNumber))
		if len(payload) > 0 {
			buf := bytes.NewBuffer(payload)
			reader := protocol.NewReader(buf, 0, false)
			protocol.SliceUint8Length(reader, &result)
		}
		return nil
	})

	return
}

// SetAllowServerConfig ..
func SetAllowServerConfig(rentalServerNumber string, configs []define.AllowListConfig, useLock bool) error {
	if useLock {
		mu.Lock()
		defer mu.Unlock()
	}

	err := serverDatabase.Update(func(tx *bbolt.Tx) error {
		buf := bytes.NewBuffer(nil)
		writer := protocol.NewWriter(buf, 0)
		protocol.SliceUint8Length(writer, &configs)
		return tx.
			Bucket([]byte(DATABASE_KEY_ALLOW_LIST_CONFIG)).
			Put([]byte(rentalServerNumber), buf.Bytes())
	})
	if err != nil {
		return fmt.Errorf("SetAllowServerConfig: 设置租赁服 %s 的配置时出现问题，原因是 %v", rentalServerNumber, err)
	}

	return nil
}
