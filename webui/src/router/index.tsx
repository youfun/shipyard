import { createSignal, createContext, useContext, onMount, onCleanup, JSX, Component, Accessor } from 'solid-js'

export interface RouteMatch {
  path: string
  params: Record<string, string>
}

export interface RouteModule {
  default?: Component
  Page?: Component
  Layout?: Component
  Error?: Component
  Loading?: Component
}

export interface RouterContextValue {
  pathname: Accessor<string>
  search: Accessor<string>
  hash: Accessor<string>
  params: Accessor<Record<string, string>>
  navigate: (url: string, options?: { replace?: boolean }) => void
}

const RouterContext = createContext<RouterContextValue>()

export function useRouter(): RouterContextValue {
  const context = useContext(RouterContext)
  if (!context) {
    throw new Error('useRouter must be used within a RouterProvider')
  }
  return context
}

export function usePathname(): Accessor<string> {
  return useRouter().pathname
}

export function useParams(): Accessor<Record<string, string>> {
  return useRouter().params
}

export function useSearchParams(): Accessor<URLSearchParams> {
  const router = useRouter()
  return () => new URLSearchParams(router.search())
}

interface RouterProviderProps {
  children: JSX.Element
}

export function RouterProvider(props: RouterProviderProps): JSX.Element {
  const [pathname, setPathname] = createSignal(window.location.pathname)
  const [search, setSearch] = createSignal(window.location.search)
  const [hash, setHash] = createSignal(window.location.hash)
  const [params, _setParams] = createSignal<Record<string, string>>({})

  const updateLocation = () => {
    setPathname(window.location.pathname)
    setSearch(window.location.search)
    setHash(window.location.hash)
  }

  const navigate = (url: string, options?: { replace?: boolean }) => {
    if (options?.replace) {
      window.history.replaceState(null, '', url)
    } else {
      window.history.pushState(null, '', url)
    }
    updateLocation()
  }

  onMount(() => {
    const handlePopstate = () => updateLocation()
    window.addEventListener('popstate', handlePopstate)
    onCleanup(() => window.removeEventListener('popstate', handlePopstate))
  })

  const context: RouterContextValue = {
    pathname,
    search,
    hash,
    params,
    navigate
  }

  return (
    <RouterContext.Provider value={context}>
      {props.children}
    </RouterContext.Provider>
  )
}

// Link component for navigation
interface LinkProps {
  href: string
  children: JSX.Element
  class?: string
  replace?: boolean
  onClick?: (e: MouseEvent) => void
}

export function Link(props: LinkProps): JSX.Element {
  const router = useRouter()

  const handleClick = (e: MouseEvent) => {
    // Allow default behavior for ctrl/cmd clicks (open in new tab)
    if (e.ctrlKey || e.metaKey || e.shiftKey) return
    
    e.preventDefault()
    props.onClick?.(e)
    router.navigate(props.href, { replace: props.replace })
  }

  return (
    <a href={props.href} class={props.class} onClick={handleClick}>
      {props.children}
    </a>
  )
}

// Route matching utilities
export function matchPath(pattern: string, pathname: string): RouteMatch | null {
  // Handle static routes
  if (!pattern.includes('[')) {
    return pattern === pathname ? { path: pattern, params: {} } : null
  }

  // Convert file-based route pattern to regex
  // /[id] -> /:id pattern
  // /[...slug] -> catch-all pattern
  const parts = pattern.split('/').filter(Boolean)
  const pathParts = pathname.split('/').filter(Boolean)
  
  const params: Record<string, string> = {}
  
  for (let i = 0; i < parts.length; i++) {
    const part = parts[i]
    const pathPart = pathParts[i]
    
    // Catch-all route [...slug]
    if (part.startsWith('[...') && part.endsWith(']')) {
      const paramName = part.slice(4, -1)
      params[paramName] = pathParts.slice(i).join('/')
      return { path: pattern, params }
    }
    
    // Dynamic route [id]
    if (part.startsWith('[') && part.endsWith(']')) {
      if (!pathPart) return null
      const paramName = part.slice(1, -1)
      params[paramName] = pathPart
      continue
    }
    
    // Static segment
    if (part !== pathPart) return null
  }
  
  // Check if all path parts were matched
  if (parts.length !== pathParts.length) return null
  
  return { path: pattern, params }
}

// Helper to convert file path to route pattern
export function filePathToRoutePattern(filePath: string): string {
  // Remove 'routes' prefix and file extension
  // Support both +page.tsx and page.tsx styles
  let pattern = filePath
    .replace(/^\.?\/?(src\/)?routes/, '')
    .replace(/\/\+?(page|layout|error|loading)\.(tsx|ts|jsx|js)$/, '')
    .replace(/\/index$/, '')
  
  // Handle root route
  if (!pattern || pattern === '') {
    pattern = '/'
  }
  
  // Ensure pattern starts with /
  if (!pattern.startsWith('/')) {
    pattern = '/' + pattern
  }
  
  return pattern
}
