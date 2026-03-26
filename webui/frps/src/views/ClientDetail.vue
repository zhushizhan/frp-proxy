<template>
  <div class="client-detail-page">
    <nav class="breadcrumb">
      <a class="breadcrumb-link" @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
      </a>
      <router-link to="/clients" class="breadcrumb-item">
        {{ t('clientDetail.clients') }}
      </router-link>
      <span class="breadcrumb-separator">/</span>
      <span class="breadcrumb-current">{{ client?.displayName || route.params.key }}</span>
    </nav>

    <div v-loading="loading" class="detail-content">
      <template v-if="client">
        <div class="header-card">
          <div class="header-main">
            <div class="header-left">
              <div class="client-avatar">
                {{ client.displayName.charAt(0).toUpperCase() }}
              </div>
              <div class="client-info">
                <div class="client-name-row">
                  <h1 class="client-name">{{ client.displayName }}</h1>
                  <el-tag v-if="client.version" size="small" type="success">
                    v{{ client.version }}
                  </el-tag>
                </div>
                <div class="client-meta">
                  <span v-if="client.ip" class="meta-item">{{ client.ip }}</span>
                  <span v-if="client.hostname" class="meta-item">{{ client.hostname }}</span>
                </div>
              </div>
            </div>
            <div class="header-right">
              <span class="status-badge" :class="client.online ? 'online' : 'offline'">
                {{ client.online ? t('status.online') : t('status.offline') }}
              </span>
            </div>
          </div>

          <div class="info-section">
            <div class="info-item">
              <span class="info-label">{{ t('clientDetail.connections') }}</span>
              <span class="info-value">{{ totalConnections }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t('clientDetail.runId') }}</span>
              <span class="info-value">{{ client.runID }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t('clientDetail.firstConnected') }}</span>
              <span class="info-value">{{ client.firstConnectedAgo }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">
                {{ client.online ? t('clientDetail.connected') : t('clientDetail.disconnected') }}
              </span>
              <span class="info-value">
                {{ client.online ? client.lastConnectedAgo : client.disconnectedAgo }}
              </span>
            </div>
          </div>
        </div>

        <div class="proxies-card">
          <div class="proxies-header">
            <div class="proxies-title">
              <h2>{{ t('clientDetail.proxies') }}</h2>
              <span class="proxies-count">{{ filteredProxies.length }}</span>
            </div>
            <el-input
              v-model="proxySearch"
              :placeholder="t('common.searchProxies')"
              :prefix-icon="Search"
              clearable
              class="proxy-search"
            />
          </div>
          <div class="proxies-body">
            <div v-if="proxiesLoading" class="loading-state">
              <el-icon class="is-loading"><Loading /></el-icon>
              <span>{{ t('clientDetail.loading') }}</span>
            </div>
            <div v-else-if="filteredProxies.length > 0" class="proxies-list">
              <ProxyCard
                v-for="proxy in filteredProxies"
                :key="proxy.name"
                :proxy="proxy"
                show-type
              />
            </div>
            <div v-else-if="clientProxies.length > 0" class="empty-state">
              <p>{{ t('clientDetail.noProxiesMatch', { search: proxySearch }) }}</p>
            </div>
            <div v-else class="empty-state">
              <p>{{ t('clientDetail.noProxiesFound') }}</p>
            </div>
          </div>
        </div>
      </template>

      <div v-else-if="!loading" class="not-found">
        <h2>{{ t('clientDetail.notFoundTitle') }}</h2>
        <p>{{ t('clientDetail.notFoundMessage') }}</p>
        <router-link to="/clients">
          <el-button type="primary">{{ t('common.backToClients') }}</el-button>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Loading, Search } from '@element-plus/icons-vue'
import { getClient } from '../api/client'
import { getProxiesByType } from '../api/proxy'
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

const client = ref<Client | null>(null)
const loading = ref(true)
const proxiesLoading = ref(false)
const allProxies = ref<BaseProxy[]>([])
const proxySearch = ref('')

let serverInfo: {
  vhostHTTPPort: number
  vhostHTTPSPort: number
  tcpmuxHTTPConnectPort: number
  subdomainHost: string
} | null = null

const goBack = () => {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/clients')
  }
}

const clientProxies = computed(() => {
  if (!client.value) return []
  return allProxies.value.filter(
    (proxy) =>
      proxy.clientID === client.value?.clientID &&
      proxy.user === client.value?.user,
  )
})

const filteredProxies = computed(() => {
  if (!proxySearch.value) return clientProxies.value
  const search = proxySearch.value.toLowerCase()
  return clientProxies.value.filter(
    (proxy) =>
      proxy.name.toLowerCase().includes(search) ||
      proxy.type.toLowerCase().includes(search),
  )
})

const totalConnections = computed(() =>
  clientProxies.value.reduce((sum, proxy) => sum + proxy.conns, 0),
)

const fetchServerInfo = async () => {
  if (serverInfo) return serverInfo
  serverInfo = await getServerInfo()
  return serverInfo
}

const fetchClient = async () => {
  const key = route.params.key as string
  if (!key) {
    loading.value = false
    return
  }
  try {
    const data = await getClient(key)
    client.value = new Client(data)
  } catch (error: any) {
    ElMessage.error(t('clientDetail.fetchFailed', { message: error.message }))
  } finally {
    loading.value = false
  }
}

const fetchProxies = async () => {
  proxiesLoading.value = true
  const proxyTypes = ['tcp', 'udp', 'http', 'https', 'tcpmux', 'stcp', 'sudp']
  const proxies: BaseProxy[] = []
  try {
    const info = await fetchServerInfo()
    for (const type of proxyTypes) {
      try {
        const json = await getProxiesByType(type)
        if (!json.proxies) continue
        if (type === 'tcp') {
          proxies.push(...json.proxies.map((proxy: any) => new TCPProxy(proxy)))
        } else if (type === 'udp') {
          proxies.push(...json.proxies.map((proxy: any) => new UDPProxy(proxy)))
        } else if (type === 'http' && info?.vhostHTTPPort) {
          proxies.push(
            ...json.proxies.map(
              (proxy: any) => new HTTPProxy(proxy, info.vhostHTTPPort, info.subdomainHost),
            ),
          )
        } else if (type === 'https' && info?.vhostHTTPSPort) {
          proxies.push(
            ...json.proxies.map(
              (proxy: any) => new HTTPSProxy(proxy, info.vhostHTTPSPort, info.subdomainHost),
            ),
          )
        } else if (type === 'tcpmux' && info?.tcpmuxHTTPConnectPort) {
          proxies.push(
            ...json.proxies.map(
              (proxy: any) =>
                new TCPMuxProxy(proxy, info.tcpmuxHTTPConnectPort, info.subdomainHost),
            ),
          )
        } else if (type === 'stcp') {
          proxies.push(...json.proxies.map((proxy: any) => new STCPProxy(proxy)))
        } else if (type === 'sudp') {
          proxies.push(...json.proxies.map((proxy: any) => new SUDPProxy(proxy)))
        }
      } catch {
        // Ignore per-type fetch errors in the combined list.
      }
    }
    allProxies.value = proxies
  } finally {
    proxiesLoading.value = false
  }
}

onMounted(() => {
  fetchClient()
  fetchProxies()
})
</script>

<style scoped>
.client-detail-page {
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  margin-bottom: 24px;
}

.breadcrumb-link {
  display: flex;
  align-items: center;
  color: var(--text-secondary);
  cursor: pointer;
  transition: color 0.2s;
  margin-right: 4px;
}

.breadcrumb-link:hover {
  color: var(--text-primary);
}

.breadcrumb-item {
  color: var(--text-secondary);
  text-decoration: none;
  transition: color 0.2s;
}

.breadcrumb-item:hover {
  color: var(--el-color-primary);
}

.breadcrumb-separator {
  color: var(--el-border-color);
}

.breadcrumb-current {
  color: var(--text-primary);
  font-weight: 500;
}

.header-card,
.proxies-card {
  background: var(--el-bg-color);
  border: 1px solid var(--header-border);
  border-radius: 12px;
  margin-bottom: 16px;
}

.header-main {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 24px;
}

.header-left {
  display: flex;
  gap: 16px;
  align-items: center;
}

.client-avatar {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 500;
  flex-shrink: 0;
}

.client-info {
  min-width: 0;
}

.client-name-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 4px;
}

.client-name {
  font-size: 20px;
  font-weight: 500;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.3;
}

.client-meta {
  display: flex;
  gap: 12px;
  font-size: 14px;
  color: var(--text-secondary);
}

.status-badge {
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
}

.status-badge.online {
  background: rgba(34, 197, 94, 0.1);
  color: #16a34a;
}

.status-badge.offline {
  background: var(--hover-bg);
  color: var(--text-secondary);
}

html.dark .status-badge.online {
  background: rgba(34, 197, 94, 0.15);
  color: #4ade80;
}

.info-section {
  display: flex;
  flex-wrap: wrap;
  gap: 16px 32px;
  padding: 16px 24px;
}

.info-item {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.info-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.info-label::after {
  content: ':';
}

.info-value {
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 500;
  word-break: break-all;
}

.proxies-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  gap: 16px;
}

.proxies-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.proxies-title h2 {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
  margin: 0;
}

.proxies-count {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  background: var(--hover-bg);
  padding: 4px 10px;
  border-radius: 6px;
}

.proxy-search {
  width: 200px;
}

.proxy-search :deep(.el-input__wrapper) {
  border-radius: 6px;
}

.proxies-body {
  padding: 16px;
}

.proxies-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px;
  color: var(--text-secondary);
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
}

.empty-state p {
  margin: 0;
}

.not-found {
  text-align: center;
  padding: 60px 20px;
}

.not-found h2 {
  font-size: 18px;
  font-weight: 500;
  color: var(--text-primary);
  margin: 0 0 8px;
}

.not-found p {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0 0 20px;
}

@media (max-width: 640px) {
  .header-main {
    flex-direction: column;
    gap: 16px;
  }

  .header-right {
    align-self: flex-start;
  }
}
</style>
