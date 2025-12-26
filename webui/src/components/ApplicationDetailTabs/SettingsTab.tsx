import { JSX } from 'solid-js'
import { useI18n } from '@i18n'
import type { Application } from '@types'

export function SettingsTab(props: { app: Application | undefined }): JSX.Element {
  const { t } = useI18n()
  
  return (
    <div class="space-y-6">
      <div class="card bg-base-100">
        <div class="card-body">
          <h3 class="card-title">{t('app_detail.settings_title')}</h3>
          <p class="text-base-content/70">
            {t('app_detail.settings_description')}
          </p>
          <div class="form-control">
            <label class="label">
              <span class="label-text">{t('app_detail.settings_label_name')}</span>
            </label>
            <input 
              type="text" 
              class="input input-bordered" 
              value={props.app?.name || ''} 
              disabled 
            />
          </div>
          <div class="form-control">
            <label class="label">
              <span class="label-text">{t('app_detail.settings_label_description')}</span>
            </label>
            <textarea 
              class="textarea textarea-bordered" 
              value={props.app?.description || ''} 
              disabled 
            />
          </div>
        </div>
      </div>
    </div>
  )
}
