import { computed, ref } from 'vue'
import { messagesEn } from './messages-en'
import { messagesZh } from './messages-zh'

export type Locale = 'en' | 'zh-CN'

const STORAGE_KEY = 'frp-webui-frps-locale'

const messages = {
  en: messagesEn,
  'zh-CN': messagesZh,
} as const

interface MessageTree {
  [key: string]: string | string[] | MessageTree
}

const detectLocale = (): Locale => {
  if (typeof window === 'undefined') {
    return 'en'
  }

  const stored = window.localStorage.getItem(STORAGE_KEY)
  if (stored === 'en' || stored === 'zh-CN') {
    return stored
  }

  const browserLocale = window.navigator.language.toLowerCase()
  return browserLocale.startsWith('zh') ? 'zh-CN' : 'en'
}

const currentLocale = ref<Locale>(detectLocale())

const syncLocale = (value: Locale) => {
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(STORAGE_KEY, value)
    document.documentElement.lang = value
  }
}

syncLocale(currentLocale.value)

const getByPath = (
  tree: MessageTree,
  path: string,
): string | string[] | undefined => {
  const segments = path.split('.')
  let current: string | string[] | MessageTree | undefined = tree
  for (const segment of segments) {
    if (!current || typeof current === 'string' || Array.isArray(current)) {
      return undefined
    }
    current = current[segment]
  }
  return current as string | string[] | undefined
}

const interpolate = (
  text: string,
  params: Record<string, string | number> = {},
) => text.replace(/\{(\w+)\}/g, (_match, key) => String(params[key] ?? `{${key}}`))

export const translate = (
  key: string,
  params: Record<string, string | number> = {},
) => {
  const message =
    getByPath(messages[currentLocale.value] as unknown as MessageTree, key) ??
    getByPath(messages.en as unknown as MessageTree, key)
  if (typeof message !== 'string') {
    return key
  }
  return interpolate(message, params)
}

export const translateList = (key: string): string[] => {
  const message =
    getByPath(messages[currentLocale.value] as unknown as MessageTree, key) ??
    getByPath(messages.en as unknown as MessageTree, key)
  return Array.isArray(message) ? message.map(String) : []
}

export const setLocale = (value: Locale) => {
  currentLocale.value = value
  syncLocale(value)
}

export const getCurrentLocale = (): Locale => currentLocale.value

export const useI18n = () => {
  const locale = computed(() => currentLocale.value)
  return {
    locale,
    setLocale,
    t: translate,
    tList: translateList,
  }
}
