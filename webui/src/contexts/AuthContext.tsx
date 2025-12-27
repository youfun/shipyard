import { createContext, useContext, createSignal, createEffect, onMount, onCleanup, JSX } from 'solid-js'
import type { Component } from 'solid-js'
import type { User } from '@types'

interface AuthContextValue {
  user: () => User | null
  isAuthenticated: () => boolean
  isLoading: () => boolean
  accessToken: () => string | null
  setAccessToken: (token: string) => void
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue>()

const TOKEN_KEY = 'shipyard_access_token'

export const AuthProvider: Component<{ children: JSX.Element }> = (props) => {
  const [user, setUser] = createSignal<User | null>(null)
  const [accessToken, setAccessTokenInternal] = createSignal<string | null>(null)
  const [isLoading, setIsLoading] = createSignal(true)

  // Initialize from localStorage
  onMount(() => {
    const stored = localStorage.getItem(TOKEN_KEY)
    if (stored) {
      setAccessTokenInternal(stored)
    } else {
      setIsLoading(false)
    }
  })

  // Monitor localStorage changes (e.g., when token is cleared by API client)
  const handleStorageChange = () => {
    const currentToken = accessToken()
    const storedToken = localStorage.getItem(TOKEN_KEY)
    
    // Only log and act if there's an actual mismatch
    if (currentToken && !storedToken) {
      console.log('[AuthContext] Token was cleared externally, updating state')
      setAccessTokenInternal(null)
    } else if (!currentToken && storedToken) {
      console.log('[AuthContext] Token was set externally, updating state')
      setAccessTokenInternal(storedToken)
    }
  }

  onMount(() => {
    // Listen for storage events from other tabs
    window.addEventListener('storage', handleStorageChange)
    
    // Check for same-tab changes less frequently and only when needed
    let lastKnownToken = accessToken()
    const checkTokenChange = () => {
      const currentToken = accessToken()
      const storedToken = localStorage.getItem(TOKEN_KEY)
      
      // Only check if there might be a change
      if (lastKnownToken !== currentToken || (!!currentToken) !== (!!storedToken)) {
        handleStorageChange()
        lastKnownToken = currentToken
      }
    }
    
    const interval = setInterval(checkTokenChange, 5000) // Check every 5 seconds instead of 1
    
    onCleanup(() => {
      window.removeEventListener('storage', handleStorageChange)
      clearInterval(interval)
    })
  })

  // Fetch user when token changes
  createEffect(() => {
    const token = accessToken()
    if (token) {
      setIsLoading(true)
      fetchUser(token)
    } else {
      setUser(null)
      setIsLoading(false)
    }
  })

  const fetchUser = async (token: string) => {
    try {
      const response = await fetch('/api/auth/me', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })

      if (response.ok) {
        const data = await response.json()
        setUser(data)
      } else {
        console.error('[AuthContext] Fetch user failed:', response.status)
        // Token is invalid, clear it
        setAccessTokenInternal(null)
        localStorage.removeItem(TOKEN_KEY)
      }
    } catch (error) {
      console.error('Failed to fetch user:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const setAccessToken = (token: string) => {
    setAccessTokenInternal(token)
    localStorage.setItem(TOKEN_KEY, token)
  }

  const logout = async () => {
    const token = accessToken()
    if (token) {
      try {
        await fetch('/api/auth/logout', {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        })
      } catch (error) {
        console.error('Logout error:', error)
      }
    }
    
    setAccessTokenInternal(null)
    setUser(null)
    localStorage.removeItem(TOKEN_KEY)
    window.location.href = '/login'
  }

  const isAuthenticated = () => !!accessToken() && !!user()

  // Make token accessible globally for components that need it
  if (typeof window !== 'undefined') {
    (window as any).getAccessToken = () => accessToken()
  }

  const contextValue: AuthContextValue = {
    user,
    isAuthenticated,
    isLoading,
    accessToken,
    setAccessToken,
    logout,
  }

  return (
    <AuthContext.Provider value={contextValue}>
      {props.children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
