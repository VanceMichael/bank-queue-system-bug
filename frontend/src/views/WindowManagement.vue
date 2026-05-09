<template>
  <div class="window-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span class="card-title">窗口管理</span>
          <el-button type="primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>
            新增窗口
          </el-button>
        </div>
      </template>
      
      <el-table :data="windows" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="窗口名称" width="150">
          <template #default="{ row }">
            <span class="window-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getWindowStatusType(row.status)" size="large">
              {{ getWindowStatusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="businessTypes" label="办理业务类型">
          <template #default="{ row }">
            <el-tag 
              v-for="type in row.businessTypes" 
              :key="type" 
              :type="getBusinessTypeTagType(type)"
              size="small"
              style="margin-right: 8px;"
            >
              {{ getBusinessTypeName(type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="currentQueue" label="当前办理" width="150">
          <template #default="{ row }">
            <el-tag v-if="row.currentQueue" type="warning" size="small">
              {{ row.currentQueue.number }}
            </el-tag>
            <span v-else style="color: #909399;">空闲</span>
          </template>
        </el-table-column>
        <el-table-column prop="lastActiveAt" label="最后活跃时间" width="180">
          <template #default="{ row }">
            {{ row.lastActiveAt ? formatTime(row.lastActiveAt) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button 
                v-if="row.status === 'closed'"
                type="success" 
                size="small" 
                @click="handleUpdateStatus(row.id, 'open')"
              >
                开放
              </el-button>
              <el-button 
                v-else-if="row.status === 'open'"
                type="warning" 
                size="small" 
                :disabled="!!row.currentQueue"
                @click="handleUpdateStatus(row.id, 'paused')"
                :title="row.currentQueue ? '窗口有客户正在办理，无法暂停' : ''"
              >
                暂停
              </el-button>
              <el-button 
                v-else
                type="success" 
                size="small" 
                @click="handleUpdateStatus(row.id, 'open')"
              >
                恢复
              </el-button>
              <el-button 
                v-if="row.status !== 'closed'"
                type="info" 
                size="small" 
                :disabled="!!row.currentQueue"
                @click="handleUpdateStatus(row.id, 'closed')"
                :title="row.currentQueue ? '窗口有客户正在办理，无法关闭' : ''"
              >
                关闭
              </el-button>
            </el-button-group>
            <el-button type="primary" size="small" @click="handleEditWindow(row)" style="margin-left: 8px;">
              编辑
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog
      v-model="showCreateDialog"
      :title="editingWindow ? '编辑窗口' : '新增窗口'"
      width="500px"
    >
      <el-form :model="windowForm" :rules="windowRules" ref="windowFormRef" label-width="100px">
        <el-form-item label="窗口名称" prop="name">
          <el-input v-model="windowForm.name" placeholder="请输入窗口名称" />
        </el-form-item>
        <el-form-item label="业务类型" prop="businessTypes">
          <el-checkbox-group v-model="windowForm.businessTypes">
            <el-checkbox 
              v-for="config in businessTypes" 
              :key="config.type" 
              :label="config.type"
            >
              {{ config.name }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitWindow">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { windowApi, queueApi } from '@/api'
import type { Window, BusinessTypeConfig, BusinessType, WindowStatus } from '@/types'

const loading = ref(false)
const windows = ref<Window[]>([])
const businessTypes = ref<BusinessTypeConfig[]>([])
const showCreateDialog = ref(false)
const editingWindow = ref<Window | null>(null)
const windowFormRef = ref<FormInstance>()

const windowForm = reactive({
  name: '',
  businessTypes: [] as BusinessType[]
})

const windowRules: FormRules = {
  name: [
    { required: true, message: '请输入窗口名称', trigger: 'blur' }
  ],
  businessTypes: [
    { required: true, message: '请选择至少一种业务类型', trigger: 'change' }
  ]
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

const getWindowStatusName = (status: WindowStatus) => windowStatusNames[status] || status
const getWindowStatusType = (status: WindowStatus) => windowStatusTypes[status] || 'info'
const getBusinessTypeName = (type: BusinessType) => businessTypeNames[type] || type
const getBusinessTypeTagType = (type: BusinessType) => businessTypeTagTypes[type] || 'info'

const formatTime = (time: string | undefined) => {
  if (!time) return '-'
  const date = new Date(time)
  if (isNaN(date.getTime())) return '-'
  return date.toLocaleString('zh-CN')
}

const fetchWindows = async () => {
  loading.value = true
  try {
    const res = await windowApi.getAllWindows()
    if (res.data.success) {
      windows.value = res.data.data
    }
  } catch (error) {
    ElMessage.error('获取窗口列表失败')
  } finally {
    loading.value = false
  }
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

const handleUpdateStatus = async (id: string, status: WindowStatus) => {
  try {
    const res = await windowApi.updateWindowStatus(id, status)
    if (res.data.success) {
      ElMessage.success(`窗口${getWindowStatusName(status)}成功`)
      fetchWindows()
    }
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const handleEditWindow = (window: Window) => {
  editingWindow.value = window
  windowForm.name = window.name
  windowForm.businessTypes = [...window.businessTypes]
  showCreateDialog.value = true
}

const handleSubmitWindow = async () => {
  if (!windowFormRef.value) return
  
  await windowFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (editingWindow.value) {
          const res = await windowApi.updateWindowBusinessTypes(
            editingWindow.value.id,
            windowForm.businessTypes
          )
          if (res.data.success) {
            ElMessage.success('窗口更新成功')
          }
        } else {
          const res = await windowApi.createWindow(
            windowForm.name,
            windowForm.businessTypes
          )
          if (res.data.success) {
            ElMessage.success('窗口创建成功')
          }
        }
        showCreateDialog.value = false
        resetForm()
        fetchWindows()
      } catch (error) {
        ElMessage.error(editingWindow.value ? '更新失败' : '创建失败')
      }
    }
  })
}

const resetForm = () => {
  editingWindow.value = null
  windowForm.name = ''
  windowForm.businessTypes = []
  windowFormRef.value?.resetFields()
}

onMounted(() => {
  fetchWindows()
  fetchBusinessTypes()
})
</script>

<style scoped>
.window-management {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.window-name {
  font-weight: 600;
  color: #303133;
}
</style>
