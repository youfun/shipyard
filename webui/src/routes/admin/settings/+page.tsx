import { JSX } from 'solid-js'
import { useI18n } from '@i18n'
import { ChangePassword } from '@components/settings/ChangePassword'
import { TwoFactorAuth } from '@components/settings/TwoFactorAuth'
import { SystemSettings } from '@components/settings/SystemSettings'

export default function SettingsPage(): JSX.Element {
  const { t } = useI18n()
  
  return (
    <div class="container mx-auto p-6 space-y-8">
      {/* Header */}
      <div>
        <h1 class="text-3xl font-bold text-base-content">{t('nav.settings')}</h1>
        <p class="text-base-content/70 mt-2">{t('security_settings.description')}</p>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* System Domain Section */}
        <div class="lg:col-span-2">
           <SystemSettings />
        </div>

        {/* Change Password Section */}
        <ChangePassword />

        {/* 2FA Section */}
        <TwoFactorAuth />
      </div>
    </div>
  )
}