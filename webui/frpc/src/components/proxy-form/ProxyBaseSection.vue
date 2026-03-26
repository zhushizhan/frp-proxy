<template>
  <div v-if="!readonly" class="field-row three-col">
    <el-form-item :label="t('common.name')" prop="name" class="field-grow">
      <el-input
        v-model="form.name"
        :disabled="editing || readonly"
        :placeholder="t('form.namePlaceholderProxy')"
      />
    </el-form-item>
    <ConfigField
      :label="t('common.type')"
      type="select"
      v-model="form.type"
      :disabled="editing || lockType"
      :options="PROXY_TYPES.map((item) => ({ label: item.toUpperCase(), value: item }))"
      prop="type"
    />
    <el-form-item :label="t('common.enabled')" class="switch-field">
      <el-switch v-model="form.enabled" size="small" />
    </el-form-item>
  </div>
  <div v-else class="field-row three-col">
    <ConfigField :label="t('common.name')" type="text" :model-value="form.name" readonly class="field-grow" />
    <ConfigField :label="t('common.type')" type="text" :model-value="form.type.toUpperCase()" readonly />
    <ConfigField :label="t('common.enabled')" type="switch" :model-value="form.enabled" readonly />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { PROXY_TYPES, type ProxyFormData } from '../../types'
import ConfigField from '../ConfigField.vue'
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
</script>

<style scoped lang="scss">
@use '@/assets/css/form-layout';
</style>
