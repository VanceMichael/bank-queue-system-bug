<template>
  <div class="statistics">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
              <el-icon :size="32"><User /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.totalCustomers || 0 }}</div>
              <div class="stat-label">今日总客户数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);">
              <el-icon :size="32"><Clock /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.averageWaitTime?.toFixed(1) || 0 }} 分钟</div>
              <div class="stat-label">平均等待时长</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);">
              <el-icon :size="32"><Timer /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ statistics.averageProcessTime?.toFixed(1) || 0 }} 分钟</div>
              <div class="stat-label">平均办理时长</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);">
              <el-icon :size="32"><TrendCharts /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ Object.keys(statistics.windowEfficiency || {}).length }}</div>
              <div class="stat-label">活跃窗口数</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span class="card-title">业务类型分布</span>
          </template>
          <div ref="businessTypeChart" style="height: 350px;"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <span class="card-title">高峰时段分析</span>
          </template>
          <div ref="peakHoursChart" style="height: 350px;"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span class="card-title">窗口效率分析</span>
          </template>
          <div ref="windowEfficiencyChart" style="height: 350px;"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <span class="card-title">业务类型统计详情</span>
          </template>
          <el-table :data="businessTypeStatsList" style="width: 100%">
            <el-table-column prop="type" label="业务类型" width="120">
              <template #default="{ row }">
                <el-tag :type="getBusinessTypeTagType(row.type)" size="large">
                  {{ getBusinessTypeName(row.type) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="count" label="办理数量" width="100">
              <template #default="{ row }">
                <strong>{{ row.count }}</strong>
              </template>
            </el-table-column>
            <el-table-column prop="averageWait" label="平均等待(分钟)">
              <template #default="{ row }">
                {{ row.averageWait?.toFixed(1) || 0 }}
              </template>
            </el-table-column>
            <el-table-column prop="averageProcess" label="平均办理(分钟)">
              <template #default="{ row }">
                {{ row.averageProcess?.toFixed(1) || 0 }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span class="card-title">数据刷新</span>
              <el-button type="primary" @click="fetchStatistics">
                <el-icon><Refresh /></el-icon>
                刷新数据
              </el-button>
            </div>
          </template>
          <div style="text-align: center; padding: 20px; color: #909399;">
            数据每 30 秒自动刷新，最后更新时间：{{ lastUpdateTime }}
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import * as echarts from 'echarts'
import { statisticsApi } from '@/api'
import type { Statistics, BusinessType, BusinessTypeStat } from '@/types'

const statistics = ref<Statistics>({} as Statistics)
const lastUpdateTime = ref('')
const businessTypeChart = ref<HTMLElement>()
const peakHoursChart = ref<HTMLElement>()
const windowEfficiencyChart = ref<HTMLElement>()
let businessTypeChartInstance: echarts.ECharts | null = null
let peakHoursChartInstance: echarts.ECharts | null = null
let windowEfficiencyChartInstance: echarts.ECharts | null = null
let refreshTimer: number | null = null

const businessTypeNames: Record<BusinessType, string> = {
  personal: '个人业务',
  corporate: '对公业务',
  vip: 'VIP业务'
}

const businessTypeTagTypes: Record<BusinessType, string> = {
  personal: 'primary',
  corporate: 'danger',
  vip: 'warning'
}

const getBusinessTypeName = (type: BusinessType) => businessTypeNames[type] || type
const getBusinessTypeTagType = (type: BusinessType) => businessTypeTagTypes[type] || 'info'

const businessTypeStatsList = ref<Array<{ type: BusinessType } & BusinessTypeStat>>([])

const updateBusinessTypeStatsList = () => {
  if (!statistics.value.businessTypeStats) {
    businessTypeStatsList.value = []
    return
  }
  
  businessTypeStatsList.value = Object.entries(statistics.value.businessTypeStats).map(([type, stat]) => ({
    type: type as BusinessType,
    ...stat
  }))
}

const fetchStatistics = async () => {
  try {
    const res = await statisticsApi.getStatistics()
    if (res.data.success) {
      statistics.value = res.data.data
      updateBusinessTypeStatsList()
      lastUpdateTime.value = new Date().toLocaleString('zh-CN')
      
      await nextTick()
      initCharts()
    }
  } catch (error) {
    console.error('Failed to fetch statistics:', error)
  }
}

const initCharts = () => {
  initBusinessTypeChart()
  initPeakHoursChart()
  initWindowEfficiencyChart()
}

const initBusinessTypeChart = () => {
  if (!businessTypeChart.value) return
  
  if (!businessTypeChartInstance) {
    businessTypeChartInstance = echarts.init(businessTypeChart.value)
  }

  const data = Object.entries(statistics.value.businessTypeStats || {}).map(([type, stat]) => ({
    name: getBusinessTypeName(type as BusinessType),
    value: stat.count || 0
  }))

  const option: echarts.EChartsOption = {
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left',
      top: 'center'
    },
    series: [
      {
        name: '业务类型',
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: true,
          formatter: '{b}: {c}'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 16,
            fontWeight: 'bold'
          }
        },
        data: data.length > 0 ? data : [
          { name: '暂无数据', value: 1 }
        ],
        color: ['#409EFF', '#f56c6c', '#e6a23c']
      }
    ]
  }

  businessTypeChartInstance.setOption(option)
}

const initPeakHoursChart = () => {
  if (!peakHoursChart.value) return
  
  if (!peakHoursChartInstance) {
    peakHoursChartInstance = echarts.init(peakHoursChart.value)
  }

  const hours = Array.from({ length: 24 }, (_, i) => `${i}:00`)
  const values = hours.map((_, i) => statistics.value.peakHours?.[i] || 0)

  const option: echarts.EChartsOption = {
    tooltip: {
      trigger: 'axis',
      formatter: '{b}<br/>客户数: {c}'
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: hours,
      axisLabel: {
        interval: 2
      }
    },
    yAxis: {
      type: 'value',
      name: '客户数'
    },
    series: [
      {
        name: '客户数',
        type: 'line',
        smooth: true,
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
            { offset: 1, color: 'rgba(64, 158, 255, 0.05)' }
          ])
        },
        lineStyle: {
          color: '#409EFF',
          width: 2
        },
        itemStyle: {
          color: '#409EFF'
        },
        data: values
      }
    ]
  }

  peakHoursChartInstance.setOption(option)
}

const initWindowEfficiencyChart = () => {
  if (!windowEfficiencyChart.value) return
  
  if (!windowEfficiencyChartInstance) {
    windowEfficiencyChartInstance = echarts.init(windowEfficiencyChart.value)
  }

  const windowNames = Object.keys(statistics.value.windowEfficiency || {})
  const efficiencies = windowNames.map(name => statistics.value.windowEfficiency?.[name] || 0)

  const option: echarts.EChartsOption = {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      },
      formatter: '{b}<br/>平均办理时长: {c} 分钟'
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: windowNames.length > 0 ? windowNames : ['暂无数据'],
      axisLabel: {
        interval: 0,
        rotate: 30
      }
    },
    yAxis: {
      type: 'value',
      name: '平均办理时长(分钟)'
    },
    series: [
      {
        name: '平均办理时长',
        type: 'bar',
        barWidth: '60%',
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#67c23a' },
            { offset: 1, color: '#95d475' }
          ]),
          borderRadius: [4, 4, 0, 0]
        },
        data: efficiencies.length > 0 ? efficiencies : [0]
      }
    ]
  }

  windowEfficiencyChartInstance.setOption(option)
}

const handleResize = () => {
  businessTypeChartInstance?.resize()
  peakHoursChartInstance?.resize()
  windowEfficiencyChartInstance?.resize()
}

onMounted(() => {
  fetchStatistics()
  refreshTimer = window.setInterval(fetchStatistics, 30000)
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  window.removeEventListener('resize', handleResize)
  businessTypeChartInstance?.dispose()
  peakHoursChartInstance?.dispose()
  windowEfficiencyChartInstance?.dispose()
})
</script>

<style scoped>
.statistics {
  height: 100%;
}

.stat-card {
  border: none;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.stat-content {
  display: flex;
  align-items: center;
}

.stat-icon {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.stat-info {
  margin-left: 16px;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
