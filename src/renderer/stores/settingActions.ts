import { getDefaultStore } from 'jotai'
import { ModelProviderEnum } from 'src/shared/types'
import * as atoms from './atoms'
import { settingsStore } from './settingsStore'

const API_BASE_URL = process.env.API_BASE_URL || ''

export function needEditSetting() {
  const settings = settingsStore.getState()

  // 激活了chatbox ai
  if (settings.licenseKey) {
    return false
  }

  if (settings.providers && Object.keys(settings.providers).length > 0) {
    const providers = settings.providers
    const keys = Object.keys(settings.providers)
    // 有任何一个供应商配置了api key
    if (keys.filter((key) => !!providers[key].apiKey).length > 0) {
      return false
    }
    // EnterAI 配置了模型（可能是从后端同步的）
    if (providers[ModelProviderEnum.EnterAI]?.models?.length) {
      return false
    }
    // Ollama / LMStudio/ custom provider 配置了至少一个模型
    if (
      keys.filter(
        (key) =>
          (key === ModelProviderEnum.Ollama ||
            key === ModelProviderEnum.LMStudio ||
            key.startsWith('custom-provider')) &&
          providers[key].models?.length
      ).length > 0
    ) {
      return false
    }
  }
  return true
}

/**
 * 异步检查是否需要编辑设置
 * 会先检查后端是否配置了 EnterAI
 */
export async function needEditSettingAsync(): Promise<boolean> {
  // 先检查本地配置
  if (!needEditSetting()) {
    return false
  }

  // 检查后端是否配置了 EnterAI
  try {
    const response = await fetch(`${API_BASE_URL}/api/config/providers`)
    if (response.ok) {
      const data = await response.json()
      const enterAIConfig = data.providers?.find(
        (p: any) => p.providerId === 'enter-ai' || p.name === 'EnterAI'
      )
      // 如果后端配置了 EnterAI（有系统 Key 和 models）
      if (enterAIConfig?.hasSystemKey && enterAIConfig?.models?.length > 0) {
        return false
      }
    }
  } catch (error) {
    console.error('Failed to check server config:', error)
  }

  return true
}

export function getLanguage() {
  return settingsStore.getState().language
}

export function getProxy() {
  return settingsStore.getState().proxy
}

export function getLicenseKey() {
  return settingsStore.getState().licenseKey
}

export function getLicenseDetail() {
  return settingsStore.getState().licenseDetail
}

export function isPaid() {
  return !!getLicenseKey()
}

export function isPro() {
  return !!getLicenseKey() && !getLicenseDetail()?.name.toLowerCase().includes('lite')
}

export function getRemoteConfig() {
  const store = getDefaultStore()
  return store.get(atoms.remoteConfigAtom)
}

export function getAutoGenerateTitle() {
  return settingsStore.getState().autoGenerateTitle
}

export function getExtensionSettings() {
  return settingsStore.getState().extension
}
