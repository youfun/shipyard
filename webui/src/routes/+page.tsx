import { JSX, createEffect } from 'solid-js'
import { useRouter } from '../router'
import { useAuth } from '../contexts/AuthContext'

// Root page - redirects to appropriate page
export default function RootPage(): JSX.Element {
  const router = useRouter()
  const auth = useAuth()

  createEffect(() => {
    if (!auth.isLoading()) {
      if (auth.isAuthenticated()) {
        router.navigate('/admin/dashboard')
      } else {
        const currentUrl = router.pathname() + router.search()
        router.navigate(`/login?redirect=${encodeURIComponent(currentUrl)}`)
      }
    }
  })

  return (
    <div class="min-h-screen flex items-center justify-center">
      <span class="loading loading-spinner loading-lg"></span>
    </div>
  )
}
