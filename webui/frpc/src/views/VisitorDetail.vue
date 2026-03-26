<template>
  <div class="visitor-detail-page">
    <div class="detail-top">
      <nav class="breadcrumb">
        <router-link to="/visitors" class="breadcrumb-link">{{ t('app.visitors') }}</router-link>
        <span class="breadcrumb-sep">&rsaquo;</span>
        <span class="breadcrumb-current">{{ visitorName }}</span>
      </nav>

      <template v-if="visitor">
        <div class="detail-header">
          <div>
            <h2 class="detail-title">{{ visitor.name }}</h2>
            <p class="header-subtitle">
              {{ activeGuide?.summary || t('visitorDetail.fallbackSummary') }}
            </p>
            <p class="header-meta">{{ t('visitorDetail.typeMeta', { type: visitor.type.toUpperCase() }) }}</p>
          </div>
          <div v-if="isStore" class="header-actions">
            <ActionButton variant="outline" size="small" @click="handleEdit">
              {{ t('common.edit') }}
            </ActionButton>
          </div>
        </div>
      </template>
    </div>

    <div v-if="notFound" class="not-found">
      <p class="empty-text">{{ t('visitorDetail.notFoundTitle') }}</p>
      <p class="empty-hint">{{ t('visitorDetail.notFoundHint', { name: visitorName }) }}</p>
      <ActionButton variant="outline" @click="router.push('/visitors')">
        {{ t('common.backToVisitors') }}
      </ActionButton>
    </div>

    <div v-else-if="visitor" v-loading="loading" class="detail-content">
      <div v-if="activeGuide" class="guide-banner">
        <div class="guide-banner-title">{{ activeGuide.label }}</div>
        <div class="guide-banner-copy">
          {{ activeGuide.caution || t('visitorDetail.fallbackGuide') }}
        </div>
      </div>

      <VisitorFormLayout
        v-if="formData"
        :model-value="formData"
        readonly
      />
    </div>

    <div v-else v-loading="loading" class="loading-area"></div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import ActionButton from '@shared/components/ActionButton.vue'
import VisitorFormLayout from '../components/visitor-form/VisitorFormLayout.vue'
import { getVisitorGuideMap } from '../content/guides'
import { getStoreVisitor, getVisitorConfig } from '../api/frpc'
import { storeVisitorToForm } from '../types'
import type { VisitorDefinition, VisitorFormData, VisitorType } from '../types'
import { useI18n } from '../i18n'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const visitorName = route.params.name as string
const visitor = ref<VisitorDefinition | null>(null)
const loading = ref(true)
const notFound = ref(false)
const isStore = ref(false)

onMounted(async () => {
  try {
    const config = await getVisitorConfig(visitorName)
    visitor.value = config

    try {
      await getStoreVisitor(visitorName)
      isStore.value = true
    } catch {
      // not store managed
    }
  } catch (err: any) {
    if (err?.status === 404 || err?.response?.status === 404) {
      notFound.value = true
    } else {
      notFound.value = true
      ElMessage.error(t('visitorEdit.loadFailed', { message: err.message }))
    }
  } finally {
    loading.value = false
  }
})

const visitorGuideMap = computed(() => getVisitorGuideMap())

const activeGuide = computed(() => {
  const type = visitor.value?.type
  if (!type) return null
  return visitorGuideMap.value[type as VisitorType] || null
})

const formData = computed<VisitorFormData | null>(() => {
  if (!visitor.value) return null
  return storeVisitorToForm(visitor.value)
})

const handleEdit = () => {
  router.push('/visitors/' + encodeURIComponent(visitorName) + '/edit')
}
</script>

<style scoped lang="scss">
.visitor-detail-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  max-width: 1080px;
  margin: 0 auto;
}

.detail-top {
  flex-shrink: 0;
  padding: $spacing-xl 28px 0;
}

.detail-content {
  flex: 1;
  overflow-y: auto;
  padding: 0 28px 120px;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  margin-bottom: 18px;
}

.breadcrumb-link {
  color: $color-text-secondary;
  text-decoration: none;

  &:hover {
    color: $color-text-primary;
  }
}

.breadcrumb-current {
  color: $color-text-primary;
  font-weight: 500;
}

.breadcrumb-sep {
  color: $color-text-light;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 22px;
}

.detail-title {
  margin: 0 0 8px;
  font-size: 26px;
  color: $color-text-primary;
}

.header-subtitle,
.header-meta {
  margin: 0;
  font-size: 14px;
  line-height: 1.7;
}

.header-subtitle {
  color: $color-text-secondary;
}

.header-meta {
  color: $color-text-muted;
}

.guide-banner {
  margin-bottom: 18px;
  padding: 16px 18px;
  border-radius: 16px;
  background: linear-gradient(135deg, #eef6f1 0%, #fbfdfc 100%);
  border: 1px solid rgba(148, 163, 184, 0.18);
}

.guide-banner-title {
  margin-bottom: 6px;
  font-size: 14px;
  font-weight: 700;
  color: $color-text-primary;
}

.guide-banner-copy {
  font-size: 14px;
  line-height: 1.7;
  color: $color-text-secondary;
}

.not-found,
.loading-area {
  text-align: center;
  padding: 60px 20px;
}

.empty-text {
  margin: 0 0 8px;
  font-size: 18px;
  font-weight: 600;
  color: $color-text-secondary;
}

.empty-hint {
  margin: 0 0 18px;
  font-size: 14px;
  color: $color-text-muted;
}

@include mobile {
  .detail-top {
    padding: $spacing-lg $spacing-lg 0;
  }

  .detail-content {
    padding: 0 $spacing-lg 120px;
  }

  .detail-header {
    flex-direction: column;
  }
}
</style>
