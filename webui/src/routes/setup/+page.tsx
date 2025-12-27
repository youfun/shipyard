import { createSignal, JSX } from 'solid-js'
import { useRouter } from '@router'
import { useSetup } from '@contexts/SetupContext'
import { useI18n } from '@i18n'
import * as setupService from '@api/services/setupService'

export default function SetupPage(): JSX.Element {
  const router = useRouter()
  const setup = useSetup()
  const { t } = useI18n()
  
  const [username, setUsername] = createSignal('')
  const [password, setPassword] = createSignal('')
  const [confirmPassword, setConfirmPassword] = createSignal('')
  const [loading, setLoading] = createSignal(false)
  const [error, setError] = createSignal('')

  const handleSubmit = async (e: Event) => {
    e.preventDefault()
    
    if (!username().trim() || !password().trim()) {
      setError(t('setup.error_empty_fields'))
      return
    }
    
    if (password() !== confirmPassword()) {
      setError(t('setup.error_password_mismatch'))
      return
    }
    
    if (password().length < 6) {
      setError(t('setup.error_password_length'))
      return
    }

    setLoading(true)
    setError('')

    try {
      const result = await setupService.setup({
        username: username().trim(),
        password: password()
      })
      
      if (result.success) {
        // Refresh setup status, SetupGuard will automatically redirect to login page
        await setup.checkSetupStatus()
      } else {
        setError(result.message || t('setup.error_setup_failed'))
      }
    } catch (err: any) {
      setError(err.response?.data?.message || t('setup.error_network'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div class="min-h-screen bg-base-200 flex items-center justify-center">
      <div class="card w-96 bg-base-100 shadow-xl">
        <div class="card-body">
          <h2 class="card-title text-center justify-center mb-6">{t('setup.title')}</h2>
          <p class="text-center text-base-content/70 mb-4">
            {t('setup.subtitle')}
          </p>
          
          <form onSubmit={handleSubmit}>
            <div class="form-control w-full mb-4">
              <label class="label">
                <span class="label-text">{t('setup.admin_username')}</span>
              </label>
              <input
                type="text"
                placeholder={t('setup.username_placeholder')}
                class="input input-bordered w-full"
                value={username()}
                onInput={(e) => setUsername(e.currentTarget.value)}
                disabled={loading()}
                required
              />
            </div>
            
            <div class="form-control w-full mb-4">
              <label class="label">
                <span class="label-text">{t('setup.password')}</span>
              </label>
              <input
                type="password"
                placeholder={t('setup.password_placeholder')}
                class="input input-bordered w-full"
                value={password()}
                onInput={(e) => setPassword(e.currentTarget.value)}
                disabled={loading()}
                required
              />
            </div>
            
            <div class="form-control w-full mb-6">
              <label class="label">
                <span class="label-text">{t('setup.confirm_password')}</span>
              </label>
              <input
                type="password"
                placeholder={t('setup.confirm_password_placeholder')}
                class="input input-bordered w-full"
                value={confirmPassword()}
                onInput={(e) => setConfirmPassword(e.currentTarget.value)}
                disabled={loading()}
                required
              />
            </div>
            
            {error() && (
              <div class="alert alert-error mb-4">
                <span>{error()}</span>
              </div>
            )}
            
            <div class="form-control">
              <button
                type="submit"
                class="btn btn-primary"
                disabled={loading()}
              >
                {loading() && <span class="loading loading-spinner"></span>}
                {loading() ? t('setup.setting_up') : t('setup.setup_button')}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}
