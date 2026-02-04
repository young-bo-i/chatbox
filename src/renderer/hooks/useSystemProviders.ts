import { useState, useEffect } from 'react'

const API_BASE_URL = process.env.API_BASE_URL || ''

export interface SystemProviderConfig {
  providerId: string
  name: string
  enabled: boolean
  apiStyle: string
  apiHost?: string
  hasSystemKey: boolean
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
 * 从后端获取系统配置的提供方列表
 * 用于控制哪些提供方对用户可见
 */
export function useSystemProviders() {
  const [enabledProviders, setEnabledProviders] = useState<string[]>([])
  const [systemProviders, setSystemProviders] = useState<SystemProviderConfig[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchProviders = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/api/config/providers`)
        if (response.ok) {
          const data = await response.json()
          const providers = data.providers || []
          setSystemProviders(providers)
          // 获取启用的提供方 ID 列表
          setEnabledProviders(providers.filter((p: SystemProviderConfig) => p.enabled).map((p: SystemProviderConfig) => p.providerId))
        } else {
          // 后端不可用时，不过滤任何提供方
          setEnabledProviders([])
          setError('Failed to load provider configuration')
        }
      } catch (err) {
        console.error('Failed to fetch system providers:', err)
        // 后端不可用时，不过滤任何提供方（显示所有）
        setEnabledProviders([])
        setError('Unable to connect to server')
      } finally {
        setLoading(false)
      }
    }

    fetchProviders()
  }, [])

  return {
    enabledProviders,
    systemProviders,
    loading,
    error,
    // 如果后端不可用或没有配置，则不过滤（显示所有）
    shouldFilter: enabledProviders.length > 0,
  }
}
