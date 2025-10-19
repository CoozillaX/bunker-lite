package database

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-lite/define"
	"bunker-lite/enhance"
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"go.etcd.io/bbolt"
)

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
	gu *g79.G79User,
	activeGu *define.ActiveG79User,
	err error,
) {
	if useLock {
		LockG79Transaction(helper.HelperToken)
		defer UnlockG79Transaction(helper.HelperToken)
	}

	if len(peAuthData) == 0 && len(saAuthData) == 0 {
		var mu *defines.MpayUser
		var protocolError *defines.ProtocolError
		if err = json.Unmarshal(helper.MpayUserData, mu); err != nil {
			return nil, nil, fmt.Errorf("RegisterActiveG79User: %v", err)
		}
		if gu, protocolError = g79.Login(engineVersion, mu); protocolError != nil {
			return nil, nil, fmt.Errorf("RegisterActiveG79User: %v", protocolError.Error())
		}
	}
	if len(peAuthData) > 0 {
		gu, err = enhance.PEAuthLogin(peAuthData)
	}
	if len(saAuthData) > 0 {
		gu, err = enhance.SaAuthLogin(engineVersion, saAuthData)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("RegisterActiveG79User: %v", err)
	}

	jsonBytes, err := json.Marshal(gu)
	if err != nil {
		return nil, nil, fmt.Errorf("RegisterActiveG79User: %v", err)
	}
	activeG79User := define.ActiveG79User{
		SessionID:         uuid.NewString(),
		G79UserData:       jsonBytes,
		G79UserExpireTime: time.Now().Unix() + 1800,
	}

	err = serverDatabase.Update(func(tx *bbolt.Tx) error {
		buf := bytes.NewBuffer(nil)
		writer := protocol.NewWriter(buf, 0)
		activeG79User.Marshal(writer)
		return tx.
			Bucket([]byte(DATABASE_KEY_ACTIVE_G79_USER)).
			Put([]byte(helper.HelperToken), buf.Bytes())
	})
	if err != nil {
		return nil, nil, fmt.Errorf("RegisterActiveG79User: %v", err)
	}

	return gu, &activeG79User, nil
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

// LoadActiveG79User ..
func LoadActiveG79User(helperToken string, useLock bool) (gu *g79.G79User, activeGu *define.ActiveG79User, found bool, err error) {
	if useLock {
		LockG79Transaction(helperToken)
		defer UnlockG79Transaction(helperToken)
	}

	_ = serverDatabase.View(func(tx *bbolt.Tx) error {
		payload := tx.Bucket([]byte(DATABASE_KEY_ACTIVE_G79_USER)).Get([]byte(helperToken))
		if len(payload) == 0 {
			return nil
		}

		buf := bytes.NewBuffer(payload)
		reader := protocol.NewReader(buf, 0, false)
		activeGu = new(define.ActiveG79User)
		activeGu.Marshal(reader)

		return nil
	})
	if activeGu == nil {
		return nil, nil, false, nil
	}
	if activeGu.G79UserExpireTime <= time.Now().Unix() {
		if err = DeleteActiveG79User(helperToken, false); err != nil {
			return nil, nil, false, fmt.Errorf("LoadActiveG79User: %v", err)
		}
		return nil, nil, false, nil
	}

	err = json.Unmarshal(activeGu.G79UserData, &gu)
	if err != nil {
		return nil, nil, false, fmt.Errorf("LoadActiveG79User: %v", err)
	}
	return gu, activeGu, true, nil
}

// LoadOrRegisterActiveG79User ..
func LoadOrRegisterActiveG79User(helper define.AuthServerHelper, engineVersion string, peAuthData string, saAuthData string, useLock bool) (
	gu *g79.G79User,
	activeGu *define.ActiveG79User,
	err error,
) {
	if useLock {
		LockG79Transaction(helper.HelperToken)
		defer UnlockG79Transaction(helper.HelperToken)
	}

	gu, activeGu, found, err := LoadActiveG79User(helper.HelperToken, false)
	if err != nil {
		return nil, nil, fmt.Errorf("LoadOrRegisterActiveG79User: %v", err)
	}
	if found {
		return gu, activeGu, nil
	}

	gu, activeGu, err = RegisterActiveG79User(helper, engineVersion, peAuthData, saAuthData, false)
	if err != nil {
		return nil, nil, fmt.Errorf("LoadOrRegisterActiveG79User: %v", err)
	}
	return gu, activeGu, nil
}

// ExtendG79UserLifeTime ..
func ExtendG79UserLifeTime(helperToken string, newG79UserToken string, sessionID string, useLock bool) (
	gu *g79.G79User,
	activeGu *define.ActiveG79User,
	err error,
) {
	if useLock {
		LockG79Transaction(helperToken)
		defer UnlockG79Transaction(helperToken)
	}

	gu, activeGu, found, err := LoadActiveG79User(helperToken, false)
	if err != nil {
		return nil, nil, fmt.Errorf("ExtendG79UserLifeTime: %v", err)
	}
	if !found {
		return nil, nil, fmt.Errorf("ExtendG79UserLifeTime: Session not found or is expired")
	}
	if sessionID != activeGu.SessionID {
		return nil, nil, fmt.Errorf(
			"ExtendG79UserLifeTime: Session ID not matched (expect = %#v, provided = %#v)",
			activeGu.SessionID, sessionID,
		)
	}

	gu.G79Token = newG79UserToken
	jsonBytes, err := json.Marshal(gu)
	if err != nil {
		return nil, nil, fmt.Errorf("ExtendG79UserLifeTime: %v", err)
	}
	activeGu.G79UserData = jsonBytes
	activeGu.G79UserExpireTime = time.Now().Unix() + 1800

	err = serverDatabase.Update(func(tx *bbolt.Tx) error {
		buf := bytes.NewBuffer(nil)
		writer := protocol.NewWriter(buf, 0)
		activeGu.Marshal(writer)
		return tx.
			Bucket([]byte(DATABASE_KEY_ACTIVE_G79_USER)).
			Put([]byte(helperToken), buf.Bytes())
	})
	if err != nil {
		return nil, nil, fmt.Errorf("ExtendG79UserLifeTime: %v", err)
	}

	return gu, activeGu, nil
}
