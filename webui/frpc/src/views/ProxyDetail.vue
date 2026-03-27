<template>
  <div class="proxy-detail-page">
    <div class="detail-top">
      <nav class="breadcrumb">
        <router-link to="/proxies" class="breadcrumb-link">{{ t('app.proxies') }}</router-link>
        <span class="breadcrumb-sep">&rsaquo;</span>
        <span class="breadcrumb-current">{{ proxyName }}</span>
      </nav>

      <template v-if="proxy">
        <div class="detail-header">
          <div>
            <div class="header-title-row">
              <h2 class="detail-title">{{ proxy.name }}</h2>
              <span class="status-pill" :class="statusClass">
                <span class="status-dot"></span>
                {{ statusLabel }}
              </span>
            </div>
            <p class="header-subtitle">
              {{ activeGuide?.summary || t('proxyDetail.fallbackSummary') }}
            </p>
            <p class="header-meta">
              {{ t('proxyDetail.sourceType', { source: displaySource, type: proxy.type.toUpperCase() }) }}
            </p>
          </div>
          <div v-if="isStore" class="header-actions">
            <ActionButton variant="outline" size="small" @click="handleEdit">
              {{ t('common.edit') }}
            </ActionButton>
          </div>
        </div>
      </template>
    </div>

    <div v-if="notFound" class="not-found">
      <p class="empty-text">{{ t('proxyDetail.notFoundTitle') }}</p>
      <p class="empty-hint">{{ t('proxyDetail.notFoundHint', { name: proxyName }) }}</p>
      <ActionButton variant="outline" @click="router.push('/proxies')">
        {{ t('common.backToProxies') }}
      </ActionButton>
    </div>

    <div v-else-if="proxy" v-loading="loading" class="detail-content">
      <div v-if="activeGuide" class="guide-banner">
        <div>
          <div class="guide-banner-title">{{ activeGuide.label }}</div>
          <div class="guide-banner-copy">
            {{ activeGuide.caution || t('proxyDetail.fallbackGuide') }}
          </div>
        </div>
      </div>

      <div v-if="proxy.remote_addr" class="connect-banner">
        <div class="connect-info">
          <span class="connect-label">{{ t('proxyDetail.remoteAddr') }}</span>
          <code class="connect-addr">{{ proxy.remote_addr }}</code>
        </div>
        <button class="copy-btn" :class="{ copied: copied }" @click="copyAddr">
          {{ copied ? t('proxyDetail.copied') : t('proxyDetail.copy') }}
        </button>
      </div>

      <div v-if="proxy.err" class="error-banner">
        <el-icon class="error-icon"><Warning /></el-icon>
        <div>
          <div class="error-title">{{ t('proxyDetail.errorTitle') }}</div>
          <div class="error-message">{{ proxy.err }}</div>
        </div>
      </div>

      <ProxyFormLayout
        v-if="formData"
        :model-value="formData"
        readonly
      />
    </div>

    <div v-else v-loading="loading" class="loading-area"></div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Warning } from '@element-plus/icons-vue'
import ActionButton from '@shared/components/ActionButton.vue'
import ProxyFormLayout from '../components/proxy-form/ProxyFormLayout.vue'
import { getProxyGuideMap } from '../content/guides'
import { getProxyConfig, getStoreProxy } from '../api/frpc'
import { useProxyStore } from '../stores/proxy'
import { storeProxyToForm } from '../types'
import type { ProxyDefinition, ProxyFormData, ProxyStatus, ProxyType } from '../types'
import { useI18n } from '../i18n'

const route = useRoute()
const router = useRouter()
const proxyStore = useProxyStore()
const { t } = useI18n()

const proxyName = route.params.name as string
const proxy = ref<ProxyStatus | null>(null)
const proxyConfig = ref<ProxyDefinition | null>(null)
const loading = ref(true)
const notFound = ref(false)
const isStore = ref(false)
const copied = ref(false)

const copyAddr = async () => {
  if (!proxy.value?.remote_addr) return
  try {
    await navigator.clipboard.writeText(proxy.value.remote_addr)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    ElMessage.error(t('proxyDetail.copyFailed'))
  }
}

onMounted(async () => {
  try {
    await proxyStore.fetchStatus()
    const found = proxyStore.proxies.find((entry) => entry.name === proxyName)

    let configDef: ProxyDefinition | null = null
    try {
      configDef = await getProxyConfig(proxyName)
      proxyConfig.value = configDef
    } catch {
      // keep fallback logic below
    }

    try {
      await getStoreProxy(proxyName)
      isStore.value = true
    } catch {
      // not store managed
    }

    if (found) {
      proxy.value = found
    } else if (configDef) {
      const block = (configDef as any)[configDef.type]
      const localIP = block?.localIP || '127.0.0.1'
      const localPort = block?.localPort
      const enabled = block?.enabled !== false
      proxy.value = {
        name: configDef.name,
        type: configDef.type,
        status: enabled ? 'waiting' : 'disabled',
        err: '',
        local_addr: localPort != null ? `${localIP}:${localPort}` : '',
        remote_addr: block?.remotePort != null ? `:${block.remotePort}` : '',
        plugin: block?.plugin?.type || '',
      }
    } else {
      notFound.value = true
    }
  } catch (err: any) {
    ElMessage.error(t('proxyEdit.loadFailed', { message: err.message }))
  } finally {
    loading.value = false
  }
})

const proxyGuideMap = computed(() => getProxyGuideMap())

const activeGuide = computed(() => {
  const type = proxy.value?.type
  if (!type) return null
  return proxyGuideMap.value[type as ProxyType] || null
})

const displaySource = computed(() => (isStore.value ? t('common.store') : t('common.config')))

const statusClass = computed(() => {
  const status = proxy.value?.status
  if (status === 'running') return 'running'
  if (status === 'error') return 'error'
  if (status === 'disabled') return 'disabled'
  return 'waiting'
})

const statusLabel = computed(() => {
  switch (proxy.value?.status) {
    case 'running':
      return t('status.running')
    case 'error':
      return t('status.error')
    case 'disabled':
      return t('status.disabled')
    default:
      return t('status.waiting')
  }
})

const formData = computed((): ProxyFormData | null => {
  if (!proxyConfig.value) return null
  return storeProxyToForm(proxyConfig.value)
})

const handleEdit = () => {
  router.push('/proxies/' + encodeURIComponent(proxyName) + '/edit')
}
</script>

<style scoped lang="scss">
.proxy-detail-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  max-width: 1080px;
  margin: 0 auto;
}

.detail-top {
  flex-shrink: 0;
  padding: $spacing-xl 28px 0;
}

.detail-content {
  flex: 1;
  overflow-y: auto;
  padding: 0 28px 120px;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  margin-bottom: 18px;
}

.breadcrumb-link {
  color: $color-text-secondary;
  text-decoration: none;

  &:hover {
    color: $color-text-primary;
  }
}

.breadcrumb-current {
  color: $color-text-primary;
  font-weight: 500;
}

.breadcrumb-sep {
  color: $color-text-light;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 22px;
}

.header-title-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.detail-title {
  margin: 0;
  font-size: 26px;
  color: $color-text-primary;
}

.header-subtitle,
.header-meta {
  margin: 0;
  font-size: 14px;
  line-height: 1.7;
}

.header-subtitle {
  color: $color-text-secondary;
}

.header-meta {
  color: $color-text-muted;
}

.connect-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 18px;
  border-radius: 12px;
  margin-bottom: 18px;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-lighter);
}

.connect-info {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.connect-label {
  font-size: 13px;
  color: $color-text-secondary;
  white-space: nowrap;
  font-weight: 500;
}

.connect-addr {
  font-size: 14px;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  color: $color-text-primary;
  word-break: break-all;
}

.copy-btn {
  flex-shrink: 0;
  padding: 5px 14px;
  border-radius: 8px;
  border: 1px solid var(--el-border-color);
  background: var(--el-bg-color);
  color: $color-text-secondary;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    border-color: var(--el-color-primary);
    color: var(--el-color-primary);
  }

  &.copied {
    border-color: var(--el-color-success);
    color: var(--el-color-success);
    background: var(--el-color-success-light-9);
  }
}

.guide-banner,
.error-banner {
  display: flex;
  gap: 12px;
  padding: 16px 18px;
  border-radius: 16px;
  margin-bottom: 18px;
}

.guide-banner {
  background: linear-gradient(135deg, #f8f5ef 0%, #fcfbf8 100%);
  border: 1px solid rgba(148, 163, 184, 0.18);
}

.guide-banner-title {
  margin-bottom: 6px;
  font-size: 14px;
  font-weight: 700;
  color: $color-text-primary;
}

.guide-banner-copy {
  font-size: 14px;
  line-height: 1.7;
  color: $color-text-secondary;
}

.error-banner {
  background: var(--color-danger-light);
  border: 1px solid rgba(245, 108, 108, 0.2);
}

.error-icon {
  color: $color-danger;
  font-size: 18px;
  margin-top: 2px;
}

.error-title {
  font-size: 14px;
  font-weight: 700;
  color: $color-danger;
  margin-bottom: 4px;
}

.error-message {
  font-size: 14px;
  color: $color-text-secondary;
}

.not-found,
.loading-area {
  text-align: center;
  padding: 60px 20px;
}

.empty-text {
  margin: 0 0 8px;
  font-size: 18px;
  font-weight: 600;
  color: $color-text-secondary;
}

.empty-hint {
  margin: 0 0 18px;
  font-size: 14px;
  color: $color-text-muted;
}

@include mobile {
  .detail-top {
    padding: $spacing-lg $spacing-lg 0;
  }

  .detail-content {
    padding: 0 $spacing-lg 120px;
  }

  .detail-header {
    flex-direction: column;
  }
}
</style>
