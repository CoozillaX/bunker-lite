package eulogist_api

import (
	"bunker-core/protocol/defines"
	"sync"
	"time"
)

const VerifyTransactionExpireSeconds = 300 // 5 min is enough for finish verify

var verifyTransactionMu = new(sync.Mutex)
var verifyTransactions = make(map[string]*VerifyTransaction)

// VerifyTransaction ..
type VerifyTransaction struct {
	MpayUser             *defines.MpayUser
	Mobile               string
	MobileVerifyCallback func(code string) (*defines.MpayUser, *defines.ProtocolError)
	ExpireUnixTime       int64
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
		MpayUser:       new(defines.MpayUser),
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
