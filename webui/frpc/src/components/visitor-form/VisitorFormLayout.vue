<template>
  <div class="visitor-form-layout">
    <ConfigSection :readonly="readonly" :title="t('form.coreSetup')">
      <VisitorBaseSection v-model="form" :readonly="readonly" :editing="editing" :lock-type="lockType" />
      <VisitorConnectionSection v-model="form" :readonly="readonly" />
    </ConfigSection>

    <ConfigSection :title="t('form.pairingNote')" :readonly="readonly">
      <p class="helper-copy">{{ t('form.visitorPairHint') }}</p>
    </ConfigSection>

    <ConfigSection
      :title="t('form.transportAndAdvanced')"
      :readonly="readonly"
      collapsible
      :has-value="advancedConfigured"
    >
      <VisitorTransportSection v-model="form" :readonly="readonly" />
      <VisitorXtcpSection v-if="form.type === 'xtcp'" v-model="form" :readonly="readonly" />
    </ConfigSection>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { VisitorFormData } from '../../types'
import ConfigSection from '../ConfigSection.vue'
import VisitorBaseSection from './VisitorBaseSection.vue'
import VisitorConnectionSection from './VisitorConnectionSection.vue'
import VisitorTransportSection from './VisitorTransportSection.vue'
import VisitorXtcpSection from './VisitorXtcpSection.vue'
import { useI18n } from '../../i18n'

const props = withDefaults(defineProps<{
  modelValue: VisitorFormData
  readonly?: boolean
  editing?: boolean
  lockType?: boolean
}>(), { readonly: false, editing: false, lockType: false })

const emit = defineEmits<{ 'update:modelValue': [value: VisitorFormData] }>()
const { t } = useI18n()

const form = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const lockType = computed(() => props.lockType)

const advancedConfigured = computed(() => {
  return Boolean(
    form.value.useEncryption ||
      form.value.useCompression ||
      form.value.keepTunnelOpen ||
      form.value.maxRetriesAnHour != null ||
      form.value.minRetryInterval != null ||
      form.value.fallbackTo ||
      form.value.fallbackTimeoutMs != null ||
      form.value.natTraversalDisableAssistedAddrs,
  )
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
