package std_api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func New(c *gin.Context) {
	c.Data(
		http.StatusOK,
		"text/plain",
		[]byte(uuid.NewString()),
	)
}
