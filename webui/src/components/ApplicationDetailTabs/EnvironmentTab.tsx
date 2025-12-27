import { For, Show, JSX, createSignal } from 'solid-js'
import { useI18n } from '@i18n'
import type { EnvironmentVariable } from '@types'
import { useApplications } from '@api/hooks'
import { toast } from 'solid-toast'

export function EnvironmentTab(props: { envVars: EnvironmentVariable[]; isLoading: boolean; appUid?: string }): JSX.Element {
  const { t } = useI18n()
  const { mutations } = useApplications()
  
  const [isModalOpen, setIsModalOpen] = createSignal(false)
  const [editingId, setEditingId] = createSignal<string | null>(null)
  const [newKey, setNewKey] = createSignal('')
  const [newValue, setNewValue] = createSignal('')
  const [isEncrypted, setIsEncrypted] = createSignal(false)

  const openCreateModal = () => {
    setEditingId(null)
    setNewKey('')
    setNewValue('')
    setIsEncrypted(false)
    setIsModalOpen(true)
  }

  const openEditModal = (envVar: EnvironmentVariable) => {
    setEditingId(envVar.uid)
    setNewKey(envVar.key)
    setNewValue(envVar.isEncrypted ? '' : envVar.value) // Don't show value if encrypted
    setIsEncrypted(envVar.isEncrypted)
    setIsModalOpen(true)
  }

  const handleSave = async () => {
    if (!props.appUid) return
    
    if (!newKey().trim() || (!newValue().trim() && !editingId())) { // Allow empty value on edit (if updating only key or encryption) - actually value is required for create, maybe optional for update if not changed? Let's assume re-entering value if encrypted is needed or keep simple.
      // If editing and encrypted, user might not change value. But here we treat it as new input.
      // Let's stick to simple validation: Key required. Value required if not encrypted or if new.
      // For simplicity, let's require value always for now, or handle empty value logic.
      if (!newKey().trim() || (!newValue().trim() && !isEncrypted())) {
         toast.error(t('login.empty_fields_error'))
         return
      }
    }
    
    // Check if key and value are present.
    if (!newKey().trim()) {
      toast.error(t('login.empty_fields_error'))
      return
    }

    try {
      if (editingId()) {
        await mutations.updateEnvVar.mutateAsync({
          envVarId: editingId()!,
          appUid: props.appUid,
          data: {
            key: newKey().trim(),
            value: newValue().trim(), // If empty and encrypted, backend might keep old value? Typically we send what we have.
            isEncrypted: isEncrypted()
          }
        })
        toast.success(t('common.save') + ' ' + t('common.yes'))
      } else {
        if (!newValue().trim()) {
           toast.error(t('login.empty_fields_error'))
           return
        }
        await mutations.createEnvVar.mutateAsync({
          uid: props.appUid,
          data: {
            key: newKey().trim(),
            value: newValue().trim(),
            isEncrypted: isEncrypted()
          }
        })
        toast.success(t('common.save') + ' ' + t('common.yes'))
      }
      setIsModalOpen(false)
      setNewKey('')
      setNewValue('')
      setIsEncrypted(false)
      setEditingId(null)
    } catch (error) {
      toast.error(t('common.error'))
    }
  }

  const handleDelete = async (envVarId: string) => {
    if (!props.appUid || !confirm(t('common.confirm_delete'))) return

    try {
      await mutations.deleteEnvVar.mutateAsync({
        envVarId,
        appUid: props.appUid
      })
      toast.success(t('common.delete') + ' ' + t('common.yes'))
    } catch (error) {
      toast.error(t('common.error'))
    }
  }

  return (
    <div>
      <div class="flex justify-end mb-4">
        <button class="btn btn-primary btn-sm" onClick={openCreateModal}>
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4 mr-1">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
          {t('common.add')}
        </button>
      </div>

      <Show when={!props.isLoading} fallback={
        <div class="flex justify-center py-8">
          <span class="loading loading-spinner loading-md"></span>
        </div>
      }>
        <Show when={props.envVars.length > 0} fallback={
          <div class="text-center py-8 text-base-content/50">
            {t('app_detail.environment_empty')}
          </div>
        }>
          <table class="table">
            <thead>
              <tr>
                <th>{t('app_detail.environment_key')}</th>
                <th>{t('app_detail.environment_value')}</th>
                <th>{t('app_detail.environment_encrypted')}</th>
                <th class="w-24">{t('common.actions')}</th>
              </tr>
            </thead>
            <tbody>
              <For each={props.envVars}>
                {(envVar) => {
                  return (
                  <tr class="hover">
                    <td class="font-mono">{envVar.key}</td>
                    <td class="font-mono">{envVar.isEncrypted ? '********' : envVar.value}</td>
                    <td>
                      <span classList={{
                        'badge': true,
                        'badge-success': envVar.isEncrypted,
                        'badge-ghost': !envVar.isEncrypted,
                      }}>
                        {envVar.isEncrypted ? t('app_detail.environment_yes') : t('app_detail.environment_no')}
                      </span>
                    </td>
                    <td>
                      <div class="flex gap-2">
                        <button 
                          class="btn btn-ghost btn-xs" 
                          onClick={() => openEditModal(envVar)}
                          title={t('common.edit')}
                        >
                           <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
                            <path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" />
                          </svg>
                        </button>
                        <button 
                          class="btn btn-ghost btn-xs text-error" 
                          onClick={() => handleDelete(envVar.uid)}
                          title={t('common.delete')}
                        >
                          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
                            <path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" />
                          </svg>
                        </button>
                      </div>
                    </td>
                  </tr>
                )}}
              </For>
            </tbody>
          </table>
        </Show>
      </Show>

      {/* Create/Edit Modal */}
      <Show when={isModalOpen()}>
        <div class="modal modal-open">
          <div class="modal-box">
            <h3 class="font-bold text-lg mb-4">
              {editingId() ? t('common.edit') : t('common.add')} {t('app_detail.tab_environment')}
            </h3>
            
            <div class="form-control w-full mb-4">
              <label class="label">
                <span class="label-text">{t('app_detail.environment_key')}</span>
              </label>
              <input 
                type="text" 
                class="input input-bordered w-full" 
                value={newKey()}
                onInput={(e) => setNewKey(e.currentTarget.value)}
                placeholder="KEY"
              />
            </div>

            <div class="form-control w-full mb-4">
              <label class="label">
                <span class="label-text">{t('app_detail.environment_value')}</span>
              </label>
              <textarea 
                class="textarea textarea-bordered h-24" 
                value={newValue()}
                onInput={(e) => setNewValue(e.currentTarget.value)}
                placeholder={isEncrypted() && editingId() ? t('app_detail.environment_encrypted_placeholder') : "VALUE"}
              ></textarea>
              <Show when={isEncrypted() && editingId()}>
                <label class="label">
                   <span class="label-text-alt text-warning">{t('app_detail.environment_encrypted_warning')}</span>
                </label>
              </Show>
            </div>

            <div class="form-control mb-6">
              <label class="label cursor-pointer justify-start gap-4">
                <span class="label-text">{t('app_detail.environment_encrypted')}</span>
                <input 
                  type="checkbox" 
                  class="toggle" 
                  checked={isEncrypted()}
                  onChange={(e) => setIsEncrypted(e.currentTarget.checked)}
                />
              </label>
            </div>

            <div class="modal-action">
              <button class="btn btn-ghost" onClick={() => setIsModalOpen(false)}>
                {t('common.cancel')}
              </button>
              <button 
                class="btn btn-primary" 
                onClick={handleSave}
                disabled={mutations.createEnvVar.isPending || mutations.updateEnvVar.isPending}
              >
                {(mutations.createEnvVar.isPending || mutations.updateEnvVar.isPending) && <span class="loading loading-spinner loading-xs"></span>}
                {t('common.save')}
              </button>
            </div>
          </div>
          <div class="modal-backdrop" onClick={() => setIsModalOpen(false)}></div>
        </div>
      </Show>
    </div>
  )
}
