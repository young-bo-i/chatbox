import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Text,
  Title,
  Stack,
  Paper,
  Badge,
  Group,
  Button,
  TextInput,
  Select,
  Switch,
  Modal,
  Textarea,
  ActionIcon,
  Alert,
  Loader,
  Tabs,
  Table,
} from '@mantine/core'
import {
  IconPlus,
  IconEdit,
  IconTrash,
  IconShieldCheck,
  IconAlertCircle,
  IconUsers,
  IconServer,
} from '@tabler/icons-react'
import { useAuthStore } from '@/stores/authStore'

export const Route = createFileRoute('/settings/admin/')({
  component: AdminSettings,
})

// 生产环境通过 Nginx 反向代理，使用相对路径；开发环境使用本地后端
const API_BASE_URL = process.env.API_BASE_URL || ''

interface Provider {
  id: number
  providerId: string
  name: string
  apiStyle: string
  apiHost?: string
  apiKey?: string
  enabled: boolean
  allowCustomKey: boolean
  models: Array<{
    modelId: string
    nickname?: string
    type?: string
    capabilities?: string[]
  }>
  isDefault: boolean
  sortOrder: number
}

interface User {
  id: number
  username: string
  role: string
  created_at: string
}

function AdminSettings() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { isAuthenticated, isAdmin, token } = useAuthStore()

  const [providers, setProviders] = useState<Provider[]>([])
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // 模态框状态
  const [modalOpen, setModalOpen] = useState(false)
  const [editingProvider, setEditingProvider] = useState<Provider | null>(null)

  // 表单状态
  const [formData, setFormData] = useState({
    providerId: '',
    name: '',
    apiStyle: 'openai',
    apiHost: '',
    apiKey: '',
    enabled: true,
    allowCustomKey: false,
    models: '',
    isDefault: false,
    sortOrder: 0,
  })

  // 检查权限
  useEffect(() => {
    if (!isAuthenticated || !isAdmin) {
      navigate({ to: '/' })
    }
  }, [isAuthenticated, isAdmin, navigate])

  // 获取数据
  useEffect(() => {
    if (isAdmin && token) {
      fetchProviders()
      fetchUsers()
    }
  }, [isAdmin, token])

  const fetchProviders = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/admin/providers`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (response.ok) {
        const data = await response.json()
        setProviders(data.providers || [])
      }
    } catch (err) {
      setError('Failed to fetch providers')
    } finally {
      setLoading(false)
    }
  }

  const fetchUsers = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/admin/users`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (response.ok) {
        const data = await response.json()
        setUsers(data.users || [])
      }
    } catch (err) {
      console.error('Failed to fetch users:', err)
    }
  }

  const handleOpenModal = (provider?: Provider) => {
    if (provider) {
      setEditingProvider(provider)
      setFormData({
        providerId: provider.providerId,
        name: provider.name,
        apiStyle: provider.apiStyle,
        apiHost: provider.apiHost || '',
        apiKey: provider.apiKey || '',
        enabled: provider.enabled,
        allowCustomKey: provider.allowCustomKey,
        models: JSON.stringify(provider.models || [], null, 2),
        isDefault: provider.isDefault,
        sortOrder: provider.sortOrder,
      })
    } else {
      setEditingProvider(null)
      setFormData({
        providerId: 'enter-ai',
        name: 'EnterAI',
        apiStyle: 'openai',
        apiHost: '',
        apiKey: '',
        enabled: true,
        allowCustomKey: false,
        models: '[]',
        isDefault: true,
        sortOrder: 0,
      })
    }
    setModalOpen(true)
  }

  const handleSaveProvider = async () => {
    try {
      let models = []
      try {
        models = JSON.parse(formData.models)
      } catch {
        alert('Invalid JSON for models')
        return
      }

      const payload = {
        providerId: formData.providerId,
        name: formData.name,
        apiStyle: formData.apiStyle,
        apiHost: formData.apiHost,
        apiKey: formData.apiKey,
        enabled: formData.enabled,
        allowCustomKey: formData.allowCustomKey,
        models,
        isDefault: formData.isDefault,
        sortOrder: formData.sortOrder,
      }

      const url = editingProvider
        ? `${API_BASE_URL}/api/admin/providers/${editingProvider.id}`
        : `${API_BASE_URL}/api/admin/providers`

      const response = await fetch(url, {
        method: editingProvider ? 'PUT' : 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      })

      if (response.ok) {
        setModalOpen(false)
        fetchProviders()
      } else {
        const data = await response.json()
        alert(data.error || 'Failed to save provider')
      }
    } catch (err) {
      alert('Failed to save provider')
    }
  }

  const handleDeleteProvider = async (id: number) => {
    if (!confirm('Are you sure you want to delete this provider?')) return

    try {
      const response = await fetch(`${API_BASE_URL}/api/admin/providers/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      })
      if (response.ok) {
        fetchProviders()
      }
    } catch (err) {
      alert('Failed to delete provider')
    }
  }

  if (!isAdmin) {
    return (
      <Alert icon={<IconAlertCircle size={16} />} color="red">
        {t('Admin access required')}
      </Alert>
    )
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader size="lg" />
      </div>
    )
  }

  return (
    <Stack gap="lg" p="md">
      <Group justify="space-between" align="center">
        <div>
          <Title order={3}>{t('Admin Panel')}</Title>
          <Text size="sm" c="dimmed">
            {t('Manage system configuration')}
          </Text>
        </div>
        <Badge color="green" leftSection={<IconShieldCheck size={14} />}>
          {t('Administrator')}
        </Badge>
      </Group>

      {error && (
        <Alert icon={<IconAlertCircle size={16} />} color="red" variant="light">
          {error}
        </Alert>
      )}

      <Tabs defaultValue="providers">
        <Tabs.List>
          <Tabs.Tab value="providers" leftSection={<IconServer size={16} />}>
            {t('Providers')}
          </Tabs.Tab>
          <Tabs.Tab value="users" leftSection={<IconUsers size={16} />}>
            {t('Users')}
          </Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="providers" pt="md">
          <Stack gap="md">
            <Group justify="space-between">
              <Text fw={500}>{t('System Providers')}</Text>
              <Button leftSection={<IconPlus size={16} />} onClick={() => handleOpenModal()}>
                {t('Add Provider')}
              </Button>
            </Group>

            {providers.length === 0 ? (
              <Paper p="xl" withBorder>
                <Text ta="center" c="dimmed">
                  {t('No providers configured')}
                </Text>
                <Text ta="center" size="sm" c="dimmed" mt="xs">
                  {t('Add a provider to get started')}
                </Text>
              </Paper>
            ) : (
              <Stack gap="xs">
                {providers.map((provider) => (
                  <Paper key={provider.id} p="md" withBorder>
                    <Group justify="space-between" align="flex-start">
                      <div>
                        <Group gap="xs">
                          <Text fw={500}>{provider.name}</Text>
                          {provider.isDefault && (
                            <Badge size="xs" color="green">
                              {t('Default')}
                            </Badge>
                          )}
                          {!provider.enabled && (
                            <Badge size="xs" color="gray">
                              {t('Disabled')}
                            </Badge>
                          )}
                        </Group>
                        <Text size="sm" c="dimmed">
                          ID: {provider.providerId} | Style: {provider.apiStyle}
                        </Text>
                        <Text size="sm" c="dimmed">
                          {t('Models')}: {provider.models?.length || 0} |{' '}
                          {provider.apiKey ? t('API Key Set') : t('No API Key')}
                        </Text>
                      </div>
                      <Group gap="xs">
                        <ActionIcon variant="subtle" onClick={() => handleOpenModal(provider)}>
                          <IconEdit size={16} />
                        </ActionIcon>
                        <ActionIcon
                          variant="subtle"
                          color="red"
                          onClick={() => handleDeleteProvider(provider.id)}
                        >
                          <IconTrash size={16} />
                        </ActionIcon>
                      </Group>
                    </Group>
                  </Paper>
                ))}
              </Stack>
            )}
          </Stack>
        </Tabs.Panel>

        <Tabs.Panel value="users" pt="md">
          <Stack gap="md">
            <Text fw={500}>{t('Registered Users')}</Text>
            <Table>
              <Table.Thead>
                <Table.Tr>
                  <Table.Th>{t('Username')}</Table.Th>
                  <Table.Th>{t('Role')}</Table.Th>
                  <Table.Th>{t('Created')}</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {users.map((user) => (
                  <Table.Tr key={user.id}>
                    <Table.Td>{user.username}</Table.Td>
                    <Table.Td>
                      <Badge color={user.role === 'admin' ? 'green' : 'blue'}>{user.role}</Badge>
                    </Table.Td>
                    <Table.Td>{new Date(user.created_at).toLocaleDateString()}</Table.Td>
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>
          </Stack>
        </Tabs.Panel>
      </Tabs>

      {/* Provider 编辑模态框 */}
      <Modal
        opened={modalOpen}
        onClose={() => setModalOpen(false)}
        title={editingProvider ? t('Edit Provider') : t('Add Provider')}
        size="lg"
      >
        <Stack gap="md">
          <TextInput
            label={t('Provider ID')}
            placeholder="enter-ai"
            value={formData.providerId}
            onChange={(e) => setFormData({ ...formData, providerId: e.target.value })}
            required
          />

          <TextInput
            label={t('Display Name')}
            placeholder="EnterAI"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            required
          />

          <Select
            label={t('API Style')}
            data={[
              { value: 'openai', label: 'OpenAI Compatible' },
              { value: 'google', label: 'Google Gemini' },
              { value: 'anthropic', label: 'Anthropic Claude' },
            ]}
            value={formData.apiStyle}
            onChange={(value) => setFormData({ ...formData, apiStyle: value || 'openai' })}
          />

          <TextInput
            label={t('API Host')}
            placeholder="https://api.openai.com"
            value={formData.apiHost}
            onChange={(e) => setFormData({ ...formData, apiHost: e.target.value })}
          />

          <TextInput
            label={t('API Key')}
            placeholder="sk-..."
            value={formData.apiKey}
            onChange={(e) => setFormData({ ...formData, apiKey: e.target.value })}
            type="password"
          />

          <Textarea
            label={t('Models (JSON)')}
            placeholder='[{"modelId": "gpt-4o", "nickname": "GPT-4o"}]'
            value={formData.models}
            onChange={(e) => setFormData({ ...formData, models: e.target.value })}
            rows={6}
            styles={{ input: { fontFamily: 'monospace' } }}
          />

          <Group>
            <Switch
              label={t('Enabled')}
              checked={formData.enabled}
              onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
            />
            <Switch
              label={t('Allow Custom Key')}
              checked={formData.allowCustomKey}
              onChange={(e) => setFormData({ ...formData, allowCustomKey: e.target.checked })}
            />
            <Switch
              label={t('Default Provider')}
              checked={formData.isDefault}
              onChange={(e) => setFormData({ ...formData, isDefault: e.target.checked })}
            />
          </Group>

          <Group justify="flex-end">
            <Button variant="subtle" onClick={() => setModalOpen(false)}>
              {t('Cancel')}
            </Button>
            <Button onClick={handleSaveProvider}>{t('Save')}</Button>
          </Group>
        </Stack>
      </Modal>
    </Stack>
  )
}
