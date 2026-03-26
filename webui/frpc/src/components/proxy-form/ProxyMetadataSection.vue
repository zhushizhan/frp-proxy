<template>
  <ConfigSection :title="t('form.metadata')" collapsible :readonly="readonly" :has-value="form.metadatas.length > 0 || form.annotations.length > 0">
    <ConfigField :label="t('form.metadatas')" type="kv" v-model="form.metadatas" :key-placeholder="t('common.key')" :value-placeholder="t('common.value')" :readonly="readonly" />
    <ConfigField :label="t('form.annotations')" type="kv" v-model="form.annotations" :key-placeholder="t('common.key')" :value-placeholder="t('common.value')" :readonly="readonly" />
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
