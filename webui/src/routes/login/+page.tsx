import { createSignal, Show, JSX } from 'solid-js'
import { useRouter } from '@router'
import { useI18n } from '@i18n'
import { useAuth } from '@contexts/AuthContext'
import * as authService from '@api/services/authService'

export default function LoginPage(): JSX.Element {
  const { t } = useI18n()
  const router = useRouter()
  const auth = useAuth()

  // State for form inputs
  const [username, setUsername] = createSignal('')
  const [password, setPassword] = createSignal('')
  const [otp, setOtp] = createSignal('')

  // State for login flow
  const [loginStep, setLoginStep] = createSignal<'credentials' | 'otp'>('credentials')
  const [temp2FAToken, setTemp2FAToken] = createSignal('')
  const [isLoading, setIsLoading] = createSignal(false)
  const [error, setError] = createSignal<string | null>(null)

  const handleLoginSuccess = (token: string) => {
    auth.setAccessToken(token)
    
    // Redirect to admin dashboard
    const urlParams = new URLSearchParams(window.location.search)
    const redirect = urlParams.get('redirect')
    
    router.navigate(redirect || '/admin/dashboard')
  }

  const handleSubmit = async (e: Event) => {
    e.preventDefault()
    setError(null)
    setIsLoading(true)

    try {
      if (loginStep() === 'credentials') {
        const result = await authService.login({ 
          username: username(), 
          password: password() 
        })
        
        if (result.two_factor_required && result.temp_2fa_token) {
          setTemp2FAToken(result.temp_2fa_token)
          setLoginStep('otp')
        } else if (result.access_token) {
          handleLoginSuccess(result.access_token)
        }
      } else {
        const result = await authService.login2FA({ 
          temp_2fa_token: temp2FAToken(), 
          otp: otp() 
        })
        
        if (result.access_token) {
          handleLoginSuccess(result.access_token)
        }
      }
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || t('login.login_failed'))
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div class="min-h-screen bg-base-200 flex items-center justify-center">
      <div class="card w-96 bg-base-100 shadow-xl">
        <div class="card-body">
          <h2 class="card-title text-center justify-center mb-6">{t('login.title')}</h2>
          
          <form onSubmit={handleSubmit}>
            <Show when={loginStep() === 'credentials'}>
              <div class="form-control w-full mb-4">
                <label class="label"><span class="label-text">{t('login.username')}</span></label>
                <input 
                  type="text" 
                  placeholder={t('login.username_placeholder')} 
                  class="input input-bordered w-full" 
                  value={username()} 
                  onInput={(e) => setUsername(e.currentTarget.value)} 
                  disabled={isLoading()} 
                  required 
                  autocomplete="username" 
                />
              </div>
              <div class="form-control w-full mb-6">
                <label class="label"><span class="label-text">{t('login.password')}</span></label>
                <input 
                  type="password" 
                  placeholder={t('login.password_placeholder')} 
                  class="input input-bordered w-full" 
                  value={password()} 
                  onInput={(e) => setPassword(e.currentTarget.value)} 
                  disabled={isLoading()} 
                  required 
                  autocomplete="current-password" 
                />
              </div>
            </Show>

            <Show when={loginStep() === 'otp'}>
              <div class="form-control w-full mb-6">
                <label class="label"><span class="label-text">6-Digit Authentication Code</span></label>
                <input 
                  type="text" 
                  placeholder="123456" 
                  class="input input-bordered w-full" 
                  value={otp()} 
                  onInput={(e) => setOtp(e.currentTarget.value)} 
                  disabled={isLoading()} 
                  required 
                  maxLength={6} 
                />
              </div>
            </Show>
            
            <Show when={error()}>
              <div class="alert alert-error mb-4">
                <span>{error()}</span>
              </div>
            </Show>
            
            <div class="form-control">
              <button type="submit" class="btn btn-primary" disabled={isLoading()}>
                {isLoading() && <span class="loading loading-spinner"></span>}
                {isLoading() ? t('login.logging_in') : (loginStep() === 'credentials' ? t('login.login_button') : 'Verify')}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}
