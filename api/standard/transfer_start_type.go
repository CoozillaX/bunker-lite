package std_api

import (
	"bunker-core/protocol/g79"
	"bunker-lite/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TransferStartTypeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func TransferStartType(c *gin.Context) {
	// get session
	session := utils.GetSessionByBearer(c)

	// get entityID
	entityID, ok := session.Load(session_key_entity_id)
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}

	// get start type
	result, err := g79.GetStartType(
		entityID.(string),
		c.Request.URL.Query().Get("content"),
	)
	if err != nil {
		c.JSON(http.StatusOK, TransferStartTypeResponse{
			Success: false,
			Message: fmt.Sprintf("TransferStartType: 获取 StartType 时出现问题, 原因是 %s", err),
		})
		return
	}

	// return result
	c.JSON(
		http.StatusOK,
		TransferStartTypeResponse{
			Success: true,
			Message: "ok",
			Data:    result,
		},
	)
}
