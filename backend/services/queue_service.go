package services

import (
	"bank-queue-system/config"
	"bank-queue-system/models"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

var businessTypeConfigs = []models.BusinessTypeConfig{
	{
		Type:        models.BusinessTypePersonal,
		Name:        "个人业务",
		Prefix:      "A",
		AverageTime: 10,
		Description: "个人存取款、转账等业务",
		IsActive:    true,
	},
	{
		Type:        models.BusinessTypeCorporate,
		Name:        "对公业务",
		Prefix:      "B",
		AverageTime: 20,
		Description: "企业开户、对公转账等业务",
		IsActive:    true,
	},
	{
		Type:        models.BusinessTypeVIP,
		Name:        "VIP业务",
		Prefix:      "V",
		AverageTime: 15,
		Description: "VIP客户专属服务",
		IsActive:    true,
	},
}

func GetBusinessTypeConfigs() []models.BusinessTypeConfig {
	return businessTypeConfigs
}

func GetBusinessTypeConfig(businessType models.BusinessType) *models.BusinessTypeConfig {
	for _, config := range businessTypeConfigs {
		if config.Type == businessType {
			return &config
		}
	}
	return nil
}

func GenerateQueueNumber(businessType models.BusinessType) (*models.QueueNumber, error) {
	businessConfig := GetBusinessTypeConfig(businessType)
	if businessConfig == nil {
		return nil, fmt.Errorf("invalid business type: %s", businessType)
	}

	counterKey := fmt.Sprintf("queue:counter:%s", businessType)
	counter, err := config.RedisClient.Incr(config.Ctx, counterKey).Result()
	if err != nil {
		return nil, err
	}

	number := fmt.Sprintf("%s%04d", businessConfig.Prefix, counter)

	isVIP := businessType == models.BusinessTypeVIP
	priority := 0
	if isVIP {
		priority = 100
	}

	queueNumber := &models.QueueNumber{
		ID:           uuid.New().String(),
		Number:       number,
		BusinessType: businessType,
		Status:       models.QueueStatusWaiting,
		CreatedAt:    time.Now(),
		IsVIP:        isVIP,
		Priority:     priority,
	}

	queueJSON, err := json.Marshal(queueNumber)
	if err != nil {
		return nil, err
	}

	queueKey := fmt.Sprintf("queue:%s", queueNumber.ID)
	err = config.RedisClient.Set(config.Ctx, queueKey, queueJSON, 24*time.Hour).Err()
	if err != nil {
		return nil, err
	}

	waitingKey := fmt.Sprintf("queue:waiting:%s", businessType)
	err = config.RedisClient.ZAdd(config.Ctx, waitingKey, &redis.Z{
		Score:  float64(int(queueNumber.CreatedAt.Unix()) - queueNumber.Priority*1000000),
		Member: queueNumber.ID,
	}).Err()
	if err != nil {
		return nil, err
	}

	return queueNumber, nil
}

func GetWaitingQueues(businessType models.BusinessType) ([]*models.QueueNumber, error) {
	waitingKey := fmt.Sprintf("queue:waiting:%s", businessType)
	queueIDs, err := config.RedisClient.ZRange(config.Ctx, waitingKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var queues []*models.QueueNumber
	for _, id := range queueIDs {
		queueKey := fmt.Sprintf("queue:%s", id)
		queueJSON, err := config.RedisClient.Get(config.Ctx, queueKey).Result()
		if err != nil {
			continue
		}

		var queue models.QueueNumber
		err = json.Unmarshal([]byte(queueJSON), &queue)
		if err != nil {
			continue
		}

		if queue.Status == models.QueueStatusWaiting {
			queues = append(queues, &queue)
		}
	}

	sort.Slice(queues, func(i, j int) bool {
		if queues[i].Priority != queues[j].Priority {
			return queues[i].Priority > queues[j].Priority
		}
		return queues[i].CreatedAt.Before(queues[j].CreatedAt)
	})

	return queues, nil
}

func GetAllWaitingQueues() ([]*models.QueueNumber, error) {
	var allQueues []*models.QueueNumber

	for _, businessConfig := range businessTypeConfigs {
		queues, err := GetWaitingQueues(businessConfig.Type)
		if err != nil {
			continue
		}
		allQueues = append(allQueues, queues...)
	}

	sort.Slice(allQueues, func(i, j int) bool {
		if allQueues[i].Priority != allQueues[j].Priority {
			return allQueues[i].Priority > allQueues[j].Priority
		}
		return allQueues[i].CreatedAt.Before(allQueues[j].CreatedAt)
	})

	return allQueues, nil
}

func GetQueueByID(id string) (*models.QueueNumber, error) {
	queueKey := fmt.Sprintf("queue:%s", id)
	queueJSON, err := config.RedisClient.Get(config.Ctx, queueKey).Result()
	if err != nil {
		return nil, err
	}

	var queue models.QueueNumber
	err = json.Unmarshal([]byte(queueJSON), &queue)
	if err != nil {
		return nil, err
	}

	return &queue, nil
}

func UpdateQueueStatus(queue *models.QueueNumber) error {
	queueJSON, err := json.Marshal(queue)
	if err != nil {
		return err
	}

	queueKey := fmt.Sprintf("queue:%s", queue.ID)
	return config.RedisClient.Set(config.Ctx, queueKey, queueJSON, 24*time.Hour).Err()
}

func EstimateWaitTime(businessType models.BusinessType) (int, error) {
	queues, err := GetWaitingQueues(businessType)
	if err != nil {
		return 0, err
	}

	businessConfig := GetBusinessTypeConfig(businessType)
	if businessConfig == nil {
		return 0, fmt.Errorf("invalid business type")
	}

	return len(queues) * businessConfig.AverageTime, nil
}

func GetQueuePosition(queueID string) (int, error) {
	queue, err := GetQueueByID(queueID)
	if err != nil {
		return 0, err
	}

	queues, err := GetWaitingQueues(queue.BusinessType)
	if err != nil {
		return 0, err
	}

	for i, q := range queues {
		if q.ID == queueID {
			return i + 1, nil
		}
	}

	return 0, nil
}
