package handlers

import (
	"bank-queue-system/models"
	"bank-queue-system/services"

	"github.com/gin-gonic/gin"
)

type CreateWindowRequest struct {
	Name          string                `json:"name" binding:"required"`
	BusinessTypes []models.BusinessType `json:"business_types" binding:"required"`
}

func CreateWindow(c *gin.Context) {
	var req CreateWindowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, "Invalid request body")
		return
	}

	window, err := services.CreateWindow(req.Name, req.BusinessTypes)
	if err != nil {
		ErrorResponse(c, err.Error())
		return
	}

	SuccessResponse(c, window)
}

func GetAllWindows(c *gin.Context) {
	windows, err := services.GetAllWindows()
	if err != nil {
		ErrorResponse(c, err.Error())
		return
	}

	SuccessResponse(c, windows)
}

func GetWindowByID(c *gin.Context) {
	id := c.Param("id")
	window, err := services.GetWindowByID(id)
	if err != nil {
		ErrorResponse(c, "Window not found")
		return
	}

	SuccessResponse(c, window)
}

type UpdateWindowStatusRequest struct {
	Status models.WindowStatus `json:"status" binding:"required"`
}

func UpdateWindowStatus(c *gin.Context) {
	id := c.Param("id")
	var req UpdateWindowStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, "Invalid request body")
		return
	}

	err := services.UpdateWindowStatus(id, req.Status)
	if err != nil {
		ErrorResponse(c, err.Error())
		return
	}

	SuccessResponse(c, gin.H{"message": "Window status updated successfully"})
}

type UpdateWindowBusinessTypesRequest struct {
	BusinessTypes []models.BusinessType `json:"business_types" binding:"required"`
}

func UpdateWindowBusinessTypes(c *gin.Context) {
	id := c.Param("id")
	var req UpdateWindowBusinessTypesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, "Invalid request body")
		return
	}

	err := services.UpdateWindowBusinessTypes(id, req.BusinessTypes)
	if err != nil {
		ErrorResponse(c, err.Error())
		return
	}

	SuccessResponse(c, gin.H{"message": "Window business types updated successfully"})
}

func CallNextQueue(c *gin.Context) {
	id := c.Param("id")
	queue, err := services.CallNextQueue(id)
	if err != nil {
		ErrorResponse(c, err.Error())
		return
	}

	SuccessResponse(c, queue)
}

func StartProcessing(c *gin.Context) {
	id := c.Param("id")
	err := services.StartProcessing(id)
	if err != nil {
		ErrorResponse(c, err.Error())
		return
	}

	SuccessResponse(c, gin.H{"message": "Started processing"})
}

func CompleteQueue(c *gin.Context) {
	id := c.Param("id")
	err := services.CompleteQueue(id)
	if err != nil {
		ErrorResponse(c, err.Error())
		return
	}

	SuccessResponse(c, gin.H{"message": "Queue completed"})
}

func MissedQueue(c *gin.Context) {
	id := c.Param("id")
	err := services.MissedQueue(id)
	if err != nil {
		ErrorResponse(c, err.Error())
		return
	}

	SuccessResponse(c, gin.H{"message": "Queue marked as missed"})
}

func RecallMissedQueue(c *gin.Context) {
	id := c.Param("id")
	err := services.RecallMissedQueue(id)
	if err != nil {
		ErrorResponse(c, err.Error())
		return
	}

	SuccessResponse(c, gin.H{"message": "Queue recalled successfully"})
}

type VIPInsertQueueRequest struct {
	BusinessType models.BusinessType `json:"business_type" binding:"required"`
}

func VIPInsertQueue(c *gin.Context) {
	var req VIPInsertQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, "Invalid request body")
		return
	}

	queue, err := services.VIPInsertQueue(req.BusinessType)
	if err != nil {
		ErrorResponse(c, err.Error())
		return
	}

	SuccessResponse(c, queue)
}
