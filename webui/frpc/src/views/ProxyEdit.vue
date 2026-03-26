<template>
  <div class="proxy-edit-page">
    <div class="edit-header">
      <nav class="breadcrumb">
        <router-link to="/proxies" class="breadcrumb-item">{{ t('app.proxies') }}</router-link>
        <span class="breadcrumb-separator">&rsaquo;</span>
        <span class="breadcrumb-current">
          {{ isEditing ? t('proxyEdit.crumbEdit') : hasSelectedType ? t('proxyEdit.crumbConfigure') : t('proxyEdit.crumbChoose') }}
        </span>
      </nav>
      <div class="header-actions">
        <ActionButton variant="outline" size="small" @click="goBack">{{ t('common.cancel') }}</ActionButton>
        <ActionButton
          v-if="canSave"
          size="small"
          :loading="saving"
          @click="handleSave"
        >
          {{ isEditing ? t('common.update') : t('common.create') }}
        </ActionButton>
      </div>
    </div>

    <div v-loading="pageLoading" class="edit-content">
      <template v-if="!isEditing && !hasSelectedType">
        <section class="chooser-hero">
          <span class="hero-kicker">{{ t('proxyEdit.stepChoose') }}</span>
          <h2 class="page-title">{{ t('proxyEdit.titleChoose') }}</h2>
          <p class="page-subtitle">{{ t('proxyEdit.subtitleChoose') }}</p>
        </section>
        <GuideTypeGrid :items="proxyGuides" @select="handleChooseType" />
      </template>

      <template v-else>
        <section v-if="activeGuide" class="guide-summary">
          <div class="guide-summary-copy">
            <span class="hero-kicker">{{ t(isEditing ? 'proxyEdit.stepEdit' : 'proxyEdit.stepConfigure') }}</span>
            <h2 class="page-title">{{ activeGuide.label }}</h2>
            <p class="page-subtitle">{{ activeGuide.summary }}</p>
            <p class="guide-caution">{{ activeGuide.caution || t('proxyEdit.defaultStoreCopy') }}</p>
            <div class="field-pills">
              <span v-for="field in activeGuide.primaryFields" :key="field" class="field-pill">
                {{ field }}
              </span>
            </div>
          </div>
          <div class="guide-summary-actions">
            <ActionButton
              v-if="!isEditing"
              variant="outline"
              size="small"
              @click="handleResetType"
            >
              {{ t('proxyEdit.changeType') }}
            </ActionButton>
            <ActionButton variant="outline" size="small" @click="router.push('/config')">
              {{ t('common.configPage') }}
            </ActionButton>
          </div>
        </section>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          @submit.prevent
        >
          <ProxyFormLayout
            v-model="form"
            :editing="isEditing"
            :lock-type="!isEditing && hasSelectedType"
          />
        </el-form>
      </template>
    </div>

    <ConfirmDialog
      v-model="leaveDialogVisible"
      :title="t('proxyEdit.unsavedTitle')"
      :message="t('proxyEdit.unsavedMessage')"
      :is-mobile="isMobile"
      @confirm="handleLeaveConfirm"
      @cancel="handleLeaveCancel"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import ActionButton from '@shared/components/ActionButton.vue'
import ConfirmDialog from '@shared/components/ConfirmDialog.vue'
import GuideTypeGrid from '../components/GuideTypeGrid.vue'
import ProxyFormLayout from '../components/proxy-form/ProxyFormLayout.vue'
import { getProxyGuides, getProxyGuideMap, PROXY_TYPE_ORDER } from '../content/guides'
import { useResponsive } from '../composables/useResponsive'
import { getStoreProxy } from '../api/frpc'
import { useProxyStore } from '../stores/proxy'
import {
  createDefaultProxyForm,
  formToStoreProxy,
  storeProxyToForm,
  type ProxyFormData,
  type ProxyType,
} from '../types'
import { useI18n } from '../i18n'

const { isMobile } = useResponsive()
const route = useRoute()
const router = useRouter()
const proxyStore = useProxyStore()
const { t } = useI18n()

const isEditing = computed(() => !!route.params.name)
const selectedType = computed<ProxyType | null>(() => {
  const value = route.query.type
  if (typeof value !== 'string') {
    return null
  }
  return PROXY_TYPE_ORDER.includes(value as ProxyType) ? (value as ProxyType) : null
})
const hasSelectedType = computed(() => selectedType.value != null)
const canSave = computed(() => isEditing.value || hasSelectedType.value)
const proxyGuides = computed(() => getProxyGuides())
const proxyGuideMap = computed(() => getProxyGuideMap())
const activeGuide = computed(() => {
  const type = isEditing.value ? form.value.type : selectedType.value
  if (!type) return null
  return proxyGuideMap.value[type]
})

const pageLoading = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = ref<ProxyFormData>(createDefaultProxyForm())
const dirty = ref(false)
const formSaved = ref(false)
const trackChanges = ref(false)

const rules: FormRules = {
  name: [
    { required: true, message: t('validation.nameRequired'), trigger: 'blur' },
    { min: 1, max: 50, message: t('validation.length1to50'), trigger: 'blur' },
  ],
  type: [{ required: true, message: t('validation.typeRequired'), trigger: 'change' }],
  localPort: [
    {
      validator: (_rule, value, callback) => {
        if (!form.value.pluginType && value == null) {
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
          ['http', 'https', 'tcpmux'].includes(form.value.type) &&
          (!value || value.length === 0) &&
          !form.value.subdomain
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
        if (form.value.healthCheckType === 'http' && !value) {
          callback(new Error(t('validation.healthPathRequired')))
          return
        }
        callback()
      },
      trigger: 'blur',
    },
  ],
}

const goBack = () => {
  router.back()
}

const handleChooseType = (type: string) => {
  router.replace({ query: { type } })
}

const handleResetType = () => {
  router.replace({ query: {} })
}

watch(
  () => form.value,
  () => {
    if (trackChanges.value) {
      dirty.value = true
    }
  },
  { deep: true },
)

watch(
  () => selectedType.value,
  async (type) => {
    if (isEditing.value) return

    trackChanges.value = false
    dirty.value = false
    formSaved.value = false

    const nextForm = createDefaultProxyForm()
    if (type) {
      nextForm.type = type
    }
    form.value = nextForm

    await nextTick()
    trackChanges.value = Boolean(type)
  },
  { immediate: true },
)

const leaveDialogVisible = ref(false)
const leaveResolve = ref<((value: boolean) => void) | null>(null)

onBeforeRouteLeave(async () => {
  if (dirty.value && !formSaved.value) {
    leaveDialogVisible.value = true
    return new Promise<boolean>((resolve) => {
      leaveResolve.value = resolve
    })
  }
})

const handleLeaveConfirm = () => {
  leaveDialogVisible.value = false
  leaveResolve.value?.(true)
}

const handleLeaveCancel = () => {
  leaveDialogVisible.value = false
  leaveResolve.value?.(false)
}

const loadProxy = async () => {
  const name = route.params.name as string
  if (!name) return

  trackChanges.value = false
  dirty.value = false
  pageLoading.value = true
  try {
    const res = await getStoreProxy(name)
    form.value = storeProxyToForm(res)
    await nextTick()
  } catch (err: any) {
    ElMessage.error(t('proxyEdit.loadFailed', { message: err.message }))
    router.push('/proxies')
  } finally {
    pageLoading.value = false
    nextTick(() => {
      trackChanges.value = true
    })
  }
}

const handleSave = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch {
    ElMessage.warning(t('proxyEdit.validationWarning'))
    return
  }

  saving.value = true
  try {
    const data = formToStoreProxy(form.value)
    if (isEditing.value) {
      await proxyStore.updateProxy(form.value.name, data)
      ElMessage.success(t('proxyEdit.updated'))
    } else {
      await proxyStore.createProxy(data)
      ElMessage.success(t('proxyEdit.created'))
    }
    formSaved.value = true
    router.push('/proxies')
  } catch (err: any) {
    ElMessage.error(t('proxyEdit.operationFailed', { message: err.message || 'Unknown error' }))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  if (isEditing.value) {
    loadProxy()
  }
})

watch(
  () => route.params.name,
  (name, oldName) => {
    if (name === oldName || !name) return
    loadProxy()
  },
)
</script>

<style scoped lang="scss">
.proxy-edit-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  max-width: 1080px;
  margin: 0 auto;
}

.edit-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  padding: $spacing-xl 28px 16px;
}

.edit-content {
  flex: 1;
  overflow-y: auto;
  padding: 0 28px 160px;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
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

.chooser-hero,
.guide-summary {
  margin-bottom: 20px;
  padding: 24px;
  border-radius: 22px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background: linear-gradient(135deg, #f8f6ef 0%, #fcfbf7 100%);
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
  background: rgba(255, 255, 255, 0.74);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.guide-summary {
  display: flex;
  justify-content: space-between;
  gap: 20px;
}

.guide-summary-copy {
  max-width: 720px;
}

.guide-summary-actions {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  flex-shrink: 0;
}

.guide-caution {
  margin: 12px 0 0;
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
  .edit-header {
    padding: $spacing-lg;
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .edit-content {
    padding: 0 $spacing-lg 120px;
  }

  .guide-summary {
    flex-direction: column;
  }

  .guide-summary-actions {
    flex-wrap: wrap;
  }
}
</style>
