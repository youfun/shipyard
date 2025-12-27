/* @refresh reload */
import { render } from 'solid-js/web'
import { QueryClient, QueryClientProvider } from '@tanstack/solid-query'
import { Toaster } from 'solid-toast'
import { FileRouter } from '@router/FileRouter'
import { I18nProvider } from '@i18n'
import { AuthProvider } from '@contexts/AuthContext'
import { SetupProvider } from '@contexts/SetupContext'
import './index.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
})

const root = document.getElementById('root')

if (!root) {
  throw new Error('Root element not found')
}

render(
  () => (
    <QueryClientProvider client={queryClient}>
      <I18nProvider>
        <SetupProvider>
          <AuthProvider>
            <FileRouter />
            <Toaster />
          </AuthProvider>
        </SetupProvider>
      </I18nProvider>
    </QueryClientProvider>
  ),
  root
)
