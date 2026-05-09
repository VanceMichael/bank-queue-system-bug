import axios from 'axios'
import type { BusinessType, QueueNumber, Window, BusinessTypeConfig, Statistics } from '@/types'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

export const queueApi = {
  getBusinessTypes: () => api.get<{ success: boolean; data: BusinessTypeConfig[] }>('/queue/business-types'),
  
  generateQueue: (businessType: BusinessType) => 
    api.post<{ success: boolean; data: QueueNumber }>('/queue/generate', { business_type: businessType }),
  
  getWaitingQueues: (businessType?: BusinessType) => 
    api.get<{ success: boolean; data: QueueNumber[] }>('/queue/waiting', {
      params: businessType ? { business_type: businessType } : {}
    }),
  
  getQueueById: (id: string) => 
    api.get<{ success: boolean; data: QueueNumber }>(`/queue/${id}`),
  
  getQueuePosition: (id: string) => 
    api.get<{ success: boolean; data: { position: number } }>(`/queue/${id}/position`),
  
  estimateWaitTime: (businessType: BusinessType) => 
    api.get<{ success: boolean; data: { wait_time_minutes: number } }>('/queue/wait-time/estimate', {
      params: { business_type: businessType }
    })
}

export const windowApi = {
  createWindow: (name: string, businessTypes: BusinessType[]) => 
    api.post<{ success: boolean; data: Window }>('/window', { name, business_types: businessTypes }),
  
  getAllWindows: () => 
    api.get<{ success: boolean; data: Window[] }>('/window'),
  
  getWindowById: (id: string) => 
    api.get<{ success: boolean; data: Window }>(`/window/${id}`),
  
  updateWindowStatus: (id: string, status: string) => 
    api.put<{ success: boolean }>(`/window/${id}/status`, { status }),
  
  updateWindowBusinessTypes: (id: string, businessTypes: BusinessType[]) => 
    api.put<{ success: boolean }>(`/window/${id}/business-types`, { business_types: businessTypes }),
  
  callNextQueue: (id: string) => 
    api.post<{ success: boolean; data: QueueNumber }>(`/window/${id}/call-next`),
  
  startProcessing: (id: string) => 
    api.post<{ success: boolean }>(`/window/${id}/start-processing`),
  
  completeQueue: (id: string) => 
    api.post<{ success: boolean }>(`/window/${id}/complete`),
  
  missedQueue: (id: string) => 
    api.post<{ success: boolean }>(`/window/${id}/missed`),
  
  recallMissedQueue: (id: string) => 
    api.post<{ success: boolean }>(`/window/recall/${id}`),
  
  vipInsertQueue: (businessType: BusinessType) => 
    api.post<{ success: boolean; data: QueueNumber }>('/window/vip-insert', { business_type: businessType })
}

export const statisticsApi = {
  getStatistics: () => 
    api.get<{ success: boolean; data: Statistics }>('/statistics')
}

export default api
