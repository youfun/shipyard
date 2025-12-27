import { createContext, useContext, createSignal, createEffect, onMount, JSX } from 'solid-js'
import type { Component } from 'solid-js'
import * as setupService from '@api/services/setupService'

interface SetupContextValue {
  setupRequired: () => boolean | null
  isCheckingSetup: () => boolean
  checkSetupStatus: () => Promise<void>
}

const SetupContext = createContext<SetupContextValue>()

export const SetupProvider: Component<{ children: JSX.Element }> = (props) => {
  const [setupRequired, setSetupRequired] = createSignal<boolean | null>(null)
  const [isCheckingSetup, setIsCheckingSetup] = createSignal(true)

  const checkSetupStatus = async () => {
    try {
      setIsCheckingSetup(true)
      const response = await setupService.checkSetupStatus()
      setSetupRequired(response.setup_required)
    } catch (error) {
      console.error('Failed to check setup status:', error)
      // If API call fails, assume setup is required (for security)
      setSetupRequired(true)
    } finally {
      setIsCheckingSetup(false)
    }
  }

  // 在组件挂载时检查设置状态
  onMount(() => {
    checkSetupStatus()
  })

  const contextValue: SetupContextValue = {
    setupRequired,
    isCheckingSetup,
    checkSetupStatus,
  }

  return (
    <SetupContext.Provider value={contextValue}>
      {props.children}
    </SetupContext.Provider>
  )
}

export const useSetup = () => {
  const context = useContext(SetupContext)
  if (!context) {
    throw new Error('useSetup must be used within a SetupProvider')
  }
  return context
}