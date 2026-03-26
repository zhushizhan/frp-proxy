<template>
  <ConfigSection :title="t('form.natTraversal')" collapsible :readonly="readonly" :has-value="form.natTraversalDisableAssistedAddrs">
    <ConfigField :label="t('form.disableAssistedAddresses')" type="switch" v-model="form.natTraversalDisableAssistedAddrs" :tip="t('form.assistedTip')" :readonly="readonly" />
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
