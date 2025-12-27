import { JSX, Show, For } from 'solid-js'
import { useRouter } from '@router'
import { useI18n } from '@i18n'
import { useDashboard } from '@api/hooks'

// Helper to format deployed_at time
const formatTime = (isoString?: string): string => {
  if (!isoString) return '-'
  const date = new Date(isoString)
  return date.toLocaleString()
}

// Helper to truncate git commit to 7 chars
const truncateCommit = (commit: string): string => {
  if (!commit) return '-'
  return commit.slice(0, 7)
}

// Helper to get status badge class
const getStatusBadgeClass = (status: string): string => {
  switch (status) {
    case 'success':
      return 'badge-success'
    case 'failed':
      return 'badge-error'
    case 'pending':
    case 'deploying':
      return 'badge-warning'
    default:
      return 'badge-ghost'
  }
}

export default function DashboardPage(): JSX.Element {
  const { t } = useI18n()
  const router = useRouter()
  const { queries } = useDashboard()

  const statsQuery = queries.getStats()
  const deploymentsQuery = queries.getRecentDeployments()

  return (
    <div class="container mx-auto p-6">
      <h1 class="text-3xl font-bold mb-6">{t('nav.dashboard')}</h1>

      {/* Stats Cards */}
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div
          class="card bg-base-100 shadow-xl cursor-pointer hover:shadow-2xl hover:scale-[1.02] transition-all duration-200"
          onClick={() => router.navigate('/admin/apps')}
        >
          <div class="card-body">
            <h2 class="card-title text-primary">{t('dashboard.cards.applications')}</h2>
            <Show when={!statsQuery.isLoading} fallback={<span class="loading loading-spinner loading-md"></span>}>
              <p class="text-4xl font-bold">{statsQuery.data?.applications_count ?? 0}</p>
            </Show>
            <p class="text-base-content/70">{t('dashboard.cards.deployed_applications')}</p>
          </div>
        </div>

        <div
          class="card bg-base-100 shadow-xl cursor-pointer hover:shadow-2xl hover:scale-[1.02] transition-all duration-200"
          onClick={() => router.navigate('/admin/ssh-management')}
        >
          <div class="card-body">
            <h2 class="card-title text-success">{t('dashboard.cards.hosts')}</h2>
            <Show when={!statsQuery.isLoading} fallback={<span class="loading loading-spinner loading-md"></span>}>
              <p class="text-4xl font-bold">{statsQuery.data?.hosts_count ?? 0}</p>
            </Show>
            <p class="text-base-content/70">{t('dashboard.cards.connected_hosts')}</p>
          </div>
        </div>

        <div class="card bg-base-100 shadow-xl">
          <div class="card-body">
            <h2 class="card-title text-info">{t('dashboard.cards.deployments')}</h2>
            <Show when={!statsQuery.isLoading} fallback={<span class="loading loading-spinner loading-md"></span>}>
              <p class="text-4xl font-bold">{statsQuery.data?.deployments_count ?? 0}</p>
            </Show>
            <p class="text-base-content/70">{t('dashboard.cards.total_deployments')}</p>
          </div>
        </div>
      </div>

      {/* Recent Deployments */}
      <div class="card bg-base-100 shadow-xl">
        <div class="card-body">
          <h2 class="card-title">{t('dashboard.recent_deployments.title')}</h2>
          <Show
            when={!deploymentsQuery.isLoading}
            fallback={
              <div class="flex justify-center py-8">
                <span class="loading loading-spinner loading-lg"></span>
              </div>
            }
          >
            <Show
              when={deploymentsQuery.data && deploymentsQuery.data.length > 0}
              fallback={
                <div class="text-center py-8 text-base-content/50">
                  {t('dashboard.recent_deployments.no_deployments')}
                </div>
              }
            >
              <div class="overflow-x-auto">
                <table class="table table-zebra w-full">
                  <thead>
                    <tr>
                      <th>{t('dashboard.recent_deployments.host')}</th>
                      <th>{t('dashboard.recent_deployments.time')}</th>
                      <th>{t('dashboard.recent_deployments.app_name')}</th>
                      <th>{t('dashboard.recent_deployments.version')}</th>
                      <th>{t('dashboard.recent_deployments.git_commit')}</th>
                      <th>{t('dashboard.recent_deployments.status')}</th>
                      <th>{t('dashboard.recent_deployments.port')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <For each={deploymentsQuery.data}>
                      {(deployment) => (
                        <tr>
                          <td>
                            <div class="font-medium">{deployment.host_name}</div>
                            <div class="text-sm text-base-content/60">{deployment.host_addr}</div>
                          </td>
                          <td class="text-sm">{formatTime(deployment.deployed_at)}</td>
                          <td class="font-medium">{deployment.app_name}</td>
                          <td><code class="text-sm">{deployment.version}</code></td>
                          <td>
                            <code class="text-sm font-mono">{truncateCommit(deployment.git_commit)}</code>
                          </td>
                          <td>
                            <span class={`badge ${getStatusBadgeClass(deployment.status)}`}>
                              {deployment.status}
                            </span>
                          </td>
                          <td>{deployment.port || '-'}</td>
                        </tr>
                      )}
                    </For>
                  </tbody>
                </table>
              </div>
            </Show>
          </Show>
        </div>
      </div>
    </div>
  )
}
