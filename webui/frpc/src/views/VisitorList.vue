<template>
  <div class="visitors-page">
    <div class="page-top">
      <section class="hero-card">
        <div class="hero-copy">
          <span class="hero-kicker">{{ t('visitorList.heroKicker') }}</span>
          <h2 class="page-title">{{ t('visitorList.heroTitle') }}</h2>
          <p class="page-subtitle">{{ t('visitorList.heroSubtitle') }}</p>
        </div>
        <div class="hero-actions">
          <ActionButton variant="outline" size="small" @click="fetchData">
            <el-icon><Refresh /></el-icon>
            {{ t('common.refresh') }}
          </ActionButton>
          <ActionButton variant="outline" size="small" @click="router.push('/settings')">
            {{ t('common.openSettings') }}
          </ActionButton>
          <ActionButton
            size="small"
            :disabled="!visitorStore.storeEnabled"
            @click="openCreateDrawer"
          >
            + {{ t('visitorList.newVisitor') }}
          </ActionButton>
          <ActionButton
            size="small"
            variant="outline"
            @click="router.push('/pairing/create')"
          >
            {{ t('pairing.pairingWizardCreate') }}
          </ActionButton>
          <ActionButton
            size="small"
            variant="outline"
            @click="router.push('/pairing/import')"
          >
            {{ t('pairing.pairingWizardImport') }}
          </ActionButton>
        </div>
      </section>
    </div>

    <div class="page-content" v-loading="visitorStore.loading">
      <section class="section">
        <div class="section-header">
          <div>
            <h3 class="section-title">{{ t('visitorList.configTitle') }}</h3>
            <p class="section-copy">{{ t('visitorList.configCopy') }}</p>
          </div>
        </div>

        <div v-if="filteredConfigVisitors.length > 0" class="visitor-list">
          <div
            v-for="visitor in filteredConfigVisitors"
            :key="visitor.name"
            class="visitor-card"
            @click="goToDetail(visitor.name)"
          >
            <div class="card-left">
              <div class="card-header">
                <span class="visitor-name">{{ visitor.name }}</span>
                <span class="type-tag">{{ visitor.type.toUpperCase() }}</span>
                <span class="source-tag">{{ t('common.config') }}</span>
              </div>
              <div class="card-meta">
                {{ getServerName(visitor) || t('visitorList.pairFallback') }}
              </div>
            </div>
          </div>
        </div>
        <div v-else class="empty-state empty-state-compact">
          <p class="empty-text">{{ t('visitorList.configEmptyTitle') }}</p>
          <p class="empty-hint">{{ t('visitorList.configEmptyHint') }}</p>
        </div>
      </section>

      <section class="section">
        <div class="section-header">
          <div>
            <h3 class="section-title">{{ t('visitorList.managedTitle') }}</h3>
            <p class="section-copy">{{ t('visitorList.managedCopy') }}</p>
          </div>
        </div>

        <div v-if="!visitorStore.storeEnabled && visitorStore.storeChecked" class="store-disabled">
          <p class="store-disabled-title">{{ t('visitorList.storeDisabledTitle') }}</p>
          <p class="store-disabled-copy">{{ t('visitorList.storeDisabledCopy') }}</p>
          <ActionButton size="small" variant="outline" @click="router.push('/settings')">
            {{ t('common.openSettings') }}
          </ActionButton>
          <pre class="config-hint">[store]
path = "./frpc_store.json"</pre>
        </div>

        <template v-else-if="visitorStore.storeEnabled">
          <div class="filter-bar">
            <el-input v-model="searchText" :placeholder="t('visitorList.filterPlaceholder')" clearable class="search-input">
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

          <div v-if="filteredVisitors.length > 0" class="visitor-list">
            <div
              v-for="visitor in filteredVisitors"
              :key="visitor.name"
              class="visitor-card"
              @click="goToDetail(visitor.name)"
            >
              <div class="card-left">
                <div class="card-header">
                  <span class="visitor-name">{{ visitor.name }}</span>
                  <span class="type-tag">{{ visitor.type.toUpperCase() }}</span>
                </div>
                <div class="card-meta">
                  {{ getServerName(visitor) || t('visitorList.pairFallback') }}
                </div>
              </div>
              <div class="card-actions">
                <ActionButton variant="outline" size="small" @click.stop="handleEdit(visitor)">
                  {{ t('common.edit') }}
                </ActionButton>
                <ActionButton variant="outline" size="small" danger @click.stop="handleDelete(visitor.name)">
                  {{ t('common.delete') }}
                </ActionButton>
              </div>
            </div>
          </div>
          <div v-else class="empty-state">
            <p class="empty-text">{{ t('visitorList.emptyTitle') }}</p>
            <p class="empty-hint">{{ t('visitorList.emptyHint') }}</p>
          </div>
        </template>

        <div v-else-if="visitorStore.storeChecked" class="empty-state">
          <p class="empty-text">{{ t('visitorList.waitingTitle') }}</p>
          <p class="empty-hint">{{ t('visitorList.waitingHint') }}</p>
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
            <span class="hero-kicker">{{ t('visitorEdit.drawerTitle') }}</span>
            <h3 class="drawer-title">
              {{ selectedType ? t('visitorEdit.drawerTitleConfigure') : t('visitorEdit.drawerTitleChoose') }}
            </h3>
            <p class="drawer-copy">
              {{ selectedType ? activeGuide?.summary : t('visitorEdit.drawerCopyChoose') }}
            </p>
          </div>
          <div class="drawer-actions">
            <ActionButton
              v-if="selectedType"
              variant="outline"
              size="small"
              @click="resetDrawerType"
            >
              {{ t('visitorEdit.changeType') }}
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
            <GuideTypeGrid :items="visitorGuides" @select="handleChooseType" />
          </template>

          <template v-else>
            <section v-if="activeGuide" class="guide-summary">
              <div class="guide-summary-copy">
                <p class="guide-caution">{{ activeGuide.caution || t('visitorEdit.defaultStoreCopy') }}</p>
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
              <VisitorFormLayout
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
      :title="t('visitorList.deleteTitle')"
      :message="deleteDialog.message"
      :confirm-text="t('common.delete')"
      danger
      :loading="deleteDialog.loading"
      :is-mobile="isMobile"
      @confirm="doDelete"
    />

    <ConfirmDialog
      v-model="closeDrawerDialogVisible"
      :title="t('visitorEdit.closeDrawerTitle')"
      :message="t('visitorEdit.closeDrawerMessage')"
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
import VisitorFormLayout from '../components/visitor-form/VisitorFormLayout.vue'
import { getVisitorGuideMap, getVisitorGuides, VISITOR_TYPE_ORDER } from '../content/guides'
import { useResponsive } from '../composables/useResponsive'
import { useVisitorStore } from '../stores/visitor'
import {
  createDefaultVisitorForm,
  formToStoreVisitor,
  type VisitorDefinition,
  type VisitorFormData,
  type VisitorType,
} from '../types'
import { useI18n } from '../i18n'

const { isMobile } = useResponsive()
const router = useRouter()
const visitorStore = useVisitorStore()
const { t } = useI18n()

const searchText = ref('')
const typeFilter = ref('')

const createDrawerVisible = ref(false)
const selectedType = ref<VisitorType | null>(null)
const createSaving = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = ref<VisitorFormData>(createDefaultVisitorForm())
const drawerDirty = ref(false)
const trackDrawerChanges = ref(false)
const closeDrawerDialogVisible = ref(false)

const deleteDialog = reactive({
  visible: false,
  message: '',
  loading: false,
  name: '',
})

const visitorGuides = computed(() => getVisitorGuides())
const visitorGuideMap = computed(() => getVisitorGuideMap())
const activeGuide = computed(() => {
  if (!selectedType.value) return null
  return visitorGuideMap.value[selectedType.value]
})

const createRules: FormRules = {
  name: [
    { required: true, message: t('validation.nameRequired'), trigger: 'blur' },
    { min: 1, max: 50, message: t('validation.length1to50'), trigger: 'blur' },
  ],
  type: [{ required: true, message: t('validation.typeRequired'), trigger: 'change' }],
  serverName: [{ required: true, message: t('validation.serverNameRequired'), trigger: 'blur' }],
  bindPort: [
    { required: true, message: t('validation.bindPortRequired'), trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value == null) {
          callback(new Error(t('validation.bindPortRequired')))
          return
        }
        if (value > 65535) {
          callback(new Error(t('validation.bindPortMax')))
          return
        }
        if (createForm.value.type === 'sudp' && value < 1) {
          callback(new Error(t('validation.bindPortSudpMin')))
          return
        }
        if (createForm.value.type !== 'sudp' && value === 0) {
          callback(new Error(t('validation.bindPortNotZero')))
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
  visitorStore.storeVisitors.forEach((visitor) => types.add(visitor.type))
  return Array.from(types)
    .sort((a, b) => VISITOR_TYPE_ORDER.indexOf(a as VisitorType) - VISITOR_TYPE_ORDER.indexOf(b as VisitorType))
    .map((type) => ({ label: type.toUpperCase(), value: type }))
})

const filteredVisitors = computed(() => {
  let list = visitorStore.storeVisitors as VisitorDefinition[]

  if (typeFilter.value) {
    list = list.filter((visitor) => visitor.type === typeFilter.value)
  }

  if (searchText.value) {
    const query = searchText.value.toLowerCase()
    list = list.filter((visitor) => visitor.name.toLowerCase().includes(query))
  }

  return list
})

const filteredConfigVisitors = computed(() => {
  let list = visitorStore.configVisitors as VisitorDefinition[]

  if (typeFilter.value) {
    list = list.filter((visitor) => visitor.type === typeFilter.value)
  }

  if (searchText.value) {
    const query = searchText.value.toLowerCase()
    list = list.filter((visitor) => visitor.name.toLowerCase().includes(query))
  }

  return list
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

const getServerName = (visitor: VisitorDefinition): string => {
  const block = (visitor as any)[visitor.type]
  return block?.serverName || ''
}

const fetchData = async () => {
  try {
    await Promise.all([
      visitorStore.fetchConfigVisitors(),
      visitorStore.fetchStoreVisitors(),
    ])
  } catch (error: any) {
    ElMessage.error(t('visitorList.refreshFailed', { message: error.message }))
  }
}

const resetDrawerState = async () => {
  selectedType.value = null
  drawerDirty.value = false
  trackDrawerChanges.value = false
  createSaving.value = false
  createForm.value = createDefaultVisitorForm()
  await nextTick()
}

const openCreateDrawer = async () => {
  if (!visitorStore.storeEnabled) {
    ElMessage.warning(t('visitorList.storeEnableWarning'))
    return
  }
  await resetDrawerState()
  createDrawerVisible.value = true
}

const handleChooseType = async (type: string) => {
  selectedType.value = type as VisitorType
  const nextForm = createDefaultVisitorForm()
  nextForm.type = type as VisitorType
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
    ElMessage.warning(t('visitorEdit.validationWarning'))
    return
  }

  createSaving.value = true
  try {
    const data = formToStoreVisitor(createForm.value)
    await visitorStore.createVisitor(data)
    ElMessage.success(t('visitorEdit.created'))
    createDrawerVisible.value = false
  } catch (err: any) {
    ElMessage.error(t('visitorEdit.operationFailed', { message: err.message || 'Unknown error' }))
  } finally {
    createSaving.value = false
  }
}

const handleEdit = (visitor: VisitorDefinition) => {
  router.push('/visitors/' + encodeURIComponent(visitor.name) + '/edit')
}

const goToDetail = (name: string) => {
  router.push('/visitors/detail/' + encodeURIComponent(name))
}

const handleDelete = (name: string) => {
  deleteDialog.name = name
  deleteDialog.message = t('visitorList.deleteMessage', { name })
  deleteDialog.visible = true
}

const doDelete = async () => {
  deleteDialog.loading = true
  try {
    await visitorStore.deleteVisitor(deleteDialog.name)
    ElMessage.success(t('visitorList.deleteSuccess'))
    deleteDialog.visible = false
  } catch (err: any) {
    ElMessage.error(t('visitorList.deleteFailed', { message: err.message || 'Unknown error' }))
  } finally {
    deleteDialog.loading = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped lang="scss">
.visitors-page {
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
  background: linear-gradient(135deg, #eef5ef 0%, #f6faf2 45%, #edf6f0 100%);
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
  color: #335c48;
  background: rgba(255, 255, 255, 0.72);
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

.visitor-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.visitor-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 16px 18px;
  border: 1px solid $color-border-light;
  border-radius: 18px;
  background: $color-bg-primary;
  cursor: pointer;
  transition: all $transition-fast;

  &:hover {
    border-color: $color-border;
    box-shadow: 0 10px 24px rgba(15, 23, 42, 0.06);
  }
}

.card-left {
  min-width: 0;
  flex: 1;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.visitor-name {
  font-size: 16px;
  font-weight: 700;
  color: $color-text-primary;
}

.type-tag {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  color: $color-text-secondary;
  background: $color-bg-secondary;
}

.source-tag {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  color: $color-text-light;
  background: rgba(15, 23, 42, 0.06);
}

.card-meta {
  font-size: 14px;
  color: $color-text-muted;
  line-height: 1.6;
}

.card-actions {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
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
  background: linear-gradient(135deg, #eff6f1 0%, #fbfdfb 100%);
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

  .hero-actions {
    flex-wrap: wrap;
  }

  .filter-bar {
    flex-wrap: wrap;
  }

  .visitor-card {
    flex-direction: column;
    align-items: stretch;
  }

  .card-actions {
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
