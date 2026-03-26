<template>
  <ConfigSection :title="t('form.transportOptions')" collapsible :readonly="readonly" :has-value="form.useEncryption || form.useCompression">
    <div class="field-row two-col">
      <ConfigField :label="t('form.useEncryption')" type="switch" v-model="form.useEncryption" :readonly="readonly" />
      <ConfigField :label="t('form.useCompression')" type="switch" v-model="form.useCompression" :readonly="readonly" />
    </div>
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
