import { JSX, Show, createSignal, createEffect, onCleanup } from 'solid-js'
import { useI18n } from '@i18n'
import { toast } from 'solid-toast'

interface LogsModalProps {
  show: boolean
  title: string
  onClose: () => void
  fetchLogs: (lines: number) => Promise<string>
  instanceUid?: string  // Optional: for real-time streaming
}

export function LogsModal(props: LogsModalProps): JSX.Element {
  const { t } = useI18n()
  const [logs, setLogs] = createSignal('')
  const [loadingLogs, setLoadingLogs] = createSignal(false)
  const [logLines, setLogLines] = createSignal(500)
  const [isRealTime, setIsRealTime] = createSignal(false)
  const [wsConnected, setWsConnected] = createSignal(false)

  let ws: WebSocket | null = null
  let logsContainerRef: HTMLDivElement | undefined
  let autoScroll = true

  const loadLogs = async () => {
    setLoadingLogs(true)
    try {
      const logContent = await props.fetchLogs(logLines())
      setLogs(logContent || t('app_detail.logs_empty'))
    } catch (e: any) {
      toast.error(e.response?.data?.error || t('common.error'))
      setLogs(t('app_detail.logs_error') + ': ' + (e.response?.data?.error || e.message))
    } finally {
      setLoadingLogs(false)
    }
  }

  const connectWebSocket = () => {
    if (!props.instanceUid) return

    // Disconnect existing connection
    disconnectWebSocket()

    // Get token from localStorage
    const token = localStorage.getItem('shipyard_access_token')
    if (!token) {
      toast.error(t('app_detail.logs_stream_error') + ': No authentication token')
      return
    }

    // Construct WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const wsUrl = `${protocol}//${host}/api/instances/${props.instanceUid}/logs/stream?lines=${logLines()}&token=${token}`

    try {
      ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        setWsConnected(true)
        setLogs('') // Clear previous logs
        console.log('WebSocket connected')
      }

      ws.onmessage = (event) => {
        const newLog = event.data
        setLogs(prev => prev + newLog)

        // Auto-scroll to bottom if enabled - use requestAnimationFrame for smoother scrolling
        if (autoScroll && logsContainerRef) {
          requestAnimationFrame(() => {
            if (logsContainerRef) {
              logsContainerRef.scrollTop = logsContainerRef.scrollHeight
            }
          })
        }
      }

      ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        toast.error(t('app_detail.logs_stream_error'))
        setWsConnected(false)
      }

      ws.onclose = () => {
        console.log('WebSocket disconnected')
        setWsConnected(false)
      }
    } catch (error) {
      console.error('Failed to create WebSocket:', error)
      toast.error(t('app_detail.logs_stream_error'))
    }
  }

  const disconnectWebSocket = () => {
    if (ws) {
      ws.close()
      ws = null
      setWsConnected(false)
    }
  }

  const toggleRealTime = () => {
    const newRealTimeState = !isRealTime()
    setIsRealTime(newRealTimeState)

    if (newRealTimeState) {
      connectWebSocket()
    } else {
      disconnectWebSocket()
      loadLogs() // Load static logs
    }
  }

  const handleRefreshLogs = () => {
    if (isRealTime()) {
      // Reconnect WebSocket
      connectWebSocket()
    } else {
      loadLogs()
    }
  }

  const handleCopyLogs = async () => {
    const textToCopy = logs()
    if (!textToCopy) return

    try {
      // Try Modern Async API
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(textToCopy)
        toast.success(t('app_detail.logs_copied'))
        return
      }
      throw new Error('Clipboard API unavailable')
    } catch (err) {
      // Fallback to legacy textarea method
      try {
        const textArea = document.createElement('textarea')
        textArea.value = textToCopy

        // Ensure textarea is not visible but part of DOM
        textArea.style.position = 'fixed'
        textArea.style.left = '-9999px'
        textArea.style.top = '0'
        document.body.appendChild(textArea)

        textArea.focus()
        textArea.select()

        const successful = document.execCommand('copy')
        document.body.removeChild(textArea)

        if (successful) {
          toast.success(t('app_detail.logs_copied'))
        } else {
          throw new Error('execCommand failed')
        }
      } catch (fallbackErr) {
        console.error('Copy failed:', fallbackErr)
        toast.error(t('app_detail.logs_copy_failed'))
      }
    }
  }

  const handleScroll = () => {
    if (!logsContainerRef) return

    // Check if user scrolled to bottom (within 50px threshold)
    const isAtBottom = logsContainerRef.scrollHeight - logsContainerRef.scrollTop - logsContainerRef.clientHeight < 50
    autoScroll = isAtBottom
  }

  // Load logs when modal is shown
  createEffect(() => {
    if (props.show) {
      if (isRealTime() && props.instanceUid) {
        connectWebSocket()
      } else {
        loadLogs()
      }
    } else {
      // Disconnect WebSocket when modal is closed
      disconnectWebSocket()
      setIsRealTime(false)
    }
  })

  // Cleanup on unmount
  onCleanup(() => {
    disconnectWebSocket()
  })

  return (
    <Show when={props.show}>
      <div class="modal modal-open">
        <div class="modal-box max-w-5xl h-[80vh] flex flex-col">
          <h3 class="font-bold text-lg mb-4">
            {props.title}
          </h3>

          <div class="flex gap-2 mb-4 flex-wrap">
            <select
              class="select select-bordered select-sm"
              value={logLines()}
              onChange={(e) => {
                setLogLines(parseInt(e.currentTarget.value))
              }}
              disabled={isRealTime()}
            >
              <option value="100">100 {t('app_detail.logs_lines')}</option>
              <option value="500">500 {t('app_detail.logs_lines')}</option>
              <option value="1000">1000 {t('app_detail.logs_lines')}</option>
              <option value="2000">2000 {t('app_detail.logs_lines')}</option>
            </select>

            <Show when={props.instanceUid}>
              <button
                class="btn btn-sm"
                classList={{
                  'btn-success': isRealTime(),
                  'btn-outline': !isRealTime()
                }}
                onClick={toggleRealTime}
              >
                <Show when={wsConnected()}>
                  <span class="inline-block w-2 h-2 bg-green-500 rounded-full mr-2 animate-pulse"></span>
                </Show>
                {isRealTime() ? t('app_detail.logs_stop_realtime') : t('app_detail.logs_start_realtime')}
              </button>
            </Show>

            <button
              class="btn btn-sm btn-primary"
              onClick={handleRefreshLogs}
              disabled={loadingLogs()}
            >
              {loadingLogs() ? t('common.loading') : t('app_detail.logs_refresh')}
            </button>
            <button
              class="btn btn-sm btn-secondary"
              onClick={handleCopyLogs}
              disabled={loadingLogs() || !logs()}
            >
              {t('app_detail.logs_copy')}
            </button>
          </div>

          <div
            ref={logsContainerRef}
            class="flex-1 overflow-auto bg-base-300 rounded p-4"
            onScroll={handleScroll}
          >
            <Show when={loadingLogs()}>
              <div class="flex justify-center items-center h-full">
                <span class="loading loading-spinner loading-lg"></span>
              </div>
            </Show>
            <Show when={!loadingLogs()}>
              <pre class="text-xs whitespace-pre-wrap font-mono">{logs()}</pre>
            </Show>
          </div>

          <div class="modal-action mt-4">
            <button class="btn" onClick={props.onClose}>
              {t('common.close')}
            </button>
          </div>
        </div>
      </div>
    </Show>
  )
}
