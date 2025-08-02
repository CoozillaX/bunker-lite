package eulogist_api

import (
	"bunker-core/protocol/defines"
	"bunker-core/protocol/mpay"
	"bunker-lite/database"
	"bunker-lite/define"
	"bunker-lite/utils"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ActionTypeAddStdEmailHelper uint8 = iota
	ActionTypeAddCustomHelper
)

// HelperAddRequest ..
type HelperAddRequest struct {
	Token      string `json:"token,omitempty"`
	ActionType uint8  `json:"action_type"`

	Email       string `json:"email,omitempty"`
	MD5Password string `json:"md5_password,omitempty"`

	AuthServerAddress string `json:"auth_server_address,omitempty"`
	AuthServerToken   string `json:"auth_server_token,omitempty"`
}

// HelperAddResponse ..
type HelperAddResponse struct {
	ErrorInfo            string `json:"error_info"`
	NetEaseRequireVerify bool   `json:"netease_require_verify"`
	VerifyURL            string `json:"verify_url"`
	Success              bool   `json:"success"`
	HelperUniqueID       string `json:"helper_unique_id"`
	GameNickName         string `json:"game_nick_name"`
	G79UserUID           string `json:"g79_user_uid"`
}

// AddHelperNormal ..
func AddHelperNormal(c *gin.Context) {
	var request HelperAddRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, HelperAddResponse{
			ErrorInfo: fmt.Sprintf("AddHelperNormal: 添加新的 MC 账号时出现问题，原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, HelperAddResponse{
			ErrorInfo: "AddHelperNormal: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	if request.ActionType == ActionTypeAddCustomHelper {
		isRepeat := false
		for _, value := range user.MultipleAuthServerAccounts {
			if value.IsStdAccount() {
				continue
			}
			if value.AuthServerAddress() != request.AuthServerAddress {
				continue
			}
			if value.AuthServerSecret() != request.AuthServerToken {
				continue
			}
			isRepeat = true
			break
		}

		if isRepeat {
			c.JSON(http.StatusOK, HelperAddResponse{
				ErrorInfo: "AddHelperNormal: 该 MC 账号已经存在, 不能重复添加",
				Success:   false,
			})
			return
		}

		if len(request.AuthServerAddress) == 0 {
			c.JSON(http.StatusOK, HelperAddResponse{
				ErrorInfo: "AddHelperNormal: 验证服务地址的长度不得为 0",
				Success:   false,
			})
			return
		}

		user.InternalIncreasingAccountID++
		account := define.CustomAuthServerAccount{}
		account.UpdateData(map[string]any{
			"internalAccountID": user.InternalIncreasingAccountID,
			"authServerAddress": request.AuthServerAddress,
			"authServerToken":   request.AuthServerToken,
		})
		user.MultipleAuthServerAccounts = append(user.MultipleAuthServerAccounts, &account)

		err = database.UpdateUserInfo(user, true)
		if err != nil {
			c.JSON(http.StatusOK, HelperAddResponse{
				ErrorInfo: fmt.Sprintf("AddHelperNormal: 添加新的 MC 账号时出现问题，原因是 %v", err),
				Success:   false,
			})
			return
		}

		c.JSON(http.StatusOK, HelperAddResponse{Success: true})
		return
	}

	emptyMD5PasswordBytes := md5.Sum([]byte{})
	emptyMD5Password := hex.EncodeToString(emptyMD5PasswordBytes[:])
	if len(request.Email) == 0 {
		c.JSON(http.StatusOK, HelperAddResponse{
			ErrorInfo: "AddHelperNormal: 邮箱地址的长度不得为 0",
			Success:   false,
		})
		return
	}
	if len(request.MD5Password) == 0 || request.MD5Password == emptyMD5Password {
		c.JSON(http.StatusOK, HelperAddResponse{
			ErrorInfo: "AddHelperNormal: 邮箱密码的长度不得为 0",
			Success:   false,
		})
		return
	}

	mu := new(defines.MpayUser)
	mu, protocolError := mpay.CreateLoginHelper(mu).PasswordLogin(
		request.Email, request.MD5Password,
		utils.GetPasswordLevel(request.MD5Password),
	)
	if protocolError != nil {
		c.JSON(http.StatusOK, HelperAddResponse{
			ErrorInfo:            fmt.Sprintf("AddHelperNormal: 添加新的 MC 账号时出现问题，原因是 %s", protocolError.Error()),
			NetEaseRequireVerify: len(protocolError.VerifyUrl) != 0,
			VerifyURL:            protocolError.VerifyUrl,
			Success:              false,
		})
		return
	}

	helperUniqueID, protocolError := database.CreateAuthHelper(mu, true)
	if protocolError != nil {
		c.JSON(http.StatusOK, HelperAddResponse{
			ErrorInfo:            fmt.Sprintf("AddHelperNormal: 添加新的 MC 账号时出现问题，原因是 %s", protocolError.Error()),
			NetEaseRequireVerify: len(protocolError.VerifyUrl) != 0,
			VerifyURL:            protocolError.VerifyUrl,
			Success:              false,
		})
		return
	}

	isRepeat := false
	helper := database.GetAuthHelperByUniqueID(helperUniqueID, true)
	for _, value := range user.MultipleAuthServerAccounts {
		val, ok := value.(*define.StdAuthServerAccount)
		if ok && val.G79UserUID() == helper.G79UserUID {
			isRepeat = true
			break
		}
	}

	if isRepeat {
		if err = database.DeleteAuthHelper(helper.HelperUniqueID, true); err != nil {
			c.JSON(http.StatusOK, HelperAddResponse{
				ErrorInfo: fmt.Sprintf("AddHelperNormal: 添加新的 MC 账号时出现问题，原因是 %v", err),
				Success:   false,
			})
			return
		}
		c.JSON(http.StatusOK, HelperAddResponse{
			ErrorInfo: "AddHelperNormal: 该 MC 账号已经存在, 不能重复添加",
			Success:   false,
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
		c.JSON(http.StatusOK, HelperAddResponse{
			ErrorInfo: fmt.Sprintf("AddHelperNormal: 添加新的 MC 账号时出现问题，原因是 %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, HelperAddResponse{
		Success:        true,
		HelperUniqueID: helper.HelperUniqueID,
		GameNickName:   helper.GameNickName,
		G79UserUID:     helper.G79UserUID,
	})
}
