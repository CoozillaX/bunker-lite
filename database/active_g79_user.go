package database

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-lite/define"
	"bunker-lite/enhance"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

const SessionExpireTimeSecond = 60 * 30

var activeGuTranMutex = new(sync.Mutex)
var activeGuTranMapping = make(map[string]*g79Transaction)

// g79Transaction ..
type g79Transaction struct {
	locker *sync.Mutex
	holder int
}

// newG79Transaction ..
func newG79Transaction() *g79Transaction {
	return &g79Transaction{
		locker: new(sync.Mutex),
		holder: 0,
	}
}

// LockG79Transaction locks the g79 user transaction that
// corresponding to helperToken. If target transaction not
// exist, then it will creates a new one.
func LockG79Transaction(helperToken string) {
	var g79UserLocker *g79Transaction
	var ok bool

	func() {
		activeGuTranMutex.Lock()
		defer activeGuTranMutex.Unlock()

		g79UserLocker, ok = activeGuTranMapping[helperToken]
		if !ok {
			g79UserLocker = newG79Transaction()
			activeGuTranMapping[helperToken] = g79UserLocker
		}
		g79UserLocker.holder++
	}()

	g79UserLocker.locker.Lock()
}

// UnlockG79Transaction unlocks the g79 user transaction
// that corresponding to helperToken. If target transaction
// not exist, or the underlying locker is already unlocked,
// then this func will be panic.
func UnlockG79Transaction(helperToken string) {
	activeGuTranMutex.Lock()
	defer activeGuTranMutex.Unlock()

	g79UserLocker, ok := activeGuTranMapping[helperToken]
	if !ok {
		// We should panic here because when here is not ok,
		// it means somewhere may have some internal error,
		// and the lock states is not completely.
		panic(fmt.Sprintf("UnlockG79Transaction: Transaction %#v not found", helperToken))
	}

	g79UserLocker.holder--
	if g79UserLocker.holder == 0 {
		delete(activeGuTranMapping, helperToken)
	}
	g79UserLocker.locker.Unlock()
}

// RegisterActiveG79User ..
func RegisterActiveG79User(helper define.AuthServerHelper, engineVersion string, peAuthData string, saAuthData string, useLock bool) (
	activeGu define.ActiveG79User,
	err error,
) {
	if useLock {
		LockG79Transaction(helper.HelperToken)
		defer UnlockG79Transaction(helper.HelperToken)
	}

	if len(peAuthData) == 0 && len(saAuthData) == 0 {
		var mu = new(defines.MpayUser)
		var protocolError *defines.ProtocolError
		if err = json.Unmarshal(helper.MpayUserData, mu); err != nil {
			return define.ActiveG79User{}, fmt.Errorf("RegisterActiveG79User: %v", err)
		}
		if activeGu.RecordG79UserData, protocolError = g79.Login(engineVersion, mu); protocolError != nil {
			return define.ActiveG79User{}, fmt.Errorf("RegisterActiveG79User: %v", protocolError.Error())
		}
		activeGu.SessionType = define.SessionTypeMpayUser
	}
	if len(peAuthData) > 0 {
		activeGu.RecordG79UserData, err = enhance.PeAuthLogin(peAuthData)
		activeGu.SessionType = define.SessionTypePeAuth
	}
	if len(saAuthData) > 0 {
		activeGu.RecordG79UserData, err = enhance.SaAuthLogin(engineVersion, saAuthData)
		activeGu.SessionType = define.SessionTypeSaAuth
	}
	if err != nil {
		return define.ActiveG79User{}, fmt.Errorf("RegisterActiveG79User: %v", err)
	}

	currentTime := time.Now().Unix()
	activeGu.SessionID = uuid.NewString()
	activeGu.SessionStartTime = currentTime
	activeGu.SessionExpireTime = currentTime + SessionExpireTimeSecond

	err = serverDatabase.Update(func(tx *bbolt.Tx) error {
		return tx.
			Bucket([]byte(DATABASE_KEY_ACTIVE_G79_USER)).
			Put([]byte(helper.HelperToken), define.EncodeActiveG79User(activeGu))
	})
	if err != nil {
		return define.ActiveG79User{}, fmt.Errorf("RegisterActiveG79User: %v", err)
	}

	return activeGu, nil
}

// DeleteActiveG79User ..
func DeleteActiveG79User(helperToken string, useLock bool) error {
	if useLock {
		LockG79Transaction(helperToken)
		defer UnlockG79Transaction(helperToken)
	}

	err := serverDatabase.Update(func(tx *bbolt.Tx) error {
		return tx.
			Bucket([]byte(DATABASE_KEY_ACTIVE_G79_USER)).
			Delete([]byte(helperToken))
	})
	if err != nil {
		return fmt.Errorf("DeleteActiveG79User: %v", err)
	}

	return nil
}

// MigrateActiveG79User ..
func MigrateActiveG79User(legacyHelperToken string, newHelperToken string, useLock bool) error {
	if useLock {
		LockG79Transaction(legacyHelperToken)
		LockG79Transaction(newHelperToken)
		defer UnlockG79Transaction(legacyHelperToken)
		defer UnlockG79Transaction(newHelperToken)
	}

	err := serverDatabase.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(DATABASE_KEY_ACTIVE_G79_USER))

		payload := bucket.Get([]byte(legacyHelperToken))
		if len(payload) == 0 {
			return nil
		}

		err := bucket.Delete([]byte(legacyHelperToken))
		if err != nil {
			return err
		}

		return bucket.Put(
			[]byte(newHelperToken),
			payload,
		)
	})
	if err != nil {
		return fmt.Errorf("MigrateActiveG79User: %v", err)
	}

	return nil
}

// LoadActiveG79User ..
func LoadActiveG79User(helperToken string, useLock bool) (activeGu define.ActiveG79User, found bool, err error) {
	if useLock {
		LockG79Transaction(helperToken)
		defer UnlockG79Transaction(helperToken)
	}

	_ = serverDatabase.View(func(tx *bbolt.Tx) error {
		payload := tx.Bucket([]byte(DATABASE_KEY_ACTIVE_G79_USER)).Get([]byte(helperToken))
		if len(payload) > 0 {
			activeGu = define.DecodeActiveG79User(payload)
		}
		return nil
	})
	if len(activeGu.SessionID) == 0 {
		return define.ActiveG79User{}, false, nil
	}
	if activeGu.SessionExpireTime <= time.Now().Unix() {
		if err = DeleteActiveG79User(helperToken, false); err != nil {
			return define.ActiveG79User{}, false, fmt.Errorf("LoadActiveG79User: %v", err)
		}
		return define.ActiveG79User{}, false, nil
	}

	activeGu.RecordG79UserData.Username, err = enhance.GetName(activeGu.RecordG79UserData)
	if err != nil {
		_ = DeleteActiveG79User(helperToken, false)
		return define.ActiveG79User{}, false, fmt.Errorf("LoadActiveG79User: %v", err)
	}
	return activeGu, true, nil
}

// LoadOrRegisterActiveG79User ..
func LoadOrRegisterActiveG79User(helper define.AuthServerHelper, engineVersion string, peAuthData string, saAuthData string, useLock bool) (
	activeGu define.ActiveG79User,
	err error,
) {
	if useLock {
		LockG79Transaction(helper.HelperToken)
		defer UnlockG79Transaction(helper.HelperToken)
	}

	activeGu, found, err := LoadActiveG79User(helper.HelperToken, false)
	if err != nil {
		return define.ActiveG79User{}, fmt.Errorf("LoadOrRegisterActiveG79User: %v", err)
	}
	if found {
		return activeGu, nil
	}

	activeGu, err = RegisterActiveG79User(helper, engineVersion, peAuthData, saAuthData, false)
	if err != nil {
		return define.ActiveG79User{}, fmt.Errorf("LoadOrRegisterActiveG79User: %v", err)
	}
	return activeGu, nil
}

// ExtendG79UserLifeTime ..
func ExtendG79UserLifeTime(helperToken string, activeGu define.ActiveG79User, useLock bool) (
	sessionExpireTime int64,
	err error,
) {
	if useLock {
		LockG79Transaction(helperToken)
		defer UnlockG79Transaction(helperToken)
	}

	activeGu.SessionExpireTime =
		time.Now().Unix() + SessionExpireTimeSecond
	err = serverDatabase.Update(func(tx *bbolt.Tx) error {
		return tx.
			Bucket([]byte(DATABASE_KEY_ACTIVE_G79_USER)).
			Put([]byte(helperToken), define.EncodeActiveG79User(activeGu))
	})
	if err != nil {
		return 0, fmt.Errorf("ExtendG79UserLifeTime: %v", err)
	}

	return activeGu.SessionExpireTime, nil
}
