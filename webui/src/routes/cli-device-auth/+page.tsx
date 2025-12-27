import { createSignal, Show, JSX, createEffect } from 'solid-js'
import { useI18n } from '../../i18n'
import { useCLI } from '../../api/hooks'
import { useAuth } from '../../contexts/AuthContext'
import { useRouter } from '../../router'

export default function CLIDeviceAuthPage(): JSX.Element {
  const { t } = useI18n()
  const { queries, mutations } = useCLI()
  const auth = useAuth()
  const router = useRouter()
  
  const [success, setSuccess] = createSignal(false)
  const [error, setError] = createSignal('')

  // Use router state for reactive URL params
  const sessionId = () => new URLSearchParams(router.search()).get('session_id')

  // Check authentication and redirect if needed
  createEffect(() => {
    if (!auth.isLoading() && !auth.isAuthenticated()) {
      // Redirect to login with current URL for redirect back
      const currentUrl = router.pathname() + router.search()
      const redirectUrl = `/login?redirect=${encodeURIComponent(currentUrl)}`
      router.navigate(redirectUrl)
    }
  })

  // Get device context - only if authenticated
  const deviceQuery = queries.getDeviceContext(() => 
    auth.isAuthenticated() ? sessionId() : null
  )
  const deviceContext = () => deviceQuery.data
  const loading = () => auth.isLoading() || deviceQuery.isPending

  const handleConfirm = (approved: boolean) => {
    const context = deviceContext()
    if (!context) return

    setError('')
    mutations.confirmAuth.mutate(
      { session_id: context.session_id, approved },
      {
        onSuccess: () => {
          setSuccess(true)
          setTimeout(() => window.close(), 3000)
        },
        onError: (err: any) => {
          setError(err.response?.data?.error || err.message)
        },
      }
    )
  }

  // Format timestamp
  const formatTimestamp = (timestamp: number) => {
    const date = new Date(timestamp * 1000)
    return date.toLocaleString(navigator.language, {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      timeZoneName: 'short'
    })
  }

  // Format OS display
  const formatOS = (os: string) => {
    const osMap: Record<string, string> = {
      'windows': '💻 Windows',
      'linux': '🐧 Linux',
      'darwin': '🍎 macOS',
      'freebsd': '👹 FreeBSD'
    }
    return osMap[os.toLowerCase()] || `💻 ${os}`
  }

  return (
    <div class="min-h-screen bg-base-200 flex items-center justify-center p-4">
      <div class="card w-full max-w-lg bg-base-100 shadow-xl">
        <div class="card-body">
          <Show when={loading()}>
            <div class="text-center space-y-4">
              <span class="loading loading-spinner loading-lg"></span>
              <p>{t('common.loading')}</p>
            </div>
          </Show>

          <Show when={!loading() && (deviceQuery.error || error())}>
            <div class="text-center space-y-4">
              <div class="text-error text-6xl">❌</div>
              <h3 class="text-lg font-semibold text-error">{t('common.error') || 'Error'}</h3>
              <p class="text-base-content/70">{(deviceQuery.error as any)?.message || error()}</p>
              <button
                class="btn btn-ghost"
                onClick={() => window.close()}
              >
                {t('cli_device_auth.cancel_close')}
              </button>
            </div>
          </Show>

          <Show when={!loading() && !(deviceQuery.error || error()) && deviceContext() && !success()}>
            <div class="space-y-6">
              <div class="text-center">
                <h2 class="text-2xl font-bold mb-2">{t('cli_device_auth.title')}</h2>
                <p class="text-lg text-primary mb-4">{t('cli_device_auth.subtitle')}</p>
                <p class="text-base-content/70">{t('cli_device_auth.confirm_info')}</p>
              </div>

              <div class="bg-base-200 p-6 rounded-lg space-y-4">
                <div class="flex items-center justify-between">
                  <span class="font-semibold">{t('cli_device_auth.device_system')}:</span>
                  <span class="text-lg">{formatOS(deviceContext()?.os || '')}</span>
                </div>

                <div class="flex items-center justify-between">
                  <span class="font-semibold">{t('cli_device_auth.device_name')}:</span>
                  <span class="font-mono bg-base-300 px-2 py-1 rounded">🆔 {deviceContext()?.device_name}</span>
                </div>

                <div class="flex items-center justify-between">
                  <span class="font-semibold">{t('cli_device_auth.ip_address')}:</span>
                  <div class="text-right">
                    <div>🌍 {deviceContext()?.public_ip}</div>
                  </div>
                </div>

                <div class="flex items-center justify-between">
                  <span class="font-semibold">{t('cli_device_auth.request_time')}:</span>
                  <span class="text-sm">🕒 {formatTimestamp(deviceContext()?.request_timestamp || 0)}</span>
                </div>

                <div class="flex items-center justify-between">
                  <span class="font-semibold">{t('cli_device_auth.authorize_app')}:</span>
                  <span class="font-semibold text-primary">🔒 {t('cli_device_auth.app_name')}</span>
                </div>
              </div>

              <Show when={error()}>
                <div class="alert alert-error">
                  <span>{error()}</span>
                </div>
              </Show>

              <div class="flex gap-3">
                <button
                  class="btn btn-error flex-1"
                  disabled={mutations.confirmAuth.isPending}
                  onClick={() => handleConfirm(false)}
                >
                  {t('cli_device_auth.deny_button')}
                </button>
                <button
                  class="btn btn-success flex-1"
                  disabled={mutations.confirmAuth.isPending}
                  onClick={() => handleConfirm(true)}
                >
                  {mutations.confirmAuth.isPending && <span class="loading loading-spinner loading-sm"></span>}
                  {mutations.confirmAuth.isPending ? t('cli_device_auth.authorizing') : t('cli_device_auth.confirm_button')}
                </button>
              </div>
            </div>
          </Show>

          <Show when={success()}>
            <div class="text-center space-y-4">
              <div class="text-success text-6xl">✅</div>
              <h3 class="text-lg font-semibold text-success">{t('cli_device_auth.success_title')}</h3>
              <p class="text-base-content/70">
                {t('cli_device_auth.success_message')}
              </p>
              <p class="text-sm text-base-content/50">
                {t('cli_device_auth.auto_close_message')}
              </p>
            </div>
          </Show>

          <Show when={!success() && !loading() && !(deviceQuery.error || error())}>
            <div class="divider">Or</div>
            
            <div class="text-center">
              <button
                class="btn btn-ghost btn-sm"
                onClick={() => window.close()}
              >
                {t('cli_device_auth.cancel_close')}
              </button>
            </div>
          </Show>
        </div>
      </div>
    </div>
  )
}
