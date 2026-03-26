<template>
  <ConfigSection
    :title="t('form.httpOptions')"
    collapsible
    :readonly="readonly"
    :has-value="form.locations.length > 0 || !!form.hostHeaderRewrite || form.requestHeaders.length > 0 || form.responseHeaders.length > 0"
  >
    <ConfigField :label="t('form.locations')" type="tags" v-model="form.locations" :placeholder="t('form.pathPlaceholder')" :readonly="readonly" />
    <ConfigField :label="t('form.hostHeaderRewrite')" type="text" v-model="form.hostHeaderRewrite" :readonly="readonly" />
    <ConfigField :label="t('form.requestHeaders')" type="kv" v-model="form.requestHeaders" :key-placeholder="t('common.key')" :value-placeholder="t('common.value')" :readonly="readonly" />
    <ConfigField :label="t('form.responseHeaders')" type="kv" v-model="form.responseHeaders" :key-placeholder="t('common.key')" :value-placeholder="t('common.value')" :readonly="readonly" />
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
