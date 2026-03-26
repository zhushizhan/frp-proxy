<template>
  <template v-if="!readonly">
    <el-form-item :label="t('form.backendMode')">
      <el-radio-group v-model="backendMode">
        <el-radio value="direct">{{ t('form.direct') }}</el-radio>
        <el-radio value="plugin">{{ t('form.plugin') }}</el-radio>
      </el-radio-group>
    </el-form-item>
  </template>

  <template v-if="backendMode === 'direct'">
    <div class="field-row two-col">
      <ConfigField :label="t('form.localIP')" type="text" v-model="form.localIP" placeholder="127.0.0.1" :readonly="readonly" />
      <ConfigField :label="t('form.localPort')" type="number" v-model="form.localPort" :min="0" :max="65535" prop="localPort" :readonly="readonly" />
    </div>
  </template>

  <template v-else>
    <div class="field-row two-col">
      <ConfigField
        :label="t('form.pluginType')"
        type="select"
        v-model="form.pluginType"
        :options="pluginList.map((plugin) => ({ label: plugin, value: plugin }))"
        :readonly="readonly"
      />
      <div></div>
    </div>

    <template v-if="['http2https', 'https2http', 'https2https', 'http2http', 'tls2raw'].includes(form.pluginType)">
      <div class="field-row two-col">
        <ConfigField :label="t('form.localAddress')" type="text" v-model="form.pluginConfig.localAddr" :placeholder="t('form.localAddressPlaceholder')" :readonly="readonly" />
        <ConfigField
          v-if="['http2https', 'https2http', 'https2https', 'http2http'].includes(form.pluginType)"
          :label="t('form.hostHeaderRewrite')"
          type="text"
          v-model="form.pluginConfig.hostHeaderRewrite"
          :readonly="readonly"
        />
        <div v-else></div>
      </div>
    </template>

    <template v-if="['http2https', 'https2http', 'https2https', 'http2http'].includes(form.pluginType)">
      <ConfigField
        :label="t('form.requestHeaders')"
        type="kv"
        v-model="pluginRequestHeaders"
        :key-placeholder="t('common.key')"
        :value-placeholder="t('common.value')"
        :readonly="readonly"
      />
    </template>

    <template v-if="['https2http', 'https2https', 'tls2raw'].includes(form.pluginType)">
      <div class="field-row two-col">
        <ConfigField :label="t('form.certificatePath')" type="text" v-model="form.pluginConfig.crtPath" :placeholder="t('form.certPathPlaceholder')" :readonly="readonly" />
        <ConfigField :label="t('form.keyPath')" type="text" v-model="form.pluginConfig.keyPath" :placeholder="t('form.keyPathPlaceholder')" :readonly="readonly" />
      </div>
    </template>

    <template v-if="['https2http', 'https2https'].includes(form.pluginType)">
      <ConfigField :label="t('form.enableHttp2')" type="switch" v-model="form.pluginConfig.enableHTTP2" :readonly="readonly" />
    </template>

    <template v-if="form.pluginType === 'http_proxy'">
      <div class="field-row two-col">
        <ConfigField :label="t('form.httpUser')" type="text" v-model="form.pluginConfig.httpUser" :readonly="readonly" />
        <ConfigField :label="t('form.httpPassword')" type="password" v-model="form.pluginConfig.httpPassword" :readonly="readonly" />
      </div>
    </template>

    <template v-if="form.pluginType === 'socks5'">
      <div class="field-row two-col">
        <ConfigField :label="t('form.username')" type="text" v-model="form.pluginConfig.username" :readonly="readonly" />
        <ConfigField :label="t('form.password')" type="password" v-model="form.pluginConfig.password" :readonly="readonly" />
      </div>
    </template>

    <template v-if="form.pluginType === 'static_file'">
      <div class="field-row two-col">
        <ConfigField :label="t('form.localPath')" type="text" v-model="form.pluginConfig.localPath" :placeholder="t('form.localPathPlaceholder')" :readonly="readonly" />
        <ConfigField :label="t('form.stripPrefix')" type="text" v-model="form.pluginConfig.stripPrefix" :readonly="readonly" />
      </div>
      <div class="field-row two-col">
        <ConfigField :label="t('form.httpUser')" type="text" v-model="form.pluginConfig.httpUser" :readonly="readonly" />
        <ConfigField :label="t('form.httpPassword')" type="password" v-model="form.pluginConfig.httpPassword" :readonly="readonly" />
      </div>
    </template>

    <template v-if="form.pluginType === 'unix_domain_socket'">
      <ConfigField :label="t('form.unixSocketPath')" type="text" v-model="form.pluginConfig.unixPath" :placeholder="t('form.socketPathPlaceholder')" :readonly="readonly" />
    </template>
  </template>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import type { ProxyFormData } from '../../types'
import ConfigField from '../ConfigField.vue'
import { useI18n } from '../../i18n'

const pluginList = [
  'http2https', 'http_proxy', 'https2http', 'https2https', 'http2http',
  'socks5', 'static_file', 'unix_domain_socket', 'tls2raw', 'virtual_net',
]

const props = withDefaults(defineProps<{
  modelValue: ProxyFormData
  readonly?: boolean
}>(), { readonly: false })

const emit = defineEmits<{ 'update:modelValue': [value: ProxyFormData] }>()
const { t } = useI18n()

const form = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const backendMode = ref<'direct' | 'plugin'>(form.value.pluginType ? 'plugin' : 'direct')
const isHydrating = ref(false)

const pluginRequestHeaders = computed({
  get() {
    const set = form.value.pluginConfig?.requestHeaders?.set
    if (!set || typeof set !== 'object') return []
    return Object.entries(set).map(([key, value]) => ({ key, value: String(value) }))
  },
  set(val: Array<{ key: string; value: string }>) {
    if (!form.value.pluginConfig) form.value.pluginConfig = {}
    if (val.length === 0) {
      delete form.value.pluginConfig.requestHeaders
    } else {
      form.value.pluginConfig.requestHeaders = {
        set: Object.fromEntries(val.map((entry) => [entry.key, entry.value])),
      }
    }
  },
})

watch(() => form.value.pluginType, (newType, oldType) => {
  if (isHydrating.value) return
  if (!oldType || !newType || newType === oldType) return
  if (form.value.pluginConfig && Object.keys(form.value.pluginConfig).length > 0) {
    form.value.pluginConfig = {}
  }
})

watch(backendMode, (mode) => {
  if (mode === 'direct') {
    form.value.pluginType = ''
    form.value.pluginConfig = {}
  } else if (!form.value.pluginType) {
    form.value.pluginType = 'http2https'
  }
})

const hydrate = () => {
  isHydrating.value = true
  backendMode.value = form.value.pluginType ? 'plugin' : 'direct'
  nextTick(() => {
    isHydrating.value = false
  })
}

watch(() => props.modelValue, () => {
  hydrate()
})

onMounted(() => {
  hydrate()
})
</script>

<style scoped lang="scss">
@use '@/assets/css/form-layout';
</style>
