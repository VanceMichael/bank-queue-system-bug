<template>
  <div class="dashboard">
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
              <el-icon :size="32"><Grid /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ windows.length }}</div>
              <div class="stat-label">开放窗口数</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="16">
        <el-card>
          <template #header>
            <span class="card-title">实时排队情况</span>
          </template>
          <el-table :data="waitingQueues" style="width: 100%" v-loading="loading">
            <el-table-column prop="number" label="排队号" width="120">
              <template #default="{ row }">
                <el-tag :type="row.isVip ? 'warning' : 'primary'" size="large">
                  {{ row.number }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="businessType" label="业务类型" width="120">
              <template #default="{ row }">
                {{ getBusinessTypeName(row.businessType) }}
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">
                  {{ getStatusName(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="createdAt" label="取号时间">
              <template #default="{ row }">
                {{ formatTime(row.createdAt) }}
              </template>
            </el-table-column>
            <el-table-column prop="waitTime" label="已等待" width="100">
              <template #default="{ row }">
                {{ calculateWaitTime(row.createdAt) }} 分钟
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <template #header>
            <span class="card-title">窗口状态</span>
          </template>
          <div class="window-list">
            <div v-for="window in windows" :key="window.id" class="window-item">
              <div class="window-header">
                <span class="window-name">{{ window.name }}</span>
                <el-tag :type="getWindowStatusType(window.status)" size="small">
                  {{ getWindowStatusName(window.status) }}
                </el-tag>
              </div>
              <div class="window-business">
                <span>办理业务：</span>
                <el-tag v-for="type in window.businessTypes" :key="type" size="small" style="margin-right: 4px;">
                  {{ getBusinessTypeName(type) }}
                </el-tag>
              </div>
              <div v-if="window.currentQueue" class="window-current">
                <span>当前办理：</span>
                <el-tag type="warning" size="small">{{ window.currentQueue.number }}</el-tag>
              </div>
            </div>
            <el-empty v-if="windows.length === 0" description="暂无窗口" :image-size="60" />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { queueApi, windowApi, statisticsApi } from '@/api'
import type { QueueNumber, Window, Statistics, BusinessType, QueueStatus, WindowStatus } from '@/types'

const loading = ref(false)
const waitingQueues = ref<QueueNumber[]>([])
const windows = ref<Window[]>([])
const statistics = ref<Statistics>({} as Statistics)
let refreshTimer: number | null = null

const businessTypeNames: Record<BusinessType, string> = {
  personal: '个人业务',
  corporate: '对公业务',
  vip: 'VIP业务'
}

const statusNames: Record<QueueStatus, string> = {
  waiting: '等待中',
  calling: '呼叫中',
  processing: '办理中',
  completed: '已完成',
  missed: '已过号'
}

const statusTypes: Record<QueueStatus, string> = {
  waiting: 'primary',
  calling: 'warning',
  processing: 'success',
  completed: 'info',
  missed: 'danger'
}

const windowStatusNames: Record<WindowStatus, string> = {
  open: '开放',
  closed: '关闭',
  paused: '暂停'
}

const windowStatusTypes: Record<WindowStatus, string> = {
  open: 'success',
  closed: 'info',
  paused: 'warning'
}

const getBusinessTypeName = (type: BusinessType) => businessTypeNames[type] || type
const getStatusName = (status: QueueStatus) => statusNames[status] || status
const getStatusType = (status: QueueStatus) => statusTypes[status] || 'info'
const getWindowStatusName = (status: WindowStatus) => windowStatusNames[status] || status
const getWindowStatusType = (status: WindowStatus) => windowStatusTypes[status] || 'info'

const formatTime = (time: string | undefined) => {
  if (!time) return '-'
  const date = new Date(time)
  if (isNaN(date.getTime())) return '-'
  return date.toLocaleString('zh-CN')
}

const calculateWaitTime = (createdAt: string | undefined) => {
  if (!createdAt) return 0
  const now = new Date()
  const created = new Date(createdAt)
  if (isNaN(created.getTime())) return 0
  const diff = Math.floor((now.getTime() - created.getTime()) / 60000)
  return diff
}

const fetchData = async () => {
  try {
    const [queuesRes, windowsRes, statsRes] = await Promise.all([
      queueApi.getWaitingQueues(),
      windowApi.getAllWindows(),
      statisticsApi.getStatistics()
    ])

    if (queuesRes.data.success) {
      waitingQueues.value = queuesRes.data.data
    }
    if (windowsRes.data.success) {
      windows.value = windowsRes.data.data.filter(w => w.status === 'open')
    }
    if (statsRes.data.success) {
      statistics.value = statsRes.data.data
    }
  } catch (error) {
    console.error('Failed to fetch data:', error)
  }
}

onMounted(() => {
  fetchData()
  refreshTimer = window.setInterval(fetchData, 5000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.dashboard {
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

.window-list {
  max-height: 400px;
  overflow-y: auto;
}

.window-item {
  padding: 16px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  margin-bottom: 12px;
}

.window-item:last-child {
  margin-bottom: 0;
}

.window-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.window-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.window-business,
.window-current {
  font-size: 14px;
  color: #606266;
  margin-top: 6px;
}
</style>
