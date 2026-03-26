<template>
  <BaseDialog
    v-model="visible"
    :title="title"
    width="400px"
    :close-on-click-modal="false"
    :append-to-body="true"
    :is-mobile="isMobile"
  >
    <p class="confirm-message">{{ message }}</p>
    <template #footer>
      <div class="dialog-footer">
        <ActionButton variant="outline" @click="handleCancel">
          {{ resolvedCancelText }}
        </ActionButton>
        <ActionButton
          :danger="danger"
          :loading="loading"
          @click="handleConfirm"
        >
          {{ resolvedConfirmText }}
        </ActionButton>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import BaseDialog from './BaseDialog.vue'
import ActionButton from '@shared/components/ActionButton.vue'
import { useI18n } from '../../frpc/src/i18n'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title: string
    message: string
    confirmText?: string
    cancelText?: string
    danger?: boolean
    loading?: boolean
    isMobile?: boolean
  }>(),
  {
    confirmText: '',
    cancelText: '',
    danger: false,
    loading: false,
    isMobile: false,
  },
)

const { t } = useI18n()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'confirm'): void
  (e: 'cancel'): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

const resolvedConfirmText = computed(() => props.confirmText || t('common.confirm'))
const resolvedCancelText = computed(() => props.cancelText || t('common.cancel'))

const handleConfirm = () => {
  emit('confirm')
}

const handleCancel = () => {
  visible.value = false
  emit('cancel')
}
</script>

<style scoped lang="scss">
.confirm-message {
  margin: 0;
  font-size: $font-size-md;
  color: $color-text-secondary;
  line-height: 1.6;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: $spacing-md;
}
</style>
