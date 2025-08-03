package eulogist_api

import (
	"bunker-lite/database"
	"bunker-lite/define"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

const (
	ActionTypeEditManageServer uint8 = iota
	ActionTypeRemoveManageServer
	ActionTypeRequestServerList
)

// ManagerChangeRequest ..
type ManagerChangeRequest struct {
	Token            string `json:"token,omitempty"`
	ActionType       uint8  `json:"action_type"`
	EulogistUserName string `json:"eulogist_user_name,omitempty"`
	ServerNumber     string `json:"server_number,omitempty"`
}

// ManagerChangeResponse ..
type ManagerChangeResponse struct {
	ErrorInfo       string   `json:"error_info"`
	Success         bool     `json:"success"`
	ServerCanManage []string `json:"server_can_manage"`
}

// ChangeManager ..
func ChangeManager(c *gin.Context) {
	var request ManagerChangeRequest

	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusOK, ManagerChangeResponse{
			ErrorInfo: fmt.Sprintf("ChangeManager: 设置租赁服管理人员的配置时出现问题, 原因是 %v", err),
			Success:   false,
		})
		return
	}

	if !database.CheckUserByToken(request.Token, true) {
		c.JSON(http.StatusOK, ManagerChangeResponse{
			ErrorInfo: "ChangeManager: 无效的赞颂者令牌",
			Success:   false,
		})
		return
	}
	user := database.GetUserByToken(request.Token, true)

	switch user.UserPermissionLevel {
	case define.UserPermissionSystem:
	case define.UserPermissionAdmin:
	default:
		c.JSON(http.StatusOK, ManagerChangeResponse{
			ErrorInfo: "ChangeManager: 权限不足",
			Success:   false,
		})
		return
	}

	if len(request.EulogistUserName) == 0 {
		c.JSON(http.StatusOK, ManagerChangeResponse{
			ErrorInfo: "ChangeManager: 提供的赞颂者用户名不得为空",
			Success:   false,
		})
		return
	}

	if request.ActionType != ActionTypeRequestServerList && len(request.ServerNumber) == 0 {
		c.JSON(http.StatusOK, ManagerChangeResponse{
			ErrorInfo: "ChangeManager: 提供的租赁服号的长度不得为 0",
			Success:   false,
		})
		return
	}

	if !database.CheckUserByName(request.EulogistUserName, true) {
		c.JSON(http.StatusOK, ManagerChangeResponse{
			ErrorInfo: fmt.Sprintf(
				"ChangeManager: 目标用户 (%s) 没有找到, 他可能恰好改名了？",
				request.EulogistUserName,
			),
			Success: false,
		})
		return
	}
	dst := database.GetUserByName(request.EulogistUserName, true)

	switch request.ActionType {
	case ActionTypeEditManageServer:
		isRepeat := slices.Contains(dst.RentalServerCanManage, request.ServerNumber)
		if !isRepeat {
			dst.RentalServerCanManage = append(dst.RentalServerCanManage, request.ServerNumber)
		}
	case ActionTypeRemoveManageServer:
		newList := make([]string, 0)
		for _, value := range dst.RentalServerCanManage {
			if value != request.ServerNumber {
				newList = append(newList, value)
			}
		}
		dst.RentalServerCanManage = newList
	}

	if request.ActionType != ActionTypeRequestServerList {
		if err = database.UpdateUserInfo(dst, true); err != nil {
			c.JSON(http.StatusOK, ManagerChangeResponse{
				ErrorInfo: fmt.Sprintf("ChangeManager: 请求或设置租赁服配置时出现问题, 原因是 %v", err),
				Success:   false,
			})
			return
		}
		c.JSON(http.StatusOK, ManagerChangeResponse{Success: true})
		return
	}

	c.JSON(http.StatusOK, ManagerChangeResponse{
		Success:         true,
		ServerCanManage: dst.RentalServerCanManage,
	})
}
