package database

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/g79"
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
	var gu *g79.G79User
	session = define.NewActiveSession()
	session.SessionID = uuid.NewString()

	if useGeneralLock {
		mu.Lock()
		defer mu.Unlock()
	}
	if useSessionLock {
		ActiveSessionTran.Lock(session.SessionID)
		defer ActiveSessionTran.Unlock(session.SessionID)
	}

	if len(peAuthData) == 0 && len(saAuthData) == 0 {
		var mu = new(defines.MpayUser)
		var protocolError *defines.ProtocolError
		if err = json.Unmarshal(helper.MpayUserData, mu); err != nil {
			return define.ActiveSession{}, fmt.Errorf("RegisterActiveSession: %v", err)
		}
		if gu, protocolError = g79.Login(engineVersion, mu); protocolError != nil {
			return define.ActiveSession{}, fmt.Errorf("RegisterActiveSession: %v", protocolError.Error())
		}
		session.SessionType = define.SessionTypeMpayUser
	}
	if len(peAuthData) > 0 {
		gu, err = enhance.PeAuthLogin(peAuthData)
		session.SessionType = define.SessionTypePeAuth
	}
	if len(saAuthData) > 0 {
		gu, err = enhance.SaAuthLogin(engineVersion, saAuthData)
		session.SessionType = define.SessionTypeSaAuth
	}
	if err != nil {
		return define.ActiveSession{}, fmt.Errorf("RegisterActiveSession: %v", err)
	}

	currentTime := time.Now().Unix()
	session.SessionStartTime = currentTime
	session.SessionExpireTime = currentTime + define.SessionExpireTimeSecond
	session.RecordG79UserData.NextRefreshTime = currentTime + pollers.PollerActiveSessionSuggestedSecond
	session.RecordG79UserData.InternalG79User = gu

	err = serverDatabase.Update(func(tx *bbolt.Tx) error {
		oldSessionID := tx.Bucket([]byte(DATABASE_KEY_AHSI_MAPPING)).Get([]byte(helper.HelperUniqueID))
		if len(oldSessionID) > 0 {
			err := tx.Bucket([]byte(DATABASE_KEY_ACTIVE_SESSION)).Delete(oldSessionID)
			if err != nil {
				return err
			}
		}

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

	_, _ = pollers.AppendSession(
		session.SessionID,
		ActiveSessionTran,
		utils.ExpectedUnixTimeNotSet,
		LoadActiveSession,
		DeleteActiveSession,
		UpdateActiveSession,
	)
	return session, nil
}

// DeleteActiveSession ..
func DeleteActiveSession(sessionID string, useSessionLock bool) error {
	if useSessionLock {
		ActiveSessionTran.Lock(sessionID)
		defer ActiveSessionTran.Unlock(sessionID)
	}

	pollers.DeleteSession(
		sessionID,
		ActiveSessionTran,
		false,
	)
	err := serverDatabase.Update(func(tx *bbolt.Tx) error {
		return tx.
			Bucket([]byte(DATABASE_KEY_ACTIVE_SESSION)).
			Delete([]byte(sessionID))
	})
	if err != nil {
		return fmt.Errorf("DeleteActiveSession: %v", err)
	}

	return nil
}

// LoadActiveSession ..
func LoadActiveSession(sessionID string, useSessionLock bool) (session define.ActiveSession, found bool, err error) {
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
		_ = DeleteActiveSession(sessionID, false)
		return define.ActiveSession{}, false, nil
	}

	session.G79User().Username, err = enhance.GetName(session.G79User())
	if err != nil {
		_ = DeleteActiveSession(sessionID, false)
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
		if useSessionLock {
			ActiveSessionTran.Lock(sessionID)
			defer ActiveSessionTran.Unlock(sessionID)
		}
		session, found, err := LoadActiveSession(sessionID, false)
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
