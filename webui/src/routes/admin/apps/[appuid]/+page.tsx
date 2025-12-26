import { createSignal, Switch, Match, JSX } from 'solid-js'
import { Link, useRouter, useSearchParams } from '@router'
import { useI18n } from '@i18n'
import { useApplications } from '@api/hooks'
import type {  CreateApplicationTokenRequest } from '@types'
import { OverviewTab } from '@components/ApplicationDetailTabs/OverviewTab'
import { DeploymentsTab } from '@components/ApplicationDetailTabs/DeploymentsTab'
import { EnvironmentTab } from '@components/ApplicationDetailTabs/EnvironmentTab'
import { DomainTab } from '@components/ApplicationDetailTabs/DomainTab'
import { TokensTab } from '@components/ApplicationDetailTabs/TokensTab'
import { SettingsTab } from '@components/ApplicationDetailTabs/SettingsTab'

export default function ApplicationDetailPage(): JSX.Element {
  const { t } = useI18n()
  const router = useRouter()
  const searchParams = useSearchParams()
  const { queries, mutations } = useApplications()

  // Get appuid from URL path
  const getAppUid = () => {
    const path = router.pathname()
    const parts = path.split('/')
    return parts[parts.length - 1]
  }
  
  const appUid = getAppUid

  // State for active tab
  const getInitialTab = () => searchParams().get('tab') || 'Overview'
  const [activeTab, setActiveTab] = createSignal(getInitialTab())

  const handleTabChange = (tab: string) => {
    setActiveTab(tab)
    const newUrl = `${router.pathname()}?tab=${tab}`
    router.navigate(newUrl, { replace: true })
  }

  // Query Application Details
  const appQuery = queries.getById(appUid)
  const deploymentsQuery = queries.getDeployments(appUid)
  const envVarsQuery = queries.getEnvVars(appUid)
  const domainsQuery = queries.getDomains(appUid)
  const tokensQuery = queries.getTokens(appUid)

  const currentApp = () => appQuery.data

  // Token handlers
  const handleCreateToken = async (data: CreateApplicationTokenRequest) => {
    const uid = appUid()
    if (!uid) return undefined
    try {
      const result = await mutations.createToken.mutateAsync({ uid, data })
      return { token: result.token }
    } catch {
      return undefined
    }
  }

  const handleDeleteToken = (tokenId: string) => {
    const uid = appUid()
    if (!uid) return
    mutations.deleteToken.mutate({ uid, tokenId })
  }

  // Navigation
  const handleBack = () => router.navigate('/admin/apps')

  return (
    <>
      <Switch>
        {/* Loading */}
        <Match when={appQuery.isPending}>
          <div class="container mx-auto p-6 flex justify-center py-12">
            <span class="loading loading-spinner loading-lg"></span>
            <p class="mt-4 text-base-content/70 ml-4">{t('app_detail.loading')}</p>
          </div>
        </Match>

        {/* Error */}
        <Match when={appQuery.isError || (!appQuery.isPending && !currentApp())}>
          <div class="container mx-auto p-6">
            <div class="alert alert-error mb-4">
              <span>{t('app_detail.error_not_found')}</span>
              <button class="btn btn-ghost btn-sm" onClick={handleBack}>{t('app_detail.action_back')}</button>
            </div>
          </div>
        </Match>

        {/* Success */}
        <Match when={true}>
          <div class="container mx-auto p-6">
            {/* Breadcrumbs */}
            <div class="breadcrumbs text-sm mb-6">
              <ul>
                <li><Link href="/admin/dashboard">{t('app_detail.breadcrumb_home')}</Link></li>
                <li><Link href="/admin/apps">{t('app_detail.breadcrumb_applications')}</Link></li>
                <li>{currentApp()?.name}</li>
              </ul>
            </div>

            {/* Tabs */}
            <div role="tablist" class="tabs tabs-bordered">
              <button 
                role="tab" 
                class="tab"
                classList={{ 'tab-active': activeTab() === 'Overview' }}
                onClick={() => handleTabChange('Overview')}
              >
                {t('app_detail.tab_overview')}
              </button>
              <button 
                role="tab" 
                class="tab"
                classList={{ 'tab-active': activeTab() === 'Deployments' }}
                onClick={() => handleTabChange('Deployments')}
              >
                {t('app_detail.tab_deployments')}
              </button>
              <button 
                role="tab" 
                class="tab"
                classList={{ 'tab-active': activeTab() === 'Environment' }}
                onClick={() => handleTabChange('Environment')}
              >
                {t('app_detail.tab_environment')}
              </button>
              <button 
                role="tab" 
                class="tab"
                classList={{ 'tab-active': activeTab() === 'Domain' }}
                onClick={() => handleTabChange('Domain')}
              >
                {t('app_detail.tab_domains')}
              </button>
              <button 
                role="tab" 
                class="tab"
                classList={{ 'tab-active': activeTab() === 'Tokens' }}
                onClick={() => handleTabChange('Tokens')}
              >
                {t('app_detail.tab_tokens')}
              </button>
              <button 
                role="tab" 
                class="tab"
                classList={{ 'tab-active': activeTab() === 'Settings' }}
                onClick={() => handleTabChange('Settings')}
              >
                {t('app_detail.tab_settings')}
              </button>
            </div>

            {/* Content */}
            <div class="p-4 border-base-300 border border-t-0 rounded-b-lg">
              <Switch>
                <Match when={activeTab() === 'Overview'}>
                  <OverviewTab app={currentApp()} />
                </Match>
                <Match when={activeTab() === 'Deployments'}>
                  <DeploymentsTab deployments={deploymentsQuery.data?.data || []} isLoading={deploymentsQuery.isPending} />
                </Match>
                <Match when={activeTab() === 'Environment'}>
                  <EnvironmentTab 
                    envVars={envVarsQuery.data?.data || []} 
                    isLoading={envVarsQuery.isPending}
                    appUid={appUid()}
                  />
                </Match>
                <Match when={activeTab() === 'Domain'}>
                  <DomainTab domains={domainsQuery.data?.data || []} isLoading={domainsQuery.isPending} />
                </Match>
                <Match when={activeTab() === 'Tokens'}>
                  <TokensTab 
                    tokens={tokensQuery.data || []} 
                    isLoading={tokensQuery.isPending}
                    onCreateToken={handleCreateToken}
                    onDeleteToken={handleDeleteToken}
                    isCreating={mutations.createToken.isPending}
                    isDeleting={mutations.deleteToken.isPending}
                  />
                </Match>
                <Match when={activeTab() === 'Settings'}>
                  <SettingsTab app={currentApp()} />
                </Match>
              </Switch>
            </div>
          </div>
        </Match>
      </Switch>
    </>
  )
}
