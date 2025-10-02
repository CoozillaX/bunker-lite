package std_api

import (
	"bunker-lite/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"bunker-core/mcp"

	"github.com/gin-gonic/gin"
)

type TransferCheckNumRequest struct {
	Data string `json:"data"`
}

type TransferCheckNumResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Value   string `json:"value"`
}

func TransferCheckNum(c *gin.Context) {
	var request TransferCheckNumRequest

	// parse request
	err := c.Bind(&request)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// get session
	session := utils.GetSessionByBearer(c)

	// get engineVersion
	engineVersion, ok := session.Load(session_key_engine_version)
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}

	// get patchVersion
	patchVersion, ok := session.Load(session_key_patch_version)
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}

	// parse fb req
	var dataList []any
	if err := json.Unmarshal([]byte(request.Data), &dataList); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if len(dataList) != 3 {
		c.Status(http.StatusBadRequest)
		return
	}

	// get check num
	result, err := mcp.GetMCPCheckNum(
		engineVersion.(string),
		patchVersion.(string),
		dataList[0].(string),
		dataList[1].(string),
		strconv.Itoa(int(dataList[2].(float64))),
	)
	if err != nil {
		c.JSON(http.StatusOK, TransferCheckNumResponse{
			Success: false,
			Message: fmt.Sprintf("TransferCheckNum: 获取 CheckNum 时出现问题, 原因是 %v", err),
		})
		return
	}

	// return result
	c.JSON(
		http.StatusOK,
		TransferCheckNumResponse{
			Success: true,
			Message: "ok",
			Value:   result,
		},
	)
}
