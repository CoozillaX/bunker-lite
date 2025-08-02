package eulogist_api

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/mpay"
	"bunker-lite/database"
	"bunker-lite/define"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const SMSTransactionExpireSeconds = 300 // 5 min is enough for finish SMS Login

const (
	ActionTypeOpenNewTransaction uint8 = iota // Open new transaction
	ActionTypeFinishVerify                    // User send/receive SMS to/from NetEase
)

const (
	ResponseTypeClientNeedSendSMS    uint8 = iota // User need send SMS to NetEase
	ResponseTypeClientNeedReceiveSMS              // User need receive SMS from NetEase
	ResponseTypeLoginSuccess                      // SMS Login success
	ResponseTypeMeetError                         // SMS Login meet error
)

var smsTransactionMu = new(sync.Mutex)
var activeSMSTransaction = make(map[string]SMSTransaction)

// SMSTransaction ..
type SMSTransaction struct {
	Mobile         string
	VerifyFunc     func(code string) (*defines.MpayUser, *defines.ProtocolError)
	ExpireUnixTime int64
}

// SMSHelperAddRequest ..
type SMSHelperAddRequest struct {
	Token           string `json:"token,omitempty"`
	TransactionUUID string `json:"transaction_uuid"`
	ActionType      uint8  `json:"action_type"`
	Mobile          string `json:"mobile,omitempty"`
	SMSVerifyCode   string `json:"sms_verify_code,omitempty"`
}

// SMSHelperAddResponse ..
type SMSHelperAddResponse struct {
	ErrorInfo      string `json:"error_info"`
	ResponseType   uint8  `json:"response_type"`
	VerifyURL      string `json:"verify_url"`
	HelperUniqueID string `json:"helper_unique_id"`
	GameNickName   string `json:"game_nick_name"`
	G79UserUID     string `json:"g79_user_uid"`
}

// AddHelperSMS ..
func AddHelperSMS(c *gin.Context) {
	var request SMSHelperAddRequest
	smsTransactionMu.Lock()
	defer smsTransactionMu.Unlock()

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ErrorInfo:    fmt.Sprintf("AddHelperSMS: 添加新的 MC 账号时出现问题，原因是 %v", err),
			ResponseType: ResponseTypeMeetError,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ErrorInfo:    "AddHelperSMS: 无效的赞颂者令牌",
			ResponseType: ResponseTypeMeetError,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	if request.ActionType == ActionTypeOpenNewTransaction {
		currentTime := time.Now()
		newMap := make(map[string]SMSTransaction)
		for key, value := range activeSMSTransaction {
			if currentTime.Unix() < value.ExpireUnixTime {
				newMap[key] = value
			}
		}
		activeSMSTransaction = newMap

		if _, ok := activeSMSTransaction[request.TransactionUUID]; ok {
			c.JSON(http.StatusOK, SMSHelperAddResponse{
				ErrorInfo:    "AddHelperSMS: 目标事务已经打开",
				ResponseType: ResponseTypeMeetError,
			})
			return
		}

		if len(request.Mobile) == 0 {
			c.JSON(http.StatusOK, SMSHelperAddResponse{
				ErrorInfo:    "AddHelperSMS: 手机号的长度不得为 0",
				ResponseType: ResponseTypeMeetError,
			})
			return
		}

		mu := new(defines.MpayUser)
		tran := SMSTransaction{
			Mobile:         request.Mobile,
			ExpireUnixTime: currentTime.Unix() + SMSTransactionExpireSeconds,
		}

		verifyFunc, protocolError := mpay.CreateLoginHelper(mu).SMSLogin(tran.Mobile)
		if protocolError != nil && len(protocolError.VerifyUrl) == 0 {
			c.JSON(http.StatusOK, SMSHelperAddResponse{
				ErrorInfo:    fmt.Sprintf("AddHelperSMS: 添加新的 MC 账号时出现问题，原因是 %v", protocolError.Error()),
				ResponseType: ResponseTypeMeetError,
			})
			return
		}
		tran.VerifyFunc = verifyFunc
		activeSMSTransaction[request.TransactionUUID] = tran

		if protocolError == nil || len(protocolError.VerifyUrl) > 0 {
			var verifyURL string
			if protocolError != nil {
				verifyURL = protocolError.VerifyUrl
			}
			c.JSON(http.StatusOK, SMSHelperAddResponse{
				ResponseType: ResponseTypeClientNeedSendSMS,
				VerifyURL:    verifyURL,
			})
			return
		}

		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ResponseType: ResponseTypeClientNeedReceiveSMS,
		})
		return
	}

	tran, ok := activeSMSTransaction[request.TransactionUUID]
	if !ok {
		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ErrorInfo:    "AddHelperSMS: 请求未找到，可能已经过期，请重试",
			ResponseType: ResponseTypeMeetError,
		})
		return
	}
	defer delete(activeSMSTransaction, request.TransactionUUID)

	mu, protocolError := tran.VerifyFunc(request.SMSVerifyCode)
	if protocolError != nil {
		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ErrorInfo:    fmt.Sprintf("AddHelperSMS: 添加新的 MC 账号时出现问题，原因是 %v (stage 1)", protocolError.Error()),
			ResponseType: ResponseTypeMeetError,
		})
		return
	}

	helperUniqueID, protocolError := database.CreateAuthHelper(mu, true)
	if protocolError != nil {
		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ErrorInfo:    fmt.Sprintf("AddHelperSMS: 添加新的 MC 账号时出现问题，原因是 %v (stage 2)", protocolError.Error()),
			ResponseType: ResponseTypeMeetError,
		})
		return
	}
	helper := database.GetAuthHelperByUniqueID(helperUniqueID, true)

	isRepeat := false
	for _, value := range user.MultipleAuthServerAccounts {
		val, ok := value.(*define.StdAuthServerAccount)
		if ok && val.G79UserUID() == helper.G79UserUID {
			isRepeat = true
			break
		}
	}

	if isRepeat {
		if err = database.DeleteAuthHelper(helper.HelperUniqueID, true); err != nil {
			c.JSON(http.StatusOK, SMSHelperAddResponse{
				ErrorInfo:    fmt.Sprintf("AddHelperSMS: 添加新的 MC 账号时出现问题，原因是 %v", err),
				ResponseType: ResponseTypeMeetError,
			})
			return
		}
		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ErrorInfo:    "AddHelperSMS: 该 MC 账号已经存在, 不能重复添加",
			ResponseType: ResponseTypeMeetError,
		})
		return
	}

	account := define.StdAuthServerAccount{}
	account.UpdateData(map[string]any{
		"gameNickName":       helper.GameNickName,
		"g79UserUID":         helper.G79UserUID,
		"authHelperUniqueID": helper.HelperUniqueID,
	})
	user.MultipleAuthServerAccounts = append(user.MultipleAuthServerAccounts, &account)

	err = database.UpdateUserInfo(user, true)
	if err != nil {
		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ErrorInfo:    fmt.Sprintf("AddHelperSMS: 添加新的 MC 账号时出现问题，原因是 %v", err),
			ResponseType: ResponseTypeMeetError,
		})
		return
	}

	c.JSON(http.StatusOK, SMSHelperAddResponse{
		ResponseType:   ResponseTypeLoginSuccess,
		HelperUniqueID: helper.HelperUniqueID,
		GameNickName:   helper.GameNickName,
		G79UserUID:     helper.G79UserUID,
	})
}
