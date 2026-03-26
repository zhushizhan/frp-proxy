<template>
  <ConfigSection :title="t('form.authentication')" :readonly="readonly">
    <template v-if="['http', 'tcpmux'].includes(form.type)">
      <div class="field-row three-col">
        <ConfigField :label="t('form.httpUser')" type="text" v-model="form.httpUser" :readonly="readonly" />
        <ConfigField :label="t('form.httpPassword')" type="password" v-model="form.httpPassword" :readonly="readonly" />
        <ConfigField :label="t('form.routeByHttpUser')" type="text" v-model="form.routeByHTTPUser" :readonly="readonly" />
      </div>
    </template>
    <template v-if="['stcp', 'sudp', 'xtcp'].includes(form.type)">
      <div class="field-row two-col">
        <ConfigField :label="t('form.secretKey')" type="password" v-model="form.secretKey" prop="secretKey" :readonly="readonly" />
        <ConfigField :label="t('form.allowUsers')" type="tags" v-model="form.allowUsers" :placeholder="t('form.usernamePlaceholder')" :readonly="readonly" />
      </div>
    </template>
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
