import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import storage from '@/storage'

export interface User {
  id: number
  username: string
  role: 'admin' | 'user'
  password_changed: boolean
  created_at: string
}

interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  isAdmin: boolean
  needChangePassword: boolean

  // Actions
  setAuth: (token: string, user: User) => void
  clearAuth: () => void
  checkAuth: () => Promise<void>
}

// 生产环境通过 Nginx 反向代理，使用相对路径；开发环境使用本地后端
const API_BASE_URL = process.env.API_BASE_URL || ''

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      isAuthenticated: false,
      isAdmin: false,
      needChangePassword: false,

      setAuth: (token: string, user: User) => {
        set({
          token,
          user,
          isAuthenticated: true,
          isAdmin: user.role === 'admin',
          needChangePassword: !user.password_changed,
        })
      },

      clearAuth: () => {
        set({
          token: null,
          user: null,
          isAuthenticated: false,
          isAdmin: false,
          needChangePassword: false,
        })
      },

      checkAuth: async () => {
        const { token } = get()
        if (!token) {
          return
        }

        try {
          const response = await fetch(`${API_BASE_URL}/api/auth/me`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          })

          if (response.ok) {
            const user = await response.json()
            set({
              user,
              isAuthenticated: true,
              isAdmin: user.role === 'admin',
              needChangePassword: !user.password_changed,
            })
          } else {
            // Token 无效，清除认证状态
            get().clearAuth()
          }
        } catch (error) {
          console.error('Failed to check auth:', error)
        }
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => ({
        getItem: async (key) => {
          const value = await storage.getItem(key, null)
          return JSON.stringify(value)
        },
        setItem: async (key, value) => {
          await storage.setItem(key, JSON.parse(value))
        },
        removeItem: async (key) => {
          await storage.removeItem(key)
        },
      })),
      partialize: (state) => ({
        token: state.token,
        user: state.user,
      }),
    }
  )
)

// API 函数
export const authAPI = {
  async login(username: string, password: string) {
    const response = await fetch(`${API_BASE_URL}/api/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Login failed')
    }

    return response.json()
  },

  async changePassword(token: string, oldPassword: string, newPassword: string) {
    const response = await fetch(`${API_BASE_URL}/api/auth/change-password`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ oldPassword, newPassword }),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Failed to change password')
    }

    return response.json()
  },
}
