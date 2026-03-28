<template>
  <div class="proxies-page">
    <div class="page-top">
      <section class="hero-card">
        <div class="hero-copy">
          <span class="hero-kicker">{{ t('proxyList.heroKicker') }}</span>
          <h2 class="page-title">{{ t('proxyList.heroTitle') }}</h2>
          <p class="page-subtitle">{{ t('proxyList.heroSubtitle') }}</p>
        </div>
        <div class="hero-actions">
          <ActionButton variant="outline" size="small" @click="refreshData">
            <el-icon><Refresh /></el-icon>
            {{ t('common.refresh') }}
          </ActionButton>
          <ActionButton variant="outline" size="small" @click="router.push('/settings')">
            {{ t('common.openSettings') }}
          </ActionButton>
          <ActionButton
            size="small"
            :disabled="!proxyStore.storeEnabled"
            @click="openCreateDrawer"
          >
            + {{ t('proxyList.newProxy') }}
          </ActionButton>
        </div>
      </section>
    </div>

    <div class="page-content" v-loading="proxyStore.storeLoading">
      <section v-if="configProxyStatuses.length > 0" class="section">
        <div class="section-header">
          <div>
            <h3 class="section-title">{{ t('proxyList.configTitle') }}</h3>
            <p class="section-copy">{{ t('proxyList.configCopy') }}</p>
          </div>
        </div>

        <div class="proxy-list">
          <ProxyCard
            v-for="proxy in configProxyStatuses"
            :key="proxy.name"
            :proxy="proxy"
            showSource
            @click="goToDetail(proxy.name)"
          />
        </div>
      </section>

      <section class="section">
        <div class="section-header">
          <div>
            <h3 class="section-title">{{ t('proxyList.managedTitle') }}</h3>
            <p class="section-copy">{{ t('proxyList.managedCopy') }}</p>
          </div>
        </div>

        <div v-if="!proxyStore.storeEnabled && proxyStore.storeChecked" class="store-disabled">
          <p class="store-disabled-title">{{ t('proxyList.storeDisabledTitle') }}</p>
          <p class="store-disabled-copy">{{ t('proxyList.storeDisabledCopy') }}</p>
          <ActionButton size="small" variant="outline" @click="router.push('/settings')">
            {{ t('common.openSettings') }}
          </ActionButton>
          <pre class="config-hint">[store]
path = "./frpc_store.json"</pre>
        </div>

        <template v-else-if="proxyStore.storeEnabled">
          <div class="filter-bar">
            <el-input v-model="searchText" :placeholder="t('proxyList.filterPlaceholder')" clearable class="search-input">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <FilterDropdown
              v-model="typeFilter"
              :label="t('common.type')"
              :options="typeOptions"
              :min-width="140"
              :is-mobile="isMobile"
            />
          </div>

          <div v-if="filteredStoreProxies.length > 0" class="proxy-list">
            <ProxyCard
              v-for="proxyDef in filteredStoreProxies"
              :key="proxyDef.name"
              :proxy="proxyStore.storeProxyWithStatus(proxyDef)"
              show-actions
              @click="goToDetail(proxyDef.name)"
              @edit="handleEdit"
              @toggle="handleToggleProxy"
              @delete="handleDeleteProxy(proxyDef.name)"
            />
          </div>
          <div v-else class="empty-state">
            <p class="empty-text">{{ t('proxyList.emptyTitle') }}</p>
            <p class="empty-hint">{{ t('proxyList.emptyHint') }}</p>
          </div>
        </template>

        <div v-else-if="proxyStore.storeChecked" class="empty-state">
          <p class="empty-text">{{ t('proxyList.waitingTitle') }}</p>
          <p class="empty-hint">{{ t('proxyList.waitingHint') }}</p>
        </div>
      </section>
    </div>

    <el-drawer
      v-model="createDrawerVisible"
      direction="rtl"
      size="min(720px, 100%)"
      :with-header="false"
      :destroy-on-close="false"
      :close-on-click-modal="!drawerDirty"
      class="create-drawer"
      @closed="handleDrawerClosed"
    >
      <div class="drawer-shell">
        <div class="drawer-header">
          <div>
            <span class="hero-kicker">{{ t('proxyEdit.drawerTitle') }}</span>
            <h3 class="drawer-title">
              {{ selectedType ? t('proxyEdit.drawerTitleConfigure') : t('proxyEdit.drawerTitleChoose') }}
            </h3>
            <p class="drawer-copy">
              {{ selectedType ? activeGuide?.summary : t('proxyEdit.drawerCopyChoose') }}
            </p>
          </div>
          <div class="drawer-actions">
            <ActionButton
              v-if="selectedType"
              variant="outline"
              size="small"
              @click="resetDrawerType"
            >
              {{ t('proxyEdit.changeType') }}
            </ActionButton>
            <ActionButton variant="outline" size="small" @click="requestCloseDrawer">
              {{ t('common.close') }}
            </ActionButton>
            <ActionButton
              v-if="selectedType"
              size="small"
              :loading="createSaving"
              @click="handleCreateSave"
            >
              {{ t('common.create') }}
            </ActionButton>
          </div>
        </div>

        <div class="drawer-content">
          <template v-if="!selectedType">
            <GuideTypeGrid :items="proxyGuides" @select="handleChooseType" />
          </template>

          <template v-else>
            <section v-if="activeGuide" class="guide-summary">
              <div class="guide-summary-copy">
                <p class="guide-caution">{{ activeGuide.caution || defaultStoreCopy }}</p>
                <div class="field-pills">
                  <span v-for="field in activeGuide.primaryFields" :key="field" class="field-pill">
                    {{ field }}
                  </span>
                </div>
              </div>
            </section>

            <el-form
              ref="createFormRef"
              :model="createForm"
              :rules="createRules"
              label-position="top"
              @submit.prevent
            >
              <ProxyFormLayout
                v-model="createForm"
                :lock-type="true"
              />
            </el-form>
          </template>
        </div>
      </div>
    </el-drawer>

    <ConfirmDialog
      v-model="deleteDialog.visible"
      :title="t('proxyList.deleteTitle')"
      :message="deleteDialog.message"
      :confirm-text="t('common.delete')"
      danger
      :loading="deleteDialog.loading"
      :is-mobile="isMobile"
      @confirm="doDelete"
    />

    <ConfirmDialog
      v-model="closeDrawerDialogVisible"
      :title="t('proxyEdit.closeDrawerTitle')"
      :message="t('proxyEdit.closeDrawerMessage')"
      :confirm-text="t('common.discard')"
      danger
      :is-mobile="isMobile"
      @confirm="confirmCloseDrawer"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import ActionButton from '@shared/components/ActionButton.vue'
import ConfirmDialog from '@shared/components/ConfirmDialog.vue'
import FilterDropdown from '@shared/components/FilterDropdown.vue'
import GuideTypeGrid from '../components/GuideTypeGrid.vue'
import ProxyCard from '../components/ProxyCard.vue'
import ProxyFormLayout from '../components/proxy-form/ProxyFormLayout.vue'
import { getProxyGuides, getProxyGuideMap, PROXY_TYPE_ORDER } from '../content/guides'
import { useResponsive } from '../composables/useResponsive'
import { useProxyStore } from '../stores/proxy'
import {
  createDefaultProxyForm,
  formToStoreProxy,
  type ProxyDefinition,
  type ProxyFormData,
  type ProxyStatus,
  type ProxyType,
} from '../types'
import { useI18n } from '../i18n'

const { isMobile } = useResponsive()
const router = useRouter()
const proxyStore = useProxyStore()
const { t } = useI18n()

const defaultStoreCopy = computed(() => t('proxyEdit.defaultStoreCopy'))
const searchText = ref('')
const typeFilter = ref('')

const createDrawerVisible = ref(false)
const selectedType = ref<ProxyType | null>(null)
const createSaving = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = ref<ProxyFormData>(createDefaultProxyForm())
const drawerDirty = ref(false)
const trackDrawerChanges = ref(false)
const closeDrawerDialogVisible = ref(false)

const deleteDialog = reactive({
  visible: false,
  message: '',
  loading: false,
  name: '',
})

const proxyGuides = computed(() => getProxyGuides())
const proxyGuideMap = computed(() => getProxyGuideMap())
const activeGuide = computed(() => {
  if (!selectedType.value) return null
  return proxyGuideMap.value[selectedType.value]
})

const createRules: FormRules = {
  name: [
    { required: true, message: t('validation.nameRequired'), trigger: 'blur' },
    { min: 1, max: 50, message: t('validation.length1to50'), trigger: 'blur' },
  ],
  type: [{ required: true, message: t('validation.typeRequired'), trigger: 'change' }],
  localPort: [
    {
      validator: (_rule, value, callback) => {
        if (!createForm.value.pluginType && value == null) {
          callback(new Error(t('validation.localPortRequired')))
          return
        }
        callback()
      },
      trigger: 'blur',
    },
  ],
  customDomains: [
    {
      validator: (_rule, value, callback) => {
        if (
          ['http', 'https', 'tcpmux'].includes(createForm.value.type) &&
          (!value || value.length === 0) &&
          !createForm.value.subdomain
        ) {
          callback(new Error(t('validation.customDomainsOrSubdomainRequired')))
          return
        }
        callback()
      },
      trigger: 'blur',
    },
  ],
  healthCheckPath: [
    {
      validator: (_rule, value, callback) => {
        if (createForm.value.healthCheckType === 'http' && !value) {
          callback(new Error(t('validation.healthPathRequired')))
          return
        }
        callback()
      },
      trigger: 'blur',
    },
  ],
}

const typeOptions = computed(() => {
  const types = new Set<string>()
  proxyStore.storeProxies.forEach((proxy) => types.add(proxy.type))
  return Array.from(types)
    .sort((a, b) => PROXY_TYPE_ORDER.indexOf(a as ProxyType) - PROXY_TYPE_ORDER.indexOf(b as ProxyType))
    .map((type) => ({ label: type.toUpperCase(), value: type }))
})

const filteredStoreProxies = computed(() => {
  let list = proxyStore.storeProxies as ProxyDefinition[]

  if (typeFilter.value) {
    list = list.filter((proxy) => proxy.type === typeFilter.value)
  }

  if (searchText.value) {
    const query = searchText.value.toLowerCase()
    list = list.filter((proxy) => proxy.name.toLowerCase().includes(query))
  }

  return list
})

const configProxyStatuses = computed(() => {
  return proxyStore.configProxies.map((proxy) => proxyStore.configProxyWithStatus(proxy))
})

watch(
  () => createForm.value,
  () => {
    if (trackDrawerChanges.value) {
      drawerDirty.value = true
    }
  },
  { deep: true },
)

const refreshData = async () => {
  try {
    await Promise.all([
      proxyStore.fetchStatus(),
      proxyStore.fetchConfigProxies(),
      proxyStore.fetchStoreProxies(),
    ])
  } catch (error: any) {
    ElMessage.error(t('proxyList.refreshFailed', { message: error.message }))
  }
}

const resetDrawerState = async () => {
  selectedType.value = null
  drawerDirty.value = false
  trackDrawerChanges.value = false
  createSaving.value = false
  createForm.value = createDefaultProxyForm()
  await nextTick()
}

const openCreateDrawer = async () => {
  if (!proxyStore.storeEnabled) {
    ElMessage.warning(t('proxyList.storeEnableWarning'))
    return
  }
  await resetDrawerState()
  createDrawerVisible.value = true
}

const handleChooseType = async (type: string) => {
  selectedType.value = type as ProxyType
  const nextForm = createDefaultProxyForm()
  nextForm.type = type as ProxyType
  createForm.value = nextForm
  await nextTick()
  trackDrawerChanges.value = true
}

const resetDrawerType = async () => {
  await resetDrawerState()
}

const requestCloseDrawer = () => {
  if (drawerDirty.value) {
    closeDrawerDialogVisible.value = true
    return
  }
  createDrawerVisible.value = false
}

const confirmCloseDrawer = () => {
  closeDrawerDialogVisible.value = false
  createDrawerVisible.value = false
}

const handleDrawerClosed = async () => {
  await resetDrawerState()
}

const handleCreateSave = async () => {
  if (!createFormRef.value) return

  try {
    await createFormRef.value.validate()
  } catch {
    ElMessage.warning(t('proxyEdit.validationWarning'))
    return
  }

  createSaving.value = true
  try {
    const data = formToStoreProxy(createForm.value)
    await proxyStore.createProxy(data)
    await proxyStore.fetchStatus()
    ElMessage.success(t('proxyEdit.created'))
    createDrawerVisible.value = false
  } catch (err: any) {
    ElMessage.error(t('proxyEdit.operationFailed', { message: err.message || 'Unknown error' }))
  } finally {
    createSaving.value = false
  }
}

const handleEdit = (proxy: ProxyStatus) => {
  router.push('/proxies/' + encodeURIComponent(proxy.name) + '/edit')
}

const goToDetail = (name: string) => {
  router.push('/proxies/detail/' + encodeURIComponent(name))
}

const handleToggleProxy = async (proxy: ProxyStatus, enabled: boolean) => {
  try {
    await proxyStore.toggleProxy(proxy.name, enabled)
    ElMessage.success(enabled ? t('proxyList.enabledSuccess') : t('proxyList.disabledSuccess'))
  } catch (err: any) {
    ElMessage.error(t('proxyList.operationFailed', { message: err.message || 'Unknown error' }))
  }
}

const handleDeleteProxy = (name: string) => {
  deleteDialog.name = name
  deleteDialog.message = t('proxyList.deleteMessage', { name })
  deleteDialog.visible = true
}

const doDelete = async () => {
  deleteDialog.loading = true
  try {
    await proxyStore.deleteProxy(deleteDialog.name)
    ElMessage.success(t('proxyList.deleteSuccess'))
    deleteDialog.visible = false
    await proxyStore.fetchStatus()
  } catch (err: any) {
    ElMessage.error(t('proxyList.deleteFailed', { message: err.message || 'Unknown error' }))
  } finally {
    deleteDialog.loading = false
  }
}

onMounted(() => {
  refreshData()
})
</script>

<style scoped lang="scss">
.proxies-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  max-width: 1080px;
  margin: 0 auto;
}

.page-top {
  flex-shrink: 0;
  padding: $spacing-xl 28px 0;
}

.page-content {
  flex: 1;
  overflow-y: auto;
  padding: 0 28px 32px;
}

.hero-card {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  padding: 24px;
  border-radius: 24px;
  background: linear-gradient(135deg, #f7f4ed 0%, #fdf7ea 52%, #f5efe2 100%);
  border: 1px solid rgba(148, 163, 184, 0.18);
  margin-bottom: 28px;
}

.hero-copy {
  max-width: 720px;
}

.hero-kicker {
  display: inline-flex;
  align-items: center;
  margin-bottom: 12px;
  padding: 5px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  color: #6b4f3b;
  background: rgba(255, 255, 255, 0.7);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.hero-actions {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  flex-shrink: 0;
  flex-wrap: wrap;
}

.section {
  margin-bottom: 28px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 16px;
  margin-bottom: 16px;
}

.section-title {
  margin: 0 0 6px;
  font-size: 18px;
  color: $color-text-primary;
}

.section-copy {
  margin: 0;
  font-size: 14px;
  line-height: 1.6;
  color: $color-text-muted;
}

.filter-bar {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 18px;
}

.proxy-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.store-disabled {
  margin-bottom: 18px;
  padding: 18px 20px;
  border-radius: 18px;
  background: $color-bg-secondary;
  border: 1px dashed $color-border;
}

.store-disabled-title {
  margin: 0 0 8px;
  font-size: 15px;
  font-weight: 700;
  color: $color-text-primary;
}

.store-disabled-copy {
  margin: 0;
  font-size: 14px;
  line-height: 1.7;
  color: $color-text-muted;
}

.config-hint {
  display: inline-block;
  margin: 14px 0 0;
  padding: 12px 14px;
  border-radius: 12px;
  background: $color-bg-primary;
  border: 1px solid $color-border-light;
  font-size: 13px;
  line-height: 1.6;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

.empty-state-compact {
  padding: 32px 20px;
}

.empty-text {
  margin: 0 0 8px;
  font-size: 18px;
  font-weight: 600;
  color: $color-text-secondary;
}

.empty-hint {
  margin: 0;
  font-size: 14px;
  color: $color-text-muted;
}

.drawer-shell {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.drawer-header {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 24px 24px 18px;
  border-bottom: 1px solid $color-border-light;
}

.drawer-title {
  margin: 0 0 8px;
  font-size: 24px;
  color: $color-text-primary;
}

.drawer-copy {
  margin: 0;
  font-size: 14px;
  line-height: 1.7;
  color: $color-text-muted;
}

.drawer-actions {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  flex-shrink: 0;
}

.drawer-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.guide-summary {
  margin-bottom: 20px;
  padding: 18px;
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background: linear-gradient(135deg, #f8f6ef 0%, #fcfbf7 100%);
}

.guide-caution {
  margin: 0;
  font-size: 14px;
  line-height: 1.7;
  color: $color-text-secondary;
}

.field-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 16px;
}

.field-pill {
  display: inline-flex;
  align-items: center;
  padding: 6px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  color: $color-text-secondary;
  background: rgba(15, 23, 42, 0.06);
}

@include mobile {
  .page-top {
    padding: $spacing-lg $spacing-lg 0;
  }

  .page-content {
    padding: 0 $spacing-lg 24px;
  }

  .hero-card {
    flex-direction: column;
    padding: 20px;
  }

  .filter-bar {
    flex-wrap: wrap;
  }

  .drawer-header {
    flex-direction: column;
  }

  .drawer-actions {
    flex-wrap: wrap;
  }

  .drawer-content {
    padding: 18px;
  }
}
</style>
