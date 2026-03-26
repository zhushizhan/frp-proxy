<template>
  <div class="visitor-edit-page">
    <div class="edit-header">
      <nav class="breadcrumb">
        <router-link to="/visitors" class="breadcrumb-item">{{ t('app.visitors') }}</router-link>
        <span class="breadcrumb-separator">&rsaquo;</span>
        <span class="breadcrumb-current">
          {{ isEditing ? t('visitorEdit.crumbEdit') : hasSelectedType ? t('visitorEdit.crumbConfigure') : t('visitorEdit.crumbChoose') }}
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
          <span class="hero-kicker">{{ t('visitorEdit.stepChoose') }}</span>
          <h2 class="page-title">{{ t('visitorEdit.titleChoose') }}</h2>
          <p class="page-subtitle">{{ t('visitorEdit.subtitleChoose') }}</p>
        </section>
        <GuideTypeGrid :items="visitorGuides" @select="handleChooseType" />
      </template>

      <template v-else>
        <section v-if="activeGuide" class="guide-summary">
          <div class="guide-summary-copy">
            <span class="hero-kicker">{{ t(isEditing ? 'visitorEdit.stepEdit' : 'visitorEdit.stepConfigure') }}</span>
            <h2 class="page-title">{{ activeGuide.label }}</h2>
            <p class="page-subtitle">{{ activeGuide.summary }}</p>
            <p class="guide-caution">{{ activeGuide.caution || t('visitorEdit.defaultStoreCopy') }}</p>
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
              {{ t('visitorEdit.changeType') }}
            </ActionButton>
            <ActionButton variant="outline" size="small" @click="router.push('/config')">
              {{ t('common.configPage') }}
            </ActionButton>
          </div>
        </section>

        <el-form
          ref="formRef"
          :model="form"
          :rules="formRules"
          label-position="top"
          @submit.prevent
        >
          <VisitorFormLayout
            v-model="form"
            :editing="isEditing"
            :lock-type="!isEditing && hasSelectedType"
          />
        </el-form>
      </template>
    </div>

    <ConfirmDialog
      v-model="leaveDialogVisible"
      :title="t('visitorEdit.unsavedTitle')"
      :message="t('visitorEdit.unsavedMessage')"
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
import VisitorFormLayout from '../components/visitor-form/VisitorFormLayout.vue'
import { getVisitorGuideMap, getVisitorGuides, VISITOR_TYPE_ORDER } from '../content/guides'
import { useResponsive } from '../composables/useResponsive'
import { getStoreVisitor } from '../api/frpc'
import { useVisitorStore } from '../stores/visitor'
import {
  createDefaultVisitorForm,
  formToStoreVisitor,
  storeVisitorToForm,
  type VisitorFormData,
  type VisitorType,
} from '../types'
import { useI18n } from '../i18n'

const { isMobile } = useResponsive()
const route = useRoute()
const router = useRouter()
const visitorStore = useVisitorStore()
const { t } = useI18n()

const isEditing = computed(() => !!route.params.name)
const selectedType = computed<VisitorType | null>(() => {
  const value = route.query.type
  if (typeof value !== 'string') {
    return null
  }
  return VISITOR_TYPE_ORDER.includes(value as VisitorType) ? (value as VisitorType) : null
})
const hasSelectedType = computed(() => selectedType.value != null)
const canSave = computed(() => isEditing.value || hasSelectedType.value)
const visitorGuides = computed(() => getVisitorGuides())
const visitorGuideMap = computed(() => getVisitorGuideMap())
const activeGuide = computed(() => {
  const type = isEditing.value ? form.value.type : selectedType.value
  if (!type) return null
  return visitorGuideMap.value[type]
})

const pageLoading = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = ref<VisitorFormData>(createDefaultVisitorForm())
const dirty = ref(false)
const formSaved = ref(false)
const trackChanges = ref(false)

const formRules: FormRules = {
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
        if (form.value.type === 'sudp' && value < 1) {
          callback(new Error(t('validation.bindPortSudpMin')))
          return
        }
        if (form.value.type !== 'sudp' && value === 0) {
          callback(new Error(t('validation.bindPortNotZero')))
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

    const nextForm = createDefaultVisitorForm()
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

const loadVisitor = async () => {
  const name = route.params.name as string
  if (!name) return

  trackChanges.value = false
  dirty.value = false
  pageLoading.value = true
  try {
    const res = await getStoreVisitor(name)
    form.value = storeVisitorToForm(res)
    await nextTick()
  } catch (err: any) {
    ElMessage.error(t('visitorEdit.loadFailed', { message: err.message }))
    router.push('/visitors')
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
    ElMessage.warning(t('visitorEdit.validationWarning'))
    return
  }

  saving.value = true
  try {
    const data = formToStoreVisitor(form.value)
    if (isEditing.value) {
      await visitorStore.updateVisitor(form.value.name, data)
      ElMessage.success(t('visitorEdit.updated'))
    } else {
      await visitorStore.createVisitor(data)
      ElMessage.success(t('visitorEdit.created'))
    }
    formSaved.value = true
    router.push('/visitors')
  } catch (err: any) {
    ElMessage.error(t('visitorEdit.operationFailed', { message: err.message || 'Unknown error' }))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  if (isEditing.value) {
    loadVisitor()
  }
})

watch(
  () => route.params.name,
  (name, oldName) => {
    if (name === oldName || !name) return
    loadVisitor()
  },
)
</script>

<style scoped lang="scss">
.visitor-edit-page {
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
  background: linear-gradient(135deg, #eff6f1 0%, #fbfdfb 100%);
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
