package pollers

import (
	"bunker-lite/service/define"
	"bunker-lite/utils"
	"context"
	"time"
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
	loadFunc func(ctx context.Context, sessionID string, extendSessionLifeTime bool, deletePollerWhenExpire bool, useSessionLock bool) (session define.ActiveSession, found bool, err error),
	deleteFunc func(sessionID string, deletePoller bool, useLock bool) error,
	updateFunc func(session define.ActiveSession, useLock bool) error,
) (alreadyHit bool, success bool) {
	callBack := func(sessionID *string) (appendToQueue bool) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		transaction.Lock(*sessionID)
		defer transaction.Unlock(*sessionID)

		session, found, err := loadFunc(ctx, *sessionID, false, false, false)
		if err != nil || !found {
			_ = deleteFunc(*sessionID, false, false)
			return false
		}

		protocolError := session.G79User().Update(ctx)
		if protocolError != nil {
			_ = deleteFunc(*sessionID, false, false)
			return false
		}

		session.RecordG79UserData.NextRefreshTime += PollerActiveSessionSuggestedSecond
		if err = updateFunc(session, false); err != nil {
			_ = deleteFunc(*sessionID, false, false)
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
