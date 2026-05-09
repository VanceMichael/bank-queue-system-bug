package services

import (
	"bank-queue-system/config"
	"bank-queue-system/models"
	"fmt"
	"strconv"
	"time"
)

func GetStatistics() (*models.Statistics, error) {
	stats := &models.Statistics{
		WindowEfficiency:  make(map[string]float64),
		PeakHours:         make(map[int]int64),
		BusinessTypeStats: make(map[models.BusinessType]models.BusinessTypeStat),
	}

	today := time.Now().Format("2006-01-02")
	totalCustomers, err := config.RedisClient.Get(config.Ctx, fmt.Sprintf("stats:date:%s", today)).Int64()
	if err != nil {
		totalCustomers = 0
	}
	stats.TotalCustomers = totalCustomers

	var totalWaitTime float64
	var totalProcessTime float64
	var waitCount int
	var processCount int

	for _, businessConfig := range businessTypeConfigs {
		waitTimes, err := config.RedisClient.LRange(config.Ctx, fmt.Sprintf("stats:wait_time:%s", businessConfig.Type), 0, -1).Result()
		if err != nil {
			continue
		}

		processTimes, err := config.RedisClient.LRange(config.Ctx, fmt.Sprintf("stats:process_time:%s", businessConfig.Type), 0, -1).Result()
		if err != nil {
			continue
		}

		var btWaitSum float64
		var btProcessSum float64
		var btCount int

		for _, wt := range waitTimes {
			val, _ := strconv.ParseFloat(wt, 64)
			btWaitSum += val
			totalWaitTime += val
			waitCount++
			btCount++
		}

		for _, pt := range processTimes {
			val, _ := strconv.ParseFloat(pt, 64)
			btProcessSum += val
			totalProcessTime += val
			processCount++
		}

		businessTypeCount, _ := config.RedisClient.Get(config.Ctx, fmt.Sprintf("stats:business:%s", businessConfig.Type)).Int64()

		btStat := models.BusinessTypeStat{
			Count:         businessTypeCount,
			AverageWait:   0,
			AverageProcess: 0,
		}

		if btCount > 0 {
			btStat.AverageWait = btWaitSum / float64(btCount)
		}
		if len(processTimes) > 0 {
			btStat.AverageProcess = btProcessSum / float64(len(processTimes))
		}

		stats.BusinessTypeStats[businessConfig.Type] = btStat
	}

	if waitCount > 0 {
		stats.AverageWaitTime = totalWaitTime / float64(waitCount)
	}
	if processCount > 0 {
		stats.AverageProcessTime = totalProcessTime / float64(processCount)
	}

	windows, err := GetAllWindows()
	if err == nil {
		for _, window := range windows {
			windowCount, _ := config.RedisClient.Get(config.Ctx, fmt.Sprintf("stats:window:count:%s", window.ID)).Int64()
			windowProcessTimes, _ := config.RedisClient.LRange(config.Ctx, fmt.Sprintf("stats:window:process:%s", window.ID), 0, -1).Result()

			var windowProcessSum float64
			for _, pt := range windowProcessTimes {
				val, _ := strconv.ParseFloat(pt, 64)
				windowProcessSum += val
			}

			efficiency := 0.0
			if windowCount > 0 {
				efficiency = windowProcessSum / float64(windowCount)
			}
			stats.WindowEfficiency[window.Name] = efficiency
		}
	}

	for hour := 0; hour < 24; hour++ {
		hourCount, _ := config.RedisClient.Get(config.Ctx, fmt.Sprintf("stats:hour:%d", hour)).Int64()
		if hourCount > 0 {
			stats.PeakHours[hour] = hourCount
		}
	}

	return stats, nil
}
