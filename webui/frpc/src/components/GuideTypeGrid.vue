<template>
  <div class="guide-grid">
    <button
      v-for="item in items"
      :key="item.type"
      class="guide-card"
      :class="{ disabled }"
      :disabled="disabled"
      type="button"
      @click="$emit('select', item.type)"
    >
      <div class="guide-card-top">
        <span class="guide-badge">{{ item.label }}</span>
        <span class="guide-type">{{ item.type.toUpperCase() }}</span>
      </div>
      <h3 class="guide-title">{{ item.summary }}</h3>
      <p class="guide-scenarios">{{ item.scenarios.join(' | ') }}</p>
      <p v-if="item.caution" class="guide-caution">{{ item.caution }}</p>
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  items: Array<{
    type: string
    label: string
    summary: string
    scenarios: string[]
    caution?: string
  }>
  disabled?: boolean
}>()

defineEmits<{
  select: [type: string]
}>()
</script>

<style scoped lang="scss">
.guide-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
}

.guide-card {
  text-align: left;
  border: 1px solid $color-border-light;
  border-radius: 18px;
  background: linear-gradient(180deg, $color-bg-primary 0%, $color-bg-secondary 100%);
  padding: 18px;
  cursor: pointer;
  transition: transform $transition-fast, border-color $transition-fast, box-shadow $transition-fast;

  &:hover:not(.disabled) {
    transform: translateY(-2px);
    border-color: $color-border;
    box-shadow: 0 12px 30px rgba(15, 23, 42, 0.08);
  }

  &.disabled {
    cursor: not-allowed;
    opacity: 0.6;
  }
}

.guide-card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}

.guide-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  background: rgba(17, 24, 39, 0.06);
  color: $color-text-primary;
}

.guide-type {
  font-size: 12px;
  font-weight: 700;
  color: $color-text-light;
  letter-spacing: 0.08em;
}

.guide-title {
  margin: 0 0 10px;
  font-size: 15px;
  line-height: 1.5;
  color: $color-text-primary;
}

.guide-scenarios,
.guide-caution {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
}

.guide-scenarios {
  color: $color-text-muted;
}

.guide-caution {
  margin-top: 12px;
  color: $color-text-secondary;
  font-weight: 500;
}
</style>
