import { createContext, useContext, createSignal, createResource, createMemo, Component, JSX, Show } from 'solid-js'
import * as i18n from '@solid-primitives/i18n'
import type { Dictionary } from './zh'

export type Locale = 'zh' | 'en'

// Dictionary loading functions
async function fetchDictionary(locale: Locale): Promise<i18n.Flatten<Dictionary>> {
  const dict = locale === 'zh' 
    ? (await import('./zh')).dict
    : (await import('./en')).dict
  return i18n.flatten(dict as Dictionary)
}

// Context type
interface I18nContextValue {
  locale: () => Locale
  setLocale: (locale: Locale) => void
  t: i18n.Translator<i18n.Flatten<Dictionary>>
  isLoading: () => boolean
}

// Create context
const I18nContext = createContext<I18nContextValue>()

// Provider component
interface I18nProviderProps {
  children: JSX.Element
}

export const I18nProvider: Component<I18nProviderProps> = (props) => {
  // Get initial locale from localStorage or default to Chinese
  const getInitialLocale = (): Locale => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem('locale')
      if (stored === 'zh' || stored === 'en') {
        return stored
      }
    }
    return 'zh'
  }

  const [locale, setLocaleInternal] = createSignal<Locale>(getInitialLocale())

  // Load dictionary based on current locale
  const [dict] = createResource(locale, fetchDictionary, {
    initialValue: {} as i18n.Flatten<Dictionary>
  })

  // Save locale to localStorage when it changes
  const setLocale = (newLocale: Locale) => {
    setLocaleInternal(newLocale)
    if (typeof window !== 'undefined') {
      localStorage.setItem('locale', newLocale)
    }
  }

  // Create translator
  const t = createMemo(() => i18n.translator(() => dict() || {}, i18n.resolveTemplate))

  // Check if dictionary is loading
  const isLoading = () => dict.loading

  const contextValue: I18nContextValue = {
    locale,
    setLocale,
    t: t(),
    isLoading
  }

  return (
    <I18nContext.Provider value={contextValue}>
      <Show when={!isLoading()} fallback={
        <div class="min-h-screen bg-base-200 flex items-center justify-center">
          <span class="loading loading-spinner loading-lg"></span>
        </div>
      }>
        {props.children}
      </Show>
    </I18nContext.Provider>
  )
}

// Hook to use i18n context
export const useI18n = () => {
  const context = useContext(I18nContext)
  if (!context) {
    throw new Error('useI18n must be used within an I18nProvider')
  }
  return context
}
