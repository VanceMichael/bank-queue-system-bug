<template>
  <div class="call-control">
    <el-row :gutter="20">
      <el-col :span="8">
        <el-card>
          <template #header>
            <span class="card-title">选择窗口</span>
          </template>
          <div class="window-selector">
            <div
              v-for="window in openWindows"
              :key="window.id"
              class="window-option"
              :class="{ active: selectedWindow?.id === window.id }"
              @click="selectWindow(window)"
            >
              <div class="window-option-header">
                <span class="window-option-name">{{ window.name }}</span>
                <el-tag :type="window.currentQueue ? 'warning' : 'success'" size="small">
                  {{ window.currentQueue ? '办理中' : '空闲' }}
                </el-tag>
              </div>
              <div class="window-option-business">
                <el-tag 
                  v-for="type in window.businessTypes" 
                  :key="type" 
                  size="small"
                  style="margin-right: 4px;"
                >
                  {{ getBusinessTypeName(type) }}
                </el-tag>
              </div>
              <div v-if="window.currentQueue" class="window-option-current">
                当前：<strong>{{ window.currentQueue.number }}</strong>
              </div>
            </div>
            <el-empty v-if="openWindows.length === 0" description="暂无开放窗口" :image-size="60" />
          </div>
        </el-card>
      </el-col>

      <el-col :span="16">
        <el-card v-if="selectedWindow">
          <template #header>
            <div class="card-header">
              <span class="card-title">{{ selectedWindow.name }} - 叫号控制</span>
              <el-tag :type="selectedWindow.currentQueue ? 'warning' : 'success'" size="large">
                {{ selectedWindow.currentQueue ? '办理中' : '空闲' }}
              </el-tag>
            </div>
          </template>

          <div v-if="selectedWindow.currentQueue" class="current-customer">
            <div class="customer-info">
              <div class="customer-number">
                {{ selectedWindow.currentQueue.number }}
              </div>
              <div class="customer-details">
                <div class="detail-item">
                  <span class="label">业务类型：</span>
                  <span class="value">{{ getBusinessTypeName(selectedWindow.currentQueue.businessType) }}</span>
                </div>
                <div class="detail-item">
                  <span class="label">取号时间：</span>
                  <span class="value">{{ formatTime(selectedWindow.currentQueue.createdAt) }}</span>
                </div>
                <div class="detail-item" v-if="selectedWindow.currentQueue.calledAt">
                  <span class="label">呼叫时间：</span>
                  <span class="value">{{ formatTime(selectedWindow.currentQueue.calledAt) }}</span>
                </div>
                <div class="detail-item">
                  <span class="label">状态：</span>
                  <el-tag :type="getStatusType(selectedWindow.currentQueue.status)">
                    {{ getStatusName(selectedWindow.currentQueue.status) }}
                  </el-tag>
                </div>
              </div>
            </div>
            <div class="action-buttons">
              <el-button
                v-if="selectedWindow.currentQueue.status === 'calling'"
                type="success"
                size="large"
                @click="handleStartProcessing"
              >
                <el-icon><Check /></el-icon>
                开始办理
              </el-button>
              <el-button
                v-if="selectedWindow.currentQueue.status === 'processing'"
                type="primary"
                size="large"
                @click="handleComplete"
              >
                <el-icon><Finished /></el-icon>
                完成办理
              </el-button>
              <el-button
                v-if="selectedWindow.currentQueue.status === 'calling'"
                type="danger"
                size="large"
                @click="handleMissed"
              >
                <el-icon><Close /></el-icon>
                过号
              </el-button>
            </div>
          </div>

          <div v-else class="no-customer">
            <el-empty description="当前无客户办理业务" :image-size="80">
              <el-button type="primary" size="large" @click="handleCallNext">
                <el-icon><Microphone /></el-icon>
                呼叫下一位
              </el-button>
            </el-empty>
          </div>
        </el-card>

        <el-card v-else style="margin-top: 20px;">
          <el-empty description="请先选择一个窗口" :image-size="100" />
        </el-card>

        <el-card style="margin-top: 20px;" v-if="selectedWindow">
          <template #header>
            <span class="card-title">等待队列</span>
          </template>
          <el-table :data="windowWaitingQueues" style="width: 100%" v-loading="loading">
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
            <el-table-column label="优先级" width="80">
              <template #default="{ row }">
                <el-tag v-if="row.isVip" type="warning" size="small">VIP</el-tag>
                <span v-else style="color: #909399;">普通</span>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <el-card style="margin-top: 20px;">
          <template #header>
            <span class="card-title">过号重呼</span>
          </template>
          <el-table :data="missedQueues" style="width: 100%" v-loading="loading">
            <el-table-column prop="number" label="排队号" width="120">
              <template #default="{ row }">
                <el-tag type="danger" size="large">{{ row.number }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="businessType" label="业务类型" width="120">
              <template #default="{ row }">
                {{ getBusinessTypeName(row.businessType) }}
              </template>
            </el-table-column>
            <el-table-column prop="createdAt" label="取号时间">
              <template #default="{ row }">
                {{ formatTime(row.createdAt) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100">
              <template #default="{ row }">
                <el-button type="primary" size="small" @click="handleRecall(row.id)">
                  重呼
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="missedQueues.length === 0" description="暂无过号记录" :image-size="60" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { windowApi, queueApi } from '@/api'
import type { Window, QueueNumber, BusinessType, QueueStatus, WindowStatus } from '@/types'

const loading = ref(false)
const windows = ref<Window[]>([])
const selectedWindow = ref<Window | null>(null)
const waitingQueues = ref<QueueNumber[]>([])
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

const getBusinessTypeName = (type: BusinessType) => businessTypeNames[type] || type
const getStatusName = (status: QueueStatus) => statusNames[status] || status
const getStatusType = (status: QueueStatus) => statusTypes[status] || 'info'

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

const openWindows = computed(() => {
  return windows.value.filter(w => w.status === 'open')
})

const windowWaitingQueues = computed(() => {
  if (!selectedWindow.value) return []
  
  return waitingQueues.value.filter(queue => 
    selectedWindow.value!.businessTypes.includes(queue.businessType)
  )
})

const missedQueues = computed(() => {
  return waitingQueues.value.filter(queue => queue.status === 'missed')
})

const fetchWindows = async () => {
  try {
    const res = await windowApi.getAllWindows()
    if (res.data.success) {
      windows.value = res.data.data
      
      if (selectedWindow.value) {
        const updated = windows.value.find(w => w.id === selectedWindow.value!.id)
        if (updated) {
          selectedWindow.value = updated
        }
      }
    }
  } catch (error) {
    console.error('Failed to fetch windows:', error)
  }
}

const fetchWaitingQueues = async () => {
  try {
    const res = await queueApi.getWaitingQueues()
    if (res.data.success) {
      waitingQueues.value = res.data.data
    }
  } catch (error) {
    console.error('Failed to fetch waiting queues:', error)
  }
}

const selectWindow = (window: Window) => {
  selectedWindow.value = window
}

const handleCallNext = async () => {
  if (!selectedWindow.value) return
  
  try {
    const res = await windowApi.callNextQueue(selectedWindow.value.id)
    if (res.data.success) {
      ElMessage.success(`呼叫 ${res.data.data.number} 号`)
      await fetchWindows()
      await fetchWaitingQueues()
    }
  } catch (error: any) {
    if (error.response?.data?.message === 'no waiting queues') {
      ElMessage.info('当前没有等待的客户')
    } else {
      ElMessage.error('呼叫失败')
    }
  }
}

const handleStartProcessing = async () => {
  if (!selectedWindow.value) return
  
  try {
    const res = await windowApi.startProcessing(selectedWindow.value.id)
    if (res.data.success) {
      ElMessage.success('开始办理')
      await fetchWindows()
    }
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const handleComplete = async () => {
  if (!selectedWindow.value) return
  
  try {
    await ElMessageBox.confirm(
      '确认完成当前客户的业务办理？',
      '完成确认',
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'success'
      }
    )
    
    const res = await windowApi.completeQueue(selectedWindow.value.id)
    if (res.data.success) {
      ElMessage.success('业务办理完成')
      await fetchWindows()
      await fetchWaitingQueues()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

const handleMissed = async () => {
  if (!selectedWindow.value) return
  
  try {
    await ElMessageBox.confirm(
      '确认标记当前客户为过号？过号客户可以重新呼叫。',
      '过号确认',
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const res = await windowApi.missedQueue(selectedWindow.value.id)
    if (res.data.success) {
      ElMessage.success('已标记为过号')
      await fetchWindows()
      await fetchWaitingQueues()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

const handleRecall = async (queueId: string) => {
  try {
    const res = await windowApi.recallMissedQueue(queueId)
    if (res.data.success) {
      ElMessage.success('过号客户已重新加入队列，将优先呼叫')
      await fetchWaitingQueues()
    }
  } catch (error) {
    ElMessage.error('重呼失败')
  }
}

onMounted(() => {
  fetchWindows()
  fetchWaitingQueues()
  refreshTimer = window.setInterval(() => {
    fetchWindows()
    fetchWaitingQueues()
  }, 5000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.call-control {
  height: 100%;
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

.window-selector {
  max-height: 500px;
  overflow-y: auto;
}

.window-option {
  padding: 16px;
  border: 2px solid #ebeef5;
  border-radius: 8px;
  margin-bottom: 12px;
  cursor: pointer;
  transition: all 0.3s;
}

.window-option:hover {
  border-color: #409EFF;
}

.window-option.active {
  border-color: #409EFF;
  background-color: #ecf5ff;
}

.window-option-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.window-option-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.window-option-business {
  margin-bottom: 8px;
}

.window-option-current {
  font-size: 14px;
  color: #e6a23c;
}

.current-customer {
  padding: 20px;
}

.customer-info {
  display: flex;
  align-items: flex-start;
  gap: 40px;
}

.customer-number {
  font-size: 72px;
  font-weight: bold;
  color: #409EFF;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
}

.customer-details {
  flex: 1;
}

.detail-item {
  padding: 12px 0;
  border-bottom: 1px solid #ebeef5;
  font-size: 16px;
}

.detail-item:last-child {
  border-bottom: none;
}

.detail-item .label {
  color: #606266;
}

.detail-item .value {
  color: #303133;
  font-weight: 500;
}

.action-buttons {
  display: flex;
  gap: 16px;
  margin-top: 30px;
  justify-content: center;
}

.no-customer {
  padding: 40px;
  text-align: center;
}
</style>
