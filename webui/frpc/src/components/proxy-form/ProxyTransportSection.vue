<template>
  <ConfigSection
    :title="t('form.transport')"
    collapsible
    :readonly="readonly"
    :has-value="form.useEncryption || form.useCompression || !!form.bandwidthLimit || (!!form.bandwidthLimitMode && form.bandwidthLimitMode !== 'client') || !!form.proxyProtocolVersion"
  >
    <div class="field-row two-col">
      <ConfigField :label="t('form.useEncryption')" type="switch" v-model="form.useEncryption" :readonly="readonly" />
      <ConfigField :label="t('form.useCompression')" type="switch" v-model="form.useCompression" :readonly="readonly" />
    </div>
    <div class="field-row three-col">
      <ConfigField :label="t('form.bandwidthLimit')" type="text" v-model="form.bandwidthLimit" placeholder="1MB" :tip="t('form.bandwidthLimitTip')" :readonly="readonly" />
      <ConfigField
        :label="t('form.bandwidthLimitMode')"
        type="select"
        v-model="form.bandwidthLimitMode"
        :options="[
          { label: t('form.client'), value: 'client' },
          { label: t('form.server'), value: 'server' },
        ]"
        :readonly="readonly"
      />
      <ConfigField
        :label="t('form.proxyProtocolVersion')"
        type="select"
        v-model="form.proxyProtocolVersion"
        :options="[
          { label: t('form.none'), value: '' },
          { label: 'v1', value: 'v1' },
          { label: 'v2', value: 'v2' },
        ]"
        :readonly="readonly"
      />
    </div>
  </ConfigSection>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ProxyFormData } from '../../types'
import ConfigSection from '../ConfigSection.vue'
import ConfigField from '../ConfigField.vue'
import { useI18n } from '../../i18n'

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
</script>

<style scoped lang="scss">
@use '@/assets/css/form-layout';
</style>
