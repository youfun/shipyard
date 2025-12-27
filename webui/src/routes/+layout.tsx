import { JSX, Show, createEffect, createSignal } from 'solid-js'
import { useRouter } from '../router'
import { useAuth } from '../contexts/AuthContext'
import { SetupGuard } from '@components/SetupGuard'

interface LayoutProps {
  children?: JSX.Element
}

// Root layout - handles auth check and redirects
export function Layout(props: LayoutProps): JSX.Element {
  const router = useRouter()
  const auth = useAuth()
  const [checked, setChecked] = createSignal(false)

  createEffect(() => {
    // Skip auth check for public routes
    const publicRoutes = ['/login', '/setup', '/cli-device-auth']
    const currentPath = router.pathname()
    const isPublicRoute = publicRoutes.some(route => currentPath.startsWith(route))
    
    // Wait for loading to finish
    if (!auth.isLoading()) {
      if (!isPublicRoute && !auth.isAuthenticated()) {
        // Redirect to login with the current URL as redirect parameter
        const currentUrl = router.pathname() + router.search()
        const redirectUrl = `/login?redirect=${encodeURIComponent(currentUrl)}`
        router.navigate(redirectUrl)
      }
      setChecked(true)
    }
  })

  return (
    <SetupGuard>
      <div class="min-h-screen bg-base-200">
        <Show when={checked()} fallback={
          <div class="min-h-screen flex items-center justify-center">
            <span class="loading loading-spinner loading-lg"></span>
          </div>
        }>
          {props.children}
        </Show>
      </div>
    </SetupGuard>
  )
}

export default Layout
