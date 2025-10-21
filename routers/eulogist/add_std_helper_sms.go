package eulogist_api

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/mpay"
	"bunker-lite/database"
	"bunker-lite/define"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ActionTypeOpenNewTransaction uint8 = iota // Open new transaction
	ActionTypeReLogin                         // Need verify -> Pass verify -> Relogin
	ActionTypeFinishLogin                     // Get SMS Code -> Finish login
)

const (
	ResponseTypeClientNeedSendSMS    uint8 = iota // User need send SMS to NetEase
	ResponseTypeClientNeedReceiveSMS              // User need receive SMS from NetEase
	ResponseTypeLoginSuccess                      // SMS Login success
	ResponseTypeMeetError                         // SMS Login meet error
)

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

// AddStdHelperSMS ..
func AddStdHelperSMS(c *gin.Context) {
	var request SMSHelperAddRequest
	var protocolError *defines.ProtocolError

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ErrorInfo:    fmt.Sprintf("AddStdHelperSMS: 添加新的 MC 账号时出现问题, 原因是 %v", err),
			ResponseType: ResponseTypeMeetError,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ErrorInfo:    "AddStdHelperSMS: 无效的赞颂者令牌",
			ResponseType: ResponseTypeMeetError,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	if request.ActionType != ActionTypeFinishLogin {
		tran := loadOrCreateVerifyTransaction(request.TransactionUUID)

		switch request.ActionType {
		case ActionTypeOpenNewTransaction:
			if len(request.Mobile) > 0 {
				tran.Mobile = request.Mobile
				break
			}
			c.JSON(http.StatusOK, SMSHelperAddResponse{
				ErrorInfo:    "AddStdHelperSMS: 手机号的长度不得为 0",
				ResponseType: ResponseTypeMeetError,
			})
			deleteVerifyTransaction(request.TransactionUUID)
			return
		case ActionTypeReLogin:
			if len(tran.Mobile) > 0 {
				break
			}
			c.JSON(http.StatusOK, SMSHelperAddResponse{
				ErrorInfo:    "AddStdHelperSMS: 请求未找到, 可能已经过期, 请重试",
				ResponseType: ResponseTypeMeetError,
			})
			deleteVerifyTransaction(request.TransactionUUID)
			return
		}

		loginHelper, protocolError := mpay.CreateLoginHelper(tran.MpayUser)
		if protocolError != nil {
			if len(protocolError.VerifyUrl) == 0 {
				c.JSON(http.StatusOK, SMSHelperAddResponse{
					ErrorInfo:    fmt.Sprintf("AddStdHelperSMS: 添加新的 MC 账号时出现问题, 原因是 %v", protocolError.Error()),
					ResponseType: ResponseTypeMeetError,
				})
			} else {
				c.JSON(http.StatusOK, SMSHelperAddResponse{
					VerifyURL:    protocolError.VerifyUrl,
					ResponseType: ResponseTypeClientNeedSendSMS,
				})
			}
			return
		}

		protocolError = loginHelper.GetSMSLoginCode(tran.Mobile)
		if protocolError != nil {
			if len(protocolError.VerifyUrl) == 0 {
				c.JSON(http.StatusOK, SMSHelperAddResponse{
					ErrorInfo:    fmt.Sprintf("AddStdHelperSMS: 添加新的 MC 账号时出现问题, 原因是 %v", protocolError.Error()),
					ResponseType: ResponseTypeMeetError,
				})
			} else {
				c.JSON(http.StatusOK, SMSHelperAddResponse{
					VerifyURL:    protocolError.VerifyUrl,
					ResponseType: ResponseTypeClientNeedSendSMS,
				})
			}
			return
		}

		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ResponseType: ResponseTypeClientNeedReceiveSMS,
		})
		return
	}

	tran := loadOrCreateVerifyTransaction(request.TransactionUUID)
	defer deleteVerifyTransaction(request.TransactionUUID)

	if len(tran.Mobile) == 0 {
		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ErrorInfo:    "AddStdHelperSMS: 请求未找到, 可能已经过期, 请重试",
			ResponseType: ResponseTypeMeetError,
		})
		return
	}

	loginHelper, protocolError := mpay.CreateLoginHelper(tran.MpayUser)
	if protocolError != nil {
		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ErrorInfo:    fmt.Sprintf("AddStdHelperSMS: 添加新的 MC 账号时出现问题, 原因是 %v (stage 1)", protocolError.Error()),
			ResponseType: ResponseTypeMeetError,
		})
		return
	}

	protocolError = loginHelper.SMSLogin(tran.Mobile, request.SMSVerifyCode)
	if protocolError != nil {
		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ErrorInfo:    fmt.Sprintf("AddStdHelperSMS: 添加新的 MC 账号时出现问题, 原因是 %v (stage 2)", protocolError.Error()),
			ResponseType: ResponseTypeMeetError,
		})
		return
	}

	helperUniqueID, protocolError := database.CreateAuthHelper(tran.MpayUser, true)
	if protocolError != nil {
		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ErrorInfo:    fmt.Sprintf("AddStdHelperSMS: 添加新的 MC 账号时出现问题, 原因是 %v (stage 3)", protocolError.Error()),
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
		if _, err = database.DeleteAuthHelper(helper.HelperUniqueID, true); err != nil {
			c.JSON(http.StatusOK, SMSHelperAddResponse{
				ErrorInfo:    fmt.Sprintf("AddStdHelperSMS: 添加新的 MC 账号时出现问题, 原因是 %v", err),
				ResponseType: ResponseTypeMeetError,
			})
			return
		}
		c.JSON(http.StatusOK, SMSHelperAddResponse{
			ErrorInfo:    "AddStdHelperSMS: 该 MC 账号已经存在, 不能重复添加",
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
			ErrorInfo:    fmt.Sprintf("AddStdHelperSMS: 添加新的 MC 账号时出现问题, 原因是 %v", err),
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
