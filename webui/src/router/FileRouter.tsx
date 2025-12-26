import { createMemo, Component, JSX, Suspense, ErrorBoundary } from 'solid-js'
import { RouterProvider, useRouter, matchPath, filePathToRoutePattern, RouteModule } from './index'

// Import all route modules using Vite's glob import
// Support +page.tsx, +layout.tsx, +error.tsx, +loading.tsx style
const routeModules = import.meta.glob<RouteModule>('/src/routes/**/*.tsx', { eager: true })

interface RouteDefinition {
  pattern: string
  page?: Component
  layout?: Component
  error?: Component
  loading?: Component
  depth: number
}

function buildRoutes(): Map<string, RouteDefinition> {
  const routes = new Map<string, RouteDefinition>()
  
  for (const [filePath, module] of Object.entries(routeModules)) {
    // Only process +page, +layout, +error, +loading files
    if (!filePath.match(/\/\+(page|layout|error|loading)\.tsx$/)) {
      continue
    }
    
    const pattern = filePathToRoutePattern(filePath)
    const existing = routes.get(pattern) || { 
      pattern, 
      depth: pattern.split('/').filter(Boolean).length 
    }
    
    const PageComponent = module.default || module.Page
    const LayoutComponent = module.default || module.Layout
    const ErrorComponent = module.default || module.Error
    const LoadingComponent = module.default || module.Loading
    
    if (filePath.includes('/+page.tsx')) {
      existing.page = PageComponent
    }
    if (filePath.includes('/+layout.tsx')) {
      existing.layout = LayoutComponent
    }
    if (filePath.includes('/+error.tsx')) {
      existing.error = ErrorComponent
    }
    if (filePath.includes('/+loading.tsx')) {
      existing.loading = LoadingComponent
    }
    
    routes.set(pattern, existing)
  }
  
  console.log('[FileRouter] Routes built:', routes)
  return routes
}

const routes = buildRoutes()

// Get layouts for a given path (including parent layouts)
function getLayoutsForPath(pathname: string): Component[] {
  const layouts: Component[] = []
  const segments = pathname.split('/').filter(Boolean)
  
  // Check root layout
  const rootRoute = routes.get('/')
  if (rootRoute?.layout) {
    layouts.push(rootRoute.layout)
  }
  
  // Check nested layouts
  let currentPath = ''
  for (const segment of segments) {
    currentPath += '/' + segment
    const route = routes.get(currentPath)
    if (route?.layout) {
      layouts.push(route.layout)
    }
  }
  
  console.log(`[FileRouter] Layouts for ${pathname}:`, layouts.length)
  return layouts
}

// Find matching route for pathname
function findMatchingRoute(pathname: string): RouteDefinition | null {
  console.log('[FileRouter] Finding route for:', pathname)
  // First try exact match
  const exact = routes.get(pathname)
  if (exact?.page) {
    console.log('[FileRouter] Exact match found:', exact)
    return exact
  }
  
  // Then try pattern matching for dynamic routes
  for (const [pattern, route] of routes) {
    const match = matchPath(pattern, pathname)
    if (match && route.page) {
      console.log('[FileRouter] Dynamic match found:', pattern, match.params)
      return route
    }
  }
  
  console.warn('[FileRouter] No match found for:', pathname)
  return null
}

// Default error fallback component
function DefaultErrorFallback(props: { error: Error; reset: () => void }): JSX.Element {
  return (
    <div class="min-h-screen flex items-center justify-center bg-base-200">
      <div class="card bg-base-100 shadow-xl max-w-md w-full">
        <div class="card-body">
          <h2 class="card-title text-error">Error</h2>
          <p class="text-base-content/70">{props.error.message}</p>
          <div class="card-actions justify-end">
            <button 
              onClick={props.reset}
              class="btn btn-primary"
            >
              Retry
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

// Default loading component
function DefaultLoading(): JSX.Element {
  return (
    <div class="min-h-screen flex items-center justify-center bg-base-200">
      <span class="loading loading-spinner loading-lg"></span>
    </div>
  )
}

// Default 404 component
function NotFound(): JSX.Element {
  return (
    <div class="min-h-screen flex items-center justify-center bg-base-200">
      <div class="card bg-base-100 shadow-xl max-w-md w-full text-center">
        <div class="card-body">
          <h1 class="text-6xl font-bold text-base-content/30 mb-4">404</h1>
          <h2 class="card-title justify-center">Page Not Found</h2>
          <p class="text-base-content/70">The page you are looking for does not exist</p>
          <div class="card-actions justify-center mt-4">
            <a href="/" class="btn btn-primary">
              Back to Home
            </a>
          </div>
        </div>
      </div>
    </div>
  )
}

interface OutletProps {
  children?: JSX.Element
}

// Outlet component to render nested content
export function Outlet(props: OutletProps): JSX.Element {
  return <>{props.children}</>
}

// Internal router view
function RouterView(): JSX.Element {
  const router = useRouter()
  
  const matchedRoute = createMemo(() => {
    const pathname = router.pathname()
    return findMatchingRoute(pathname)
  })
  
  const layouts = createMemo(() => {
    return getLayoutsForPath(router.pathname())
  })
  
  // Nest content within layouts - REACTIVE
  const view = createMemo(() => {
    const route = matchedRoute()
    const layoutComponents = layouts()
    
    let content: JSX.Element
    
    // 1. Render Page
    if (!route || !route.page) {
      content = <NotFound />
    } else {
      const PageComponent = route.page
      const ErrorComponent = route.error || DefaultErrorFallback
      const LoadingComponent = route.loading || DefaultLoading
      
      content = (
        <ErrorBoundary fallback={(err, reset) => <ErrorComponent error={err} reset={reset} />}>
          <Suspense fallback={<LoadingComponent />}>
            <PageComponent />
          </Suspense>
        </ErrorBoundary>
      )
    }
    
    // 2. Wrap with Layouts (Innermost -> Outermost)
    for (let i = layoutComponents.length - 1; i >= 0; i--) {
      const LayoutComponent = layoutComponents[i] as Component<{ children?: JSX.Element }>
      const currentContent = content
      content = (
        <LayoutComponent>
          <Outlet>{currentContent}</Outlet>
        </LayoutComponent>
      )
    }
    
    return content
  })
  
  return <>{view()}</>
}

interface FileRouterProps {
  children?: JSX.Element
}

// Main FileRouter component
export function FileRouter(props: FileRouterProps): JSX.Element {
  return (
    <RouterProvider>
      <RouterView />
      {props.children}
    </RouterProvider>
  )
}

export default FileRouter
