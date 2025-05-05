package apiHandlersHealth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
}

func InitializeHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
