import { JSX, Show, createEffect } from 'solid-js'
import { useRouter } from '@router'
import { useSetup } from '@contexts/SetupContext'
import { useI18n } from '@i18n'

interface SetupGuardProps {
  children: JSX.Element
}

/**
 * SetupGuard 组件负责检查系统初始化状态并引导用户到正确的页面
 * - 如果系统需要初始化且用户不在 /setup 页面，重定向到 /setup
 * - 如果系统已初始化且用户在 /setup 页面，重定向到 /login
 */
export function SetupGuard(props: SetupGuardProps): JSX.Element {
  const router = useRouter()
  const setup = useSetup()
  const { t } = useI18n()

  createEffect(() => {
    const setupRequired = setup.setupRequired()
    const isCheckingSetup = setup.isCheckingSetup()
    const currentPath = router.pathname()

    // 等待检查完成
    if (isCheckingSetup || setupRequired === null) {
      return
    }

    console.log('[SetupGuard] Setup required:', setupRequired, 'Current path:', currentPath)

    if (setupRequired) {
      // 系统需要初始化，但用户不在设置页面
      if (currentPath !== '/setup') {
        console.log('[SetupGuard] Redirecting to setup')
        router.navigate('/setup', { replace: true })
      }
    } else {
      // 系统已经初始化，但用户在设置页面
      if (currentPath === '/setup') {
        console.log('[SetupGuard] Setup already completed, redirecting to login')
        router.navigate('/login', { replace: true })
      }
    }
  })

  return (
    <Show 
      when={!setup.isCheckingSetup()} 
      fallback={
        <div class="min-h-screen flex items-center justify-center bg-base-200">
          <div class="text-center">
            <span class="loading loading-spinner loading-lg"></span>
            <div class="mt-4 text-base-content/70">{t('setup_guard.checking_status')}</div>
          </div>
        </div>
      }
    >
      {props.children}
    </Show>
  )
}

export default SetupGuard