<template>
  <ConfigSection :title="t('form.connection')" :readonly="readonly">
    <div class="field-row two-col">
      <ConfigField :label="t('form.serverName')" type="text" v-model="form.serverName" prop="serverName" :placeholder="t('form.proxyToVisitPlaceholder')" :readonly="readonly" />
      <ConfigField :label="t('form.serverUser')" type="text" v-model="form.serverUser" :placeholder="t('form.sameUserPlaceholder')" :readonly="readonly" />
    </div>
    <ConfigField :label="t('form.secretKey')" type="password" v-model="form.secretKey" :placeholder="t('form.sharedSecretPlaceholder')" :readonly="readonly" />
    <div class="field-row two-col">
      <ConfigField :label="t('form.bindAddress')" type="text" v-model="form.bindAddr" placeholder="127.0.0.1" :readonly="readonly" />
      <ConfigField :label="t('form.bindPort')" type="number" v-model="form.bindPort" :min="bindPortMin" :max="65535" prop="bindPort" :readonly="readonly" />
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

const bindPortMin = computed(() => (form.value.type === 'sudp' ? 1 : undefined))
</script>

<style scoped lang="scss">
@use '@/assets/css/form-layout';
</style>
