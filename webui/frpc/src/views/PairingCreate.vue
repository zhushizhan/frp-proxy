<template>
  <div class="pairing-create-page">
    <div class="edit-header">
      <nav class="breadcrumb">
        <router-link to="/visitors" class="breadcrumb-item">{{ t('app.visitors') }}</router-link>
        <span class="breadcrumb-separator">&rsaquo;</span>
        <span class="breadcrumb-current">{{ t('pairing.createTitle') }}</span>
      </nav>
      <div class="header-actions">
        <ActionButton variant="outline" size="small" @click="$router.push('/visitors')">{{ t('common.cancel') }}</ActionButton>
      </div>
    </div>

    <div class="edit-content">
      <!-- Step indicator -->
      <div class="steps-bar">
        <div v-for="(label, i) in stepLabels" :key="i"
          class="step-item"
          :class="{ active: step === i + 1, done: step > i + 1 }">
          <span class="step-dot">{{ step > i + 1 ? '✓' : i + 1 }}</span>
          <span class="step-label">{{ label }}</span>
        </div>
      </div>

      <!-- Step 1: 填写我的配置 -->
      <template v-if="step === 1">
        <section class="chooser-hero">
          <span class="hero-kicker">{{ t('pairing.step1Kicker') }}</span>
          <h2 class="page-title">{{ t('pairing.step1Title') }}</h2>
          <p class="page-subtitle">{{ t('pairing.step1Subtitle') }}</p>
        </section>
        <el-form :model="form" label-position="top" class="pairing-form">
          <div class="field-row two-col">
            <el-form-item :label="t('common.name')" required>
              <el-input v-model="form.name" :placeholder="t('pairing.namePlaceholder')" />
            </el-form-item>
            <el-form-item :label="t('common.type')">
              <el-select v-model="form.type" style="width:100%">
                <el-option label="STCP" value="stcp" />
                <el-option label="XTCP" value="xtcp" />
                <el-option label="SUDP" value="sudp" />
              </el-select>
            </el-form-item>
          </div>
          <el-form-item :label="t('pairing.label')">
            <el-input v-model="form.label" :placeholder="t('pairing.labelPlaceholder')" />
          </el-form-item>
          <div class="field-row two-col">
            <el-form-item :label="t('form.localIP')" required>
              <el-input v-model="form.localIP" placeholder="127.0.0.1" />
            </el-form-item>
            <el-form-item :label="t('form.localPort')" required>
              <el-input-number v-model="form.localPort" :min="1" :max="65535" style="width:100%" />
            </el-form-item>
          </div>
          <el-form-item :label="t('form.secretKey')">
            <el-input v-model="form.secretKey" show-password>
              <template #append>
                <el-button @click="refreshKey">{{ t('pairing.regenerate') }}</el-button>
              </template>
            </el-input>
          </el-form-item>
        </el-form>
        <div class="step-actions">
          <ActionButton :disabled="!step1Valid" @click="goStep2">{{ t('pairing.nextStep') }}</ActionButton>
        </div>
      </template>

      <!-- Step 2: 分享给对方 -->
      <template v-if="step === 2">
        <section class="chooser-hero">
          <span class="hero-kicker">{{ t('pairing.step2Kicker') }}</span>
          <h2 class="page-title">{{ t('pairing.step2Title') }}</h2>
          <p class="page-subtitle">{{ t('pairing.step2Subtitle') }}</p>
        </section>

        <div class="share-card">
          <div class="share-meta">
            <span class="share-type-badge">{{ form.type.toUpperCase() }}</span>
            <span class="share-name">{{ form.name }}</span>
            <span v-if="form.label" class="share-label">{{ form.label }}</span>
          </div>

          <el-tabs v-model="shareTab" class="share-tabs">
            <el-tab-pane :label="t('pairing.shareCodeTab')" name="code">
              <div class="share-code-block">
                <el-input
                  type="textarea"
                  :value="shareCode"
                  readonly
                  :rows="4"
                  resize="none"
                />
                <ActionButton size="small" @click="copyShareCode">{{ t('pairing.copyCode') }}</ActionButton>
              </div>
              <p class="share-hint">{{ t('pairing.shareCodeHint') }}</p>
            </el-tab-pane>
            <el-tab-pane :label="t('pairing.shareJsonTab')" name="json">
              <div class="share-code-block">
                <el-input
                  type="textarea"
                  :value="shareJson"
                  readonly
                  :rows="8"
                  resize="none"
                />
                <ActionButton size="small" @click="copyShareJson">{{ t('pairing.copyJson') }}</ActionButton>
              </div>
            </el-tab-pane>
          </el-tabs>
        </div>

        <div class="step-actions">
          <ActionButton variant="outline" @click="step = 1">{{ t('common.back') }}</ActionButton>
          <ActionButton :loading="saving" @click="saveAndFinish">{{ t('pairing.saveAndFinish') }}</ActionButton>
        </div>
      </template>

      <!-- Step 3: 完成 -->
      <template v-if="step === 3">
        <div class="done-card">
          <div class="done-icon">✓</div>
          <h2>{{ t('pairing.doneTitle') }}</h2>
          <p>{{ t('pairing.doneCopy') }}</p>
          <div class="done-actions">
            <ActionButton variant="outline" @click="$router.push('/visitors')">{{ t('common.backToVisitors') }}</ActionButton>
            <ActionButton @click="$router.push('/pairing/import')">{{ t('pairing.goImport') }}</ActionButton>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useProxyStore } from '../stores/proxy'
import { getSettings } from '../api/frpc'
import { generateSecretKey, encodePairConfig } from '../utils/pairing'
import type { PairSharePayload, PairingProxyType } from '../types/pairing'
import { useI18n } from '../i18n'
import ActionButton from '@shared/components/ActionButton.vue'

const { t } = useI18n()
const proxyStore = useProxyStore()

const step = ref(1)
const shareTab = ref('code')
const saving = ref(false)
const serverAddr = ref('')
const serverUser = ref('')

const form = ref({
  name: '',
  type: 'stcp' as PairingProxyType,
  label: '',
  localIP: '127.0.0.1',
  localPort: 22,
  secretKey: generateSecretKey(),
})

const stepLabels = computed(() => [
  t('pairing.step1Label'),
  t('pairing.step2Label'),
  t('pairing.step3Label'),
])

const step1Valid = computed(() =>
  form.value.name.trim() !== '' &&
  form.value.localPort > 0 &&
  form.value.secretKey.trim() !== ''
)

const sharePayload = computed<PairSharePayload>(() => ({
  v: 1,
  type: form.value.type,
  serverName: form.value.name,
  secretKey: form.value.secretKey,
  serverAddr: serverAddr.value,
  serverUser: serverUser.value || undefined,
  label: form.value.label || undefined,
}))

const shareCode = computed(() => encodePairConfig(sharePayload.value))
const shareJson = computed(() => JSON.stringify(sharePayload.value, null, 2))

function refreshKey() {
  form.value.secretKey = generateSecretKey()
}

async function goStep2() {
  if (!step1Valid.value) return
  step.value = 2
}

function copyShareCode() {
  navigator.clipboard.writeText(shareCode.value).then(() => {
    ElMessage.success(t('pairing.codeCopied'))
  })
}

function copyShareJson() {
  navigator.clipboard.writeText(shareJson.value).then(() => {
    ElMessage.success(t('pairing.jsonCopied'))
  })
}

async function saveAndFinish() {
  saving.value = true
  try {
    // Save the proxy (A side) - must use nested ProxyDefinition format
    const proxyType = form.value.type
    const block: Record<string, any> = {
      secretKey: form.value.secretKey,
    }
    if (form.value.localIP && form.value.localIP !== '127.0.0.1') {
      block.localIP = form.value.localIP
    }
    if (form.value.localPort) {
      block.localPort = form.value.localPort
    }
    await proxyStore.createProxy({
      name: form.value.name,
      type: proxyType,
      [proxyType]: block,
    })
    step.value = 3
    ElMessage.success(t('pairing.proxySaved'))
  } catch (e: any) {
    ElMessage.error(t('pairing.saveFailed') + ': ' + (e?.message || e))
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try {
    const settings = await getSettings()
    serverAddr.value = settings.serverAddr || ''
    serverUser.value = settings.user || ''
  } catch (_) {
    // ignore
  }
})
</script>

<style scoped lang="scss">
@use '../assets/css/var';

$color-text-secondary: #64748b;
$spacing-lg: 24px;

.pairing-create-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.edit-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-lg 32px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.07);
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
  &:hover { text-decoration: underline; }
}

.breadcrumb-separator { color: $color-text-secondary; }

.breadcrumb-current {
  font-weight: 600;
  color: #1e293b;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.edit-content {
  padding: 32px 32px 120px;
  max-width: 760px;
  width: 100%;
  margin: 0 auto;
}

.steps-bar {
  display: flex;
  gap: 0;
  margin-bottom: 36px;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 16px;
}

.step-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 20px 0 0;
  font-size: 13px;
  color: $color-text-secondary;

  &.active {
    color: #16a34a;
    font-weight: 700;
    .step-dot { background: #16a34a; color: #fff; }
  }
  &.done {
    color: #16a34a;
    .step-dot { background: #16a34a; color: #fff; }
  }
}

.step-dot {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: #e2e8f0;
  color: #64748b;
  font-size: 11px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chooser-hero {
  margin-bottom: 28px;
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
  color: #1e293b;
  margin: 0 0 8px;
}

.page-subtitle {
  font-size: 14px;
  color: $color-text-secondary;
  margin: 0;
  line-height: 1.6;
}

.pairing-form {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 24px;
}

.field-row {
  display: grid;
  gap: 16px;
  &.two-col { grid-template-columns: 1fr 1fr; }
}

.step-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
}

.share-card {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 16px;
}

.share-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}

.share-type-badge {
  background: #dcfce7;
  color: #16a34a;
  font-size: 11px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 999px;
  letter-spacing: 0.06em;
}

.share-name {
  font-weight: 600;
  font-size: 15px;
  color: #1e293b;
}

.share-label {
  font-size: 13px;
  color: $color-text-secondary;
}

.share-code-block {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: flex-start;
}

.share-hint {
  font-size: 12px;
  color: $color-text-secondary;
  margin-top: 8px;
}

.share-tabs {
  margin-top: 8px;
}

.done-card {
  text-align: center;
  padding: 60px 24px;
  background: #fff;
  border: 1px solid #e2e8f0;
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

.done-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  margin-top: 24px;
}
</style>
