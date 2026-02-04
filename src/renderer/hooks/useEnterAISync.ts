import { useEffect, useRef } from 'react'
import { ModelProviderEnum, type ProviderModelInfo } from 'src/shared/types'
import { settingsStore } from '@/stores/settingsStore'
import { useAuthStore } from '@/stores/authStore'

const API_BASE_URL = process.env.API_BASE_URL || ''

interface SystemProviderConfig {
  id: number
  providerId: string
  name: string
  apiStyle: string
  apiHost: string
  apiKey?: string
  hasSystemKey?: boolean
  enabled: boolean
  allowCustomKey: boolean
  models?: Array<{
    modelId: string
    nickname?: string
    capabilities?: string[]
    contextWindow?: number
    maxOutput?: number
    type?: string
  }>
}

/**
 * 同步 EnterAI 配置从后端到本地存储
 * - 应用启动时自动同步
 * - 管理员保存配置后触发同步
 */
export function useEnterAISync() {
  const { isAdmin, token } = useAuthStore()
  const syncedRef = useRef(false)

  useEffect(() => {
    // 只同步一次
    if (syncedRef.current) return

    const syncEnterAIConfig = async () => {
      try {
        // 获取后端配置
        const endpoint = isAdmin && token
          ? `${API_BASE_URL}/api/admin/providers`
          : `${API_BASE_URL}/api/config/providers`
        
        const headers: Record<string, string> = {}
        if (isAdmin && token) {
          headers['Authorization'] = `Bearer ${token}`
        }
        
        const response = await fetch(endpoint, { headers })
        if (!response.ok) return

        const data = await response.json()
        const enterAIConfig: SystemProviderConfig = data.providers?.find(
          (p: any) => p.providerId === 'enter-ai' || p.name === 'EnterAI'
        )

        if (!enterAIConfig) return

        // 将后端配置同步到本地存储
        const models: ProviderModelInfo[] = enterAIConfig.models?.map((m) => ({
          modelId: m.modelId,
          nickname: m.nickname,
          capabilities: m.capabilities as any,
          contextWindow: m.contextWindow,
          maxOutput: m.maxOutput,
          type: m.type as any,
        })) || []

        // 更新本地设置
        const currentSettings = settingsStore.getState()
        const currentProviderSettings = currentSettings.providers?.[ModelProviderEnum.EnterAI] || {}
        
        settingsStore.setState({
          providers: {
            ...currentSettings.providers,
            [ModelProviderEnum.EnterAI]: {
              ...currentProviderSettings,
              apiHost: enterAIConfig.apiHost || currentProviderSettings.apiHost || '',
              // 管理员可以获取到 API Key，普通用户使用系统配置（不需要本地存储 key）
              apiKey: isAdmin ? (enterAIConfig.apiKey || '') : currentProviderSettings.apiKey,
              models: models.length > 0 ? models : currentProviderSettings.models,
            },
          },
        })

        syncedRef.current = true
        console.log('EnterAI config synced from server')
      } catch (error) {
        console.error('Failed to sync EnterAI config:', error)
      }
    }

    // 延迟执行，确保其他初始化完成
    const timer = setTimeout(syncEnterAIConfig, 500)
    return () => clearTimeout(timer)
  }, [isAdmin, token])

  // 手动触发同步
  const syncNow = async () => {
    syncedRef.current = false
    // 触发重新同步
  }

  return { syncNow }
}

/**
 * 全局同步函数，供管理员保存后调用
 */
export async function syncEnterAIFromServer(token?: string | null, isAdmin?: boolean) {
  try {
    const endpoint = isAdmin && token
      ? `${API_BASE_URL}/api/admin/providers`
      : `${API_BASE_URL}/api/config/providers`
    
    const headers: Record<string, string> = {}
    if (isAdmin && token) {
      headers['Authorization'] = `Bearer ${token}`
    }
    
    const response = await fetch(endpoint, { headers })
    if (!response.ok) return false

    const data = await response.json()
    const enterAIConfig: SystemProviderConfig = data.providers?.find(
      (p: any) => p.providerId === 'enter-ai' || p.name === 'EnterAI'
    )

    if (!enterAIConfig) return false

    const models: ProviderModelInfo[] = enterAIConfig.models?.map((m) => ({
      modelId: m.modelId,
      nickname: m.nickname,
      capabilities: m.capabilities as any,
      contextWindow: m.contextWindow,
      maxOutput: m.maxOutput,
      type: m.type as any,
    })) || []

    const currentSettings = settingsStore.getState()
    const currentProviderSettings = currentSettings.providers?.[ModelProviderEnum.EnterAI] || {}
    
    settingsStore.setState({
      providers: {
        ...currentSettings.providers,
        [ModelProviderEnum.EnterAI]: {
          ...currentProviderSettings,
          apiHost: enterAIConfig.apiHost || '',
          apiKey: isAdmin ? (enterAIConfig.apiKey || '') : currentProviderSettings.apiKey,
          models: models.length > 0 ? models : currentProviderSettings.models,
        },
      },
    })

    return true
  } catch (error) {
    console.error('Failed to sync EnterAI config:', error)
    return false
  }
}
