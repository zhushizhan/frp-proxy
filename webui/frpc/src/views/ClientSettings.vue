<template>
  <div class="settings-page">
    <div class="page-header">
      <div class="title-section">
        <h1 class="page-title">{{ t('clientSettings.title') }}</h1>
        <p class="page-subtitle">{{ t('clientSettings.subtitle') }}</p>
        <p v-if="form.configPath" class="config-path">
          {{ t('common.configPage') }}: <code>{{ form.configPath }}</code>
        </p>
      </div>
      <div class="page-actions">
        <ActionButton variant="outline" size="small" @click="fetchData">
          {{ t('common.refresh') }}
        </ActionButton>
        <ActionButton size="small" @click="handleSave">
          {{ t('clientSettings.saveRestart') }}
        </ActionButton>
      </div>
    </div>

    <el-alert
      :title="t('clientSettings.restartAlert')"
      type="warning"
      :closable="false"
      show-icon
      class="restart-alert"
    />

    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
      status-icon
      scroll-to-error
      class="settings-form"
    >
      <el-card shadow="hover" class="settings-card">
        <template #header>{{ t('clientSettings.sections.connection') }}</template>
        <div class="grid three">
          <el-form-item :label="label('serverAddr')" prop="serverAddr">
            <el-input v-model="form.serverAddr" />
          </el-form-item>
          <el-form-item :label="label('serverPort')" prop="serverPort">
            <el-input-number v-model="form.serverPort" :min="1" :max="65535" controls-position="right" />
          </el-form-item>
          <el-form-item :label="label('loginFailExit')">
            <el-switch v-model="form.loginFailExit" />
          </el-form-item>
        </div>
        <div class="grid three">
          <el-form-item :label="label('user')">
            <el-input v-model="form.user" />
          </el-form-item>
          <el-form-item :label="label('clientID')">
            <el-input v-model="form.clientID" />
          </el-form-item>
          <el-form-item :label="label('dnsServer')">
            <el-input v-model="form.dnsServer" placeholder="8.8.8.8" />
          </el-form-item>
        </div>
        <el-form-item :label="label('natHoleStunServer')">
          <el-input v-model="form.natHoleStunServer" placeholder="stun.easyvoip.com:3478" />
        </el-form-item>
      </el-card>

      <el-card shadow="hover" class="settings-card">
        <template #header>{{ t('clientSettings.sections.authStore') }}</template>
        <div class="grid three">
          <el-form-item :label="label('authMethod')">
            <el-select v-model="form.authMethod">
              <el-option :label="t('clientSettings.options.token')" value="token" />
              <el-option :label="t('clientSettings.options.oidc')" value="oidc" />
            </el-select>
          </el-form-item>
          <el-form-item :label="label('authAdditionalScopes')">
            <el-checkbox-group v-model="form.authAdditionalScopes" class="checkbox-group">
              <el-checkbox label="HeartBeats">HeartBeats</el-checkbox>
              <el-checkbox label="NewWorkConns">NewWorkConns</el-checkbox>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item :label="label('storePath')">
            <el-input v-model="form.storePath" placeholder="./frpc_store.json" />
          </el-form-item>
        </div>
        <div class="field-help">{{ t('clientSettings.helpers.store') }}</div>

        <template v-if="isTokenAuth">
          <div class="grid two">
            <el-form-item :label="label('authTokenSource')">
              <el-select v-model="form.authTokenSourceType">
                <el-option :label="t('clientSettings.options.inline')" value="inline" />
                <el-option :label="t('clientSettings.options.file')" value="file" />
                <el-option :label="t('clientSettings.options.exec')" value="exec" :disabled="!supportsExecTokenSource" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="usesInlineToken" :label="label('authToken')" prop="authToken">
              <el-input v-model="form.authToken" show-password />
            </el-form-item>
            <el-form-item
              v-else-if="usesFileTokenSource"
              :label="label('authTokenSourceFile')"
              prop="authTokenSourceFile"
            >
              <div class="path-upload-row">
                <el-input v-model="form.authTokenSourceFile" placeholder="./token.txt" />
                <el-button
                  :loading="uploadLoading.authTokenSourceFile"
                  @click="pickAndUploadFile('authTokenSourceFile')"
                >
                  {{ t('common.upload') }}
                </el-button>
              </div>
            </el-form-item>
            <el-form-item v-else :label="label('authTokenSource')">
              <el-input :model-value="t('clientSettings.options.exec')" disabled />
            </el-form-item>
          </div>
        </template>

        <template v-else>
          <div class="grid three">
            <el-form-item :label="label('oidcClientID')" prop="oidcClientID">
              <el-input v-model="form.oidcClientID" />
            </el-form-item>
            <el-form-item :label="label('oidcClientSecret')">
              <el-input v-model="form.oidcClientSecret" show-password />
            </el-form-item>
            <el-form-item :label="label('oidcAudience')">
              <el-input v-model="form.oidcAudience" />
            </el-form-item>
          </div>
          <div class="grid three">
            <el-form-item :label="label('oidcScope')">
              <el-input v-model="form.oidcScope" />
            </el-form-item>
            <el-form-item :label="label('oidcTokenEndpointURL')" prop="oidcTokenEndpointURL">
              <el-input v-model="form.oidcTokenEndpointURL" placeholder="https://issuer.example.com/oauth/token" />
            </el-form-item>
            <el-form-item :label="label('oidcProxyURL')">
              <el-input v-model="form.oidcProxyURL" placeholder="http://proxy.example.com:8080" />
            </el-form-item>
          </div>
          <div class="grid two">
            <el-form-item :label="label('oidcTrustedCaFile')">
              <div class="path-upload-row">
                <el-input v-model="form.oidcTrustedCaFile" placeholder="./ca.crt" />
                <el-button
                  :loading="uploadLoading.oidcTrustedCaFile"
                  @click="pickAndUploadFile('oidcTrustedCaFile')"
                >
                  {{ t('common.upload') }}
                </el-button>
              </div>
            </el-form-item>
            <el-form-item :label="label('oidcInsecureSkipVerify')">
              <el-switch v-model="form.oidcInsecureSkipVerify" />
            </el-form-item>
          </div>
        </template>
      </el-card>

      <el-card shadow="hover" class="settings-card">
        <template #header>{{ t('clientSettings.sections.adminUI') }}</template>
        <div class="grid two">
          <el-form-item :label="label('webServerAddr')" prop="webServerAddr">
            <el-input v-model="form.webServerAddr" />
          </el-form-item>
          <el-form-item :label="label('webServerPort')" prop="webServerPort">
            <el-input-number v-model="form.webServerPort" :min="1" :max="65535" controls-position="right" />
          </el-form-item>
        </div>
        <div class="grid two">
          <el-form-item :label="label('webServerUser')">
            <el-input v-model="form.webServerUser" />
          </el-form-item>
          <el-form-item :label="label('webServerPassword')">
            <el-input v-model="form.webServerPassword" show-password />
          </el-form-item>
        </div>
        <el-form-item :label="label('webServerPprofEnable')">
          <el-switch v-model="form.webServerPprofEnable" />
        </el-form-item>
      </el-card>

      <el-card shadow="hover" class="settings-card">
        <template #header>{{ t('clientSettings.sections.logging') }}</template>
        <div class="grid four">
          <el-form-item :label="label('logTo')">
            <el-input v-model="form.logTo" placeholder="./frpc.log" />
          </el-form-item>
          <el-form-item :label="label('logLevel')">
            <el-select v-model="form.logLevel">
              <el-option v-for="option in logLevelOptions" :key="option" :label="option" :value="option" />
            </el-select>
          </el-form-item>
          <el-form-item :label="label('logMaxDays')">
            <el-input-number v-model="form.logMaxDays" :min="0" controls-position="right" />
          </el-form-item>
          <el-form-item :label="label('logDisablePrintColor')">
            <el-switch v-model="form.logDisablePrintColor" />
          </el-form-item>
        </div>
      </el-card>

      <el-card shadow="hover" class="settings-card">
        <template #header>{{ t('clientSettings.sections.transport') }}</template>
        <div class="grid four">
          <el-form-item :label="label('transportProtocol')">
            <el-select v-model="form.transportProtocol">
              <el-option :label="t('clientSettings.options.tcp')" value="tcp" />
              <el-option :label="t('clientSettings.options.kcp')" value="kcp" />
              <el-option :label="t('clientSettings.options.quic')" value="quic" />
              <el-option :label="t('clientSettings.options.websocket')" value="websocket" />
              <el-option :label="t('clientSettings.options.wss')" value="wss" />
            </el-select>
          </el-form-item>
          <el-form-item :label="label('transportPoolCount')">
            <el-input-number v-model="form.transportPoolCount" :min="0" controls-position="right" />
          </el-form-item>
          <el-form-item :label="label('transportTcpMux')">
            <el-switch v-model="form.transportTcpMux" />
          </el-form-item>
          <el-form-item :label="label('transportTcpMuxKeepaliveInterval')">
            <el-input-number v-model="form.transportTcpMuxKeepaliveInterval" :min="-1" controls-position="right" />
          </el-form-item>
        </div>
        <div class="grid four">
          <el-form-item :label="label('transportDialServerTimeout')">
            <el-input-number v-model="form.transportDialServerTimeout" :min="0" controls-position="right" />
          </el-form-item>
          <el-form-item :label="label('transportDialServerKeepalive')">
            <el-input-number v-model="form.transportDialServerKeepalive" :min="-1" controls-position="right" />
          </el-form-item>
          <el-form-item :label="label('transportHeartbeatInterval')">
            <el-input-number v-model="form.transportHeartbeatInterval" controls-position="right" />
          </el-form-item>
          <el-form-item :label="label('transportHeartbeatTimeout')">
            <el-input-number v-model="form.transportHeartbeatTimeout" controls-position="right" />
          </el-form-item>
        </div>
        <div class="grid three">
          <el-form-item :label="label('transportConnectServerLocalIP')">
            <el-input v-model="form.transportConnectServerLocalIP" placeholder="0.0.0.0" />
          </el-form-item>
          <el-form-item :label="label('transportProxyURL')">
            <el-input v-model="form.transportProxyURL" placeholder="http://proxy.example.com:8080" />
          </el-form-item>
          <el-form-item :label="label('udpPacketSize')">
            <el-input-number v-model="form.udpPacketSize" :min="0" controls-position="right" />
          </el-form-item>
        </div>
        <div class="grid two">
          <el-form-item :label="label('transportTlsEnable')">
            <el-switch v-model="form.transportTlsEnable" />
          </el-form-item>
          <el-form-item :label="label('transportTlsDisableCustomTLSFirstByte')">
            <el-switch v-model="form.transportTlsDisableCustomTLSFirstByte" />
          </el-form-item>
        </div>
        <div class="grid four">
          <el-form-item :label="label('transportTlsCertFile')">
            <div class="path-upload-row">
              <el-input v-model="form.transportTlsCertFile" placeholder="./client.crt" />
              <el-button
                :loading="uploadLoading.transportTlsCertFile"
                @click="pickAndUploadFile('transportTlsCertFile')"
              >
                {{ t('common.upload') }}
              </el-button>
            </div>
          </el-form-item>
          <el-form-item :label="label('transportTlsKeyFile')">
            <div class="path-upload-row">
              <el-input v-model="form.transportTlsKeyFile" placeholder="./client.key" />
              <el-button
                :loading="uploadLoading.transportTlsKeyFile"
                @click="pickAndUploadFile('transportTlsKeyFile')"
              >
                {{ t('common.upload') }}
              </el-button>
            </div>
          </el-form-item>
          <el-form-item :label="label('transportTlsTrustedCaFile')">
            <div class="path-upload-row">
              <el-input v-model="form.transportTlsTrustedCaFile" placeholder="./ca.crt" />
              <el-button
                :loading="uploadLoading.transportTlsTrustedCaFile"
                @click="pickAndUploadFile('transportTlsTrustedCaFile')"
              >
                {{ t('common.upload') }}
              </el-button>
            </div>
          </el-form-item>
          <el-form-item :label="label('transportTlsServerName')">
            <el-input v-model="form.transportTlsServerName" placeholder="frps.example.com" />
          </el-form-item>
        </div>
        <div class="grid three">
          <el-form-item :label="label('transportQuicKeepalivePeriod')">
            <el-input-number v-model="form.transportQuicKeepalivePeriod" :min="0" controls-position="right" />
          </el-form-item>
          <el-form-item :label="label('transportQuicMaxIdleTimeout')">
            <el-input-number v-model="form.transportQuicMaxIdleTimeout" :min="0" controls-position="right" />
          </el-form-item>
          <el-form-item :label="label('transportQuicMaxIncomingStreams')">
            <el-input-number v-model="form.transportQuicMaxIncomingStreams" :min="0" controls-position="right" />
          </el-form-item>
        </div>
        <div class="field-help">{{ t('clientSettings.helpers.tls') }}</div>
        <div class="field-help">{{ t('clientSettings.helpers.restart') }}</div>
      </el-card>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormItemRule, FormRules } from 'element-plus'
import ActionButton from '@shared/components/ActionButton.vue'
import { getSettings, putSettings, uploadFile as uploadClientFile } from '../api/frpc'
import { useI18n } from '../i18n'
import type { ClientSettings } from '../types'

const { t } = useI18n()

const formRef = ref<FormInstance>()
const saving = ref(false)
const supportsExecTokenSource = ref(false)
const uploadLoading = reactive<Record<string, boolean>>({})
const logLevelOptions = ['trace', 'debug', 'info', 'warn', 'error']

const label = (key: string) => t(`clientSettings.labels.${key}`)

const form = reactive<ClientSettings>({
  configPath: '',
  user: '',
  clientID: '',
  serverAddr: '127.0.0.1',
  serverPort: 7000,
  natHoleStunServer: 'stun.easyvoip.com:3478',
  dnsServer: '',
  loginFailExit: true,
  authMethod: 'token',
  authAdditionalScopes: [],
  authToken: '',
  authTokenSourceType: 'inline',
  authTokenSourceFile: '',
  oidcClientID: '',
  oidcClientSecret: '',
  oidcAudience: '',
  oidcScope: '',
  oidcTokenEndpointURL: '',
  oidcTrustedCaFile: '',
  oidcInsecureSkipVerify: false,
  oidcProxyURL: '',
  webServerAddr: '127.0.0.1',
  webServerPort: 7400,
  webServerUser: 'admin',
  webServerPassword: 'admin',
  webServerPprofEnable: false,
  storePath: '',
  logTo: './frpc.log',
  logLevel: 'info',
  logMaxDays: 3,
  logDisablePrintColor: false,
  transportProtocol: 'tcp',
  transportPoolCount: 1,
  transportTcpMux: true,
  transportTcpMuxKeepaliveInterval: 30,
  transportDialServerTimeout: 10,
  transportDialServerKeepalive: 7200,
  transportConnectServerLocalIP: '',
  transportProxyURL: '',
  transportHeartbeatInterval: -1,
  transportHeartbeatTimeout: -1,
  transportTlsEnable: true,
  transportTlsDisableCustomTLSFirstByte: true,
  transportTlsCertFile: '',
  transportTlsKeyFile: '',
  transportTlsTrustedCaFile: '',
  transportTlsServerName: '',
  transportQuicKeepalivePeriod: 10,
  transportQuicMaxIdleTimeout: 30,
  transportQuicMaxIncomingStreams: 100000,
  udpPacketSize: 1500,
})

const isTokenAuth = computed(() => form.authMethod !== 'oidc')
const usesInlineToken = computed(
  () => isTokenAuth.value && form.authTokenSourceType !== 'file' && form.authTokenSourceType !== 'exec',
)
const usesFileTokenSource = computed(
  () => isTokenAuth.value && form.authTokenSourceType === 'file',
)

const rules = computed<FormRules>(() => ({
  serverAddr: [{ required: true, message: t('clientSettings.validation.serverAddrRequired'), trigger: 'blur' }],
  serverPort: [{ validator: validateServerPort, trigger: ['blur', 'change'] }],
  webServerAddr: [{ required: true, message: t('clientSettings.validation.adminAddrRequired'), trigger: 'blur' }],
  webServerPort: [{ validator: validateAdminPort, trigger: ['blur', 'change'] }],
  authToken: [{ validator: validateAuthToken, trigger: ['blur', 'change'] }],
  authTokenSourceFile: [{ validator: validateAuthTokenSourceFile, trigger: ['blur', 'change'] }],
  oidcClientID: [{ validator: validateOIDCClientID, trigger: ['blur', 'change'] }],
  oidcTokenEndpointURL: [{ validator: validateOIDCTokenEndpointURL, trigger: ['blur', 'change'] }],
}))

const fetchData = async () => {
  try {
    const payload = await getSettings()
    supportsExecTokenSource.value = payload.authTokenSourceType === 'exec'
    Object.assign(form, {
      ...payload,
      authMethod: payload.authMethod || 'token',
      authAdditionalScopes: payload.authAdditionalScopes ?? [],
      authTokenSourceType: payload.authTokenSourceType || 'inline',
    })
  } catch (error: any) {
    ElMessage.error(t('clientSettings.loadFailed', { message: error.message }))
  }
}

const handleSave = async () => {
  if (!(await validateForm())) {
    ElMessage.warning(t('clientSettings.validation.fixErrors'))
    return
  }

  saving.value = true
  try {
    await putSettings({ ...form })
    ElMessage.success(t('clientSettings.saveSuccess'))
  } catch (error: any) {
    ElMessage.error(t('clientSettings.saveFailed', { message: error.message }))
  } finally {
    saving.value = false
  }
}

onMounted(fetchData)

async function pickAndUploadFile(field: UploadableField) {
  const input = document.createElement('input')
  input.type = 'file'
  input.onchange = async () => {
    const file = input.files?.[0]
    if (!file) {
      return
    }

    uploadLoading[field] = true
    try {
      const targetPath = String(form[field] || '')
      const resp = await uploadClientFile(targetPath, file)
      form[field] = resp.savedPath as ClientSettings[UploadableField]
      ElMessage.success(t('clientSettings.uploadSuccess', { path: resp.savedPath }))
    } catch (error: any) {
      ElMessage.error(t('clientSettings.uploadFailed', { message: error.message }))
    } finally {
      uploadLoading[field] = false
    }
  }
  input.click()
}

async function validateForm() {
  if (!formRef.value) {
    return true
  }
  try {
    await formRef.value.validate()
    return true
  } catch {
    return false
  }
}

function validateServerPort(
  _rule: FormItemRule,
  value: number,
  callback: (error?: Error) => void,
) {
  if (typeof value !== 'number' || value < 1 || value > 65535) {
    callback(new Error(t('clientSettings.validation.serverPortRange')))
    return
  }
  callback()
}

function validateAdminPort(
  _rule: FormItemRule,
  value: number,
  callback: (error?: Error) => void,
) {
  if (typeof value !== 'number' || value < 1 || value > 65535) {
    callback(new Error(t('clientSettings.validation.adminPortRange')))
    return
  }
  callback()
}

function validateAuthToken(
  _rule: FormItemRule,
  value: string,
  callback: (error?: Error) => void,
) {
  if (usesInlineToken.value && !String(value || '').trim()) {
    callback(new Error(t('clientSettings.validation.authTokenRequired')))
    return
  }
  callback()
}

function validateAuthTokenSourceFile(
  _rule: FormItemRule,
  value: string,
  callback: (error?: Error) => void,
) {
  if (usesFileTokenSource.value && !String(value || '').trim()) {
    callback(new Error(t('clientSettings.validation.authTokenSourceFileRequired')))
    return
  }
  callback()
}

function validateOIDCClientID(
  _rule: FormItemRule,
  value: string,
  callback: (error?: Error) => void,
) {
  if (!isTokenAuth.value && !String(value || '').trim()) {
    callback(new Error(t('clientSettings.validation.oidcClientIDRequired')))
    return
  }
  callback()
}

function validateOIDCTokenEndpointURL(
  _rule: FormItemRule,
  value: string,
  callback: (error?: Error) => void,
) {
  if (!isTokenAuth.value && !String(value || '').trim()) {
    callback(new Error(t('clientSettings.validation.oidcTokenEndpointRequired')))
    return
  }
  callback()
}

type UploadableField =
  | 'authTokenSourceFile'
  | 'oidcTrustedCaFile'
  | 'transportTlsCertFile'
  | 'transportTlsKeyFile'
  | 'transportTlsTrustedCaFile'
</script>

<style scoped lang="scss">
.settings-page {
  height: 100%;
  overflow-y: auto;
  padding: $spacing-xl 40px;
  max-width: 1080px;
  margin: 0 auto;
  @include flex-column;
  gap: $spacing-lg;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: $spacing-lg;
}

.title-section {
  @include flex-column;
  gap: $spacing-sm;
}

.page-actions {
  display: flex;
  gap: $spacing-sm;
  flex-wrap: wrap;
}

.config-path {
  margin: 0;
  font-size: $font-size-sm;
  color: $color-text-muted;
}

.restart-alert {
  margin-bottom: 4px;
}

.settings-form {
  @include flex-column;
  gap: $spacing-md;
}

.settings-card {
  border-radius: 12px;
}

.grid {
  display: grid;
  gap: $spacing-md;
}

.grid.four {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.grid.three {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.grid.two {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.field-help {
  margin-top: 6px;
  font-size: 12px;
  line-height: 1.6;
  color: $color-text-muted;
}

.checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: $spacing-sm;
}

.path-upload-row {
  display: flex;
  gap: $spacing-sm;
  align-items: center;
}

.path-upload-row .el-input {
  flex: 1;
}

@media (max-width: 1200px) {
  .grid.four {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 900px) {
  .page-header {
    flex-direction: column;
  }

  .grid.four,
  .grid.three,
  .grid.two {
    grid-template-columns: 1fr;
  }
}

@include mobile {
  .settings-page {
    padding: $spacing-xl $spacing-lg;
  }
}
</style>
