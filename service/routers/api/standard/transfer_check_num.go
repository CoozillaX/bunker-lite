package std_api

import (
	"bunker-lite/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

	// get tradition session
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

	// get platform
	platform, ok := session.Load(session_key_platform)
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}

	// parse fb req
	var dataList []any
	decoder := json.NewDecoder(strings.NewReader(request.Data))
	decoder.UseNumber()
	if err := decoder.Decode(&dataList); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if len(dataList) != 3 {
		c.Status(http.StatusBadRequest)
		return
	}
	mcpData, ok := dataList[0].(string)
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}
	salt, ok := dataList[1].(string)
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}
	uid, ok := dataList[2].(json.Number)
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}

	// get check num
	result, err := mcp.GetMCPCheckNum(
		engineVersion.(string),
		patchVersion.(string),
		mcpData,
		salt,
		uid.String(),
		platform.(string),
	)
	if err != nil {
		c.JSON(http.StatusOK, TransferCheckNumResponse{
			Message: fmt.Sprintf("TransferCheckNum: 获取 CheckNum 时出现问题, 原因是 %v", err),
			Success: false,
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
