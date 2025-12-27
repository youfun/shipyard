import { Component, createSignal, createEffect } from 'solid-js'
import { useI18n } from '@i18n'
import { useSystemSettings } from '@api/hooks'
import { toast } from 'solid-toast'

export const SystemSettings: Component = () => {
  const { t } = useI18n()
  const { settingsQuery, updateSettingsMutation } = useSystemSettings()
  const [domain, setDomain] = createSignal('')

  createEffect(() => {
    if (settingsQuery.data) {
      setDomain(settingsQuery.data.domain || '')
    }
  })

  const handleSave = async (e: Event) => {
    e.preventDefault()
    if (!domain().trim()) {
      toast.error(t('system_settings.error_empty_domain'))
      return
    }

    try {
      const result = await updateSettingsMutation.mutateAsync({ domain: domain().trim() }) as any
      if (result && result.caddy_error) {
        toast.error(result.message, { duration: 5000 })
      } else {
        toast.success(t('system_settings.success_message'))
      }
    } catch (error) {
      toast.error(t('system_settings.error_update_failed'))
    }
  }

  return (
    <div class="card bg-base-100 shadow-xl border border-base-300">
      <div class="card-body">
        <h2 class="card-title flex items-center gap-2">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6 text-primary">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0121 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0112 16.5c-3.162 0-6.133-.815-8.716-2.247m17.432 0c-.185.72-.486 1.405-.884 2.033m-16.548-2.033a8.959 8.959 0 01-.284-2.253c0-.778.099-1.533.284-2.253m0 0A17.919 17.919 0 0012 7.5a17.919 17.919 0 008.716 2.247m0 0c.185.72.284 1.475.284 2.253" />
          </svg>
          {t('system_settings.title')}
        </h2>
        <p class="text-base-content/60 text-sm mb-4">{t('system_settings.description')}</p>
        
        <form onSubmit={handleSave} class="space-y-4">
          <div class="form-control w-full">
            <label class="label">
              <span class="label-text font-medium">{t('system_settings.domain')}</span>
            </label>
            <input 
              type="text" 
              placeholder="admin.example.com" 
              class="input input-bordered w-full" 
              value={domain()}
              onInput={(e) => setDomain(e.currentTarget.value)}
            />
            <label class="label">
              <span class="label-text-alt text-base-content/50">
                {t('system_settings.domain_placeholder')}
              </span>
            </label>
          </div>

          <div class="card-actions justify-end mt-2">
            <button 
              type="submit" 
              class="btn btn-primary"
              disabled={updateSettingsMutation.isPending}
            >
              {updateSettingsMutation.isPending && <span class="loading loading-spinner loading-xs"></span>}
              {t('common.save')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
