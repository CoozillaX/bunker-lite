package api

import (
	"bunker-lite/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"bunker-core/mcp"
)

type TransferCheckNumRequest struct {
	Data string `json:"data"`
}

type TransferCheckNumResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Value   string `json:"value"`
}

func TransferCheckNum(w http.ResponseWriter, r *http.Request) {
	// check method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// get session
	session := utils.GetSessionByBearer(r)
	// parse request
	var req TransferCheckNumRequest
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// get engineVersion
	engineVersion, ok := session.Load(session_key_engine_version)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// get patchVersion
	patchVersion, ok := session.Load(session_key_patch_version)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// get platform
	platform, ok := session.Load(session_key_platform)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// parse fb req
	var dataList []any
	decoder := json.NewDecoder(strings.NewReader(req.Data))
	decoder.UseNumber()
	if err := decoder.Decode(&dataList); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(dataList) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	mcpData, ok := dataList[0].(string)
	if !ok || mcpData == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	salt, ok := dataList[1].(string)
	if !ok || salt == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	uid, ok := dataList[2].(json.Number)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
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
		json.NewEncoder(w).Encode(&TransferCheckNumResponse{
			Success: false,
			Message: fmt.Sprintf("获取 CheckNum 失败: %s", err.Error()),
		})
		return
	}
	// return result
	json.NewEncoder(w).Encode(&TransferCheckNumResponse{
		Success: true,
		Message: "ok",
		Value:   result,
	})
}
