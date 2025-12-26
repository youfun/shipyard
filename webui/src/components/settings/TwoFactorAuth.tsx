import { createSignal, Show, For, JSX } from 'solid-js'
import { toast } from 'solid-toast'
import { useAuth } from '@api/hooks'
import { QRCodeSVG } from 'solid-qr-code'
import type { Setup2FAResponse } from '@types'

export function TwoFactorAuth(): JSX.Element {
  const { queries, mutations } = useAuth()

  // 2FA State
  const [setupInfo, setSetupInfo] = createSignal<Setup2FAResponse | null>(null)
  const [otp, setOtp] = createSignal('')
  const [recoveryCodes, setRecoveryCodes] = createSignal<string[]>([])
  
  // Disable 2FA State
  const [disablePassword, setDisablePassword] = createSignal('')
  const [disableOtp, setDisableOtp] = createSignal('')

  // Current User Data
  const userQuery = queries.getCurrentUser()
  const is2FAEnabled = () => userQuery.data?.two_factor_enabled

  // --- Handlers ---

  const handleSetup2FA = () => {
    mutations.setup2FA.mutate(undefined, {
      onSuccess: (data) => {
        setSetupInfo(data)
        const modal = document.getElementById('enable_2fa_modal') as HTMLDialogElement
        modal?.showModal()
      },
      onError: (error: any) => {
        toast.error(error.response?.data?.error || error.message || 'Failed to setup 2FA')
      }
    })
  }

  const handleEnable2FA = (e: Event) => {
    e.preventDefault()
    const info = setupInfo()
    if (!info || !otp()) return

    mutations.enable2FA.mutate(
      { secret: info.secret, otp: otp() },
      {
        onSuccess: (data) => {
          const modal = document.getElementById('enable_2fa_modal') as HTMLDialogElement
          modal?.close()
          setRecoveryCodes(data.recovery_codes)
          setSetupInfo(null)
          setOtp('')
          toast.success('2FA enabled successfully')
        },
        onError: (error: any) => {
          toast.error(error.response?.data?.error || error.message || 'Failed to verify 2FA')
        }
      }
    )
  }

  const handleOpenDisableModal = () => {
    setDisablePassword('')
    setDisableOtp('')
    const modal = document.getElementById('disable_2fa_modal') as HTMLDialogElement
    modal?.showModal()
  }

  const handleDisable2FA = (e: Event) => {
    e.preventDefault()
    if (!disablePassword() || !disableOtp()) return

    mutations.disable2FA.mutate(
      { password: disablePassword(), otp: disableOtp() },
      {
        onSuccess: () => {
          const modal = document.getElementById('disable_2fa_modal') as HTMLDialogElement
          modal?.close()
          toast.success('2FA disabled successfully')
        },
        onError: (error: any) => {
          toast.error(error.response?.data?.error || error.message || 'Failed to disable 2FA')
        }
      }
    )
  }

  const handlePrintRecoveryCodes = () => {
    const codes = recoveryCodes().join('\n')
    const blob = new Blob([codes], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'shipyard-recovery-codes.txt'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  return (
    <>
      <div class="space-y-6">
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h2 class="card-title mb-4">Two-Factor Authentication</h2>
            
            <Show when={userQuery.isLoading}>
              <div class="flex justify-center py-4">
                <span class="loading loading-spinner"></span>
              </div>
            </Show>

            <Show when={!userQuery.isLoading}>
              <Show when={is2FAEnabled()}>
                <div class="alert alert-success mb-4">
                  <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                  <span>Status: Enabled</span>
                </div>
                <p class="text-sm mb-4">Two-Factor Authentication is currently active on your account.</p>
                <button 
                  onClick={handleOpenDisableModal} 
                  class="btn btn-error btn-outline"
                  disabled={mutations.disable2FA.isPending}
                >
                  Disable 2FA
                </button>
              </Show>

              <Show when={!is2FAEnabled()}>
                <div class="alert alert-warning mb-4">
                  <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
                  <span>Status: Not Enabled</span>
                </div>
                <p class="text-sm mb-4">Enhance your account security by enabling 2FA.</p>
                <button 
                  onClick={handleSetup2FA} 
                  class="btn btn-primary"
                  disabled={mutations.setup2FA.isPending}
                >
                  {mutations.setup2FA.isPending && <span class="loading loading-spinner"></span>}
                  Enable 2FA
                </button>
              </Show>
            </Show>
          </div>
        </div>

        {/* Recovery Codes Display */}
        <Show when={recoveryCodes().length > 0}>
          <div class="card bg-success text-success-content shadow">
            <div class="card-body">
              <h3 class="card-title">2FA Enabled Successfully!</h3>
              <p class="font-bold">Please save these recovery codes in a safe place. You will not be shown them again.</p>
              <div class="grid grid-cols-2 gap-2 my-4 font-mono bg-base-100 text-base-content p-4 rounded-lg">
                <For each={recoveryCodes()}>
                  {(code) => <div class="text-center">{code}</div>}
                </For>
              </div>
              <button onClick={handlePrintRecoveryCodes} class="btn btn-sm btn-ghost border-current">
                Download Codes
              </button>
            </div>
          </div>
        </Show>
      </div>

      {/* Enable 2FA Modal */}
      <dialog id="enable_2fa_modal" class="modal">
        <div class="modal-box">
          <form method="dialog">
            <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">✕</button>
          </form>
          <h3 class="font-bold text-lg">Enable Two-Factor Authentication</h3>
          
          <Show when={setupInfo()}>
            <div class="mt-6 space-y-6">
              <div>
                <p class="font-semibold mb-2">Step 1: Scan QR Code</p>
                <p class="text-sm text-base-content/70 mb-4">Scan this image with your authenticator app (e.g., Google Authenticator).</p>
                <div class="flex justify-center bg-white p-4 rounded-lg border">
                  <QRCodeSVG value={setupInfo()!.qr_code_url} size={200} />
                </div>
                <div class="mt-4 text-center">
                  <p class="text-xs text-base-content/50 mb-1">Or manually enter this secret:</p>
                  <code class="bg-base-200 px-2 py-1 rounded font-mono text-sm select-all">{setupInfo()!.secret}</code>
                </div>
              </div>
              
              <div class="divider"></div>
              
              <form onSubmit={handleEnable2FA}>
                <p class="font-semibold mb-2">Step 2: Verify Code</p>
                <p class="text-sm text-base-content/70 mb-4">Enter the 6-digit code from your app to complete setup.</p>
                <div class="form-control">
                  <input 
                    type="text" 
                    value={otp()} 
                    onInput={(e) => setOtp(e.currentTarget.value)} 
                    class="input input-bordered w-full text-center tracking-widest text-lg font-mono" 
                    placeholder="000000" 
                    maxLength={6} 
                    required 
                  />
                </div>
                <button 
                  type="submit" 
                  disabled={mutations.enable2FA.isPending || otp().length !== 6} 
                  class="btn btn-primary w-full mt-4"
                >
                  {mutations.enable2FA.isPending ? 'Verifying...' : 'Verify & Enable'}
                </button>
              </form>
            </div>
          </Show>
        </div>
      </dialog>

      {/* Disable 2FA Modal */}
      <dialog id="disable_2fa_modal" class="modal">
        <div class="modal-box">
          <h3 class="font-bold text-lg text-error">Disable Two-Factor Authentication</h3>
          <p class="py-4">For your security, please confirm your password and enter a current 2FA code to disable this feature.</p>
          
          <form onSubmit={handleDisable2FA}>
            <div class="form-control w-full mb-4">
              <label class="label"><span class="label-text">Current Password</span></label>
              <input 
                type="password" 
                value={disablePassword()} 
                onInput={(e) => setDisablePassword(e.currentTarget.value)} 
                class="input input-bordered w-full" 
                required 
              />
            </div>
            
            <div class="form-control w-full mb-6">
              <label class="label"><span class="label-text">6-Digit Authentication Code</span></label>
              <input 
                type="text" 
                value={disableOtp()} 
                onInput={(e) => setDisableOtp(e.currentTarget.value)} 
                class="input input-bordered w-full text-center tracking-widest font-mono" 
                required 
                maxLength={6} 
                placeholder="000000"
              />
            </div>
            
            <div class="modal-action">
              <button type="button" class="btn" onClick={() => {
                const modal = document.getElementById('disable_2fa_modal') as HTMLDialogElement;
                modal?.close();
              }}>Cancel</button>
              <button type="submit" class="btn btn-error" disabled={mutations.disable2FA.isPending}>
                {mutations.disable2FA.isPending ? 'Disabling...' : 'Confirm Disable'}
              </button>
            </div>
          </form>
        </div>
        <form method="dialog" class="modal-backdrop">
          <button>close</button>
        </form>
      </dialog>
    </>
  )
}
