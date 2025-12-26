import { For, Show, JSX, createSignal } from 'solid-js'
import { useI18n } from '@i18n'
import { toast } from 'solid-toast'
import type { ApplicationToken, CreateApplicationTokenRequest } from '@types'

interface TokensTabProps {
  tokens: ApplicationToken[]
  isLoading: boolean
  onCreateToken: (data: CreateApplicationTokenRequest) => Promise<{ token: string } | undefined>
  onDeleteToken: (tokenId: string) => void
  isCreating?: boolean
  isDeleting?: boolean
}

export function TokensTab(props: TokensTabProps): JSX.Element {
  const { t } = useI18n()
  
  const [showCreateModal, setShowCreateModal] = createSignal(false)
  const [tokenName, setTokenName] = createSignal('')
  const [expiresIn, setExpiresIn] = createSignal<string>('')
  const [newToken, setNewToken] = createSignal<string | null>(null)
  const [tokenToDelete, setTokenToDelete] = createSignal<string | null>(null)

  const formatDate = (dateStr: string | undefined) => {
    if (!dateStr) return '-'
    try {
      return new Date(dateStr).toLocaleString()
    } catch {
      return dateStr
    }
  }

  const isExpired = (expiresAt: string | undefined) => {
    if (!expiresAt) return false
    return new Date(expiresAt) < new Date()
  }

  const handleCreateToken = async () => {
    if (!tokenName().trim()) return

    let expiresAt: string | undefined
    if (expiresIn()) {
      const daysToAdd = parseInt(expiresIn(), 10)
      if (!isNaN(daysToAdd) && daysToAdd > 0) {
        // Use milliseconds for accurate date calculation across month boundaries
        const expirationDate = new Date(Date.now() + daysToAdd * 24 * 60 * 60 * 1000)
        expiresAt = expirationDate.toISOString()
      }
    }

    const result = await props.onCreateToken({
      name: tokenName(),
      expires_at: expiresAt,
    })

    if (result?.token) {
      setNewToken(result.token)
      setTokenName('')
      setExpiresIn('')
    }
  }

  const handleCloseCreateModal = () => {
    setShowCreateModal(false)
    setTokenName('')
    setExpiresIn('')
    setNewToken(null)
  }

  const handleDeleteToken = (tokenId: string) => {
    setTokenToDelete(tokenId)
  }

  const confirmDelete = () => {
    const tokenId = tokenToDelete()
    if (tokenId) {
      props.onDeleteToken(tokenId)
      setTokenToDelete(null)
    }
  }

  const copyToClipboard = async (text: string) => {
    let success = false
    if (navigator.clipboard && navigator.clipboard.writeText) {
      try {
        await navigator.clipboard.writeText(text)
        success = true
      } catch {
        // Fall through to fallback
      }
    }
    if (!success) {
      // Fallback for older browsers or when clipboard API fails
      const textarea = document.createElement('textarea')
      textarea.value = text
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.focus()
      textarea.select()
      try {
        success = document.execCommand('copy')
      } finally {
        document.body.removeChild(textarea)
      }
    }
    if (success) {
      toast.success(t('app_detail.tokens_copied'))
    } else {
      toast.error(t('app_detail.tokens_copy_failed'))
    }
  }

  return (
    <div>
      {/* Header with Create Button */}
      <div class="flex justify-between items-center mb-4">
        <div>
          <h3 class="text-lg font-semibold">{t('app_detail.tokens_title')}</h3>
          <p class="text-sm text-base-content/60">
            {t('app_detail.tokens_description')}
          </p>
        </div>
        <button 
          class="btn btn-primary btn-sm"
          onClick={() => setShowCreateModal(true)}
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          {t('app_detail.tokens_create')}
        </button>
      </div>

      {/* Development Notice */}
      <div class="alert alert-info mb-4">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
        </svg>
        <span>{t('app_detail.tokens_development_notice')}</span>
      </div>

      {/* Tokens List */}
      <Show when={!props.isLoading} fallback={
        <div class="flex justify-center py-8">
          <span class="loading loading-spinner loading-md"></span>
        </div>
      }>
        <Show when={props.tokens.length > 0} fallback={
          <div class="text-center py-8 text-base-content/50">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto mb-4 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
            </svg>
            <p>{t('app_detail.tokens_empty')}</p>
            <p class="text-sm mt-2">{t('app_detail.tokens_empty_description')}</p>
          </div>
        }>
          <div class="overflow-x-auto">
            <table class="table">
              <thead>
                <tr>
                  <th>{t('app_detail.tokens_table_name')}</th>
                  <th>{t('app_detail.tokens_table_created')}</th>
                  <th>{t('app_detail.tokens_table_expires')}</th>
                  <th>{t('app_detail.tokens_table_last_used')}</th>
                  <th>{t('app_detail.tokens_table_actions')}</th>
                </tr>
              </thead>
              <tbody>
                <For each={props.tokens}>
                  {(token) => (
                    <tr class="hover">
                      <td class="font-medium">{token.name}</td>
                      <td class="text-sm">{formatDate(token.created_at)}</td>
                      <td>
                        <Show when={token.expires_at} fallback={
                          <span class="badge badge-ghost badge-sm">{t('app_detail.tokens_never')}</span>
                        }>
                          <span classList={{
                            'badge badge-sm': true,
                            'badge-error': isExpired(token.expires_at),
                            'badge-warning': !isExpired(token.expires_at),
                          }}>
                            {isExpired(token.expires_at) ? t('app_detail.tokens_expired') : formatDate(token.expires_at)}
                          </span>
                        </Show>
                      </td>
                      <td class="text-sm text-base-content/60">
                        {token.last_used_at ? formatDate(token.last_used_at) : t('app_detail.tokens_never_used')}
                      </td>
                      <td>
                        <button 
                          class="btn btn-ghost btn-xs text-error"
                          onClick={() => handleDeleteToken(token.uid)}
                          disabled={props.isDeleting}
                        >
                          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                          </svg>
                        </button>
                      </td>
                    </tr>
                  )}
                </For>
              </tbody>
            </table>
          </div>
        </Show>
      </Show>

      {/* Create Token Modal */}
      <Show when={showCreateModal()}>
        <div class="modal modal-open">
          <div class="modal-box">
            <h3 class="font-bold text-lg mb-4">
              {newToken() ? t('app_detail.tokens_modal_success_title') : t('app_detail.tokens_modal_create_title')}
            </h3>
            
            <Show when={!newToken()} fallback={
              <div class="space-y-4">
                <div class="alert alert-warning">
                  <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                  </svg>
                  <span>{t('app_detail.tokens_warning_copy')}</span>
                </div>
                <div class="form-control">
                  <label class="label">
                    <span class="label-text">{t('app_detail.tokens_label_token')}</span>
                  </label>
                  <div class="join w-full">
                    <input 
                      type="text" 
                      class="input input-bordered join-item flex-1 font-mono text-sm" 
                      value={newToken() || ''} 
                      readonly 
                    />
                    <button 
                      class="btn join-item"
                      onClick={() => copyToClipboard(newToken() || '')}
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                      </svg>
                      {t('app_detail.tokens_action_copy')}
                    </button>
                  </div>
                </div>
                <div class="modal-action">
                  <button class="btn btn-primary" onClick={handleCloseCreateModal}>{t('app_detail.tokens_action_done')}</button>
                </div>
              </div>
            }>
              <div class="space-y-4">
                <div class="form-control">
                  <label class="label">
                    <span class="label-text">{t('app_detail.tokens_label_name')}</span>
                  </label>
                  <input 
                    type="text" 
                    class="input input-bordered" 
                    placeholder={t('app_detail.tokens_placeholder_name') || ''}
                    value={tokenName()}
                    onInput={(e) => setTokenName(e.currentTarget.value)}
                    maxLength={100}
                  />
                  <label class="label">
                    <span class="label-text-alt">{t('app_detail.tokens_hint_name')}</span>
                  </label>
                </div>
                
                <div class="form-control">
                  <label class="label">
                    <span class="label-text">{t('app_detail.tokens_label_expiration')}</span>
                  </label>
                  <select 
                    class="select select-bordered"
                    value={expiresIn()}
                    onChange={(e) => setExpiresIn(e.currentTarget.value)}
                  >
                    <option value="">{t('app_detail.tokens_option_never')}</option>
                    <option value="7">{t('app_detail.tokens_option_7days')}</option>
                    <option value="30">{t('app_detail.tokens_option_30days')}</option>
                    <option value="90">{t('app_detail.tokens_option_90days')}</option>
                    <option value="180">{t('app_detail.tokens_option_180days')}</option>
                    <option value="365">{t('app_detail.tokens_option_1year')}</option>
                  </select>
                </div>

                <div class="modal-action">
                  <button class="btn btn-ghost" onClick={handleCloseCreateModal}>{t('app_detail.tokens_action_cancel')}</button>
                  <button 
                    class="btn btn-primary" 
                    onClick={handleCreateToken}
                    disabled={!tokenName().trim() || props.isCreating}
                  >
                    {props.isCreating ? (
                      <span class="loading loading-spinner loading-sm"></span>
                    ) : t('app_detail.tokens_action_create')}
                  </button>
                </div>
              </div>
            </Show>
          </div>
          <div class="modal-backdrop bg-black/50" onClick={handleCloseCreateModal}></div>
        </div>
      </Show>

      {/* Delete Confirmation Modal */}
      <Show when={tokenToDelete()}>
        <div class="modal modal-open">
          <div class="modal-box">
            <h3 class="font-bold text-lg">{t('app_detail.tokens_delete_title')}</h3>
            <p class="py-4">
              {t('app_detail.tokens_delete_message')}
            </p>
            <div class="modal-action">
              <button class="btn btn-ghost" onClick={() => setTokenToDelete(null)}>{t('app_detail.tokens_delete_cancel')}</button>
              <button 
                class="btn btn-error" 
                onClick={confirmDelete}
                disabled={props.isDeleting}
              >
                {props.isDeleting ? (
                  <span class="loading loading-spinner loading-sm"></span>
                ) : t('app_detail.tokens_delete_confirm')}
              </button>
            </div>
          </div>
          <div class="modal-backdrop bg-black/50" onClick={() => setTokenToDelete(null)}></div>
        </div>
      </Show>
    </div>
  )
}
