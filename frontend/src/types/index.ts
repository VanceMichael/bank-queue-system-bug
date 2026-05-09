export type BusinessType = 'personal' | 'corporate' | 'vip'

export type QueueStatus = 'waiting' | 'calling' | 'processing' | 'completed' | 'missed'

export type WindowStatus = 'open' | 'closed' | 'paused'

export interface BusinessTypeConfig {
  type: BusinessType
  name: string
  prefix: string
  averageTime: number
  description: string
  isActive: boolean
}

export interface QueueNumber {
  id: string
  number: string
  businessType: BusinessType
  status: QueueStatus
  windowId?: string
  createdAt: string
  calledAt?: string
  completedAt?: string
  isVip: boolean
  priority: number
}

export interface Window {
  id: string
  name: string
  status: WindowStatus
  businessTypes: BusinessType[]
  currentQueue?: QueueNumber
  lastActiveAt?: string
}

export interface BusinessTypeStat {
  count: number
  averageWait: number
  averageProcess: number
}

export interface Statistics {
  totalCustomers: number
  averageWaitTime: number
  averageProcessTime: number
  windowEfficiency: Record<string, number>
  peakHours: Record<number, number>
  businessTypeStats: Record<BusinessType, BusinessTypeStat>
}
