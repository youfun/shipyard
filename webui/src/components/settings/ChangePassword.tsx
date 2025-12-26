import { createSignal, JSX } from 'solid-js'
import { toast } from 'solid-toast'
import { useI18n } from '@i18n'
import { useAuth } from '@api/hooks'

export function ChangePassword(): JSX.Element {
  const { t } = useI18n()
  const { mutations } = useAuth()

  const [currentPassword, setCurrentPassword] = createSignal('')
  const [newPassword, setNewPassword] = createSignal('')
  const [confirmPassword, setConfirmPassword] = createSignal('')

  const handlePasswordSubmit = async (e: Event) => {
    e.preventDefault()
    
    if (!currentPassword().trim() || !newPassword().trim()) {
      toast.error(t('change_password.error_empty_fields'))
      return
    }
    
    if (newPassword().length < 6) {
      toast.error(t('change_password.error_password_length'))
      return
    }
    
    if (newPassword() !== confirmPassword()) {
      toast.error(t('change_password.error_password_mismatch'))
      return
    }

    mutations.changePassword.mutate(
      {
        current_password: currentPassword(),
        new_password: newPassword(),
      },
      {
        onSuccess: () => {
          toast.success(t('change_password.success_message'))
          setCurrentPassword('')
          setNewPassword('')
          setConfirmPassword('')
        },
        onError: (error: any) => {
          toast.error(error.response?.data?.error || error.message || t('change_password.error_change_failed'))
        },
      }
    )
  }

  return (
    <div class="card bg-base-100 shadow h-fit">
      <div class="card-body">
        <h2 class="card-title mb-4">{t('change_password.title')}</h2>
        <form onSubmit={handlePasswordSubmit}>
          <div class="form-control w-full mb-4">
            <label class="label">
              <span class="label-text">{t('change_password.current_password')}</span>
            </label>
            <input
              type="password"
              placeholder={t('change_password.current_password_placeholder')}
              class="input input-bordered w-full"
              value={currentPassword()}
              onInput={(e) => setCurrentPassword(e.currentTarget.value)}
              disabled={mutations.changePassword.isPending}
              required
            />
          </div>
          
          <div class="form-control w-full mb-4">
            <label class="label">
              <span class="label-text">{t('change_password.new_password')}</span>
            </label>
            <input
              type="password"
              placeholder={t('change_password.new_password_placeholder')}
              class="input input-bordered w-full"
              value={newPassword()}
              onInput={(e) => setNewPassword(e.currentTarget.value)}
              disabled={mutations.changePassword.isPending}
              required
            />
          </div>
          
          <div class="form-control w-full mb-6">
            <label class="label">
              <span class="label-text">{t('change_password.confirm_password')}</span>
            </label>
            <input
              type="password"
              placeholder={t('change_password.confirm_password_placeholder')}
              class="input input-bordered w-full"
              value={confirmPassword()}
              onInput={(e) => setConfirmPassword(e.currentTarget.value)}
              disabled={mutations.changePassword.isPending}
              required
            />
          </div>
          
          <div class="form-control">
            <button
              type="submit"
              class="btn btn-primary"
              disabled={mutations.changePassword.isPending}
            >
              {mutations.changePassword.isPending && <span class="loading loading-spinner"></span>}
              {mutations.changePassword.isPending ? t('change_password.changing') : t('change_password.change_button')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
