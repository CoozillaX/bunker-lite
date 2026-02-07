package database

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
	"bunker-core/protocol/mpay/android"
	"bunker-lite/enhance"
	"bunker-lite/service/define"
	"bunker-lite/service/pollers"
	"bunker-lite/utils"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

var ActiveSessionTran = utils.NewTransaction()

// initActiveSessionPoller ..
func initActiveSessionPoller() error {
	err := serverDatabase.Update(func(tx *bbolt.Tx) error {
		sessionBucket := tx.Bucket([]byte(DATABASE_KEY_ACTIVE_SESSION))
		sessionToDelete := make([]string, 0)

		sessionBucket.ForEach(func(k, v []byte) error {
			session := define.DecodeActiveSession(v)
			alreadyHit, success := pollers.AppendSession(
				session.SessionID,
				ActiveSessionTran,
				session.RecordG79UserData.NextRefreshTime,
				LoadActiveSession,
				DeleteActiveSession,
				UpdateActiveSession,
			)
			if !alreadyHit && !success {
				sessionToDelete = append(sessionToDelete, string(k))
			}
			return nil
		})

		for _, sessionID := range sessionToDelete {
			if err := sessionBucket.Delete([]byte(sessionID)); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("initActiveSessionPoller: %v", err))
	}
	return nil
}

// RegisterActiveSession ..
func RegisterActiveSession(
	helper define.AuthServerHelper,
	engineVersion string,
	peAuthData string,
	saAuthData string,
	useGeneralLock bool,
	useSessionLock bool,
) (
	session define.ActiveSession,
	err error,
) {
	// prepare
	var g79User *g79.G79User
	var success bool
	session = define.NewActiveSession()
	session.SessionID = uuid.NewString()

	// do lock
	if useGeneralLock {
		mu.Lock()
		defer mu.Unlock()
	}
	if useSessionLock {
		ActiveSessionTran.Lock(session.SessionID)
		defer ActiveSessionTran.Unlock(session.SessionID)
	}

	// check login status and update info after operate
	if time.Now().Unix()-helper.TryLoginUnixTime < 3600 && helper.LoginFailedCount >= 3 {
		return define.ActiveSession{}, fmt.Errorf("RegisterActiveSession: 该辅助用户已因登录多次失败而被锁定, 请等待至多一小时或联系管理员以解除锁定")
	}
	defer func() {
		// prepare
		currentTime := time.Now().Unix()
		// update login status
		if success || currentTime-helper.TryLoginUnixTime >= 3600 {
			helper.LoginFailedCount = 0
		}
		if !success {
			helper.LoginFailedCount++
		}
		helper.TryLoginUnixTime = currentTime
		// update helper info
		UpdateHelperInfo(helper, false)
	}()

	// g79 login
	if len(peAuthData) == 0 && len(saAuthData) == 0 {
		var mu = new(android.AndroidMpayUser)
		var protocolError *defines.ProtocolError
		if err = json.Unmarshal(helper.MpayUserData, mu); err != nil {
			return define.ActiveSession{}, fmt.Errorf("RegisterActiveSession: %v", err)
		}
		if err = mu.UpdateGameInfoByEngineVersion(engineVersion); err != nil {
			return define.ActiveSession{}, fmt.Errorf("RegisterActiveSession: %v", err)
		}
		if g79User, protocolError = g79.Login(mu); protocolError != nil {
			return define.ActiveSession{}, fmt.Errorf("RegisterActiveSession: %v", protocolError.Error())
		}
		session.SessionType = define.SessionTypeMpayUser
	}
	if len(peAuthData) > 0 {
		g79User, err = enhance.PeAuthLogin(peAuthData)
		session.SessionType = define.SessionTypePeAuth
	}
	if len(saAuthData) > 0 {
		g79User, err = enhance.SaAuthLogin(engineVersion, saAuthData)
		session.SessionType = define.SessionTypeSaAuth
	}
	if err != nil {
		return define.ActiveSession{}, fmt.Errorf("RegisterActiveSession: %v", err)
	}

	// delete old session if exists
	oldSessionID, found, err := GetSessionIDByHelperUniqueID(helper.HelperUniqueID, false)
	if err != nil {
		return define.ActiveSession{}, fmt.Errorf("RegisterActiveSession: %v", err)
	}
	if found {
		go DeleteActiveSession(oldSessionID, true, true)
	}

	// set session info
	currentTime := time.Now().Unix()
	session.SessionStartTime = currentTime
	session.SessionExpireTime = currentTime + define.SessionExpireTimeSecond
	session.RecordG79UserData.NextRefreshTime = currentTime + pollers.PollerActiveSessionSuggestedSecond
	session.RecordG79UserData.InternalG79User = g79User

	// update session to underlying database
	err = serverDatabase.Update(func(tx *bbolt.Tx) error {
		err := tx.Bucket([]byte(DATABASE_KEY_ACTIVE_SESSION)).Put(
			[]byte(session.SessionID),
			define.EncodeActiveSession(session),
		)
		if err != nil {
			return err
		}

		err = tx.Bucket([]byte(DATABASE_KEY_AHSI_MAPPING)).Put(
			[]byte(helper.HelperUniqueID),
			[]byte(session.SessionID),
		)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return define.ActiveSession{}, fmt.Errorf("RegisterActiveSession: %v", err)
	}

	// start poller and return
	_, _ = pollers.AppendSession(
		session.SessionID,
		ActiveSessionTran,
		utils.ExpectedUnixTimeNotSet,
		LoadActiveSession,
		DeleteActiveSession,
		UpdateActiveSession,
	)
	success = true
	return session, nil
}

// DeleteActiveSession ..
func DeleteActiveSession(sessionID string, deletePoller bool, useSessionLock bool) error {
	// do lock
	if useSessionLock {
		ActiveSessionTran.Lock(sessionID)
		defer ActiveSessionTran.Unlock(sessionID)
	}

	// delete poller if required
	if deletePoller {
		pollers.DeleteSession(
			sessionID,
			ActiveSessionTran,
			false,
		)
	}

	// delete session from underlying database
	err := serverDatabase.Update(func(tx *bbolt.Tx) error {
		return tx.
			Bucket([]byte(DATABASE_KEY_ACTIVE_SESSION)).
			Delete([]byte(sessionID))
	})
	if err != nil {
		return fmt.Errorf("DeleteActiveSession: %v", err)
	}

	// return
	return nil
}

// LoadActiveSession ..
func LoadActiveSession(
	sessionID string,
	extendSessionLifeTime bool,
	deletePollerWhenExpire bool,
	useSessionLock bool,
) (
	session define.ActiveSession,
	found bool,
	err error,
) {
	if useSessionLock {
		ActiveSessionTran.Lock(sessionID)
		defer ActiveSessionTran.Unlock(sessionID)
	}

	_ = serverDatabase.View(func(tx *bbolt.Tx) error {
		payload := tx.Bucket([]byte(DATABASE_KEY_ACTIVE_SESSION)).Get([]byte(sessionID))
		if len(payload) > 0 {
			session = define.DecodeActiveSession(payload)
		}
		return nil
	})
	if len(session.SessionID) == 0 {
		return define.ActiveSession{}, false, nil
	}
	if session.SessionExpireTime <= time.Now().Unix() {
		_ = DeleteActiveSession(sessionID, deletePollerWhenExpire, false)
		return define.ActiveSession{}, false, nil
	}
	if extendSessionLifeTime {
		if time.Now().Unix()-session.SessionStartTime < define.SessionMaxLifeTimeSecond {
			session.SessionExpireTime = time.Now().Unix() + define.SessionExpireTimeSecond
			if err = UpdateActiveSession(session, false); err != nil {
				_ = DeleteActiveSession(sessionID, deletePollerWhenExpire, false)
				return define.ActiveSession{}, false, fmt.Errorf("LoadActiveSession: %v", err)
			}
		}
	}

	session.G79User().Username, err = enhance.GetName(session.G79User())
	if err != nil {
		_ = DeleteActiveSession(sessionID, deletePollerWhenExpire, false)
		return define.ActiveSession{}, false, fmt.Errorf("LoadActiveSession: %v", err)
	}
	return session, true, nil
}

// GetSessionIDByHelperUniqueID ..
func GetSessionIDByHelperUniqueID(helperUniqueID string, useLock bool) (sessionID string, found bool, err error) {
	if useLock {
		mu.Lock()
		defer mu.Unlock()
	}

	err = serverDatabase.Update(func(tx *bbolt.Tx) error {
		payload := tx.
			Bucket([]byte(DATABASE_KEY_AHSI_MAPPING)).
			Get([]byte(helperUniqueID))
		if len(payload) > 0 {
			sessionID = string(payload)
		}
		return nil
	})
	if err != nil {
		return "", false, fmt.Errorf("GetSessionIDByHelper: %v", err)
	}

	return sessionID, len(sessionID) > 0, nil
}

// LoadOrRegisterActiveSession ..
func LoadOrRegisterActiveSession(
	helper define.AuthServerHelper,
	engineVersion string,
	peAuthData string,
	saAuthData string,
	extendSessionLifeTime bool,
	deletePollerWhenExpire bool,
	useGeneralLock bool,
	useSessionLock bool,
) (
	session define.ActiveSession,
	err error,
) {
	if useGeneralLock {
		mu.Lock()
		defer mu.Unlock()
	}

	sessionID, found, err := GetSessionIDByHelperUniqueID(helper.HelperUniqueID, false)
	if found {
		session, found, err := LoadActiveSession(sessionID, extendSessionLifeTime, deletePollerWhenExpire, useSessionLock)
		if err != nil {
			return define.ActiveSession{}, fmt.Errorf("LoadOrRegisterActiveSession: %v", err)
		}
		if found {
			return session, nil
		}
	}

	session, err = RegisterActiveSession(
		helper,
		engineVersion,
		peAuthData,
		saAuthData,
		false,
		true,
	)
	if err != nil {
		return define.ActiveSession{}, fmt.Errorf("LoadOrRegisterActiveSession: %v", err)
	}
	return session, nil
}

// UpdateActiveSession ..
func UpdateActiveSession(session define.ActiveSession, useSessionLock bool) error {
	if useSessionLock {
		ActiveSessionTran.Lock(session.SessionID)
		defer ActiveSessionTran.Unlock(session.SessionID)
	}

	err := serverDatabase.Update(func(tx *bbolt.Tx) error {
		return tx.
			Bucket([]byte(DATABASE_KEY_ACTIVE_SESSION)).
			Put([]byte(session.SessionID), define.EncodeActiveSession(session))
	})
	if err != nil {
		return fmt.Errorf("UpdateActiveSession: %v", err)
	}

	return nil
}
