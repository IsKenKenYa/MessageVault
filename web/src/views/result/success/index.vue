<template>
  <ArtResultPage
    type="success"
    :title="t('resultPage.success.title')"
    :message="t('resultPage.success.message')"
    iconCode="ri:check-fill"
  >
    <template #content>
      <p>{{ t('resultPage.success.content') }}</p>
    </template>
    <template #buttons>
      <ElButton type="primary" v-ripple>{{ t('resultPage.actions.backEdit') }}</ElButton>
      <ElButton v-ripple>{{ t('resultPage.actions.view') }}</ElButton>
      <ElButton v-ripple>{{ t('resultPage.actions.print') }}</ElButton>
    </template>
  </ArtResultPage>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'ResultSuccess' })
  const { t } = useI18n()
</script>
