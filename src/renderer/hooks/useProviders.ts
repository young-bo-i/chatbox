import { useCallback, useMemo } from 'react'
import { SystemProviders } from 'src/shared/defaults'
import { ModelProviderEnum, type ProviderInfo } from 'src/shared/types'
import { useSettingsStore } from '@/stores/settingsStore'
import useChatboxAIModels from './useChatboxAIModels'
import { useSystemProviders } from './useSystemProviders'

export const useProviders = () => {
  const { chatboxAIModels } = useChatboxAIModels()
  const { systemProviders } = useSystemProviders()
  const { setSettings, ...settings } = useSettingsStore((state) => state)
  const providerSettingsMap = settings.providers

  // 检查 EnterAI 是否在后端配置了
  const enterAISystemConfig = useMemo(
    () => systemProviders.find((p) => p.providerId === 'enter-ai'),
    [systemProviders]
  )

  const allProviderBaseInfos = useMemo(
    () => [...SystemProviders, ...(settings.customProviders || [])],
    [settings.customProviders]
  )
  const providers = useMemo(
    () =>
      allProviderBaseInfos
        .map((p) => {
          const providerSettings = providerSettingsMap?.[p.id]
          if (p.id === ModelProviderEnum.ChatboxAI && settings.licenseKey) {
            return {
              ...p,
              ...providerSettings,
              models: chatboxAIModels,
            }
          } else if (p.id === ModelProviderEnum.EnterAI) {
            // EnterAI: 如果后端配置了（有系统 Key 和 models），或者本地有配置，则包含
            if (enterAISystemConfig?.hasSystemKey || providerSettings?.apiKey || providerSettings?.models?.length) {
              return {
                models: p.defaultSettings?.models,
                ...p,
                ...providerSettings,
              } as ProviderInfo
            }
            return null
          } else if (
            (!p.isCustom && providerSettings?.apiKey) ||
            ((p.isCustom || p.id === ModelProviderEnum.Ollama || p.id === ModelProviderEnum.LMStudio) &&
              providerSettings?.models?.length)
          ) {
            return {
              // 如果没有自定义 models 列表，使用 defaultSettings，否则被自定义的列表（可能有添加或删除部分 model）覆盖, 不能包含用户排除过的 models
              models: p.defaultSettings?.models,
              ...p,
              ...providerSettings,
            } as ProviderInfo
          } else {
            return null
          }
        })
        .filter((p) => !!p),
    [providerSettingsMap, allProviderBaseInfos, chatboxAIModels, settings.licenseKey, enterAISystemConfig]
  )

  const favoritedModels = useMemo(
    () =>
      settings.favoritedModels
        ?.map((m) => {
          const provider = providers.find((p) => p.id === m.provider)
          const model = (provider?.models || provider?.defaultSettings?.models)?.find((mm) => mm.modelId === m.model)

          if (provider && model) {
            return {
              provider,
              model,
            }
          }
        })
        .filter((fm) => !!fm),
    [settings.favoritedModels, providers]
  )

  const favoriteModel = useCallback(
    (provider: string, model: string) => {
      setSettings({
        favoritedModels: [
          ...(settings.favoritedModels || []),
          {
            provider,
            model,
          },
        ],
      })
    },
    [settings, setSettings]
  )

  const unfavoriteModel = useCallback(
    (provider: string, model: string) => {
      setSettings({
        favoritedModels: (settings.favoritedModels || []).filter((m) => m.provider !== provider || m.model !== model),
      })
    },
    [settings, setSettings]
  )

  const isFavoritedModel = useCallback(
    (provider: string, model: string) =>
      !!favoritedModels?.find((m) => m.provider?.id === provider && m.model?.modelId === model),
    [favoritedModels]
  )

  return {
    providers,
    favoritedModels,
    favoriteModel,
    unfavoriteModel,
    isFavoritedModel,
  }
}
