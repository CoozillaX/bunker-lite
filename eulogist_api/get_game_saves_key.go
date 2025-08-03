package eulogist_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"bunker-lite/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GameSavesKeyRequest ..
type GameSavesKeyRequest struct {
	Token              string `json:"token,omitempty"`
	RentalServerNumber string `json:"rental_server_number,omitempty"`
}

// GameSavesKeyResponse ..
type GameSavesKeyResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
	AESCipher []byte `json:"encrypted_aes_cipher"`
}

// GetGameSavesKey ..
func GetGameSavesKey(c *gin.Context) {
	var request GameSavesKeyRequest
	var enableEncrypt bool

	requestRaw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, GameSavesKeyResponse{
			ErrorInfo: fmt.Sprintf("GetGameSavesKey: 获取存档解密密钥时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	decrypted, err := utils.DecryptPKCS1v15(define.GameSavesEncryptKey, requestRaw)
	if err == nil {
		requestRaw = decrypted
		enableEncrypt = true
	}

	err = json.Unmarshal(requestRaw, &request)
	if err != nil {
		c.JSON(http.StatusOK, GameSavesKeyResponse{
			ErrorInfo: fmt.Sprintf("GetGameSavesKey: 获取存档解密密钥时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, GameSavesKeyResponse{
			ErrorInfo: "GetGameSavesKey: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	if len(request.RentalServerNumber) == 0 {
		c.JSON(http.StatusOK, GameSavesKeyResponse{
			ErrorInfo: "GetGameSavesKey: 提供的租赁服号不得为空",
			Success:   false,
		})
		return
	}

	if enableEncrypt {
		aesCipher, err := database.GetOrCreateGameSavesKey(user.UserUniqueID, request.RentalServerNumber, true)
		if err != nil {
			c.JSON(http.StatusOK, GameSavesKeyResponse{
				ErrorInfo: fmt.Sprintf("GetGameSavesKey: 获取存档解密密钥时出现问题, 原因是 %v", err),
				Success:   false,
			})
			return
		}

		resp := GameSavesKeyResponse{
			Success:   true,
			AESCipher: aesCipher,
		}

		jsonBytes, err := json.Marshal(resp)
		if err != nil {
			c.JSON(http.StatusOK, GameSavesKeyResponse{
				ErrorInfo: fmt.Sprintf("GetGameSavesKey: 获取存档解密密钥时出现问题, 原因是 %v", err),
				Success:   false,
			})
			return
		}

		encrypted, err := utils.EncryptPKCS1v15(&define.GameSavesEncryptKey.PublicKey, jsonBytes)
		if err != nil {
			c.JSON(http.StatusOK, GameSavesKeyResponse{
				ErrorInfo: fmt.Sprintf("GetGameSavesKey: 获取存档解密密钥时出现问题, 原因是 %v", err),
				Success:   false,
			})
			return
		}

		c.Data(http.StatusOK, "application/octet-stream", encrypted)
		return
	}

	canDownloadKey := user.CanGetGameSavesKeyCipher
	configs := database.GetAllowServerConfig(request.RentalServerNumber, true)
	for _, value := range configs {
		if value.EulogistUserUniqueID == user.UserUniqueID {
			canDownloadKey = true
			break
		}
	}

	if !canDownloadKey {
		c.JSON(http.StatusOK, GameSavesKeyResponse{
			ErrorInfo: "GetGameSavesKey: 您没有权限得到目标租赁服的解密密钥, 请联系对应租赁服管理人员",
			Success:   false,
		})
		return
	}

	aesCipher, err := database.GetOrCreateGameSavesKey(user.UserUniqueID, request.RentalServerNumber, true)
	if err != nil {
		c.JSON(http.StatusOK, GameSavesKeyResponse{
			ErrorInfo: fmt.Sprintf("GetGameSavesKey: 获取存档解密密钥时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, GameSavesKeyResponse{
		Success:   true,
		AESCipher: aesCipher,
	})
}
