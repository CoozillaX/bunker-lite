package api

import (
	"bunker-core/protocol/g79"
	"bunker-lite/utils"
	"encoding/json"
	"fmt"
	"net/http"
)

type TransferStartTypeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func TransferStartType(w http.ResponseWriter, r *http.Request) {
	// check method
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	// get session
	session := utils.GetSessionByBearer(r)
	// get entityID
	entityID, ok := session.Load(session_key_entity_id)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// get start type
	result, err := g79.GetStartType(entityID.(string), r.URL.Query().Get("content"))
	if err != nil {
		json.NewEncoder(w).Encode(&TransferStartTypeResponse{
			Success: false,
			Message: fmt.Sprintf("获取 StartType 失败: %s", err.Error()),
		})
		return
	}
	// return result
	json.NewEncoder(w).Encode(&TransferStartTypeResponse{
		Success: true,
		Message: "ok",
		Data:    result,
	})
}
