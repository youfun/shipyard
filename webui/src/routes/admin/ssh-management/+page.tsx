import { createSignal, Show, For, JSX } from 'solid-js'
import { toast } from 'solid-toast'
import { useI18n } from '@i18n'
import { useSSHHosts } from '@api/hooks'
import type { SSHHost, SSHHostRequest } from '@types'

export default function SSHManagementPage(): JSX.Element {
  const { t } = useI18n()
  const { queries, mutations } = useSSHHosts()
  
  // Modal states
  const [showCreateModal, setShowCreateModal] = createSignal(false)
  const [showEditModal, setShowEditModal] = createSignal(false)
  const [showDeleteModal, setShowDeleteModal] = createSignal(false)
  const [selectedHost, setSelectedHost] = createSignal<SSHHost | null>(null)

  // Form state
  const [formData, setFormData] = createSignal<SSHHostRequest>({
    name: '',
    addr: '',
    port: 22,
    user: '',
    password: '',
    private_key: '',
  })

  // API query for SSH hosts
  const hostsQuery = queries.getAll()

  // Helper functions
  const hosts = () => hostsQuery.data?.data || []
  const isLoading = () => hostsQuery.isPending
  const isMutating = () => mutations.create.isPending || mutations.update.isPending || mutations.delete.isPending

  const resetForm = () => {
    setFormData({
      name: '',
      addr: '',
      port: 22,
      user: '',
      password: '',
      private_key: '',
    })
  }

  const openCreateModal = () => {
    resetForm()
    setShowCreateModal(true)
  }

  const openEditModal = (host: SSHHost) => {
    setSelectedHost(host)
    setFormData({
      name: host.name,
      addr: host.addr,
      port: host.port,
      user: host.user,
      password: '',
      private_key: '',
    })
    setShowEditModal(true)
  }

  const openDeleteModal = (host: SSHHost) => {
    setSelectedHost(host)
    setShowDeleteModal(true)
  }

  const handleCreate = () => {
    const data = formData()
    if (!data.name || !data.addr || !data.user) {
      toast.error('Name, address, and user are required')
      return
    }

    if (!data.password && !data.private_key) {
      toast.error('Either password or private key must be provided')
      return
    }

    mutations.create.mutate(data, {
      onSuccess: () => {
        setShowCreateModal(false)
        resetForm()
        toast.success('SSH host created successfully')
      },
      onError: (error: any) => {
        toast.error(error.response?.data?.error || error.message || 'Failed to create SSH host')
      },
    })
  }

  const handleUpdate = () => {
    const host = selectedHost()
    if (!host) return

    const data = formData()
    if (!data.name || !data.addr || !data.user) {
      toast.error('Name, address, and user are required')
      return
    }

    mutations.update.mutate(
      { uid: host.uid, data },
      {
        onSuccess: () => {
          setShowEditModal(false)
          setSelectedHost(null)
          resetForm()
          toast.success('SSH host updated successfully')
        },
        onError: (error: any) => {
          toast.error(error.response?.data?.error || error.message || 'Failed to update SSH host')
        },
      }
    )
  }

  const handleDelete = () => {
    const host = selectedHost()
    if (!host) return

    mutations.delete.mutate(host.uid, {
      onSuccess: () => {
        setShowDeleteModal(false)
        setSelectedHost(null)
        toast.success('SSH host deleted successfully')
      },
      onError: (error: any) => {
        toast.error(error.response?.data?.error || error.message || 'Failed to delete SSH host')
      },
    })
  }

  return (
    <div class="container mx-auto p-6">
      {/* Header */}
      <div class="mb-6 flex justify-between items-center">
        <div>
          <h1 class="text-3xl font-bold text-base-content">{t('ssh.title')}</h1>
          <p class="text-base-content/70 mt-2">{t('ssh.description')}</p>
        </div>
        
        <button 
          class="btn btn-primary"
          onClick={openCreateModal}
        >
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
          {t('ssh.add_host')}
        </button>
      </div>

      <SSHHostTable
        hosts={hosts()}
        isLoading={isLoading()}
        onEdit={openEditModal}
        onDelete={openDeleteModal}
      />

      {/* Create Modal */}
      <Show when={showCreateModal()}>
        <div class="modal modal-open">
          <div class="modal-box">
            <h3 class="font-bold text-lg mb-4">{t('ssh.add_host')}</h3>
            <SSHHostForm
              data={formData()}
              onChange={setFormData}
              disabled={isMutating()}
            />
            <div class="modal-action">
              <button class="btn btn-ghost" onClick={() => setShowCreateModal(false)} disabled={isMutating()}>
                {t('common.cancel')}
              </button>
              <button class="btn btn-primary" onClick={handleCreate} disabled={isMutating()}>
                {isMutating() && <span class="loading loading-spinner loading-sm"></span>}
                {t('common.save')}
              </button>
            </div>
          </div>
          <div class="modal-backdrop" onClick={() => setShowCreateModal(false)}></div>
        </div>
      </Show>

      {/* Edit Modal */}
      <Show when={showEditModal()}>
        <div class="modal modal-open">
          <div class="modal-box">
            <h3 class="font-bold text-lg mb-4">{t('ssh.edit_host')}</h3>
            <SSHHostForm
              data={formData()}
              onChange={setFormData}
              disabled={isMutating()}
              isEdit
            />
            <div class="modal-action">
              <button class="btn btn-ghost" onClick={() => setShowEditModal(false)} disabled={isMutating()}>
                {t('common.cancel')}
              </button>
              <button class="btn btn-primary" onClick={handleUpdate} disabled={isMutating()}>
                {isMutating() && <span class="loading loading-spinner loading-sm"></span>}
                {t('common.save')}
              </button>
            </div>
          </div>
          <div class="modal-backdrop" onClick={() => setShowEditModal(false)}></div>
        </div>
      </Show>

      {/* Delete Modal */}
      <Show when={showDeleteModal()}>
        <div class="modal modal-open">
          <div class="modal-box">
            <h3 class="font-bold text-lg mb-4">{t('ssh.confirm_delete')}</h3>
            <p>{t('ssh.delete_warning').replace('{name}', selectedHost()?.name || '')}</p>
            <div class="modal-action">
              <button class="btn btn-ghost" onClick={() => setShowDeleteModal(false)} disabled={isMutating()}>
                {t('common.cancel')}
              </button>
              <button class="btn btn-error" onClick={handleDelete} disabled={isMutating()}>
                {isMutating() && <span class="loading loading-spinner loading-sm"></span>}
                {t('common.delete')}
              </button>
            </div>
          </div>
          <div class="modal-backdrop" onClick={() => setShowDeleteModal(false)}></div>
        </div>
      </Show>
    </div>
  )
}

// SSH Host Table Component
function SSHHostTable(props: {
  hosts: SSHHost[]
  isLoading: boolean
  onEdit: (host: SSHHost) => void
  onDelete: (host: SSHHost) => void
}): JSX.Element {
  const { t } = useI18n()

  return (
    <Show when={!props.isLoading} fallback={
      <div class="flex justify-center py-12">
        <span class="loading loading-spinner loading-lg"></span>
      </div>
    }>
      <Show when={props.hosts.length > 0} fallback={
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body text-center">
            <h2 class="card-title justify-center">{t('ssh.no_hosts')}</h2>
            <p class="text-base-content/70">
              Click the button above to add your first SSH host.
            </p>
          </div>
        </div>
      }>
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body p-0">
            <table class="table">
              <thead>
                <tr>
                  <th>{t('ssh.name')}</th>
                  <th>{t('ssh.address')}</th>
                  <th>{t('ssh.user')}</th>
                  <th>{t('ssh.port')}</th>
                  <th>Status</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                <For each={props.hosts}>
                  {(host) => (
                    <tr class="hover">
                      <td class="font-bold">{host.name}</td>
                      <td>{host.addr}</td>
                      <td>{host.user}</td>
                      <td>{host.port}</td>
                      <td>
                        <span classList={{
                          'badge': true,
                          'badge-success': host.status === 'connected',
                          'badge-error': host.status === 'disconnected',
                          'badge-warning': host.status === 'unknown',
                        }}>
                          {host.status || 'unknown'}
                        </span>
                      </td>
                      <td>
                        <div class="flex gap-2">
                          <button 
                            class="btn btn-sm btn-ghost"
                            onClick={() => props.onEdit(host)}
                          >
                            {t('common.edit')}
                          </button>
                          <button 
                            class="btn btn-sm btn-ghost text-error"
                            onClick={() => props.onDelete(host)}
                          >
                            {t('common.delete')}
                          </button>
                        </div>
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
  )
}

// SSH Host Form Component
function SSHHostForm(props: {
  data: SSHHostRequest
  onChange: (data: SSHHostRequest) => void
  disabled: boolean
  isEdit?: boolean
}): JSX.Element {
  const { t } = useI18n()

  const updateField = (field: keyof SSHHostRequest, value: string | number) => {
    props.onChange({ ...props.data, [field]: value })
  }

  return (
    <div class="space-y-4">
      <div class="form-control">
        <label class="label">
          <span class="label-text">{t('ssh.name')}</span>
        </label>
        <input
          type="text"
          class="input input-bordered"
          placeholder={t('ssh.name_placeholder')}
          value={props.data.name}
          onInput={(e) => updateField('name', e.currentTarget.value)}
          disabled={props.disabled}
        />
      </div>

      <div class="form-control">
        <label class="label">
          <span class="label-text">{t('ssh.address')}</span>
        </label>
        <input
          type="text"
          class="input input-bordered"
          placeholder={t('ssh.address_placeholder')}
          value={props.data.addr}
          onInput={(e) => updateField('addr', e.currentTarget.value)}
          disabled={props.disabled}
        />
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div class="form-control">
          <label class="label">
            <span class="label-text">{t('ssh.user')}</span>
          </label>
          <input
            type="text"
            class="input input-bordered"
            placeholder={t('ssh.user_placeholder')}
            value={props.data.user}
            onInput={(e) => updateField('user', e.currentTarget.value)}
            disabled={props.disabled}
          />
        </div>

        <div class="form-control">
          <label class="label">
            <span class="label-text">{t('ssh.port')}</span>
          </label>
          <input
            type="number"
            class="input input-bordered"
            value={props.data.port || 22}
            onInput={(e) => updateField('port', parseInt(e.currentTarget.value) || 22)}
            disabled={props.disabled}
          />
        </div>
      </div>

      <div class="form-control">
        <label class="label">
          <span class="label-text">{t('ssh.password')}</span>
        </label>
        <input
          type="password"
          class="input input-bordered"
          placeholder={t('ssh.password_placeholder')}
          value={props.data.password || ''}
          onInput={(e) => updateField('password', e.currentTarget.value)}
          disabled={props.disabled}
        />
      </div>

      <div class="form-control">
        <label class="label">
          <span class="label-text">{t('ssh.private_key')}</span>
        </label>
        <textarea
          class="textarea textarea-bordered h-24"
          placeholder={t('ssh.private_key_placeholder')}
          value={props.data.private_key || ''}
          onInput={(e) => updateField('private_key', e.currentTarget.value)}
          disabled={props.disabled}
        />
      </div>
    </div>
  )
}
