package pollers

import (
	"bunker-lite/service/define"
	"bunker-lite/utils"
	"context"
)

const (
	PollerActiveSessionSuggestedSecond = 35 * 60
	PollerActiveSessionCheckTimeSecond = 2
)

var activeSession *utils.TimeScheduler[string]

func init() {
	activeSession = utils.NewTimeScheduler[string](
		context.Background(),
		PollerActiveSessionSuggestedSecond,
		PollerActiveSessionCheckTimeSecond,
	)
}

// AppendSession ..
func AppendSession(
	sessionID string,
	transaction *utils.Transaction,
	expectedUnixTime int64,
	loadFunc func(sessionID string, extendSessionLifeTime bool, useLock bool) (session define.ActiveSession, found bool, err error),
	deleteFunc func(sessionID string, useLock bool) error,
	updateFunc func(session define.ActiveSession, useLock bool) error,
) (alreadyHit bool, success bool) {
	callBack := func(sessionID *string) (appendToQueue bool) {
		transaction.Lock(*sessionID)
		defer transaction.Unlock(*sessionID)

		session, found, err := loadFunc(*sessionID, false, false)
		if err != nil || !found {
			_ = deleteFunc(*sessionID, false)
			return false
		}

		protocolError := session.G79User().Update()
		if protocolError != nil {
			_ = deleteFunc(*sessionID, false)
			return false
		}

		session.RecordG79UserData.NextRefreshTime += PollerActiveSessionSuggestedSecond
		if err = updateFunc(session, false); err != nil {
			_ = deleteFunc(*sessionID, false)
			return false
		}

		return true
	}

	return activeSession.Append(
		sessionID,
		&sessionID,
		callBack,
		expectedUnixTime,
	)
}

// DeleteSession ..
func DeleteSession(sessionID string, transaction *utils.Transaction, useLock bool) {
	if useLock {
		transaction.Lock(sessionID)
		defer transaction.Unlock(sessionID)
	}
	activeSession.Delete(sessionID)
}
