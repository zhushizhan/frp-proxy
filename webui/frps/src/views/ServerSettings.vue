<template>
  <div class="server-settings">
    <div class="page-header">
      <div class="title-block">
        <h1 class="page-title">{{ t('serverSettings.title') }}</h1>
        <p class="page-subtitle">{{ t('serverSettings.subtitle') }}</p>
        <p class="config-path" v-if="form.configPath">
          {{ t('common.configFile') }}: <code>{{ form.configPath }}</code>
        </p>
      </div>
      <div class="actions">
        <el-button :disabled="saving || restarting" @click="fetchData">
          {{ t('common.refresh') }}
        </el-button>
        <el-button type="primary" :loading="saving || restarting" :disabled="restarting" @click="handleSave">
          {{ restarting ? t('serverSettings.actions.waitingForRestart') : t('common.saveAndRestart') }}
        </el-button>
      </div>
    </div>

    <el-alert
      :title="t('serverSettings.restartAlert')"
      type="warning"
      :closable="false"
      show-icon
      class="restart-alert"
    />

    <el-alert
      v-if="restartState.active"
      :title="restartAlertTitle"
      :description="restartAlertDescription"
      :type="restartState.polling ? 'info' : 'success'"
      :closable="false"
      show-icon
      class="restart-alert"
    />

    <div v-if="restartState.targetUrl" class="restart-link-row">
      <a :href="restartState.targetUrl" class="restart-link">
        {{ restartState.targetUrl }}
      </a>
      <el-button text @click="openRestartTarget">
        {{ t('serverSettings.actions.openNewAddress') }}
      </el-button>
    </div>

    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
      status-icon
      scroll-to-error
      class="settings-form"
    >
      <div class="settings-shell">
        <div class="common-section">
      <el-card shadow="hover" class="settings-card">
        <template #header>
          <div class="card-header-row">
            <span class="card-header-title">{{ sectionLabel('coreListeners') }}</span>
            <div class="section-badges">
              <el-tag size="small" effect="plain">{{ t('serverSettings.badges.common') }}</el-tag>
              <el-tag size="small" type="danger" effect="plain">
                {{ t('serverSettings.badges.highImpact') }}
              </el-tag>
            </div>
          </div>
        </template>
        <el-alert
          :title="t('serverSettings.impact.listeners')"
          type="warning"
          :closable="false"
          show-icon
          class="inline-alert"
        />
        <div class="grid three">
          <el-form-item :label="fieldLabel('bindAddr')" prop="bindAddr">
            <el-input v-model="form.bindAddr" />
          </el-form-item>
          <el-form-item :label="fieldLabel('bindPort')" prop="bindPort">
            <el-input-number v-model="form.bindPort" :min="1" :max="65535" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('proxyBindAddr')" prop="proxyBindAddr">
            <el-input v-model="form.proxyBindAddr" />
          </el-form-item>
        </div>
        <div class="grid three">
          <el-form-item :label="fieldLabel('kcpBindPort')">
            <el-input-number v-model="form.kcpBindPort" :min="0" :max="65535" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('quicBindPort')">
            <el-input-number v-model="form.quicBindPort" :min="0" :max="65535" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('tcpmuxHTTPConnectPort')">
            <el-input-number
              v-model="form.tcpmuxHTTPConnectPort"
              :min="0"
              :max="65535"
              controls-position="right"
            />
          </el-form-item>
        </div>
        <div class="field-help">{{ t('serverSettings.helpers.listenerPorts') }}</div>
      </el-card>

      <el-card shadow="hover" class="settings-card">
        <template #header>
          <div class="card-header-row">
            <span class="card-header-title">{{ sectionLabel('virtualHost') }}</span>
            <div class="section-badges">
              <el-tag size="small" effect="plain">{{ t('serverSettings.badges.common') }}</el-tag>
            </div>
          </div>
        </template>
        <el-alert
          :title="t('serverSettings.alerts.vhost')"
          type="info"
          :closable="false"
          show-icon
          class="inline-alert"
        />
        <div class="grid three">
          <el-form-item :label="fieldLabel('vhostHTTPPort')">
            <el-input-number v-model="form.vhostHTTPPort" :min="0" :max="65535" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('vhostHTTPSPort')">
            <el-input-number v-model="form.vhostHTTPSPort" :min="0" :max="65535" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('subdomainHost')">
            <el-input v-model="form.subdomainHost" placeholder="frps.example.com" />
          </el-form-item>
        </div>
        <div class="grid two">
          <el-form-item :label="fieldLabel('vhostHTTPTimeout')">
            <el-input-number v-model="form.vhostHTTPTimeout" :min="0" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('tcpmuxPassthrough')">
            <el-switch v-model="form.tcpmuxPassthrough" />
          </el-form-item>
        </div>
      </el-card>
        </div>

        <div class="advanced-panel">
          <div class="advanced-panel-header">
            <div class="advanced-panel-copy">
              <h2 class="advanced-panel-title">{{ t('serverSettings.layout.advancedTitle') }}</h2>
              <p class="advanced-panel-subtitle">{{ t('serverSettings.layout.advancedSubtitle') }}</p>
            </div>
            <el-button text @click="toggleAdvancedSections">
              {{ advancedOpen ? t('serverSettings.layout.collapseAll') : t('serverSettings.layout.expandAll') }}
            </el-button>
          </div>

          <el-collapse-transition>
            <div v-show="advancedOpen" class="advanced-panel-body">

      <el-card shadow="hover" class="settings-card">
        <template #header>
          <div class="card-header-row">
            <span class="card-header-title">{{ sectionLabel('authentication') }}</span>
            <div class="section-badges">
              <el-tag size="small" effect="plain">{{ t('serverSettings.badges.advanced') }}</el-tag>
              <el-tag size="small" type="danger" effect="plain">
                {{ t('serverSettings.badges.highImpact') }}
              </el-tag>
            </div>
          </div>
        </template>
        <el-alert
          :title="t('serverSettings.impact.auth')"
          type="warning"
          :closable="false"
          show-icon
          class="inline-alert"
        />
        <div class="grid three">
          <el-form-item :label="fieldLabel('authMethod')">
            <el-select v-model="form.authMethod">
              <el-option
                v-for="option in authMethodOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="fieldLabel('tlsForce')">
            <el-switch v-model="form.tlsForce" />
          </el-form-item>
          <el-form-item :label="fieldLabel('authAdditionalScopes')">
            <el-checkbox-group v-model="form.authAdditionalScopes" class="checkbox-group">
              <el-checkbox label="HeartBeats">HeartBeats</el-checkbox>
              <el-checkbox label="NewWorkConns">NewWorkConns</el-checkbox>
            </el-checkbox-group>
          </el-form-item>
        </div>

        <template v-if="isTokenAuth">
          <div class="grid two">
            <el-form-item :label="fieldLabel('tokenSource')">
              <el-select v-model="form.authTokenSourceType">
                <el-option
                  v-for="option in tokenSourceOptions"
                  :key="option.value"
                  :label="option.label"
                  :value="option.value"
                  :disabled="option.value === 'exec' && !supportsExecTokenSource"
                />
              </el-select>
            </el-form-item>
            <el-form-item
              v-if="usesInlineToken"
              :label="fieldLabel('authToken')"
              prop="authToken"
            >
              <el-input v-model="form.authToken" show-password />
            </el-form-item>
            <el-form-item
              v-else-if="usesFileTokenSource"
              :label="fieldLabel('tokenSourceFile')"
              prop="authTokenSourceFile"
            >
              <div class="path-upload-row">
                <el-input v-model="form.authTokenSourceFile" placeholder="/etc/frp/token" />
                <el-button
                  :loading="uploadLoading.authTokenSourceFile"
                  @click="pickAndUploadFile('authTokenSourceFile')"
                >
                  {{ t('common.upload') }}
                </el-button>
              </div>
            </el-form-item>
            <el-form-item v-else :label="fieldLabel('tokenSource')">
              <el-input :model-value="t('serverSettings.alerts.execPreserved')" disabled />
            </el-form-item>
          </div>
          <el-alert
            v-if="usesExecTokenSource"
            :title="t('serverSettings.alerts.execPreserved')"
            type="warning"
            :closable="false"
            show-icon
            class="inline-alert"
          />
        </template>

        <template v-else>
          <div class="grid two">
            <el-form-item :label="fieldLabel('oidcIssuer')" prop="oidcIssuer">
              <el-input v-model="form.oidcIssuer" placeholder="https://issuer.example.com" />
            </el-form-item>
            <el-form-item :label="fieldLabel('oidcAudience')">
              <el-input v-model="form.oidcAudience" placeholder="frps-server" />
            </el-form-item>
          </div>
          <div class="grid two">
            <el-form-item :label="fieldLabel('oidcSkipExpiryCheck')">
              <el-switch v-model="form.oidcSkipExpiryCheck" />
            </el-form-item>
            <el-form-item :label="fieldLabel('oidcSkipIssuerCheck')">
              <el-switch v-model="form.oidcSkipIssuerCheck" />
            </el-form-item>
          </div>
        </template>

        <el-form-item :label="fieldLabel('allowPorts')">
          <el-input v-model="form.allowPorts" placeholder="1000-2000,3000,4000-5000" />
          <div class="field-help">{{ t('serverSettings.helpers.allowPorts') }}</div>
        </el-form-item>
      </el-card>

      <el-card shadow="hover" class="settings-card">
        <template #header>{{ sectionLabel('transport') }}</template>
        <div class="grid four">
          <el-form-item :label="fieldLabel('tcpMux')">
            <el-switch v-model="form.tcpMux" />
          </el-form-item>
          <el-form-item :label="fieldLabel('tcpMuxKeepaliveInterval')">
            <el-input-number v-model="form.tcpMuxKeepaliveInterval" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('maxPoolCount')">
            <el-input-number v-model="form.maxPoolCount" :min="0" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('heartbeatTimeout')">
            <el-input-number v-model="form.heartbeatTimeout" controls-position="right" />
          </el-form-item>
        </div>
        <div class="grid four">
          <el-form-item :label="fieldLabel('tcpKeepAlive')">
            <el-input-number v-model="form.tcpKeepAlive" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('quicKeepalivePeriod')">
            <el-input-number v-model="form.quicKeepalivePeriod" :min="0" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('quicMaxIdleTimeout')">
            <el-input-number v-model="form.quicMaxIdleTimeout" :min="0" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('quicMaxIncomingStreams')">
            <el-input-number v-model="form.quicMaxIncomingStreams" :min="0" controls-position="right" />
          </el-form-item>
        </div>
                <div class="grid three">
                  <el-form-item :label="fieldLabel('transportTLSCertFile')">
                    <div class="path-upload-row">
                      <el-input v-model="form.transportTLSCertFile" placeholder="./server.crt" />
                      <el-button
                        :loading="uploadLoading.transportTLSCertFile"
                        @click="pickAndUploadFile('transportTLSCertFile')"
                      >
                        {{ t('common.upload') }}
                      </el-button>
                    </div>
                  </el-form-item>
                  <el-form-item :label="fieldLabel('transportTLSKeyFile')">
                    <div class="path-upload-row">
                      <el-input v-model="form.transportTLSKeyFile" placeholder="./server.key" />
                      <el-button
                        :loading="uploadLoading.transportTLSKeyFile"
                        @click="pickAndUploadFile('transportTLSKeyFile')"
                      >
                        {{ t('common.upload') }}
                      </el-button>
                    </div>
                  </el-form-item>
                  <el-form-item :label="fieldLabel('transportTLSTrustedCaFile')">
                    <div class="path-upload-row">
                      <el-input v-model="form.transportTLSTrustedCaFile" placeholder="./ca.crt" />
                      <el-button
                        :loading="uploadLoading.transportTLSTrustedCaFile"
                        @click="pickAndUploadFile('transportTLSTrustedCaFile')"
                      >
                        {{ t('common.upload') }}
                      </el-button>
                    </div>
                  </el-form-item>
                </div>
        <div class="field-help">{{ t('serverSettings.helpers.transportTLS') }}</div>
      </el-card>

      <el-card shadow="hover" class="settings-card">
        <template #header>{{ sectionLabel('runtime') }}</template>
        <div class="grid four">
          <el-form-item :label="fieldLabel('maxPortsPerClient')">
            <el-input-number v-model="form.maxPortsPerClient" :min="0" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('userConnTimeout')">
            <el-input-number v-model="form.userConnTimeout" :min="0" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('udpPacketSize')">
            <el-input-number v-model="form.udpPacketSize" :min="0" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('natHoleAnalysisDataReserveHours')">
            <el-input-number
              v-model="form.natHoleAnalysisDataReserveHours"
              :min="0"
              controls-position="right"
            />
          </el-form-item>
        </div>
        <div class="grid two">
          <el-form-item :label="fieldLabel('detailedErrorsToClient')">
            <el-switch v-model="form.detailedErrorsToClient" />
          </el-form-item>
          <el-form-item :label="fieldLabel('enablePrometheus')">
            <el-switch v-model="form.enablePrometheus" />
          </el-form-item>
        </div>
      </el-card>

      <el-card shadow="hover" class="settings-card">
        <template #header>
          <div class="card-header-row">
            <span class="card-header-title">{{ sectionLabel('dashboard') }}</span>
            <div class="section-badges">
              <el-tag size="small" effect="plain">{{ t('serverSettings.badges.advanced') }}</el-tag>
              <el-tag size="small" type="danger" effect="plain">
                {{ t('serverSettings.badges.highImpact') }}
              </el-tag>
            </div>
          </div>
        </template>
        <el-alert
          :title="t('serverSettings.impact.dashboard')"
          type="warning"
          :closable="false"
          show-icon
          class="inline-alert"
        />
        <div class="grid two">
          <el-form-item :label="fieldLabel('dashboardAddr')" prop="dashboardAddr">
            <el-input v-model="form.dashboardAddr" />
          </el-form-item>
          <el-form-item :label="fieldLabel('dashboardPort')" prop="dashboardPort">
            <el-input-number v-model="form.dashboardPort" :min="1" :max="65535" controls-position="right" />
          </el-form-item>
        </div>
        <div class="grid two">
          <el-form-item :label="fieldLabel('dashboardUser')">
            <el-input v-model="form.dashboardUser" />
          </el-form-item>
          <el-form-item :label="fieldLabel('dashboardPassword')">
            <el-input v-model="form.dashboardPassword" show-password />
          </el-form-item>
        </div>
        <div class="grid two">
          <el-form-item :label="fieldLabel('dashboardAssetsDir')">
            <el-input v-model="form.dashboardAssetsDir" placeholder="./static" />
          </el-form-item>
          <el-form-item :label="fieldLabel('dashboardPprofEnable')">
            <el-switch v-model="form.dashboardPprofEnable" />
          </el-form-item>
        </div>
                <div class="grid three">
                  <el-form-item :label="fieldLabel('dashboardTLSCertFile')">
                    <div class="path-upload-row">
                      <el-input v-model="form.dashboardTLSCertFile" placeholder="./server.crt" />
                      <el-button
                        :loading="uploadLoading.dashboardTLSCertFile"
                        @click="pickAndUploadFile('dashboardTLSCertFile')"
                      >
                        {{ t('common.upload') }}
                      </el-button>
                    </div>
                  </el-form-item>
                  <el-form-item :label="fieldLabel('dashboardTLSKeyFile')">
                    <div class="path-upload-row">
                      <el-input v-model="form.dashboardTLSKeyFile" placeholder="./server.key" />
                      <el-button
                        :loading="uploadLoading.dashboardTLSKeyFile"
                        @click="pickAndUploadFile('dashboardTLSKeyFile')"
                      >
                        {{ t('common.upload') }}
                      </el-button>
                    </div>
                  </el-form-item>
                  <el-form-item :label="fieldLabel('dashboardTLSTrustedCaFile')">
                    <div class="path-upload-row">
                      <el-input v-model="form.dashboardTLSTrustedCaFile" placeholder="./ca.crt" />
                      <el-button
                        :loading="uploadLoading.dashboardTLSTrustedCaFile"
                        @click="pickAndUploadFile('dashboardTLSTrustedCaFile')"
                      >
                        {{ t('common.upload') }}
                      </el-button>
                    </div>
                  </el-form-item>
                </div>
        <div class="field-help">{{ t('serverSettings.helpers.dashboardTLS') }}</div>
      </el-card>

      <el-card shadow="hover" class="settings-card">
        <template #header>{{ sectionLabel('logging') }}</template>
        <div class="grid four">
          <el-form-item :label="fieldLabel('logTo')">
            <el-input v-model="form.logTo" placeholder="console or ./frps.log" />
          </el-form-item>
          <el-form-item :label="fieldLabel('logLevel')">
            <el-select v-model="form.logLevel">
              <el-option v-for="option in logLevelOptions" :key="option" :label="option" :value="option" />
            </el-select>
          </el-form-item>
          <el-form-item :label="fieldLabel('logMaxDays')">
            <el-input-number v-model="form.logMaxDays" :min="0" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('logDisablePrintColor')">
            <el-switch v-model="form.logDisablePrintColor" />
          </el-form-item>
        </div>
      </el-card>

      <el-card shadow="hover" class="settings-card">
        <template #header>
          <div class="section-header">
            <span>{{ sectionLabel('plugins') }}</span>
            <el-button size="small" @click="addHTTPPlugin">
              {{ t('serverSettings.actions.addPlugin') }}
            </el-button>
          </div>
        </template>
        <div class="field-help plugin-help">{{ t('serverSettings.helpers.plugins') }}</div>
        <div class="field-help plugin-help">{{ t('serverSettings.helpers.pluginOps') }}</div>
        <div class="preset-panel">
          <div class="preset-panel-title">{{ t('serverSettings.presetTitle') }}</div>
          <div class="preset-grid">
            <button
              v-for="preset in pluginPresets"
              :key="preset.key"
              type="button"
              class="preset-card"
              @click="addHTTPPluginPreset(preset.key)"
            >
              <span class="preset-name">{{ preset.name }}</span>
              <span class="preset-description">{{ preset.description }}</span>
            </button>
          </div>
        </div>

        <div v-if="form.httpPlugins.length === 0" class="plugin-empty">
          <el-empty :description="t('serverSettings.emptyPlugins')" />
        </div>

        <div v-else class="plugin-list">
          <div
            v-for="(plugin, index) in form.httpPlugins"
            :key="`plugin-${index}`"
            class="plugin-card"
          >
            <div class="plugin-card-header">
              <span class="plugin-card-title">
                {{ plugin.name || `${t('serverSettings.fields.pluginName')} #${index + 1}` }}
              </span>
              <el-button size="small" type="danger" plain @click="removeHTTPPlugin(index)">
                {{ t('serverSettings.actions.removePlugin') }}
              </el-button>
            </div>
            <div class="grid three">
              <el-form-item
                :label="fieldLabel('pluginName')"
                :prop="`httpPlugins.${index}.name`"
                :rules="pluginNameRules(index)"
              >
                <el-input v-model="plugin.name" placeholder="user-manager" />
              </el-form-item>
              <el-form-item
                :label="fieldLabel('pluginAddr')"
                :prop="`httpPlugins.${index}.addr`"
                :rules="pluginAddrRules"
              >
                <el-input v-model="plugin.addr" placeholder="127.0.0.1:9000" />
              </el-form-item>
              <el-form-item
                :label="fieldLabel('pluginPath')"
                :prop="`httpPlugins.${index}.path`"
                :rules="pluginPathRules"
              >
                <el-input v-model="plugin.path" placeholder="/handler" />
              </el-form-item>
            </div>
            <div class="grid two">
              <el-form-item
                :label="fieldLabel('pluginOps')"
                :prop="`httpPlugins.${index}.ops`"
                :rules="pluginOpsRules"
              >
                <el-checkbox-group v-model="plugin.ops" class="checkbox-group">
                  <el-checkbox
                    v-for="option in pluginOpOptions"
                    :key="option"
                    :label="option"
                  >
                    {{ option }}
                  </el-checkbox>
                </el-checkbox-group>
              </el-form-item>
              <el-form-item :label="fieldLabel('pluginTlsVerify')">
                <el-switch v-model="plugin.tlsVerify" />
              </el-form-item>
            </div>
          </div>
        </div>
      </el-card>

      <el-card shadow="hover" class="settings-card">
        <template #header>{{ sectionLabel('ssh') }}</template>
        <div class="grid two">
          <el-form-item :label="fieldLabel('sshTunnelGatewayPort')">
            <el-input-number v-model="form.sshTunnelGatewayPort" :min="0" :max="65535" controls-position="right" />
          </el-form-item>
          <el-form-item :label="fieldLabel('sshAutoGenKeyPath')">
            <el-input v-model="form.sshAutoGenKeyPath" placeholder="./.autogen_ssh_key" />
          </el-form-item>
        </div>
                <div class="grid two">
                  <el-form-item :label="fieldLabel('sshPrivateKeyFile')">
                    <div class="path-upload-row">
                      <el-input v-model="form.sshPrivateKeyFile" placeholder="/path/to/id_rsa" />
                      <el-button
                        :loading="uploadLoading.sshPrivateKeyFile"
                        @click="pickAndUploadFile('sshPrivateKeyFile')"
                      >
                        {{ t('common.upload') }}
                      </el-button>
                    </div>
                  </el-form-item>
                  <el-form-item :label="fieldLabel('sshAuthorizedKeysFile')">
                    <div class="path-upload-row">
                      <el-input v-model="form.sshAuthorizedKeysFile" placeholder="/path/to/authorized_keys" />
                      <el-button
                        :loading="uploadLoading.sshAuthorizedKeysFile"
                        @click="pickAndUploadFile('sshAuthorizedKeysFile')"
                      >
                        {{ t('common.upload') }}
                      </el-button>
                    </div>
                  </el-form-item>
                </div>
        <div class="field-help">{{ t('serverSettings.helpers.ssh') }}</div>
      </el-card>

      <el-card shadow="hover" class="settings-card">
        <template #header>{{ sectionLabel('misc') }}</template>
        <el-form-item :label="fieldLabel('custom404Page')">
          <el-input v-model="form.custom404Page" placeholder="./404.html" />
        </el-form-item>
      </el-card>
            </div>
          </el-collapse-transition>
        </div>
      </div>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormItemRule, FormRules } from 'element-plus'
import { getServerSettings, updateServerSettings, uploadFile as uploadServerFile } from '../api/server'
import { useI18n } from '../i18n'
import type { HTTPPluginSettings, ServerSettings } from '../types/server'

const { t } = useI18n()

type PluginPresetKey =
  | 'loginGuard'
  | 'proxyLifecycle'
  | 'accessGate'
  | 'fullAudit'

const formRef = ref<FormInstance>()
const saving = ref(false)
const restarting = ref(false)
const supportsExecTokenSource = ref(false)
const advancedOpen = ref(false)
const uploadLoading = reactive<Record<string, boolean>>({})
const logLevelOptions = ['trace', 'debug', 'info', 'warn', 'error']
const pluginOpOptions = [
  'Login',
  'NewProxy',
  'CloseProxy',
  'Ping',
  'NewWorkConn',
  'NewUserConn',
]
const restartState = reactive({
  active: false,
  polling: false,
  targetUrl: '',
  changedOrigin: false,
})

const authMethodOptions = computed(() => [
  { label: t('serverSettings.options.token'), value: 'token' },
  { label: t('serverSettings.options.oidc'), value: 'oidc' },
])

const tokenSourceOptions = computed(() => [
  { label: t('serverSettings.options.inlineToken'), value: 'inline' },
  { label: t('serverSettings.options.file'), value: 'file' },
  { label: t('serverSettings.helpers.execOption'), value: 'exec' },
])

const pluginPresets = computed(() => [
  {
    key: 'loginGuard' as PluginPresetKey,
    name: t('serverSettings.presets.loginGuard.name'),
    description: t('serverSettings.presets.loginGuard.description'),
  },
  {
    key: 'proxyLifecycle' as PluginPresetKey,
    name: t('serverSettings.presets.proxyLifecycle.name'),
    description: t('serverSettings.presets.proxyLifecycle.description'),
  },
  {
    key: 'accessGate' as PluginPresetKey,
    name: t('serverSettings.presets.accessGate.name'),
    description: t('serverSettings.presets.accessGate.description'),
  },
  {
    key: 'fullAudit' as PluginPresetKey,
    name: t('serverSettings.presets.fullAudit.name'),
    description: t('serverSettings.presets.fullAudit.description'),
  },
])

const sectionLabel = (key: string) => t(`serverSettings.sections.${key}`)
const fieldLabel = (key: string) => t(`serverSettings.fields.${key}`)

const restartAlertTitle = computed(() =>
  restartState.changedOrigin
    ? t('serverSettings.restartState.redirectTitle')
    : t('serverSettings.restartState.pollingTitle'),
)

const restartAlertDescription = computed(() => {
  if (restartState.changedOrigin) {
    return t('serverSettings.restartState.redirectDescription')
  }
  return t('serverSettings.restartState.pollingDescription')
})

const form = reactive<ServerSettings>({
  configPath: '',
  autoRestart: true,
  bindAddr: '0.0.0.0',
  bindPort: 7000,
  proxyBindAddr: '0.0.0.0',
  kcpBindPort: 0,
  quicBindPort: 0,
  vhostHTTPPort: 0,
  vhostHTTPSPort: 0,
  vhostHTTPTimeout: 60,
  tcpmuxHTTPConnectPort: 0,
  tcpmuxPassthrough: false,
  subdomainHost: '',
  authMethod: 'token',
  authAdditionalScopes: [],
  authToken: '',
  authTokenSourceType: 'inline',
  authTokenSourceFile: '',
  oidcIssuer: '',
  oidcAudience: '',
  oidcSkipExpiryCheck: false,
  oidcSkipIssuerCheck: false,
  tlsForce: false,
  transportTLSCertFile: '',
  transportTLSKeyFile: '',
  transportTLSTrustedCaFile: '',
  tcpMux: true,
  tcpMuxKeepaliveInterval: 30,
  maxPoolCount: 5,
  heartbeatTimeout: -1,
  tcpKeepAlive: 7200,
  quicKeepalivePeriod: 10,
  quicMaxIdleTimeout: 30,
  quicMaxIncomingStreams: 100000,
  maxPortsPerClient: 0,
  userConnTimeout: 10,
  udpPacketSize: 1500,
  natHoleAnalysisDataReserveHours: 168,
  detailedErrorsToClient: true,
  allowPorts: '',
  enablePrometheus: false,
  dashboardAddr: '127.0.0.1',
  dashboardPort: 7500,
  dashboardUser: 'admin',
  dashboardPassword: 'admin',
  dashboardAssetsDir: '',
  dashboardPprofEnable: false,
  dashboardTLSCertFile: '',
  dashboardTLSKeyFile: '',
  dashboardTLSTrustedCaFile: '',
  logTo: 'console',
  logLevel: 'info',
  logMaxDays: 3,
  logDisablePrintColor: false,
  sshTunnelGatewayPort: 0,
  sshPrivateKeyFile: '',
  sshAutoGenKeyPath: './.autogen_ssh_key',
  sshAuthorizedKeysFile: '',
  httpPlugins: [],
  custom404Page: '',
})

const isTokenAuth = computed(() => form.authMethod !== 'oidc')
const usesInlineToken = computed(
  () => isTokenAuth.value && form.authTokenSourceType === 'inline',
)
const usesFileTokenSource = computed(
  () => isTokenAuth.value && form.authTokenSourceType === 'file',
)
const usesExecTokenSource = computed(
  () => isTokenAuth.value && form.authTokenSourceType === 'exec',
)

const toggleAdvancedSections = () => {
  advancedOpen.value = !advancedOpen.value
}

const rules = computed<FormRules>(() => ({
  bindAddr: [
    { required: true, message: t('serverSettings.validation.bindAddrRequired'), trigger: 'blur' },
  ],
  proxyBindAddr: [
    { required: true, message: t('serverSettings.validation.proxyBindAddrRequired'), trigger: 'blur' },
  ],
  bindPort: [
    { validator: validateBindPort, trigger: ['blur', 'change'] },
  ],
  dashboardAddr: [
    { required: true, message: t('serverSettings.validation.dashboardAddrRequired'), trigger: 'blur' },
  ],
  dashboardPort: [
    { validator: validateDashboardPort, trigger: ['blur', 'change'] },
  ],
  authToken: [
    { validator: validateAuthToken, trigger: ['blur', 'change'] },
  ],
  authTokenSourceFile: [
    { validator: validateTokenSourceFile, trigger: ['blur', 'change'] },
  ],
  oidcIssuer: [
    { validator: validateOIDCIssuer, trigger: ['blur', 'change'] },
  ],
}))

const pluginAddrRules: FormItemRule[] = [
  {
    required: true,
    message: t('serverSettings.validation.pluginAddrRequired'),
    trigger: 'blur',
  },
]

const pluginPathRules: FormItemRule[] = [
  {
    validator: (_rule, value: string, callback) => {
      const normalized = String(value || '').trim()
      if (!normalized) {
        callback(new Error(t('serverSettings.validation.pluginPathRequired')))
        return
      }
      if (!normalized.startsWith('/')) {
        callback(new Error(t('serverSettings.validation.pluginPathFormat')))
        return
      }
      callback()
    },
    trigger: ['blur', 'change'],
  },
]

const pluginOpsRules: FormItemRule[] = [
  {
    validator: (_rule, value: string[], callback) => {
      if (!Array.isArray(value) || value.length === 0) {
        callback(new Error(t('serverSettings.validation.pluginOpsRequired')))
        return
      }
      callback()
    },
    trigger: 'change',
  },
]

const applySettings = (payload: ServerSettings) => {
  supportsExecTokenSource.value = payload.authTokenSourceType === 'exec'
  Object.assign(form, {
    ...payload,
    authMethod: payload.authMethod || 'token',
    authAdditionalScopes: payload.authAdditionalScopes ?? [],
    authTokenSourceType: payload.authTokenSourceType || 'inline',
    httpPlugins: (payload.httpPlugins ?? []).map((plugin) => ({
      ...plugin,
      ops: [...(plugin.ops ?? [])],
    })),
  })
  nextTick(() => {
    formRef.value?.clearValidate()
  })
}

const addHTTPPlugin = () => {
  form.httpPlugins.push(createEmptyHTTPPlugin())
  advancedOpen.value = true
  nextTick(() => {
    formRef.value?.clearValidate()
  })
}

const removeHTTPPlugin = (index: number) => {
  form.httpPlugins.splice(index, 1)
  nextTick(() => {
    formRef.value?.clearValidate()
  })
}

const addHTTPPluginPreset = (key: PluginPresetKey) => {
  form.httpPlugins.push(createHTTPPluginPreset(key))
  advancedOpen.value = true
  nextTick(() => {
    formRef.value?.clearValidate()
  })
}

const fetchData = async () => {
  try {
    const payload = await getServerSettings()
    applySettings(payload)
    if (restartState.active) {
      ElMessage.success(t('serverSettings.restartState.recovered'))
      resetRestartState()
    }
  } catch (error: any) {
    ElMessage.error(t('serverSettings.loadFailed', { message: error.message }))
  }
}

const handleSave = async () => {
  if (!(await validateForm())) {
    ElMessage.warning(t('serverSettings.validation.fixErrors'))
    return
  }
  saving.value = true
  try {
    const restartPlan = buildRestartPlan()
    await updateServerSettings({ ...form })
    beginRestartState(restartPlan)
    ElMessage.success(
      restartPlan.changedOrigin
        ? t('serverSettings.restartState.redirectQueued')
        : t('serverSettings.saveSuccess'),
    )
    void waitForServerRestart(restartPlan)
  } catch (error: any) {
    resetRestartState()
    ElMessage.error(t('serverSettings.saveFailed', { message: error.message }))
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
      const resp = await uploadServerFile(targetPath, file)
      form[field] = resp.savedPath as ServerSettings[UploadableField]
      ElMessage.success(t('serverSettings.uploadSuccess', { path: resp.savedPath }))
    } catch (error: any) {
      ElMessage.error(t('serverSettings.uploadFailed', { message: error.message }))
    } finally {
      uploadLoading[field] = false
    }
  }
  input.click()
}

function buildRestartPlan() {
  const current = new URL(window.location.href)
  const hostname = normalizeDashboardHost(form.dashboardAddr, current.hostname)
  const protocol = inferDashboardProtocol()
  const targetOrigin = `${protocol}//${hostname}:${form.dashboardPort}`
  const currentOrigin = window.location.origin
  return {
    targetUrl: `${targetOrigin}/webui/#/settings`,
    changedOrigin: targetOrigin !== currentOrigin,
  }
}

function normalizeDashboardHost(addr: string, fallbackHost: string) {
  const normalized = String(addr || '').trim()
  if (!normalized || normalized === '0.0.0.0' || normalized === '::' || normalized === '[::]') {
    return fallbackHost
  }
  return normalized
}

function inferDashboardProtocol() {
  if (
    String(form.dashboardTLSCertFile || '').trim() ||
    String(form.dashboardTLSKeyFile || '').trim() ||
    String(form.dashboardTLSTrustedCaFile || '').trim()
  ) {
    return 'https:'
  }
  return window.location.protocol
}

function beginRestartState(plan: { targetUrl: string; changedOrigin: boolean }) {
  restartState.active = true
  restartState.polling = true
  restartState.targetUrl = plan.targetUrl
  restartState.changedOrigin = plan.changedOrigin
  restarting.value = true
}

function resetRestartState() {
  restartState.active = false
  restartState.polling = false
  restartState.targetUrl = ''
  restartState.changedOrigin = false
  restarting.value = false
}

async function waitForServerRestart(plan: { targetUrl: string; changedOrigin: boolean }) {
  if (plan.changedOrigin) {
    restartState.polling = false
    restarting.value = false
    window.setTimeout(() => {
      window.location.assign(plan.targetUrl)
    }, 1500)
    return
  }

  for (let attempt = 0; attempt < 20; attempt += 1) {
    await sleep(attempt < 3 ? 1000 : 1500)
    try {
      const payload = await getServerSettings()
      applySettings(payload)
      ElMessage.success(t('serverSettings.restartState.recovered'))
      resetRestartState()
      return
    } catch {
      // Keep polling while frps is restarting.
    }
  }

  restartState.polling = false
  restarting.value = false
  ElMessage.warning(t('serverSettings.restartState.refreshManual'))
}

function openRestartTarget() {
  if (!restartState.targetUrl) {
    return
  }
  window.location.assign(restartState.targetUrl)
}

function sleep(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms))
}

function createEmptyHTTPPlugin(): HTTPPluginSettings {
  return {
    name: '',
    addr: '',
    path: '',
    ops: [],
    tlsVerify: false,
  }
}

function createHTTPPluginPreset(key: PluginPresetKey): HTTPPluginSettings {
  const presetMap: Record<PluginPresetKey, Omit<HTTPPluginSettings, 'name'> & { baseName: string }> = {
    loginGuard: {
      baseName: 'login-guard',
      addr: '127.0.0.1:9000',
      path: '/handler',
      ops: ['Login'],
      tlsVerify: false,
    },
    proxyLifecycle: {
      baseName: 'proxy-lifecycle',
      addr: '127.0.0.1:9001',
      path: '/handler',
      ops: ['NewProxy', 'CloseProxy'],
      tlsVerify: false,
    },
    accessGate: {
      baseName: 'access-gate',
      addr: '127.0.0.1:9002',
      path: '/handler',
      ops: ['Login', 'Ping', 'NewWorkConn', 'NewUserConn'],
      tlsVerify: false,
    },
    fullAudit: {
      baseName: 'full-audit',
      addr: '127.0.0.1:9003',
      path: '/handler',
      ops: [...pluginOpOptions],
      tlsVerify: false,
    },
  }
  const preset = presetMap[key]
  return {
    name: createUniquePluginName(preset.baseName),
    addr: preset.addr,
    path: preset.path,
    ops: [...preset.ops],
    tlsVerify: preset.tlsVerify,
  }
}

function createUniquePluginName(baseName: string): string {
  const normalized = baseName.trim() || 'plugin'
  const existing = new Set(
    form.httpPlugins.map((plugin) => plugin.name.trim()).filter(Boolean),
  )
  if (!existing.has(normalized)) {
    return normalized
  }
  let index = 2
  while (existing.has(`${normalized}-${index}`)) {
    index += 1
  }
  return `${normalized}-${index}`
}

function pluginNameRules(index: number): FormItemRule[] {
  return [
    {
      validator: (_rule, value: string, callback) => {
        const normalized = String(value || '').trim()
        if (!normalized) {
          callback(new Error(t('serverSettings.validation.pluginNameRequired')))
          return
        }
        const duplicate = form.httpPlugins.some(
          (plugin, pluginIndex) =>
            pluginIndex !== index && plugin.name.trim() === normalized,
        )
        if (duplicate) {
          callback(new Error(t('serverSettings.validation.pluginNameUnique')))
          return
        }
        callback()
      },
      trigger: ['blur', 'change'],
    },
  ]
}

async function validateForm(): Promise<boolean> {
  if (!formRef.value) {
    return true
  }
  try {
    await formRef.value.validate()
    return true
  } catch {
    advancedOpen.value = true
    await nextTick()
    return false
  }
}

function validateBindPort(_rule: FormItemRule, value: number, callback: (error?: Error) => void) {
  if (typeof value !== 'number' || value < 1 || value > 65535) {
    callback(new Error(t('serverSettings.validation.bindPortRange')))
    return
  }
  callback()
}

function validateDashboardPort(
  _rule: FormItemRule,
  value: number,
  callback: (error?: Error) => void,
) {
  if (typeof value !== 'number' || value < 1 || value > 65535) {
    callback(new Error(t('serverSettings.validation.dashboardPortRange')))
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
    callback(new Error(t('serverSettings.validation.authTokenRequired')))
    return
  }
  callback()
}

function validateTokenSourceFile(
  _rule: FormItemRule,
  value: string,
  callback: (error?: Error) => void,
) {
  if (usesFileTokenSource.value && !String(value || '').trim()) {
    callback(new Error(t('serverSettings.validation.tokenSourceFileRequired')))
    return
  }
  callback()
}

function validateOIDCIssuer(
  _rule: FormItemRule,
  value: string,
  callback: (error?: Error) => void,
) {
  if (!isTokenAuth.value && !String(value || '').trim()) {
    callback(new Error(t('serverSettings.validation.oidcIssuerRequired')))
    return
  }
  callback()
}

type UploadableField =
  | 'authTokenSourceFile'
  | 'transportTLSCertFile'
  | 'transportTLSKeyFile'
  | 'transportTLSTrustedCaFile'
  | 'dashboardTLSCertFile'
  | 'dashboardTLSKeyFile'
  | 'dashboardTLSTrustedCaFile'
  | 'sshPrivateKeyFile'
  | 'sshAuthorizedKeysFile'
</script>

<style scoped>
.server-settings {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  align-items: flex-start;
}

.title-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.restart-alert {
  margin-bottom: 4px;
}

.restart-link-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  margin-top: -4px;
}

.restart-link {
  font-size: 13px;
  color: var(--el-color-primary);
  text-decoration: none;
  word-break: break-all;
}

.restart-link:hover {
  text-decoration: underline;
}

.settings-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.settings-shell {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.common-section {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 16px;
  align-items: start;
}

.settings-card {
  border-radius: 12px;
}

.card-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.card-header-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.grid {
  display: grid;
  gap: 16px;
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
  color: #909399;
}

.inline-alert {
  margin-bottom: 16px;
}

.section-badges {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.config-path code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.advanced-panel {
  border: 1px solid var(--el-border-color-light);
  border-radius: 16px;
  padding: 16px;
  background: linear-gradient(
    180deg,
    var(--el-bg-color) 0%,
    var(--el-fill-color-extra-light) 100%
  );
}

.advanced-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 12px;
}

.advanced-panel-copy {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.advanced-panel-title {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}

.advanced-panel-subtitle {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--el-text-color-secondary);
}

.advanced-panel-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.plugin-help {
  margin-bottom: 10px;
}

.plugin-empty {
  padding: 8px 0;
}

.plugin-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.preset-panel {
  margin-bottom: 18px;
}

.preset-panel-title {
  margin-bottom: 10px;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
}

.preset-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
}

.preset-card {
  text-align: left;
  border: 1px solid var(--el-border-color-light);
  background: var(--el-bg-color);
  border-radius: 12px;
  padding: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.preset-card:hover {
  border-color: var(--el-color-primary-light-5);
  transform: translateY(-1px);
}

.preset-name {
  display: block;
  margin-bottom: 6px;
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.preset-description {
  display: block;
  font-size: 12px;
  line-height: 1.5;
  color: var(--el-text-color-secondary);
}

.plugin-card {
  border: 1px solid var(--el-border-color-light);
  border-radius: 12px;
  padding: 16px;
  background: var(--el-bg-color-page);
}

.plugin-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.plugin-card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.path-upload-row {
  display: flex;
  gap: 10px;
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
  .page-header,
  .advanced-panel-header {
    flex-direction: column;
  }

  .grid.four,
  .grid.three,
  .grid.two {
    grid-template-columns: 1fr;
  }
}
</style>
