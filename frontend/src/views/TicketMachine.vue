<template>
  <div class="ticket-machine">
    <el-row :gutter="20">
      <el-col :span="16">
        <el-card>
          <template #header>
            <span class="card-title">取号机</span>
          </template>
          <div class="ticket-machine-content">
            <div class="welcome-text">
              <h2>欢迎光临</h2>
              <p>请选择您需要办理的业务类型</p>
            </div>
            
            <div class="business-types">
              <div
                v-for="config in businessTypes"
                :key="config.type"
                class="business-type-card"
                :class="{ disabled: !config.isActive }"
                @click="handleSelectBusiness(config.type)"
              >
                <div class="business-icon" :class="config.type">
                  <el-icon :size="48">
                    <User v-if="config.type === 'personal'" />
                    <OfficeBuilding v-else-if="config.type === 'corporate'" />
                    <Star v-else />
                  </el-icon>
                </div>
                <div class="business-info">
                  <h3>{{ config.name }}</h3>
                  <p>{{ config.description }}</p>
                  <div class="business-stats">
                    <span>当前等待：<strong>{{ getWaitingCount(config.type) }}</strong> 人</span>
                    <span>预计等待：<strong>{{ getEstimatedWaitTime(config.type) }}</strong> 分钟</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="8">
        <el-card>
          <template #header>
            <span class="card-title">当前排队号</span>
          </template>
          <div class="current-ticket" v-if="currentTicket">
            <div class="ticket-number">
              {{ currentTicket.number }}
            </div>
            <div class="ticket-info">
              <div class="info-item">
                <span class="label">业务类型</span>
                <span class="value">{{ getBusinessTypeName(currentTicket.businessType) }}</span>
              </div>
              <div class="info-item">
                <span class="label">取号时间</span>
                <span class="value">{{ formatTime(currentTicket.createdAt) }}</span>
              </div>
              <div class="info-item">
                <span class="label">当前位置</span>
                <span class="value highlight">{{ queuePosition }} 位</span>
              </div>
              <div class="info-item">
                <span class="label">预计等待</span>
                <span class="value highlight">{{ estimatedWaitTime }} 分钟</span>
              </div>
            </div>
            <el-button type="primary" size="large" @click="printTicket" style="width: 100%; margin-top: 20px;">
              <el-icon><Printer /></el-icon>
              打印排队小票
            </el-button>
          </div>
          <el-empty v-else description="请选择业务类型取号" :image-size="80" />
        </el-card>

        <el-card style="margin-top: 20px;">
          <template #header>
            <span class="card-title">VIP 快速通道</span>
          </template>
          <div class="vip-section">
            <p style="color: #606266; margin-bottom: 16px;">VIP 客户可享受优先叫号服务</p>
            <el-button type="warning" size="large" @click="handleVIPInsert" style="width: 100%;">
              <el-icon><Star /></el-icon>
              VIP 取号
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { queueApi } from '@/api'
import type { BusinessTypeConfig, QueueNumber, BusinessType } from '@/types'

const businessTypes = ref<BusinessTypeConfig[]>([])
const currentTicket = ref<QueueNumber | null>(null)
const queuePosition = ref(0)
const estimatedWaitTime = ref(0)
const waitingCounts = ref<Record<BusinessType, number>>({
  personal: 0,
  corporate: 0,
  vip: 0
})
const estimatedWaitTimes = ref<Record<BusinessType, number>>({
  personal: 0,
  corporate: 0,
  vip: 0
})
let refreshTimer: number | null = null

const businessTypeNames: Record<BusinessType, string> = {
  personal: '个人业务',
  corporate: '对公业务',
  vip: 'VIP业务'
}

const getBusinessTypeName = (type: BusinessType) => businessTypeNames[type] || type
const getWaitingCount = (type: BusinessType) => waitingCounts.value[type] || 0
const getEstimatedWaitTime = (type: BusinessType) => estimatedWaitTimes.value[type] || 0

const formatTime = (time: string | undefined) => {
  if (!time) return '-'
  const date = new Date(time)
  if (isNaN(date.getTime())) return '-'
  return date.toLocaleString('zh-CN')
}

const fetchBusinessTypes = async () => {
  try {
    const res = await queueApi.getBusinessTypes()
    if (res.data.success) {
      businessTypes.value = res.data.data
    }
  } catch (error) {
    console.error('Failed to fetch business types:', error)
  }
}

const fetchWaitingCounts = async () => {
  try {
    const types: BusinessType[] = ['personal', 'corporate', 'vip']
    for (const type of types) {
      const [queuesRes, waitTimeRes] = await Promise.all([
        queueApi.getWaitingQueues(type),
        queueApi.estimateWaitTime(type)
      ])
      
      if (queuesRes.data.success) {
        waitingCounts.value[type] = queuesRes.data.data.length
      }
      if (waitTimeRes.data.success) {
        estimatedWaitTimes.value[type] = waitTimeRes.data.data.wait_time_minutes
      }
    }
  } catch (error) {
    console.error('Failed to fetch waiting counts:', error)
  }
}

const fetchQueuePosition = async () => {
  if (!currentTicket.value) return
  
  try {
    const [positionRes, waitTimeRes] = await Promise.all([
      queueApi.getQueuePosition(currentTicket.value.id),
      queueApi.estimateWaitTime(currentTicket.value.businessType)
    ])
    
    if (positionRes.data.success) {
      queuePosition.value = positionRes.data.data.position
    }
    if (waitTimeRes.data.success) {
      estimatedWaitTime.value = waitTimeRes.data.data.wait_time_minutes
    }
  } catch (error) {
    console.error('Failed to fetch queue position:', error)
  }
}

const handleSelectBusiness = async (type: BusinessType) => {
  try {
    const res = await queueApi.generateQueue(type)
    if (res.data.success) {
      currentTicket.value = res.data.data
      ElMessage.success('取号成功！')
      await fetchQueuePosition()
      await fetchWaitingCounts()
    }
  } catch (error) {
    ElMessage.error('取号失败，请重试')
  }
}

const handleVIPInsert = async () => {
  try {
    await ElMessageBox.confirm(
      '确认使用 VIP 快速通道取号？VIP 客户将享受优先叫号服务。',
      'VIP 取号确认',
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const res = await queueApi.generateQueue('vip')
    if (res.data.success) {
      currentTicket.value = res.data.data
      ElMessage.success('VIP 取号成功！您将享受优先叫号服务。')
      await fetchQueuePosition()
      await fetchWaitingCounts()
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('VIP 取号失败，请重试')
    }
  }
}

const printTicket = () => {
  if (!currentTicket.value) return
  
  const printContent = `
    <div style="text-align: center; font-family: Arial; padding: 20px;">
      <h2 style="margin: 0;">银行排队叫号系统</h2>
      <hr style="margin: 10px 0;">
      <div style="font-size: 48px; font-weight: bold; margin: 20px 0;">
        ${currentTicket.value.number}
      </div>
      <div style="font-size: 14px; color: #666;">
        <p>业务类型：${getBusinessTypeName(currentTicket.value.businessType)}</p>
        <p>取号时间：${formatTime(currentTicket.value.createdAt)}</p>
        <p>当前位置：第 ${queuePosition.value} 位</p>
        <p>预计等待：${estimatedWaitTime.value} 分钟</p>
      </div>
      <hr style="margin: 10px 0;">
      <p style="font-size: 12px; color: #999;">请耐心等待叫号</p>
    </div>
  `
  
  const printWindow = window.open('', '_blank')
  if (printWindow) {
    printWindow.document.write(printContent)
    printWindow.document.close()
    printWindow.print()
  }
}

onMounted(() => {
  fetchBusinessTypes()
  fetchWaitingCounts()
  refreshTimer = window.setInterval(() => {
    fetchWaitingCounts()
    fetchQueuePosition()
  }, 5000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.ticket-machine {
  height: 100%;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.ticket-machine-content {
  padding: 20px;
}

.welcome-text {
  text-align: center;
  margin-bottom: 40px;
}

.welcome-text h2 {
  font-size: 28px;
  color: #303133;
  margin-bottom: 10px;
}

.welcome-text p {
  font-size: 16px;
  color: #606266;
}

.business-types {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
}

.business-type-card {
  display: flex;
  align-items: center;
  padding: 24px;
  border: 2px solid #ebeef5;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s;
}

.business-type-card:hover {
  border-color: #409EFF;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.2);
  transform: translateY(-2px);
}

.business-type-card.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.business-icon {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 20px;
  color: white;
}

.business-icon.personal {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.business-icon.corporate {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.business-icon.vip {
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
}

.business-info h3 {
  font-size: 20px;
  color: #303133;
  margin-bottom: 8px;
}

.business-info p {
  font-size: 14px;
  color: #606266;
  margin-bottom: 12px;
}

.business-stats {
  display: flex;
  gap: 20px;
  font-size: 14px;
  color: #909399;
}

.business-stats strong {
  color: #409EFF;
  font-size: 16px;
}

.current-ticket {
  text-align: center;
}

.ticket-number {
  font-size: 64px;
  font-weight: bold;
  color: #409EFF;
  margin: 20px 0;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
}

.ticket-info {
  text-align: left;
  padding: 0 20px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #ebeef5;
}

.info-item:last-child {
  border-bottom: none;
}

.info-item .label {
  color: #606266;
}

.info-item .value {
  color: #303133;
  font-weight: 500;
}

.info-item .value.highlight {
  color: #409EFF;
  font-size: 18px;
}

.vip-section {
  text-align: center;
}
</style>
