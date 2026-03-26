<template>
  <ConfigSection
    :title="t('form.xtcpOptions')"
    collapsible
    :readonly="readonly"
    :has-value="form.protocol !== 'quic' || form.keepTunnelOpen || form.maxRetriesAnHour != null || form.minRetryInterval != null || !!form.fallbackTo || form.fallbackTimeoutMs != null"
  >
    <ConfigField :label="t('form.protocol')" type="select" v-model="form.protocol" :options="[{ label: t('form.quic'), value: 'quic' }, { label: t('form.kcp'), value: 'kcp' }]" :readonly="readonly" />
    <ConfigField :label="t('form.keepTunnelOpen')" type="switch" v-model="form.keepTunnelOpen" :readonly="readonly" />
    <div class="field-row two-col">
      <ConfigField :label="t('form.maxRetriesPerHour')" type="number" v-model="form.maxRetriesAnHour" :min="0" :readonly="readonly" />
      <ConfigField :label="t('form.minRetryInterval')" type="number" v-model="form.minRetryInterval" :min="0" :readonly="readonly" />
    </div>
    <div class="field-row two-col">
      <ConfigField :label="t('form.fallbackTo')" type="text" v-model="form.fallbackTo" :placeholder="t('form.fallbackVisitorPlaceholder')" :readonly="readonly" />
      <ConfigField :label="t('form.fallbackTimeoutMs')" type="number" v-model="form.fallbackTimeoutMs" :min="0" :readonly="readonly" />
    </div>
  </ConfigSection>

  <ConfigSection :title="t('form.natTraversal')" collapsible :readonly="readonly" :has-value="form.natTraversalDisableAssistedAddrs">
    <ConfigField :label="t('form.disableAssistedAddresses')" type="switch" v-model="form.natTraversalDisableAssistedAddrs" :tip="t('form.assistedTip')" :readonly="readonly" />
  </ConfigSection>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { VisitorFormData } from '../../types'
import ConfigSection from '../ConfigSection.vue'
import ConfigField from '../ConfigField.vue'
import { useI18n } from '../../i18n'

const props = withDefaults(defineProps<{
  modelValue: VisitorFormData
  readonly?: boolean
}>(), { readonly: false })

const emit = defineEmits<{ 'update:modelValue': [value: VisitorFormData] }>()
const { t } = useI18n()

const form = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})
</script>

<style scoped lang="scss">
@use '@/assets/css/form-layout';
</style>
