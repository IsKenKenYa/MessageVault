<template>
  <ArtResultPage
    type="fail"
    :title="t('resultPage.fail.title')"
    :message="t('resultPage.fail.message')"
    iconCode="ri:close-fill"
  >
    <template #content>
      <p>{{ t('resultPage.fail.intro') }}</p>
      <p>
        <ArtSvgIcon icon="ri:close-circle-line" class="text-red-500 mr-1" />
        <span>{{ t('resultPage.fail.accountFrozen') }}</span>
      </p>
      <p>
        <ArtSvgIcon icon="ri:close-circle-line" class="text-red-500 mr-1" />
        <span>{{ t('resultPage.fail.notEligible') }}</span>
      </p>
    </template>
    <template #buttons>
      <ElButton type="primary" v-ripple>{{ t('resultPage.actions.backEdit') }}</ElButton>
      <ElButton v-ripple>{{ t('resultPage.actions.view') }}</ElButton>
    </template>
  </ArtResultPage>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'ResultFail' })
  const { t } = useI18n()
</script>
