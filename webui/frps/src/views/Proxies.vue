<template>
  <div class="proxies-page">
    <div class="page-header">
      <div class="header-top">
        <div class="title-section">
          <h1 class="page-title">{{ t('proxies.title') }}</h1>
          <p class="page-subtitle">{{ t('proxies.subtitle') }}</p>
        </div>

        <div class="actions-section">
          <ActionButton variant="outline" size="small" @click="fetchData">
            {{ t('common.refresh') }}
          </ActionButton>

          <ActionButton variant="outline" size="small" danger @click="showClearDialog = true">
            {{ t('proxies.clearOffline') }}
          </ActionButton>
        </div>
      </div>

      <div class="filter-section">
        <div class="search-row">
          <el-input
            v-model="searchText"
            :placeholder="t('common.searchProxies')"
            :prefix-icon="Search"
            clearable
            class="main-search"
          />

          <PopoverMenu
            :model-value="selectedClientKey"
            :width="220"
            placement="bottom-end"
            selectable
            filterable
            :filter-placeholder="t('common.searchClients')"
            :display-value="selectedClientLabel"
            clearable
            class="client-filter"
            @update:model-value="onClientFilterChange($event as string)"
          >
            <template #default="{ filterText }">
              <PopoverMenuItem value="">
                {{ t('common.allClients') }}
              </PopoverMenuItem>
              <PopoverMenuItem
                v-if="clientIDFilter && !selectedClientInList"
                :value="selectedClientKey"
              >
                {{ t('proxies.clientMissing', { label: missingClientLabel }) }}
              </PopoverMenuItem>
              <PopoverMenuItem
                v-for="client in filteredClientOptions(filterText)"
                :key="client.key"
                :value="client.key"
              >
                {{ client.label }}
              </PopoverMenuItem>
            </template>
          </PopoverMenu>
        </div>

        <div class="type-tabs">
          <button
            v-for="type in proxyTypes"
            :key="type.value"
            class="type-tab"
            :class="{ active: activeType === type.value }"
            @click="activeType = type.value"
          >
            {{ type.label }}
          </button>
        </div>
      </div>
    </div>

    <div v-loading="loading" class="proxies-content">
      <div v-if="filteredProxies.length > 0" class="proxies-list">
        <ProxyCard
          v-for="proxy in filteredProxies"
          :key="proxy.name"
          :proxy="proxy"
        />
      </div>
      <div v-else-if="!loading" class="empty-state">
        <el-empty :description="t('proxies.empty')" />
      </div>
    </div>

    <ConfirmDialog
      v-model="showClearDialog"
      :title="t('proxies.clearDialogTitle')"
      :message="t('proxies.clearDialogMessage')"
      :confirm-text="t('proxies.clearDialogConfirm')"
      danger
      @confirm="handleClearConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import ActionButton from '@shared/components/ActionButton.vue'
import ConfirmDialog from '@shared/components/ConfirmDialog.vue'
import PopoverMenu from '@shared/components/PopoverMenu.vue'
import PopoverMenuItem from '@shared/components/PopoverMenuItem.vue'
import { getClients } from '../api/client'
import {
  clearOfflineProxies as apiClearOfflineProxies,
  getProxiesByType,
} from '../api/proxy'
import { getServerInfo } from '../api/server'
import ProxyCard from '../components/ProxyCard.vue'
import { useI18n } from '../i18n'
import { Client } from '../utils/client'
import {
  BaseProxy,
  HTTPProxy,
  HTTPSProxy,
  STCPProxy,
  SUDPProxy,
  TCPMuxProxy,
  TCPProxy,
  UDPProxy,
} from '../utils/proxy'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const proxyTypes = [
  { label: 'TCP', value: 'tcp' },
  { label: 'UDP', value: 'udp' },
  { label: 'HTTP', value: 'http' },
  { label: 'HTTPS', value: 'https' },
  { label: 'TCPMUX', value: 'tcpmux' },
  { label: 'STCP', value: 'stcp' },
  { label: 'SUDP', value: 'sudp' },
]

const activeType = ref((route.params.type as string) || 'tcp')
const proxies = ref<BaseProxy[]>([])
const clients = ref<Client[]>([])
const loading = ref(false)
const searchText = ref('')
const showClearDialog = ref(false)
const clientIDFilter = ref((route.query.clientID as string) || '')
const userFilter = ref((route.query.user as string) || '')

const clientOptions = computed(() =>
  clients.value
    .map((client) => ({
      key: client.key,
      clientID: client.clientID,
      user: client.user,
      label: client.user ? `${client.user}.${client.clientID}` : client.clientID,
    }))
    .sort((a, b) => a.label.localeCompare(b.label)),
)

const selectedClientKey = computed(() => {
  if (!clientIDFilter.value) return ''
  const client = clientOptions.value.find(
    (item) =>
      item.clientID === clientIDFilter.value &&
      item.user === userFilter.value,
  )
  return client?.key || `${userFilter.value}:${clientIDFilter.value}`
})

const selectedClientLabel = computed(() => {
  if (!clientIDFilter.value) return t('common.allClients')
  const client = clientOptions.value.find(
    (item) =>
      item.clientID === clientIDFilter.value &&
      item.user === userFilter.value,
  )
  return (
    client?.label ||
    `${userFilter.value ? `${userFilter.value}.` : ''}${clientIDFilter.value}`
  )
})

const missingClientLabel = computed(
  () => `${userFilter.value ? `${userFilter.value}.` : ''}${clientIDFilter.value}`,
)

const filteredClientOptions = (filterText: string) => {
  if (!filterText) return clientOptions.value
  const search = filterText.toLowerCase()
  return clientOptions.value.filter((client) =>
    client.label.toLowerCase().includes(search),
  )
}

const selectedClientInList = computed(() => {
  if (!clientIDFilter.value) return true
  return clientOptions.value.some(
    (client) =>
      client.clientID === clientIDFilter.value &&
      client.user === userFilter.value,
  )
})

const filteredProxies = computed(() => {
  let result = proxies.value

  if (clientIDFilter.value) {
    result = result.filter(
      (proxy) =>
        proxy.clientID === clientIDFilter.value &&
        proxy.user === userFilter.value,
    )
  }

  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    result = result.filter((proxy) => proxy.name.toLowerCase().includes(search))
  }

  return result
})

const onClientFilterChange = (key: string) => {
  if (key) {
    const client = clientOptions.value.find((item) => item.key === key)
    if (client) {
      router.replace({
        query: { ...route.query, clientID: client.clientID, user: client.user },
      })
    }
  } else {
    const query = { ...route.query }
    delete query.clientID
    delete query.user
    router.replace({ query })
  }
}

const fetchClients = async () => {
  try {
    const json = await getClients()
    clients.value = json.map((data) => new Client(data))
  } catch {
    // Ignore client list failures for filter dropdown.
  }
}

let serverInfo: {
  vhostHTTPPort: number
  vhostHTTPSPort: number
  tcpmuxHTTPConnectPort: number
  subdomainHost: string
} | null = null

const fetchServerInfo = async () => {
  if (serverInfo) return serverInfo
  serverInfo = await getServerInfo()
  return serverInfo
}

const fetchData = async () => {
  loading.value = true
  proxies.value = []

  try {
    const type = activeType.value
    const json = await getProxiesByType(type)

    if (type === 'tcp') {
      proxies.value = json.proxies.map((proxy: any) => new TCPProxy(proxy))
    } else if (type === 'udp') {
      proxies.value = json.proxies.map((proxy: any) => new UDPProxy(proxy))
    } else if (type === 'http') {
      const info = await fetchServerInfo()
      if (info?.vhostHTTPPort) {
        proxies.value = json.proxies.map(
          (proxy: any) => new HTTPProxy(proxy, info.vhostHTTPPort, info.subdomainHost),
        )
      }
    } else if (type === 'https') {
      const info = await fetchServerInfo()
      if (info?.vhostHTTPSPort) {
        proxies.value = json.proxies.map(
          (proxy: any) => new HTTPSProxy(proxy, info.vhostHTTPSPort, info.subdomainHost),
        )
      }
    } else if (type === 'tcpmux') {
      const info = await fetchServerInfo()
      if (info?.tcpmuxHTTPConnectPort) {
        proxies.value = json.proxies.map(
          (proxy: any) =>
            new TCPMuxProxy(proxy, info.tcpmuxHTTPConnectPort, info.subdomainHost),
        )
      }
    } else if (type === 'stcp') {
      proxies.value = json.proxies.map((proxy: any) => new STCPProxy(proxy))
    } else if (type === 'sudp') {
      proxies.value = json.proxies.map((proxy: any) => new SUDPProxy(proxy))
    }
  } catch (error: any) {
    ElMessage({
      showClose: true,
      message: t('proxies.fetchFailed', { message: error.message }),
      type: 'error',
    })
  } finally {
    loading.value = false
  }
}

const clearOfflineProxies = async () => {
  try {
    await apiClearOfflineProxies()
    ElMessage({
      message: t('proxies.clearSuccess'),
      type: 'success',
    })
    fetchData()
  } catch (error: any) {
    ElMessage({
      message: t('proxies.clearFailed', { message: error.message }),
      type: 'warning',
    })
  }
}

const handleClearConfirm = async () => {
  showClearDialog.value = false
  await clearOfflineProxies()
}

watch(activeType, (newType) => {
  router.replace({ params: { type: newType }, query: route.query })
  fetchData()
})

watch(
  () => [route.query.clientID, route.query.user],
  ([newClientID, newUser]) => {
    clientIDFilter.value = (newClientID as string) || ''
    userFilter.value = (newUser as string) || ''
  },
)

fetchData()
fetchClients()
</script>

<style scoped>
.proxies-page {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.page-header {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.header-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 20px;
}

.title-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.page-title {
  font-size: 28px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin: 0;
  line-height: 1.2;
}

.page-subtitle {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  margin: 0;
}

.actions-section {
  display: flex;
  gap: 12px;
}

.filter-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-top: 8px;
}

.search-row {
  display: flex;
  gap: 16px;
  width: 100%;
  align-items: center;
}

.main-search {
  flex: 1;
}

.main-search :deep(.el-input__wrapper),
.client-filter :deep(.el-input__wrapper) {
  height: 32px;
  border-radius: 8px;
}

.client-filter {
  width: 240px;
}

.type-tabs {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 4px;
}

.type-tab {
  padding: 6px 16px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
  background: var(--el-bg-color);
  color: var(--el-text-color-regular);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  text-transform: uppercase;
}

.type-tab:hover {
  background: var(--el-fill-color-light);
}

.type-tab.active {
  background: var(--el-fill-color-darker);
  color: var(--el-text-color-primary);
  border-color: var(--el-fill-color-darker);
}

.proxies-content {
  min-height: 200px;
}

.proxies-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.empty-state {
  padding: 60px 0;
}

@media (max-width: 768px) {
  .search-row {
    flex-direction: column;
  }

  .client-filter {
    width: 100%;
  }
}
</style>
