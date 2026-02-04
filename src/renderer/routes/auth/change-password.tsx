import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { createFileRoute } from '@tanstack/react-router'
import { Button, PasswordInput, Text, Title, Stack, Paper, Alert } from '@mantine/core'
import { IconAlertCircle, IconInfoCircle } from '@tabler/icons-react'
import { useTranslation } from 'react-i18next'
import { useAuthStore, authAPI } from '@/stores/authStore'

export const Route = createFileRoute('/auth/change-password')({
  component: ChangePasswordPage,
})

function ChangePasswordPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { token, user, setAuth, needChangePassword } = useAuthStore()

  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 如果未登录，跳转到登录页面
  if (!token || !user) {
    navigate({ to: '/auth/login' })
    return null
  }

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    if (newPassword !== confirmPassword) {
      setError(t('Passwords do not match'))
      return
    }

    if (newPassword.length < 6) {
      setError(t('Password must be at least 6 characters'))
      return
    }

    if (oldPassword === newPassword) {
      setError(t('New password must be different from old password'))
      return
    }

    setLoading(true)

    try {
      const { token: newToken, user: updatedUser } = await authAPI.changePassword(token, oldPassword, newPassword)
      setAuth(newToken, updatedUser)
      navigate({ to: '/' })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to change password')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex items-center justify-center min-h-screen bg-[var(--chatbox-background-secondary)]">
      <Paper shadow="md" p="xl" radius="md" className="w-full max-w-md">
        <form onSubmit={handleChangePassword}>
          <Stack gap="md">
            <Title order={2} ta="center">
              {t('Change Password')}
            </Title>

            {needChangePassword && (
              <Alert icon={<IconInfoCircle size={16} />} color="blue" variant="light">
                {t('For security reasons, please change your default password before continuing.')}
              </Alert>
            )}

            <Text c="dimmed" size="sm" ta="center">
              {t('Please enter your new password')}
            </Text>

            {error && (
              <Alert icon={<IconAlertCircle size={16} />} color="red" variant="light">
                {error}
              </Alert>
            )}

            <PasswordInput
              label={t('Current Password')}
              placeholder={t('Enter your current password')}
              value={oldPassword}
              onChange={(e) => setOldPassword(e.target.value)}
              required
            />

            <PasswordInput
              label={t('New Password')}
              placeholder={t('Enter your new password')}
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              required
              minLength={6}
            />

            <PasswordInput
              label={t('Confirm New Password')}
              placeholder={t('Confirm your new password')}
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
            />

            <Button type="submit" fullWidth loading={loading}>
              {t('Change Password')}
            </Button>

            {!needChangePassword && (
              <Button variant="subtle" fullWidth onClick={() => navigate({ to: '/' })}>
                {t('Cancel')}
              </Button>
            )}
          </Stack>
        </form>
      </Paper>
    </div>
  )
}
