package handlers

import (
	"bank-queue-system/services"

	"github.com/gin-gonic/gin"
)

func GetStatistics(c *gin.Context) {
	stats, err := services.GetStatistics()
	if err != nil {
		ErrorResponse(c, err.Error())
		return
	}

	SuccessResponse(c, stats)
}
