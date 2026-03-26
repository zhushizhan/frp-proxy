<template>
  <div class="proxy-form-layout">
    <ConfigSection :readonly="readonly" :title="t('form.coreSetup')">
      <ProxyBaseSection v-model="form" :readonly="readonly" :editing="editing" :lock-type="lockType" />
      <ProxyRemoteSection
        v-if="['tcp', 'udp', 'http', 'https', 'tcpmux'].includes(form.type)"
        v-model="form"
        :readonly="readonly"
      />
      <ProxyBackendSection v-model="form" :readonly="readonly" />
    </ConfigSection>

    <ConfigSection v-if="pairingNote" :title="t('form.accessPattern')" :readonly="readonly">
      <p class="helper-copy">{{ pairingNote }}</p>
    </ConfigSection>

    <ConfigSection v-if="httpNote" :title="t('form.domainRouting')" :readonly="readonly">
      <p class="helper-copy">{{ httpNote }}</p>
    </ConfigSection>

    <ConfigSection
      v-if="showTrafficAndAccess"
      :title="t('form.trafficAndAccess')"
      :readonly="readonly"
      collapsible
      :has-value="trafficAccessConfigured"
    >
      <ProxyAuthSection
        v-if="['http', 'tcpmux', 'stcp', 'sudp', 'xtcp'].includes(form.type)"
        v-model="form"
        :readonly="readonly"
      />
      <ProxyHttpSection v-if="form.type === 'http'" v-model="form" :readonly="readonly" />
    </ConfigSection>

    <ConfigSection
      :title="t('form.runtimeAndHealth')"
      :readonly="readonly"
      collapsible
      :has-value="runtimeConfigured"
    >
      <ProxyTransportSection v-model="form" :readonly="readonly" />
      <ProxyHealthSection v-model="form" :readonly="readonly" />
      <ProxyLoadBalanceSection v-model="form" :readonly="readonly" />
      <ProxyNatSection v-if="form.type === 'xtcp'" v-model="form" :readonly="readonly" />
    </ConfigSection>

    <ConfigSection
      :title="t('form.metadataAndLabels')"
      :readonly="readonly"
      collapsible
      :has-value="metadataConfigured"
    >
      <ProxyMetadataSection v-model="form" :readonly="readonly" />
    </ConfigSection>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ProxyFormData } from '../../types'
import ConfigSection from '../ConfigSection.vue'
import ProxyAuthSection from './ProxyAuthSection.vue'
import ProxyBackendSection from './ProxyBackendSection.vue'
import ProxyBaseSection from './ProxyBaseSection.vue'
import ProxyHealthSection from './ProxyHealthSection.vue'
import ProxyHttpSection from './ProxyHttpSection.vue'
import ProxyLoadBalanceSection from './ProxyLoadBalanceSection.vue'
import ProxyMetadataSection from './ProxyMetadataSection.vue'
import ProxyNatSection from './ProxyNatSection.vue'
import ProxyRemoteSection from './ProxyRemoteSection.vue'
import ProxyTransportSection from './ProxyTransportSection.vue'
import { useI18n } from '../../i18n'

const props = withDefaults(defineProps<{
  modelValue: ProxyFormData
  readonly?: boolean
  editing?: boolean
  lockType?: boolean
}>(), { readonly: false, editing: false, lockType: false })

const emit = defineEmits<{ 'update:modelValue': [value: ProxyFormData] }>()
const { t } = useI18n()

const form = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const lockType = computed(() => props.lockType)

const pairingNote = computed(() => {
  if (!['stcp', 'sudp', 'xtcp'].includes(form.value.type)) {
    return ''
  }
  return t('form.directAccessHint')
})

const httpNote = computed(() => {
  if (form.value.type === 'http') return t('form.httpProxyHint')
  if (form.value.type === 'https') return t('form.httpsProxyHint')
  if (form.value.type === 'tcpmux') return t('form.tcpmuxHint')
  return ''
})

const trafficAccessConfigured = computed(() => {
  return Boolean(
    form.value.httpUser ||
      form.value.httpPassword ||
      form.value.secretKey ||
      form.value.allowUsers.length ||
      form.value.locations.length ||
      form.value.requestHeaders.length ||
      form.value.responseHeaders.length ||
      form.value.routeByHTTPUser,
  )
})

const showTrafficAndAccess = computed(() => {
  return ['http', 'tcpmux', 'stcp', 'sudp', 'xtcp'].includes(form.value.type)
})

const runtimeConfigured = computed(() => {
  return Boolean(
    form.value.useEncryption ||
      form.value.useCompression ||
      form.value.bandwidthLimit ||
      form.value.proxyProtocolVersion ||
      form.value.loadBalancerGroup ||
      form.value.healthCheckType ||
      form.value.natTraversalDisableAssistedAddrs,
  )
})

const metadataConfigured = computed(() => {
  return form.value.metadatas.length > 0 || form.value.annotations.length > 0
})
</script>

<style scoped lang="scss">
.helper-copy {
  margin: 0;
  font-size: $font-size-sm;
  line-height: 1.7;
  color: $color-text-secondary;
}
</style>
