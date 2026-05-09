package services

import (
	"bank-queue-system/config"
	"bank-queue-system/models"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

func CreateWindow(name string, businessTypes []models.BusinessType) (*models.Window, error) {
	window := &models.Window{
		ID:            uuid.New().String(),
		Name:          name,
		Status:        models.WindowStatusClosed,
		BusinessTypes: businessTypes,
	}

	windowJSON, err := json.Marshal(window)
	if err != nil {
		return nil, err
	}

	windowKey := fmt.Sprintf("window:%s", window.ID)
	err = config.RedisClient.Set(config.Ctx, windowKey, windowJSON, 0).Err()
	if err != nil {
		return nil, err
	}

	err = config.RedisClient.SAdd(config.Ctx, "windows", window.ID).Err()
	if err != nil {
		return nil, err
	}

	return window, nil
}

func GetWindowByID(id string) (*models.Window, error) {
	windowKey := fmt.Sprintf("window:%s", id)
	windowJSON, err := config.RedisClient.Get(config.Ctx, windowKey).Result()
	if err != nil {
		return nil, err
	}

	var window models.Window
	err = json.Unmarshal([]byte(windowJSON), &window)
	if err != nil {
		return nil, err
	}

	return &window, nil
}

func GetAllWindows() ([]*models.Window, error) {
	windowIDs, err := config.RedisClient.SMembers(config.Ctx, "windows").Result()
	if err != nil {
		return nil, err
	}

	var windows []*models.Window
	for _, id := range windowIDs {
		window, err := GetWindowByID(id)
		if err != nil {
			continue
		}
		windows = append(windows, window)
	}

	return windows, nil
}

func UpdateWindow(window *models.Window) error {
	windowJSON, err := json.Marshal(window)
	if err != nil {
		return err
	}

	windowKey := fmt.Sprintf("window:%s", window.ID)
	return config.RedisClient.Set(config.Ctx, windowKey, windowJSON, 0).Err()
}

func UpdateWindowStatus(windowID string, status models.WindowStatus) error {
	window, err := GetWindowByID(windowID)
	if err != nil {
		return err
	}

	if (status == models.WindowStatusPaused || status == models.WindowStatusClosed) && window.CurrentQueue != nil {
		return fmt.Errorf("window has customer in process, cannot pause or close")
	}

	window.Status = status
	now := time.Now()
	window.LastActiveAt = &now

	return UpdateWindow(window)
}

func UpdateWindowBusinessTypes(windowID string, businessTypes []models.BusinessType) error {
	window, err := GetWindowByID(windowID)
	if err != nil {
		return err
	}

	window.BusinessTypes = businessTypes
	return UpdateWindow(window)
}

func CallNextQueue(windowID string) (*models.QueueNumber, error) {
	window, err := GetWindowByID(windowID)
	if err != nil {
		return nil, err
	}

	if window.Status != models.WindowStatusOpen {
		return nil, fmt.Errorf("window is not open")
	}

	if window.CurrentQueue != nil {
		return nil, fmt.Errorf("window is already processing a queue")
	}

	var nextQueue *models.QueueNumber

	for _, businessType := range window.BusinessTypes {
		waitingKey := fmt.Sprintf("queue:waiting:%s", businessType)
		queueIDs, err := config.RedisClient.ZRange(config.Ctx, waitingKey, 0, -1).Result()
		if err != nil {
			continue
		}

		for _, id := range queueIDs {
			queue, err := GetQueueByID(id)
			if err != nil {
				continue
			}

			if queue.Status == models.QueueStatusWaiting {
				nextQueue = queue
				break
			}
		}

		if nextQueue != nil {
			break
		}
	}

	if nextQueue == nil {
		return nil, fmt.Errorf("no waiting queues")
	}

	now := time.Now()
	nextQueue.Status = models.QueueStatusCalling
	nextQueue.WindowID = windowID
	nextQueue.CalledAt = &now

	err = UpdateQueueStatus(nextQueue)
	if err != nil {
		return nil, err
	}

	waitingKey := fmt.Sprintf("queue:waiting:%s", nextQueue.BusinessType)
	err = config.RedisClient.ZRem(config.Ctx, waitingKey, nextQueue.ID).Err()
	if err != nil {
		return nil, err
	}

	window.CurrentQueue = nextQueue
	window.LastActiveAt = &now
	err = UpdateWindow(window)
	if err != nil {
		return nil, err
	}

	return nextQueue, nil
}

func StartProcessing(windowID string) error {
	window, err := GetWindowByID(windowID)
	if err != nil {
		return err
	}

	if window.CurrentQueue == nil {
		return fmt.Errorf("no queue is being called")
	}

	queue, err := GetQueueByID(window.CurrentQueue.ID)
	if err != nil {
		return err
	}

	queue.Status = models.QueueStatusProcessing
	err = UpdateQueueStatus(queue)
	if err != nil {
		return err
	}

	now := time.Now()
	window.CurrentQueue = queue
	window.LastActiveAt = &now
	return UpdateWindow(window)
}

func CompleteQueue(windowID string) error {
	window, err := GetWindowByID(windowID)
	if err != nil {
		return err
	}

	if window.CurrentQueue == nil {
		return fmt.Errorf("no queue is being processed")
	}

	queue, err := GetQueueByID(window.CurrentQueue.ID)
	if err != nil {
		return err
	}

	now := time.Now()
	queue.Status = models.QueueStatusCompleted
	queue.CompletedAt = &now
	err = UpdateQueueStatus(queue)
	if err != nil {
		return err
	}

	window.CurrentQueue = nil
	window.LastActiveAt = &now
	return UpdateWindow(window)
}

func MissedQueue(windowID string) error {
	window, err := GetWindowByID(windowID)
	if err != nil {
		return err
	}

	if window.CurrentQueue == nil {
		return fmt.Errorf("no queue is being called")
	}

	queue, err := GetQueueByID(window.CurrentQueue.ID)
	if err != nil {
		return err
	}

	queue.Status = models.QueueStatusMissed
	err = UpdateQueueStatus(queue)
	if err != nil {
		return err
	}

	window.CurrentQueue = nil
	now := time.Now()
	window.LastActiveAt = &now
	return UpdateWindow(window)
}

func RecallMissedQueue(queueID string) error {
	queue, err := GetQueueByID(queueID)
	if err != nil {
		return err
	}

	if queue.Status != models.QueueStatusMissed {
		return fmt.Errorf("queue is not missed")
	}

	queue.Status = models.QueueStatusWaiting
	err = UpdateQueueStatus(queue)
	if err != nil {
		return err
	}

	waitingKey := fmt.Sprintf("queue:waiting:%s", queue.BusinessType)
	return config.RedisClient.ZAdd(config.Ctx, waitingKey, &redis.Z{
		Score:  float64(queue.Priority*1000000 + int(queue.CreatedAt.Unix())),
		Member: queue.ID,
	}).Err()
}

func VIPInsertQueue(businessType models.BusinessType) (*models.QueueNumber, error) {
	queue, err := GenerateQueueNumber(businessType)
	if err != nil {
		return nil, err
	}

	queue.Priority = 200
	err = UpdateQueueStatus(queue)
	if err != nil {
		return nil, err
	}

	waitingKey := fmt.Sprintf("queue:waiting:%s", businessType)
	err = config.RedisClient.ZAdd(config.Ctx, waitingKey, &redis.Z{
		Score:  float64(queue.Priority*1000000 + int(queue.CreatedAt.Unix())),
		Member: queue.ID,
	}).Err()
	if err != nil {
		return nil, err
	}

	return queue, nil
}

func RecordQueueCompletion(queue *models.QueueNumber) error {
	dateKey := fmt.Sprintf("stats:date:%s", time.Now().Format("2006-01-02"))
	err := config.RedisClient.Incr(config.Ctx, dateKey).Err()
	if err != nil {
		return err
	}

	hourKey := fmt.Sprintf("stats:hour:%d", time.Now().Hour())
	err = config.RedisClient.Incr(config.Ctx, hourKey).Err()
	if err != nil {
		return err
	}

	businessTypeKey := fmt.Sprintf("stats:business:%s", queue.BusinessType)
	err = config.RedisClient.Incr(config.Ctx, businessTypeKey).Err()
	if err != nil {
		return err
	}

	if queue.CalledAt != nil && queue.CompletedAt != nil {
		waitTime := queue.CalledAt.Sub(queue.CreatedAt).Minutes()
		processTime := queue.CompletedAt.Sub(*queue.CalledAt).Minutes()

		waitTimeKey := fmt.Sprintf("stats:wait_time:%s", queue.BusinessType)
		err = config.RedisClient.LPush(config.Ctx, waitTimeKey, waitTime).Err()
		if err != nil {
			return err
		}

		processTimeKey := fmt.Sprintf("stats:process_time:%s", queue.BusinessType)
		err = config.RedisClient.LPush(config.Ctx, processTimeKey, processTime).Err()
		if err != nil {
			return err
		}

		if queue.WindowID != "" {
			windowWaitKey := fmt.Sprintf("stats:window:wait:%s", queue.WindowID)
			err = config.RedisClient.LPush(config.Ctx, windowWaitKey, waitTime).Err()
			if err != nil {
				return err
			}

			windowProcessKey := fmt.Sprintf("stats:window:process:%s", queue.WindowID)
			err = config.RedisClient.LPush(config.Ctx, windowProcessKey, processTime).Err()
			if err != nil {
				return err
			}

			windowCountKey := fmt.Sprintf("stats:window:count:%s", queue.WindowID)
			err = config.RedisClient.Incr(config.Ctx, windowCountKey).Err()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
