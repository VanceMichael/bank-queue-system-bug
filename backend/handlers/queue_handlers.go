package handlers

import (
	"bank-queue-system/models"
	"bank-queue-system/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Message: message,
	})
}

func GetBusinessTypes(c *gin.Context) {
	configs := services.GetBusinessTypeConfigs()
	SuccessResponse(c, configs)
}

type GenerateQueueRequest struct {
	BusinessType models.BusinessType `json:"business_type" binding:"required"`
}

func GenerateQueue(c *gin.Context) {
	var req GenerateQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, "Invalid request body")
		return
	}

	queue, err := services.GenerateQueueNumber(req.BusinessType)
	if err != nil {
		ErrorResponse(c, err.Error())
		return
	}

	SuccessResponse(c, queue)
}

func GetWaitingQueues(c *gin.Context) {
	businessType := c.Query("business_type")
	var queues []*models.QueueNumber
	var err error

	if businessType != "" {
		queues, err = services.GetWaitingQueues(models.BusinessType(businessType))
	} else {
		queues, err = services.GetAllWaitingQueues()
	}

	if err != nil {
		ErrorResponse(c, err.Error())
		return
	}

	SuccessResponse(c, queues)
}

func GetQueueByID(c *gin.Context) {
	id := c.Param("id")
	queue, err := services.GetQueueByID(id)
	if err != nil {
		ErrorResponse(c, "Queue not found")
		return
	}

	SuccessResponse(c, queue)
}

func EstimateWaitTime(c *gin.Context) {
	businessType := c.Query("business_type")
	if businessType == "" {
		ErrorResponse(c, "Business type is required")
		return
	}

	waitTime, err := services.EstimateWaitTime(models.BusinessType(businessType))
	if err != nil {
		ErrorResponse(c, err.Error())
		return
	}

	SuccessResponse(c, gin.H{"wait_time_minutes": waitTime})
}

func GetQueuePosition(c *gin.Context) {
	id := c.Param("id")
	position, err := services.GetQueuePosition(id)
	if err != nil {
		ErrorResponse(c, err.Error())
		return
	}

	SuccessResponse(c, gin.H{"position": position})
}
