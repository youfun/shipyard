import { createSignal, JSX, createEffect, Show, onMount } from 'solid-js'
import { Link, useRouter, usePathname } from '@router'
import { useAuth } from '@contexts/AuthContext'
import { useI18n } from '@i18n'
import { fetchSystemStatus, type SystemStatus } from '@api/services/systemService'

interface AdminLayoutProps {
  children?: JSX.Element
}

// Language Switcher Component
function LanguageSwitcher() {
  const { locale, setLocale, t } = useI18n()

  return (
    <div class="dropdown dropdown-end">
      <div tabindex="0" role="button" class="btn btn-ghost btn-sm">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M10.5 21l5.25-11.25L21 21m-9-3h7.5M3 5.621a48.474 48.474 0 016-.371m0 0c1.12 0 2.233.038 3.334.114M9 5.25V3m3.334 2.364C11.176 10.658 7.69 15.08 3 17.502m9.334-12.138c.896.061 1.785.147 2.666.257m-4.589 8.495a18.023 18.023 0 01-3.827-5.802" />
        </svg>
        {locale() === 'zh' ? 'ZH' : 'EN'}
      </div>
      <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-32">
        <li><a onClick={() => setLocale('zh')} classList={{ 'active': locale() === 'zh' }}>{t('language.chinese')}</a></li>
        <li><a onClick={() => setLocale('en')} classList={{ 'active': locale() === 'en' }}>{t('language.english')}</a></li>
      </ul>
    </div>
  )
}

export default function AdminLayout(props: AdminLayoutProps): JSX.Element {
  const auth = useAuth()
  const { t } = useI18n()
  const router = useRouter()
  const pathname = usePathname()
  const [sidebarOpen, setSidebarOpen] = createSignal(false)
  const [systemStatus, setSystemStatus] = createSignal<SystemStatus | null>(null)

  // Fetch system status on mount
  onMount(async () => {
    try {
      const status = await fetchSystemStatus()
      setSystemStatus(status)
    } catch (error) {
      console.error('Failed to fetch system status:', error)
    }
  })

  createEffect(() => {
    // Redirect to login if not authenticated and not loading
    if (!auth.isLoading() && !auth.isAuthenticated()) {
      // Save current URL for redirect after login
      const currentUrl = router.pathname() + router.search()
      const redirectUrl = `/login?redirect=${encodeURIComponent(currentUrl)}`
      router.navigate(redirectUrl)
    }
  })

  const handleLogout = () => {
    void auth.logout()
  }

  // Get current page from the current route
  const getCurrentPage = () => {
    const path = pathname()
    if (path === '/admin/dashboard') return 'dashboard'
    if (path.startsWith('/admin/apps')) return 'applications'
    if (path.startsWith('/admin/ssh-management')) return 'ssh-management'
    if (path.startsWith('/admin/settings')) return 'settings'
    return path.replace('/admin/', '')
  }

  const menuItems = [
    { id: 'dashboard', name: String(t('nav.dashboard') || 'Dashboard'), href: '/admin/dashboard', icon: 'M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z' },
    { id: 'applications', name: String(t('nav.applications') || 'Applications'), href: '/admin/apps', icon: 'M3.375 19.5h17.25m-17.25 0a1.125 1.125 0 01-1.125-1.125M3.375 19.5h7.5c.621 0 1.125-.504 1.125-1.125m-9.75 0V5.625m0 12.75v-1.5c0-.621.504-1.125 1.125-1.125m18.375 2.625V5.625m0 12.75c0 .621-.504 1.125-1.125 1.125m1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125m0 3.75h-7.5A1.125 1.125 0 0112 18.375m9.75-12.75c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125m19.5 0v1.5c0 .621-.504 1.125-1.125 1.125M2.25 5.625v1.5c0 .621.504 1.125 1.125 1.125m0 0h17.25m-17.25 0h7.5c.621 0 1.125.504 1.125 1.125M3.375 8.25v1.5c0 .621.504 1.125 1.125 1.125m17.25-2.625h-7.5c-.621 0-1.125.504-1.125 1.125' },
    { id: 'ssh-management', name: String(t('nav.ssh_management') || 'SSH Management'), href: '/admin/ssh-management', icon: 'M6.75 7.5l3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0021 18V6a2.25 2.25 0 00-2.25-2.25H5.25A2.25 2.25 0 003 6v12a2.25 2.25 0 002.25 2.25z' },
    { id: 'settings', name: String(t('nav.settings') || 'Settings'), href: '/admin/settings', icon: 'M10.5 21l5.25-11.25L21 21m-9-3h7.5M3 5.621a48.474 48.474 0 016-.371m0 0c1.12 0 2.233.038 3.334.114M9 5.25V3m3.334 2.364C11.176 10.658 7.69 15.08 3 17.502m9.334-12.138c.896.061 1.785.147 2.666.257m-4.589 8.495a18.023 18.023 0 01-3.827-5.802' },
  ]

  return (
    <Show 
      when={!auth.isLoading() && auth.isAuthenticated()} 
      fallback={
        <div class="min-h-screen flex items-center justify-center">
          <span class="loading loading-spinner loading-lg"></span>
        </div>
      }
    >
      <div class="drawer lg:drawer-open">
        <input 
          id="drawer-toggle" 
          type="checkbox" 
          class="drawer-toggle" 
          checked={sidebarOpen()}
          onChange={(e) => setSidebarOpen(e.currentTarget.checked)}
        />
        
        <div class="drawer-content flex flex-col">
          {/* Top Navigation */}
          <div class="navbar bg-base-100 border-b border-base-300">
            <div class="flex-none lg:hidden">
              <label for="drawer-toggle" class="btn btn-square btn-ghost">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-6 h-6 stroke-current">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path>
                </svg>
              </label>
            </div>
            
            <div class="flex-1">
              <a class="btn btn-ghost text-xl">{t('nav.admin_title')}</a>
            </div>
            
            <div class="flex-none">
              <LanguageSwitcher />
              <div class="dropdown dropdown-end ml-2">
                <div tabindex="0" role="button" class="btn btn-ghost btn-circle avatar">
                  <div class="w-10 rounded-full bg-base-300 flex items-center justify-center">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-6 h-6 stroke-current">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>
                    </svg>
                  </div>
                </div>
                <ul tabindex="0" class="mt-3 z-[1] p-2 shadow menu menu-sm dropdown-content bg-base-100 rounded-box w-52">
                  <li class="menu-title">
                    <span>{auth.user()?.username}</span>
                  </li>
                  {/* <li><Link href="/admin/change-password">{t('nav.change_password')}</Link></li> */}
                  <li><a onClick={handleLogout}>{t('nav.logout')}</a></li>
                </ul>
              </div>
            </div>
          </div>

          {/* Main Content */}
          <main class="flex-1 overflow-y-auto bg-base-200">
            {props.children}
          </main>
        </div>
        
        {/* Sidebar */}
        <div class="drawer-side">
          <label for="drawer-toggle" aria-label="close sidebar" class="drawer-overlay"></label>
          <aside class="min-h-full w-64 bg-base-100 border-r border-base-300">
            <div class="p-4">
              <div class="text-xl font-bold text-center">{t('nav.system_title')}</div>
            </div>
            
            <ul class="menu p-4 w-full">
              {menuItems.map(item => (
                <li>
                  <Link 
                    href={item.href}
                    class={getCurrentPage() === item.id ? 'active' : ''}
                    onClick={() => setSidebarOpen(false)}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-5 h-5 stroke-current">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={item.icon}></path>
                    </svg>
                    {item.name}
                  </Link>
                </li>
              ))}
            </ul>
            
            {/* Version Info at Bottom */}
            <div class="absolute bottom-0 left-0 right-0 p-4 border-t border-base-300">
              <Show when={systemStatus()}>
                <div class="text-xs text-base-content/60 text-center space-y-1">
                  <div class="flex items-center justify-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 3.104v5.714a2.25 2.25 0 01-.659 1.591L5 14.5M9.75 3.104c-.251.023-.501.05-.75.082m.75-.082a24.301 24.301 0 014.5 0m0 0v5.714c0 .597.237 1.17.659 1.591L19.8 15.3M14.25 3.104c.251.023.501.05.75.082M19.8 15.3l-1.57.393A9.065 9.065 0 0112 15a9.065 9.065 0 00-6.23-.693L5 14.5m14.8.8l1.402 1.402c1.232 1.232.65 3.318-1.067 3.611A48.309 48.309 0 0112 21c-2.773 0-5.491-.235-8.135-.687-1.718-.293-2.3-2.379-1.067-3.61L5 14.5" />
                    </svg>
                    <span>{t('nav.version')}: {systemStatus()?.version || 'Unknown'}</span>
                  </div>
                </div>
              </Show>
            </div>
          </aside>
        </div>
      </div>
    </Show>
  )
}