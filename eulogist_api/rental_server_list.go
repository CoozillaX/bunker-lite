package eulogist_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ActionTypeEditRentalServer uint8 = iota
	ActionTypeRemoveRentalServer
)

// RentalServerListRequest ..
type RentalServerListRequest struct {
	Token          string `json:"token,omitempty"`
	ActionType     uint8  `json:"action_type"`
	ServerNumber   string `json:"server_number,omitempty"`
	ServerPassCode string `json:"server_passcode,omitempty"`
}

// RentalServerListResponse ..
type RentalServerListResponse struct {
	ErrorInfo string `json:"error_info"`
	Success   bool   `json:"success"`
}

// RentalServerList ..
func RentalServerList(c *gin.Context) {
	var request RentalServerListRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, RentalServerListResponse{
			ErrorInfo: fmt.Sprintf("RentalServerList: 请求或设置租赁服配置时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, RentalServerListResponse{
			ErrorInfo: "RentalServerList: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	if len(request.ServerNumber) == 0 {
		c.JSON(http.StatusOK, RentalServerListResponse{
			ErrorInfo: "RentalServerList: 提供的租赁服号的长度不得为 0",
			Success:   false,
		})
		return
	}

	switch request.ActionType {
	case ActionTypeEditRentalServer:
		isRepeat := false
		for index, value := range user.RentalServerConfig {
			if value.ServerNumber == request.ServerNumber {
				user.RentalServerConfig[index].ServerPassCode = request.ServerPassCode
				isRepeat = true
				break
			}
		}
		if !isRepeat {
			user.RentalServerConfig = append(user.RentalServerConfig, define.RentalServerConfig{
				ServerNumber:   request.ServerNumber,
				ServerPassCode: request.ServerPassCode,
			})
		}
	case ActionTypeRemoveRentalServer:
		newList := make([]define.RentalServerConfig, 0)
		for _, value := range user.RentalServerConfig {
			if value.ServerNumber != request.ServerNumber {
				newList = append(newList, value)
			}
		}
		user.RentalServerConfig = newList
	}

	err = database.UpdateUserInfo(user, true)
	if err != nil {
		c.JSON(http.StatusOK, RentalServerListResponse{
			ErrorInfo: fmt.Sprintf("RentalServerList: 请求或设置租赁服配置时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	c.JSON(http.StatusOK, RentalServerListResponse{Success: true})
}
