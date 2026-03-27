<template>
  <div class="pairing-import-page">
    <div class="edit-header">
      <nav class="breadcrumb">
        <router-link to="/visitors" class="breadcrumb-item">{{ t('app.visitors') }}</router-link>
        <span class="breadcrumb-separator">&rsaquo;</span>
        <span class="breadcrumb-current">{{ t('pairing.importTitle') }}</span>
      </nav>
      <div class="header-actions">
        <ActionButton variant="outline" size="small" @click="$router.push('/visitors')">{{ t('common.cancel') }}</ActionButton>
      </div>
    </div>

    <div class="edit-content">
      <!-- Step 1: 导入分享码 -->
      <template v-if="step === 1">
        <section class="chooser-hero">
          <span class="hero-kicker">{{ t('pairing.importStep1Kicker') }}</span>
          <h2 class="page-title">{{ t('pairing.importStep1Title') }}</h2>
          <p class="page-subtitle">{{ t('pairing.importStep1Subtitle') }}</p>
        </section>

        <el-tabs v-model="importMode" class="import-tabs">
          <el-tab-pane :label="t('pairing.importByCode')" name="code">
            <div class="import-code-area">
              <el-input
                v-model="shareCode"
                type="textarea"
                :rows="4"
                :placeholder="t('pairing.importCodePlaceholder')"
                @input="decodeError = ''"
              />
              <p v-if="decodeError" class="error-text">{{ decodeError }}</p>
            </div>
          </el-tab-pane>
          <el-tab-pane :label="t('pairing.importByJson')" name="json">
            <div class="import-code-area">
              <el-input
                v-model="jsonText"
                type="textarea"
                :rows="6"
                :placeholder="t('pairing.importJsonPlaceholder')"
                @input="decodeError = ''"
              />
              <p v-if="decodeError" class="error-text">{{ decodeError }}</p>
            </div>
          </el-tab-pane>
        </el-tabs>

        <div class="step-actions">
          <ActionButton :loading="parsing" @click="handleParse">{{ t('pairing.importParse') }}</ActionButton>
        </div>
      </template>

      <!-- Step 2: 填写本地访问端口 -->
      <template v-if="step === 2 && payload">
        <section class="chooser-hero">
          <span class="hero-kicker">{{ t('pairing.importStep2Kicker') }}</span>
          <h2 class="page-title">{{ t('pairing.importStep2Title') }}</h2>
          <p class="page-subtitle">{{ t('pairing.importStep2Subtitle') }}</p>
        </section>

        <!-- 分享码信息预览 -->
        <div class="info-card">
          <div class="info-row">
            <span class="info-label">{{ t('pairing.sharedProxyName') }}</span>
            <span class="info-value">{{ payload.serverName }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">{{ t('common.type') }}</span>
            <span class="info-value type-badge">{{ payload.type.toUpperCase() }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">{{ t('pairing.sharedServer') }}</span>
            <span class="info-value">{{ payload.serverAddr }}</span>
          </div>
          <div v-if="payload.label" class="info-row">
            <span class="info-label">{{ t('pairing.label') }}</span>
            <span class="info-value">{{ payload.label }}</span>
          </div>
        </div>

        <el-form :model="visitorForm" label-position="top" class="pairing-form">
          <div class="field-row two-col">
            <el-form-item :label="t('common.name')" required>
              <el-input v-model="visitorForm.name" :placeholder="t('pairing.visitorNamePlaceholder')" />
            </el-form-item>
            <el-form-item :label="t('pairing.bindPort')" required>
              <el-input-number
                v-model="visitorForm.bindPort"
                :min="1"
                :max="65535"
                style="width:100%"
                controls-position="right"
              />
            </el-form-item>
          </div>
          <el-form-item :label="t('pairing.bindAddr')">
            <el-input v-model="visitorForm.bindAddr" placeholder="127.0.0.1" />
          </el-form-item>
        </el-form>

        <div class="step-actions">
          <ActionButton variant="outline" @click="step = 1">{{ t('common.cancel') }}</ActionButton>
          <ActionButton :loading="saving" @click="handleSave">{{ t('pairing.importSave') }}</ActionButton>
        </div>
      </template>

      <!-- Step 3: 完成 + 访问地址 -->
      <template v-if="step === 3">
        <div class="done-card">
          <div class="done-icon">✓</div>
          <h3>{{ t('pairing.importDoneTitle') }}</h3>
          <p class="done-subtitle">{{ t('pairing.importDoneSubtitle') }}</p>

          <div class="access-block">
            <p class="access-label">{{ t('pairing.accessAddress') }}</p>
            <div class="access-url-row">
              <el-input v-model="accessUrl" readonly class="access-url-input" />
              <ActionButton size="small" @click="copyAccessUrl">{{ t('pairing.copyAddress') }}</ActionButton>
              <ActionButton
                v-if="canOpenUrl"
                size="small"
                variant="outline"
                @click="openAccessUrl"
              >{{ t('pairing.openAddress') }}</ActionButton>
            </div>
            <p class="access-hint">{{ t('pairing.accessHint') }}</p>
          </div>

          <div class="done-actions">
            <ActionButton variant="outline" @click="$router.push('/visitors')">{{ t('common.backToVisitors') }}</ActionButton>
            <ActionButton @click="resetAndImportAnother">{{ t('pairing.importAnother') }}</ActionButton>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import ActionButton from '../components/ActionButton.vue'
import { useVisitorStore } from '../stores/visitor'
import { decodePairConfig, buildAccessUrl } from '../utils/pairing'
import type { PairSharePayload } from '../types/pairing'
import { useI18n } from '../i18n'

const { t } = useI18n()
const visitorStore = useVisitorStore()

const step = ref(1)
const importMode = ref<'code' | 'json'>('code')
const shareCode = ref('')
const jsonText = ref('')
const decodeError = ref('')
const parsing = ref(false)
const saving = ref(false)
const payload = ref<PairSharePayload | null>(null)
const accessUrl = ref('')

const visitorForm = ref({
  name: '',
  bindAddr: '127.0.0.1',
  bindPort: 6000,
})

const canOpenUrl = computed(() => {
  return accessUrl.value.startsWith('http://') || accessUrl.value.startsWith('https://')
})

const handleParse = async () => {
  parsing.value = true
  decodeError.value = ''
  try {
    let result
    if (importMode.value === 'code') {
      const code = shareCode.value.trim()
      if (!code) {
        decodeError.value = t('pairing.importCodeEmpty')
        return
      }
      result = decodePairConfig(code)
    } else {
      const text = jsonText.value.trim()
      if (!text) {
        decodeError.value = t('pairing.importJsonEmpty')
        return
      }
      try {
        const parsed = JSON.parse(text)
        result = { ok: true as const, payload: parsed }
      } catch {
        result = { ok: false as const, error: 'Invalid JSON format' }
      }
    }

    if (!result.ok) {
      decodeError.value = result.error
      return
    }

    payload.value = result.payload
    // Pre-fill visitor name
    visitorForm.value.name = `visitor-${result.payload.serverName}`
    step.value = 2
  } finally {
    parsing.value = false
  }
}

const handleSave = async () => {
  if (!payload.value) return
  if (!visitorForm.value.name.trim()) {
    ElMessage.error(t('pairing.visitorNameRequired'))
    return
  }
  if (!visitorForm.value.bindPort || visitorForm.value.bindPort < 1) {
    ElMessage.error(t('pairing.bindPortRequired'))
    return
  }

  saving.value = true
  try {
    const p = payload.value
    const visitorDef: any = {
      name: visitorForm.value.name,
      type: p.type,
      [p.type]: {
        name: visitorForm.value.name,
        type: p.type,
        enabled: true,
        serverName: p.serverName,
        serverUser: p.serverUser || '',
        secretKey: p.secretKey,
        bindAddr: visitorForm.value.bindAddr || '127.0.0.1',
        bindPort: visitorForm.value.bindPort,
      },
    }

    await visitorStore.createVisitor(visitorDef)

    // Build access URL
    accessUrl.value = buildAccessUrl(
      visitorForm.value.bindAddr || '127.0.0.1',
      visitorForm.value.bindPort,
      p.type,
    )

    step.value = 3
    ElMessage.success(t('pairing.importSaveSuccess'))
  } catch (err: any) {
    ElMessage.error(`${t('pairing.importSaveFailed')}: ${err?.message || err}`)
  } finally {
    saving.value = false
  }
}

const copyAccessUrl = () => {
  navigator.clipboard.writeText(accessUrl.value).then(() => {
    ElMessage.success(t('pairing.copySuccess'))
  }).catch(() => {
    ElMessage.error(t('pairing.copyFailed'))
  })
}

const openAccessUrl = () => {
  window.open(accessUrl.value, '_blank')
}

const resetAndImportAnother = () => {
  step.value = 1
  importMode.value = 'code'
  shareCode.value = ''
  jsonText.value = ''
  decodeError.value = ''
  payload.value = null
  accessUrl.value = ''
  visitorForm.value = { name: '', bindAddr: '127.0.0.1', bindPort: 6000 }
}
</script>

<style scoped lang="scss">
@use '@/assets/css/var' as *;
@use '@/assets/css/form-layout';

.pairing-import-page {
  min-height: 100vh;
  background: $color-bg;
}

.edit-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-xl $spacing-2xl;
  border-bottom: 1px solid $color-border;
  background: $color-surface;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
}

.breadcrumb-item {
  color: $color-text-secondary;
  text-decoration: none;
  &:hover { color: $color-primary; }
}

.breadcrumb-separator { color: $color-text-secondary; }
.breadcrumb-current { color: $color-text-primary; font-weight: 600; }

.header-actions {
  display: flex;
  gap: 10px;
}

.edit-content {
  max-width: 720px;
  margin: 0 auto;
  padding: $spacing-2xl;
}

.chooser-hero {
  margin-bottom: 32px;
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

.page-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 8px;
  color: $color-text-primary;
}

.page-subtitle {
  font-size: 15px;
  color: $color-text-secondary;
  line-height: 1.6;
  margin: 0;
}

.import-tabs {
  margin-bottom: 24px;
}

.import-code-area {
  padding: 16px 0 8px;
}

.error-text {
  color: #ef4444;
  font-size: 13px;
  margin-top: 8px;
}

.step-actions {
  display: flex;
  gap: 12px;
  margin-top: 32px;
}

.info-card {
  background: $color-surface;
  border: 1px solid $color-border;
  border-radius: 12px;
  padding: 20px 24px;
  margin-bottom: 28px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.info-label {
  width: 120px;
  font-size: 13px;
  color: $color-text-secondary;
  flex-shrink: 0;
}

.info-value {
  font-size: 14px;
  font-weight: 500;
  color: $color-text-primary;
}

.type-badge {
  padding: 2px 10px;
  border-radius: 999px;
  background: #e0f2fe;
  color: #0369a1;
  font-size: 12px;
  font-weight: 700;
}

.pairing-form {
  margin-bottom: 24px;
}

.done-card {
  text-align: center;
  padding: 48px 24px;
  background: $color-surface;
  border: 1px solid $color-border;
  border-radius: 16px;
}

.done-icon {
  width: 56px;
  height: 56px;
  background: #dcfce7;
  color: #16a34a;
  font-size: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px;
}

.done-subtitle {
  color: $color-text-secondary;
  font-size: 14px;
  margin: 8px 0 32px;
}

.access-block {
  background: #f8fafc;
  border: 1px solid $color-border;
  border-radius: 12px;
  padding: 20px 24px;
  margin-bottom: 32px;
  text-align: left;
}

.access-label {
  font-size: 13px;
  font-weight: 600;
  color: $color-text-secondary;
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.access-url-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.access-url-input {
  flex: 1;
  font-family: monospace;
}

.access-hint {
  font-size: 12px;
  color: $color-text-secondary;
  margin-top: 10px;
}

.done-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}
</style>
