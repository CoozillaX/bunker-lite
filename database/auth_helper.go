package database

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/gameinfo"
	"bunker-lite/define"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"go.etcd.io/bbolt"
)

// CheckAuthHelperByToken ..
func CheckAuthHelperByUniqueID(uniqueID string, useLock bool) (found bool) {
	if useLock {
		mu.RLock()
		defer mu.RUnlock()
	}

	_ = serverDatabase.View(func(tx *bbolt.Tx) error {
		payload := tx.
			Bucket([]byte(DATABASE_KEY_AUTH_HELPER)).
			Get([]byte(uniqueID))
		found = len(payload) > 0
		return nil
	})

	return
}

// CheckAuthHelperByToken ..
func CheckAuthHelperByToken(token string, useLock bool) (found bool) {
	if useLock {
		mu.RLock()
		defer mu.RUnlock()
	}

	_ = serverDatabase.View(func(tx *bbolt.Tx) error {
		payload := tx.
			Bucket([]byte(DATABASE_KEY_TTAH_MAPPING)).
			Get([]byte(token))
		found = len(payload) > 0
		return nil
	})

	return
}

// GetAuthHelperByUniqueID ..
func GetAuthHelperByUniqueID(uniqueID string, useLock bool) (helper define.AuthServerHelper) {
	if useLock {
		mu.RLock()
		defer mu.RUnlock()
	}

	_ = serverDatabase.View(func(tx *bbolt.Tx) error {
		payload := tx.
			Bucket([]byte(DATABASE_KEY_AUTH_HELPER)).
			Get([]byte(uniqueID))

		buf := bytes.NewBuffer(payload)
		reader := protocol.NewReader(buf, 0, false)
		helper.Marshal(reader)

		return nil
	})

	return
}

// GetAuthHelperByToken ..
func GetAuthHelperByToken(token string, useLock bool) (helper define.AuthServerHelper) {
	if useLock {
		mu.RLock()
		defer mu.RUnlock()
	}

	_ = serverDatabase.View(func(tx *bbolt.Tx) error {
		uniqueID := tx.Bucket([]byte(DATABASE_KEY_TTAH_MAPPING)).Get([]byte(token))
		payload := tx.Bucket([]byte(DATABASE_KEY_AUTH_HELPER)).Get(uniqueID)

		buf := bytes.NewBuffer(payload)
		reader := protocol.NewReader(buf, 0, false)
		helper.Marshal(reader)

		return nil
	})

	return
}

// CreateAuthHelper ..
func CreateAuthHelper(mpayUser *defines.MpayUser, useLock bool) (uniqueID string, protocolError *defines.ProtocolError) {
	if useLock {
		mu.Lock()
		defer mu.Unlock()
	}

	gu, protocolError := g79.Login(gameinfo.DefaultEngineVersion, mpayUser)
	if protocolError != nil {
		return "", protocolError
	}
	mpayUserBytes, err := json.Marshal(mpayUser)
	if err != nil {
		return "", &defines.ProtocolError{
			Message: fmt.Sprintf("CreateAuthHelper: 创建 MC 账号时出现问题，原因是 %v", err),
		}
	}

	helper := define.AuthServerHelper{
		HelperUniqueID: uuid.NewString(),
		HelperToken:    uuid.NewString(),
		GameNickName:   gu.Username,
		G79UserUID:     gu.Uid,
		MpayUserData:   mpayUserBytes,
	}
	err = serverDatabase.Update(func(tx *bbolt.Tx) error {
		buf := bytes.NewBuffer(nil)
		writer := protocol.NewWriter(buf, 0)
		helper.Marshal(writer)

		err = tx.
			Bucket([]byte(DATABASE_KEY_AUTH_HELPER)).
			Put(
				[]byte(helper.HelperUniqueID),
				buf.Bytes(),
			)
		if err != nil {
			return err
		}

		err = tx.Bucket([]byte(DATABASE_KEY_TTAH_MAPPING)).
			Put(
				[]byte(helper.HelperToken),
				[]byte(helper.HelperUniqueID),
			)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", &defines.ProtocolError{
			Message: fmt.Sprintf("CreateAuthHelper: 创建 MC 账号时出现问题，原因是 %v", err),
		}
	}

	return helper.HelperUniqueID, nil
}

// GetHelperBasicInfo ..
func GetHelperBasicInfo(uniqueID string, useLock bool) (nickName string, G79UserUID string, protocolError *defines.ProtocolError) {
	var mpayUser defines.MpayUser

	if useLock {
		mu.Lock()
		defer mu.Unlock()
	}

	if !CheckAuthHelperByUniqueID(uniqueID, false) {
		return "", "", &defines.ProtocolError{
			Message: "GetHelperBasicInfo: 无法找到目标 MC 账号",
		}
	}
	helper := GetAuthHelperByUniqueID(uniqueID, false)

	err := json.Unmarshal(helper.MpayUserData, &mpayUser)
	if err != nil {
		return "", "", &defines.ProtocolError{
			Message: fmt.Sprintf("GetHelperBasicInfo: 查询 MC 账号信息时出现问题，原因是 %v", err),
		}
	}

	gu, protocolError := g79.Login(gameinfo.DefaultEngineVersion, &mpayUser)
	if protocolError != nil {
		return "", "", protocolError
	}
	mpayUserBytes, err := json.Marshal(mpayUser)
	if err != nil {
		return "", "", &defines.ProtocolError{
			Message: fmt.Sprintf("GetHelperBasicInfo: 查询 MC 账号信息时出现问题，原因是 %v", err),
		}
	}

	helper.GameNickName = gu.Username
	helper.G79UserUID = gu.Uid
	helper.MpayUserData = mpayUserBytes

	err = serverDatabase.Update(func(tx *bbolt.Tx) error {
		buf := bytes.NewBuffer(nil)
		writer := protocol.NewWriter(buf, 0)
		helper.Marshal(writer)

		err = tx.
			Bucket([]byte(DATABASE_KEY_AUTH_HELPER)).
			Put(
				[]byte(helper.HelperUniqueID),
				buf.Bytes(),
			)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", "", &defines.ProtocolError{
			Message: fmt.Sprintf("UpdateAuthHelper: 更新 MC 账号信息时出现问题，原因是 %v", err),
		}
	}

	return helper.GameNickName, helper.G79UserUID, nil
}

// UpdateHelperToken ..
func UpdateHelperToken(uniqueID string, useLock bool) error {
	if useLock {
		mu.Lock()
		defer mu.Unlock()
	}

	if !CheckAuthHelperByUniqueID(uniqueID, false) {
		return fmt.Errorf("UpdateHelperToken: 目标 MC 账号不存在")
	}

	helper := GetAuthHelperByUniqueID(uniqueID, false)
	legacyToken := helper.HelperToken
	helper.HelperToken = uuid.NewString()

	err := serverDatabase.Update(func(tx *bbolt.Tx) error {
		buf := bytes.NewBuffer(nil)
		writer := protocol.NewWriter(buf, 0)
		helper.Marshal(writer)

		err := tx.
			Bucket([]byte(DATABASE_KEY_AUTH_HELPER)).
			Put(
				[]byte(helper.HelperUniqueID),
				buf.Bytes(),
			)
		if err != nil {
			return err
		}

		err = tx.Bucket([]byte(DATABASE_KEY_TTAH_MAPPING)).Delete([]byte(legacyToken))
		if err != nil {
			return err
		}

		err = tx.
			Bucket([]byte(DATABASE_KEY_TTAH_MAPPING)).
			Put(
				[]byte(helper.HelperToken),
				[]byte(helper.HelperUniqueID),
			)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("UpdateHelperToken: 更新 MC 账号的令牌时出现问题，原因是 %v", err)
	}

	return nil
}

// DeleteAuthHelper ..
func DeleteAuthHelper(uniqueID string, useLock bool) error {
	if useLock {
		mu.Lock()
		defer mu.Unlock()
	}

	if !CheckAuthHelperByUniqueID(uniqueID, false) {
		return fmt.Errorf("DeleteAuthHelper: 目标 MC 账号不存在")
	}
	helper := GetAuthHelperByUniqueID(uniqueID, false)

	err := serverDatabase.Update(func(tx *bbolt.Tx) error {
		err := tx.Bucket([]byte(DATABASE_KEY_AUTH_HELPER)).Delete([]byte(helper.HelperUniqueID))
		if err != nil {
			return err
		}
		err = tx.Bucket([]byte(DATABASE_KEY_TTAH_MAPPING)).Delete([]byte(helper.HelperToken))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("DeleteAuthHelper: 删除 MC 账号时出现问题，原因是 %v", err)
	}

	return nil
}
