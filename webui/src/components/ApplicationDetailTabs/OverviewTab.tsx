import { JSX, For, Show, createSignal } from 'solid-js'
import { useI18n } from '@i18n'
import type { Application } from '@types'
import { useApplications } from '@api/hooks'
import { toast } from 'solid-toast'
import { fetchInstanceLogs } from '@api/services/applicationService'
import { LogsModal } from '@components/ui/LogsModal'

export function OverviewTab(props: { app: Application | undefined }): JSX.Element {
  const { t } = useI18n()
  const { mutations } = useApplications()
  const [showLogsModal, setShowLogsModal] = createSignal(false)
  const [currentInstanceUid, setCurrentInstanceUid] = createSignal('')
  const [currentInstanceName, setCurrentInstanceName] = createSignal('')

  const handleStop = async (uid: string) => {
    if (!props.app?.uid) return
    try {
      await mutations.stopInstance.mutateAsync({ uid, appUid: props.app.uid })
      toast.success('Success')
    } catch (e: any) {
      toast.error(e.response?.data?.error || t('common.error'))
    }
  }

  const handleStart = async (uid: string) => {
    if (!props.app?.uid) return
    try {
      await mutations.startInstance.mutateAsync({ uid, appUid: props.app.uid })
      toast.success('Success')
    } catch (e: any) {
      toast.error(e.response?.data?.error || t('common.error'))
    }
  }

  const handleRestart = async (uid: string) => {
    if (!props.app?.uid) return
    try {
      await mutations.restartInstance.mutateAsync({ uid, appUid: props.app.uid })
      toast.success('Success')
    } catch (e: any) {
      toast.error(e.response?.data?.error || t('common.error'))
    }
  }

  const handleViewLogs = (uid: string, hostName: string) => {
    setCurrentInstanceUid(uid)
    setCurrentInstanceName(hostName)
    setShowLogsModal(true)
  }

  const fetchLogsForInstance = async (lines: number) => {
    const response = await fetchInstanceLogs(currentInstanceUid(), lines)
    return response.logs
  }

  return (
    <div class="space-y-6">
      <div class="card bg-base-100">
        <div class="card-body">
          <h3 class="card-title">{t('app_detail.overview_title')}</h3>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <span class="text-base-content/70">{t('app_detail.overview_name')}:</span>
              <p class="font-semibold">{props.app?.name || t('app_detail.overview_none')}</p>
            </div>
            <div>
              <span class="text-base-content/70">{t('app_detail.overview_status')}:</span>
              <p class="font-semibold">{props.app?.status || t('app_detail.overview_none')}</p>
            </div>
            <div>
              <span class="text-base-content/70">{t('app_detail.overview_domain')}:</span>
              <p class="font-semibold">{props.app?.primary_domain || t('app_detail.overview_none')}</p>
            </div>
            <div>
              <span class="text-base-content/70">{t('app_detail.overview_target_port')}:</span>
              <p class="font-semibold">{props.app?.active_port || t('app_detail.overview_none')}</p>
            </div>
            <div>
              <span class="text-base-content/70">{t('app_detail.overview_created')}:</span>
              <p class="font-semibold">{props.app?.created_at || t('app_detail.overview_none')}</p>
            </div>
            <div>
              <span class="text-base-content/70">{t('app_detail.overview_last_deployed')}:</span>
              <p class="font-semibold">{props.app?.last_deployed_at || t('app_detail.overview_none')}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="card bg-base-100">
        <div class="card-body">
          <h3 class="card-title">{t('app_detail.instances_title')}</h3>
          <div class="overflow-x-auto">
            <table class="table">
              <thead>
                <tr>
                  <th>{t('app_detail.instance_host')}</th>
                  <th>{t('app_detail.instance_status')}</th>
                  <th>{t('app_detail.instance_port')}</th>
                  <th>{t('app_detail.instance_actions')}</th>
                </tr>
              </thead>
              <tbody>
                <For each={props.app?.instances}>
                  {(instance) => (
                    <tr>
                      <td>{instance.host_name}{instance.host_addr ? ` (${instance.host_addr})` : ''}</td>
                      <td>
                        <span classList={{
                          'badge': true,
                          'badge-success': instance.status === 'running' || instance.status === 'linked',
                          'badge-error': instance.status === 'stopped',
                          'badge-warning': instance.status !== 'running' && instance.status !== 'linked' && instance.status !== 'stopped',
                        }}>
                          {instance.status}
                        </span>
                      </td>
                      <td>{instance.active_port}</td>
                      <td class="flex gap-2">
                        <Show when={instance.status !== 'running' && instance.status !== 'linked'}>
                          <button
                            class="btn btn-xs btn-success"
                            onClick={() => handleStart(instance.uid)}
                            disabled={mutations.startInstance.isPending}
                          >
                            {t('app_detail.action_start')}
                          </button>
                        </Show>
                        <Show when={instance.status === 'running' || instance.status === 'linked'}>
                          <button
                            class="btn btn-xs btn-error"
                            onClick={() => handleStop(instance.uid)}
                            disabled={mutations.stopInstance.isPending}
                          >
                            {t('app_detail.action_stop')}
                          </button>
                          <button
                            class="btn btn-xs btn-warning"
                            onClick={() => handleRestart(instance.uid)}
                            disabled={mutations.restartInstance.isPending}
                          >
                            {t('app_detail.action_restart')}
                          </button>
                        </Show>
                        <button
                          class="btn btn-xs btn-info"
                          onClick={() => handleViewLogs(instance.uid, instance.host_name)}
                        >
                          {t('app_detail.action_view_logs')}
                        </button>
                      </td>
                    </tr>
                  )}
                </For>
                <Show when={!props.app?.instances || props.app.instances.length === 0}>
                  <tr>
                    <td colspan="4" class="text-center text-base-content/70 py-4">
                      {t('app_detail.deployments_empty')}
                    </td>
                  </tr>
                </Show>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <LogsModal
        show={showLogsModal()}
        title={`${t('app_detail.logs_title')} - ${currentInstanceName()}`}
        onClose={() => setShowLogsModal(false)}
        fetchLogs={fetchLogsForInstance}
        instanceUid={currentInstanceUid()}
      />
    </div>
  )
}
