import { JSX, For, Show } from 'solid-js'
import { Link } from '@router'
import { useI18n } from '@i18n'
import { useApplications } from '@api/hooks'

export default function ApplicationsPage(): JSX.Element {
  const { t } = useI18n()
  const { queries } = useApplications()

  const appsQuery = queries.getAll()

  const apps = () => appsQuery.data?.data || []
  const isLoading = () => appsQuery.isPending

  return (
    <div class="container mx-auto p-6">
      <div class="flex justify-between items-center mb-6">
        <h1 class="text-3xl font-bold">{t('app_list.title')}</h1>
      </div>

      <Show when={!isLoading()} fallback={
        <div class="flex justify-center py-12">
          <span class="loading loading-spinner loading-lg"></span>
        </div>
      }>
        <Show when={apps().length > 0} fallback={
          <div class="card bg-base-100 shadow-xl">
            <div class="card-body text-center">
              <h2 class="card-title justify-center">{t('app_list.empty_title')}</h2>
              <p class="text-base-content/70">
                {t('app_list.empty_description')}
              </p>
            </div>
          </div>
        }>
          <div class="card bg-base-100 shadow-xl">
            <div class="card-body p-0">
              <table class="table">
                <thead>
                  <tr>
                    <th>{t('app_list.table_name')}</th>
                    <th>{t('app_list.table_last_deployment')}</th>
                    <th>{t('app_list.table_linked_host')}</th>
                    <th>{t('app_list.table_actions')}</th>
                  </tr>
                </thead>
                <tbody>
                  <For each={apps()}>
                    {(app) => (
                      <tr class="hover">
                        <td class="font-bold">
                          <Link
                            href={`/admin/apps/${app.uid}`}
                            class="link link-primary hover:link-hover"
                          >
                            {app.name}
                          </Link>
                        </td>
                        <td>{app.last_deployed_at || '-'}</td>
                        <td>{app.linked_host || '-'}</td>
                        <td>
                          <Link
                            href={`/admin/apps/${app.uid}`}
                            class="btn btn-sm btn-ghost"
                          >
                            {t('app_list.action_details')}
                          </Link>
                        </td>
                      </tr>
                    )}
                  </For>
                </tbody>
              </table>
            </div>
          </div>
        </Show>
      </Show>
    </div>
  )
}
