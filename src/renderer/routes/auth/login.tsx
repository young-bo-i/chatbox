import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { createFileRoute } from '@tanstack/react-router'
import { Button, TextInput, PasswordInput, Text, Title, Stack, Paper, Anchor, Alert } from '@mantine/core'
import { IconAlertCircle } from '@tabler/icons-react'
import { useTranslation } from 'react-i18next'
import { useAuthStore, authAPI } from '@/stores/authStore'

export const Route = createFileRoute('/auth/login')({
  component: LoginPage,
})

function LoginPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const setAuth = useAuthStore((state) => state.setAuth)

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setLoading(true)

    try {
      const { token, user } = await authAPI.login(username, password)
      setAuth(token, user)
      
      // 如果需要修改密码，跳转到修改密码页面
      if (!user.password_changed) {
        navigate({ to: '/auth/change-password' })
      } else {
        navigate({ to: '/' })
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex items-center justify-center min-h-screen bg-[var(--chatbox-background-secondary)]">
      <Paper shadow="md" p="xl" radius="md" className="w-full max-w-md">
        <form onSubmit={handleLogin}>
          <Stack gap="md">
            <Title order={2} ta="center">
              {t('Login')}
            </Title>

            <Text c="dimmed" size="sm" ta="center">
              {t('Sign in to your account')}
            </Text>

            {error && (
              <Alert icon={<IconAlertCircle size={16} />} color="red" variant="light">
                {error}
              </Alert>
            )}

            <TextInput
              label={t('Username')}
              placeholder={t('Enter your username')}
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
            />

            <PasswordInput
              label={t('Password')}
              placeholder={t('Enter your password')}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />

            <Button type="submit" fullWidth loading={loading}>
              {t('Sign in')}
            </Button>

            <Text ta="center" size="sm">
              <Anchor component="button" type="button" onClick={() => navigate({ to: '/' })}>
                {t('Continue without login')}
              </Anchor>
            </Text>
          </Stack>
        </form>
      </Paper>
    </div>
  )
}
