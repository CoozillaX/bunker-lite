package database

import (
	"bunker-lite/define"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

// CheckUserByName ..
func CheckUserByUniqueID(uniqueID string, useLock bool) (found bool) {
	if useLock {
		mu.RLock()
		defer mu.RUnlock()
	}

	_ = serverDatabase.View(func(tx *bbolt.Tx) error {
		payload := tx.
			Bucket([]byte(DATABASE_KEY_EULOGIST_USER)).
			Get([]byte(uniqueID))
		found = len(payload) > 0
		return nil
	})

	return
}

// CheckUserByName ..
func CheckUserByName(name string, useLock bool) (found bool) {
	if useLock {
		mu.RLock()
		defer mu.RUnlock()
	}

	_ = serverDatabase.View(func(tx *bbolt.Tx) error {
		payload := tx.
			Bucket([]byte(DATABASE_KEY_NTEU_MAPPING)).
			Get([]byte(name))
		found = len(payload) > 0
		return nil
	})

	return
}

// CheckUserByToken ..
func CheckUserByToken(token string, useLock bool) (found bool) {
	if useLock {
		mu.RLock()
		defer mu.RUnlock()
	}

	_ = serverDatabase.View(func(tx *bbolt.Tx) error {
		payload := tx.
			Bucket([]byte(DATABSE_KEY_TTEU_MAPPING)).
			Get([]byte(token))
		found = len(payload) > 0
		return nil
	})

	return
}

// GetUserByUniqueID ..
func GetUserByUniqueID(uniqueID string, useLock bool) (user define.EulogistUser) {
	if useLock {
		mu.RLock()
		defer mu.RUnlock()
	}
	_ = serverDatabase.View(func(tx *bbolt.Tx) error {
		payload := tx.Bucket([]byte(DATABASE_KEY_EULOGIST_USER)).Get([]byte(uniqueID))
		user = define.DecodeEulogistUser(payload)
		return nil
	})
	return
}

// GetUserByName ..
func GetUserByName(name string, useLock bool) (user define.EulogistUser) {
	if useLock {
		mu.RLock()
		defer mu.RUnlock()
	}
	_ = serverDatabase.View(func(tx *bbolt.Tx) error {
		uniqueID := tx.Bucket([]byte(DATABASE_KEY_NTEU_MAPPING)).Get([]byte(name))
		payload := tx.Bucket([]byte(DATABASE_KEY_EULOGIST_USER)).Get(uniqueID)
		user = define.DecodeEulogistUser(payload)
		return nil
	})
	return
}

// GetUserByToken ..
func GetUserByToken(token string, useLock bool) (user define.EulogistUser) {
	if useLock {
		mu.RLock()
		defer mu.RUnlock()
	}
	_ = serverDatabase.View(func(tx *bbolt.Tx) error {
		uniqueID := tx.Bucket([]byte(DATABSE_KEY_TTEU_MAPPING)).Get([]byte(token))
		payload := tx.Bucket([]byte(DATABASE_KEY_EULOGIST_USER)).Get(uniqueID)
		user = define.DecodeEulogistUser(payload)
		return nil
	})
	return
}

// CreateUser ..
func CreateUser(name string, passwordSum256 []byte, permissionLevel uint8, useLock bool) error {
	if useLock {
		mu.Lock()
		defer mu.Unlock()
	}

	if CheckUserByName(name, false) {
		return fmt.Errorf("CreateUser: 名为 %s 的用户已存在", name)
	}
	eulogistUser := define.EulogistUser{
		UserUniqueID:        uuid.NewString(),
		UserName:            name,
		UserPermissionLevel: permissionLevel,
		UserPasswordSum256:  passwordSum256,
		EulogistToken:       uuid.NewString(),
	}

	err := serverDatabase.Update(func(tx *bbolt.Tx) error {
		err := tx.
			Bucket([]byte(DATABASE_KEY_EULOGIST_USER)).
			Put(
				[]byte(eulogistUser.UserUniqueID),
				define.EncodeEulogistUser(eulogistUser),
			)
		if err != nil {
			return err
		}

		err = tx.
			Bucket([]byte(DATABASE_KEY_NTEU_MAPPING)).
			Put(
				[]byte(eulogistUser.UserName),
				[]byte(eulogistUser.UserUniqueID),
			)
		if err != nil {
			return err
		}

		err = tx.
			Bucket([]byte(DATABSE_KEY_TTEU_MAPPING)).
			Put(
				[]byte(eulogistUser.EulogistToken),
				[]byte(eulogistUser.UserUniqueID),
			)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("CreateUser: 创建赞颂者账户失败, 原因是 %s", err)
	}

	return nil
}

// UpdateUserName ..
func UpdateUserName(name string, newName string, useLock bool) error {
	if useLock {
		mu.Lock()
		defer mu.Unlock()
	}

	if !CheckUserByName(name, false) {
		return fmt.Errorf("UpdateUserName: 没有找到名为 %s 的用户", name)
	}
	if CheckUserByName(newName, false) {
		return fmt.Errorf("UpdateUserName: 名为 %s 的用户已经存在", newName)
	}
	user := GetUserByName(name, false)
	user.UserName = newName

	err := serverDatabase.Update(func(tx *bbolt.Tx) error {
		err := tx.Bucket([]byte(DATABASE_KEY_EULOGIST_USER)).Put(
			[]byte(user.UserUniqueID),
			define.EncodeEulogistUser(user),
		)
		if err != nil {
			return err
		}

		bucket := tx.Bucket([]byte(DATABASE_KEY_NTEU_MAPPING))
		if err = bucket.Delete([]byte(name)); err != nil {
			return err
		}
		if err = bucket.Put([]byte(newName), []byte(user.UserUniqueID)); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("UpdateUserName: 更改赞颂者账户的名称时出现问题, 原因是 %s", err)
	}

	return nil
}

// UpdateUserToken ..
func UpdateUserToken(token string, newToken string, useLock bool) error {
	if useLock {
		mu.Lock()
		defer mu.Unlock()
	}

	if !CheckUserByToken(token, false) {
		return fmt.Errorf("UpdateUserToken: 目标用户没有找到")
	}
	if CheckUserByToken(newToken, false) {
		return fmt.Errorf("UpdateUserToken: 新的赞颂者令牌已经被其他人使用过了")
	}
	user := GetUserByToken(token, false)
	user.EulogistToken = newToken

	err := serverDatabase.Update(func(tx *bbolt.Tx) error {
		err := tx.Bucket([]byte(DATABASE_KEY_EULOGIST_USER)).Put(
			[]byte(user.UserUniqueID),
			define.EncodeEulogistUser(user),
		)
		if err != nil {
			return err
		}

		bucket := tx.Bucket([]byte(DATABSE_KEY_TTEU_MAPPING))
		if err = bucket.Delete([]byte(token)); err != nil {
			return err
		}
		if err = bucket.Put([]byte(user.EulogistToken), []byte(user.UserUniqueID)); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("UpdateUserToken: 更新赞颂者账户令牌时出现问题, 原因是 %s", err)
	}

	return nil
}

// UpdateUserInfo ..
func UpdateUserInfo(user define.EulogistUser, useLock bool) error {
	if useLock {
		mu.Lock()
		defer mu.Unlock()
	}

	if !CheckUserByUniqueID(user.UserUniqueID, false) {
		return fmt.Errorf("UpdateUserInfo: 没有找到目标用户")
	}
	recordedUser := GetUserByUniqueID(user.UserUniqueID, false)

	if recordedUser.UserName != user.UserName {
		err := UpdateUserName(recordedUser.UserName, user.UserName, false)
		if err != nil {
			return fmt.Errorf("UpdateUserInfo: 更新赞颂者用户信息时出现错误, 原因是 %s", err)
		}
	}
	if recordedUser.EulogistToken != user.EulogistToken {
		err := UpdateUserToken(recordedUser.EulogistToken, user.EulogistToken, false)
		if err != nil {
			return fmt.Errorf("UpdateUserInfo: 更新赞颂者用户信息时出现错误, 原因是 %s", err)
		}
	}

	err := serverDatabase.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(DATABASE_KEY_EULOGIST_USER)).Put(
			[]byte(user.UserUniqueID),
			define.EncodeEulogistUser(user),
		)
	})
	if err != nil {
		return fmt.Errorf("UpdateUserInfo: 更新赞颂者用户信息时出现错误, 原因是 %s", err)
	}

	return nil
}

// ListUsers ..
func ListUsers(filterString string, useLock bool) (hitUserName []string) {
	if useLock {
		mu.RLock()
		defer mu.RUnlock()
	}

	filterString = strings.ToLower(filterString)
	_ = serverDatabase.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(DATABASE_KEY_EULOGIST_USER))
		_ = bucket.ForEach(func(k, v []byte) error {
			userName := define.DecodeEulogistUser(v).UserName
			if len(filterString) == 0 || strings.Contains(strings.ToLower(userName), filterString) {
				hitUserName = append(hitUserName, userName)
			}
			return nil
		})
		return nil
	})

	return
}
