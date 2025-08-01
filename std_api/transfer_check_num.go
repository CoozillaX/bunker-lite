package std_api

import (
	"bunker-lite/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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
	// parse fb req
	var dataList []any
	if err := json.Unmarshal([]byte(req.Data), &dataList); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(dataList) != 3 {
		w.WriteHeader(http.StatusBadRequest)
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
