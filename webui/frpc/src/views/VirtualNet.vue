<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ t('virtualNet.title') }}</h1>
        <p class="page-subtitle">{{ t('virtualNet.subtitle') }}</p>
      </div>
    </div>

    <!-- Notice banner -->
    <el-alert
      :title="t('virtualNet.alphaNotice')"
      type="warning"
      :closable="false"
      show-icon
      class="alpha-alert"
    />

    <!-- Node role selector -->
    <div class="section">
      <div class="section-header">
        <div>
          <h3 class="section-title">{{ t('virtualNet.roleTitle') }}</h3>
          <p class="section-copy">{{ t('virtualNet.roleCopy') }}</p>
        </div>
      </div>
      <div class="role-grid">
        <button
          class="role-card"
          :class="{ active: form.role === 'server' }"
          @click="form.role = 'server'"
        >
          <div class="role-icon">&#x1F5A5;</div>
          <div class="role-label">{{ t('virtualNet.roleServer') }}</div>
          <div class="role-desc">{{ t('virtualNet.roleServerDesc') }}</div>
        </button>
        <button
          class="role-card"
          :class="{ active: form.role === 'client' }"
          @click="form.role = 'client'"
        >
          <div class="role-icon">&#x1F4BB;</div>
          <div class="role-label">{{ t('virtualNet.roleClient') }}</div>
          <div class="role-desc">{{ t('virtualNet.roleClientDesc') }}</div>
        </button>
      </div>
    </div>

    <!-- Form -->
    <div class="section">
      <div class="section-header">
        <div>
          <h3 class="section-title">{{ t('virtualNet.configTitle') }}</h3>
        </div>
      </div>
      <el-form :model="form" label-position="top" class="vnet-form">
        <el-form-item :label="t('virtualNet.fieldName')">
          <el-input v-model="form.name" :placeholder="form.role === 'server' ? 'vnet-server' : 'vnet-visitor'" />
          <div class="field-hint">{{ t('virtualNet.fieldNameHint') }}</div>
        </el-form-item>

        <el-form-item :label="t('virtualNet.fieldVirtualIP')">
          <el-input v-model="form.virtualIP" placeholder="100.86.0.1/24" />
          <div class="field-hint">{{ t('virtualNet.fieldVirtualIPHint') }}</div>
        </el-form-item>

        <el-form-item :label="t('virtualNet.fieldSecretKey')">
          <el-input v-model="form.secretKey" type="password" show-password :placeholder="t('virtualNet.fieldSecretKeyPlaceholder')" />
          <div class="field-hint">{{ t('virtualNet.fieldSecretKeyHint') }}</div>
        </el-form-item>

        <!-- Client-only fields -->
        <template v-if="form.role === 'client'">
          <el-form-item :label="t('virtualNet.fieldServerName')">
            <el-input v-model="form.serverName" placeholder="vnet-server" />
            <div class="field-hint">{{ t('virtualNet.fieldServerNameHint') }}</div>
          </el-form-item>

          <el-form-item :label="t('virtualNet.fieldDestIP')">
            <el-input v-model="form.destIP" placeholder="100.86.0.1" />
            <div class="field-hint">{{ t('virtualNet.fieldDestIPHint') }}</div>
          </el-form-item>
        </template>
      </el-form>
    </div>

    <!-- Generated config preview -->
    <div class="section" v-if="generatedConfig">
      <div class="section-header">
        <div>
          <h3 class="section-title">{{ t('virtualNet.previewTitle') }}</h3>
          <p class="section-copy">{{ t('virtualNet.previewCopy') }}</p>
        </div>
        <button class="icon-btn" @click="copyConfig" :title="t('virtualNet.copyConfig')">
          <span v-if="copied">&#x2713;</span>
          <span v-else>&#x1F4CB;</span>
        </button>
      </div>
      <pre class="config-preview">{{ generatedConfig }}</pre>
    </div>

    <!-- Actions -->
    <div class="action-bar">
      <el-button @click="generate" type="primary" :disabled="!canGenerate">
        {{ t('virtualNet.btnGenerate') }}
      </el-button>
      <el-button @click="applyToStore" type="success" :disabled="!generatedConfig || applying" :loading="applying">
        {{ t('virtualNet.btnApply') }}
      </el-button>
      <el-button @click="resetForm">
        {{ t('virtualNet.btnReset') }}
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from '../i18n'
import {
  createStoreProxy,
  createStoreVisitor,
} from '../api/frpc'

const { t } = useI18n()

const form = ref({
  role: 'server' as 'server' | 'client',
  name: '',
  virtualIP: '',
  secretKey: '',
  serverName: '',
  destIP: '',
})

const generatedConfig = ref('')
const applying = ref(false)
const copied = ref(false)

const canGenerate = computed(() => {
  if (!form.value.name || !form.value.virtualIP || !form.value.secretKey) return false
  if (form.value.role === 'client' && (!form.value.serverName || !form.value.destIP)) return false
  return true
})

function generate() {
  const f = form.value
  if (f.role === 'server') {
    generatedConfig.value = [
      `# frpc.toml - VirtualNet server side`,
      `featureGates = { VirtualNet = true }`,
      `virtualNet.address = "${f.virtualIP}"`,
      ``,
      `[[proxies]]`,
      `name = "${f.name}"`,
      `type = "stcp"`,
      `secretKey = "${f.secretKey}"`,
      `[proxies.plugin]`,
      `type = "virtual_net"`,
    ].join('\n')
  } else {
    generatedConfig.value = [
      `# frpc.toml - VirtualNet client side`,
      `featureGates = { VirtualNet = true }`,
      `virtualNet.address = "${f.virtualIP}"`,
      ``,
      `[[visitors]]`,
      `name = "${f.name}"`,
      `type = "stcp"`,
      `serverName = "${f.serverName}"`,
      `secretKey = "${f.secretKey}"`,
      `bindPort = -1`,
      `[visitors.plugin]`,
      `type = "virtual_net"`,
      `destinationIP = "${f.destIP}"`,
    ].join('\n')
  }
}

async function applyToStore() {
  applying.value = true
  try {
    const f = form.value
    if (f.role === 'server') {
      await createStoreProxy({
        name: f.name,
        type: 'stcp',
        stcp: {
          secretKey: f.secretKey,
          plugin: { type: 'virtual_net' },
        },
      })
    } else {
      await createStoreVisitor({
        name: f.name,
        type: 'stcp',
        stcp: {
          serverName: f.serverName,
          secretKey: f.secretKey,
          bindPort: -1,
          plugin: { type: 'virtual_net', destinationIP: f.destIP },
        },
      })
    }
    ElMessage.success(t('virtualNet.applySuccess'))
  } catch (e: any) {
    ElMessage.error(t('virtualNet.applyFailed') + ': ' + (e?.message ?? e))
  } finally {
    applying.value = false
  }
}

async function copyConfig() {
  try {
    await navigator.clipboard.writeText(generatedConfig.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    ElMessage.warning(t('virtualNet.copyFailed'))
  }
}

function resetForm() {
  form.value = { role: 'server', name: '', virtualIP: '', secretKey: '', serverName: '', destIP: '' }
  generatedConfig.value = ''
}
</script>

<style lang="scss" scoped>
.page {
  height: 100%;
  overflow-y: auto;
  padding: $spacing-xl;
  @include flex-column;
  gap: $spacing-lg;

  @include mobile {
    padding: $spacing-lg;
  }
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: $spacing-md;
}

.alpha-alert {
  border-radius: $radius-md;
}

.section {
  background: $color-bg-primary;
  border: 1px solid $color-border-light;
  border-radius: $radius-md;
  padding: $spacing-lg;
  @include flex-column;
  gap: $spacing-md;
}

.section-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: $spacing-md;
}

.section-title {
  font-size: $font-size-lg;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
  margin: 0;
}

.section-copy {
  font-size: $font-size-sm;
  color: $color-text-muted;
  margin: 4px 0 0;
}

.role-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: $spacing-md;

  @include mobile {
    grid-template-columns: 1fr;
  }
}

.role-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-lg;
  border: 2px solid $color-border-light;
  border-radius: $radius-md;
  background: $color-bg-secondary;
  cursor: pointer;
  transition: all $transition-fast;
  text-align: center;

  &:hover {
    border-color: $color-border;
    background: $color-bg-hover;
  }

  &.active {
    border-color: #606266;
    background: $color-bg-hover;
  }
}

.role-icon {
  font-size: 28px;
}

.role-label {
  font-size: $font-size-md;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
}

.role-desc {
  font-size: $font-size-sm;
  color: $color-text-muted;
}

.vnet-form {
  max-width: 560px;
}

.field-hint {
  font-size: $font-size-xs;
  color: $color-text-muted;
  margin-top: 4px;
}

.config-preview {
  background: $color-bg-tertiary;
  border: 1px solid $color-border-light;
  border-radius: $radius-md;
  padding: $spacing-md;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: $font-size-sm;
  line-height: 1.6;
  overflow-x: auto;
  white-space: pre;
  color: $color-text-primary;
  margin: 0;
}

.action-bar {
  display: flex;
  gap: $spacing-md;
  flex-wrap: wrap;
}
</style>
