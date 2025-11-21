package eulogist_api

import (
	"bunker-core/protocol/mpay"
	"sync"
	"time"
)

const VerifyTransactionExpireSeconds = 300 // 5 min is enough for finish verify

var verifyTransactionMu = new(sync.Mutex)
var verifyTransactions = make(map[string]*VerifyTransaction)

// VerifyTransaction ..
type VerifyTransaction struct {
	LoginHelpder   *mpay.LoginHelper
	Mobile         string
	ExpireUnixTime int64
}

// loadOrCreateVerifyTransaction ..
func loadOrCreateVerifyTransaction(uniqueID string) *VerifyTransaction {
	verifyTransactionMu.Lock()
	defer verifyTransactionMu.Unlock()

	currentTime := time.Now()
	newMap := make(map[string]*VerifyTransaction)
	for key, value := range verifyTransactions {
		if currentTime.Unix() < value.ExpireUnixTime {
			newMap[key] = value
		}
	}
	verifyTransactions = newMap

	tran, ok := verifyTransactions[uniqueID]
	if ok {
		return tran
	}

	tran = &VerifyTransaction{
		LoginHelpder:   mpay.CreateLoginHelper(nil),
		ExpireUnixTime: currentTime.Unix() + VerifyTransactionExpireSeconds,
	}
	verifyTransactions[uniqueID] = tran
	return tran
}

// deleteVerifyTransaction ..
func deleteVerifyTransaction(uniqueID string) {
	verifyTransactionMu.Lock()
	defer verifyTransactionMu.Unlock()
	delete(verifyTransactions, uniqueID)
}
