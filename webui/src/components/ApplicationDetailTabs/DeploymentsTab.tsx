import { For, Show, JSX, createSignal } from 'solid-js'
import { useI18n } from '@i18n'
import type { DeploymentHistory } from '@types'
import { LogsModal } from '@components/ui/LogsModal'
import { fetchDeploymentLogs } from '@api/services/applicationService'

export function DeploymentsTab(props: { deployments: DeploymentHistory[]; isLoading: boolean }): JSX.Element {
  const { t } = useI18n()
  const [showLogsModal, setShowLogsModal] = createSignal(false)
  const [currentDeploymentUid, setCurrentDeploymentUid] = createSignal('')
  const [currentDeploymentVersion, setCurrentDeploymentVersion] = createSignal('')

  const handleViewLogs = (uid: string, version: string) => {
    setCurrentDeploymentUid(uid)
    setCurrentDeploymentVersion(version)
    setShowLogsModal(true)
  }

  const fetchLogsForDeployment = async (_lines: number) => {
    const response = await fetchDeploymentLogs(currentDeploymentUid())
    return response.logs
  }

  return (
    <div>
      <Show when={!props.isLoading} fallback={
        <div class="flex justify-center py-8">
          <span class="loading loading-spinner loading-md"></span>
        </div>
      }>
        <Show when={props.deployments.length > 0} fallback={
          <div class="text-center py-8 text-base-content/50">
            {t('app_detail.deployments_empty')}
          </div>
        }>
          <table class="table">
            <thead>
              <tr>
                <th>{t('app_detail.deployments_version')}</th>
                <th>{t('app_detail.deployments_status')}</th>
                <th>{t('app_detail.deployments_host')}</th>
                <th>{t('app_detail.deployments_port')}</th>
                <th>{t('app_detail.deployments_created')}</th>
                <th>{t('app_detail.instance_actions')}</th>
              </tr>
            </thead>
            <tbody>
              <For each={props.deployments}>
                {(deployment) => (
                  <tr class="hover">
                    <td>{deployment.version}</td>
                    <td>
                      <span classList={{
                        'badge': true,
                        'badge-success': deployment.status === 'success',
                        'badge-error': deployment.status === 'failed',
                        'badge-warning': deployment.status === 'pending',
                      }}>
                        {deployment.status}
                      </span>
                    </td>
                    <td>{deployment.host_name}</td>
                    <td>{deployment.port || '-'}</td>
                    <td>{deployment.created_at}</td>
                    <td>
                      <button 
                        class="btn btn-xs btn-info" 
                        onClick={() => handleViewLogs(deployment.uid, deployment.version)}
                      >
                        {t('app_detail.action_view_logs')}
                      </button>
                    </td>
                  </tr>
                )}
              </For>
            </tbody>
          </table>
        </Show>
      </Show>

      <LogsModal
        show={showLogsModal()}
        title={`${t('app_detail.deployment_logs_title')} - ${currentDeploymentVersion()}`}
        onClose={() => setShowLogsModal(false)}
        fetchLogs={fetchLogsForDeployment}
      />
    </div>
  )
}
