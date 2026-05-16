<template>
  <div class="page-grid">
    <section class="toolbar">
      <ElInput
        v-model="query"
        :placeholder="t('dashboardPage.search.placeholder')"
        clearable
        @keyup.enter="load"
      />
      <ElButton type="primary" @click="load">{{ t('common.search') }}</ElButton>
    </section>

    <ElCard class="panel" shadow="never">
      <div class="panel-header">
        <h3>{{ t('dashboardPage.search.title') }}</h3>
        <span>{{ t('dashboardPage.search.matches', { count: results.length }) }}</span>
      </div>
      <ElEmpty v-if="!loading && !results.length" :description="t('common.noData')" />
      <div v-else class="result-list" v-loading="loading">
        <article v-for="item in results" :key="item.event_id" class="result-item">
          <div class="result-top">
            <strong>{{ item.content_summary || item.type }}</strong>
            <ElTag size="small">{{ item.type }}</ElTag>
          </div>
          <p class="meta">{{ item.timestamp }} · {{ item.direction }}</p>
          <p class="participants">{{ item.participants.join(', ') }}</p>
        </article>
      </div>
      <div class="pagination-wrapper" v-if="results.length > 0">
        <ElPagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { fetchSearch } from '@/api/commory'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'SearchRecords' })

  const { t } = useI18n()
  const query = ref('')
  const loading = ref(false)
  const results = ref<Api.Commory.TimelineItem[]>([])
  const currentPage = ref(1)
  const pageSize = 20
  const total = ref(0)

  const load = async () => {
    loading.value = true
    try {
      const offset = (currentPage.value - 1) * pageSize
      results.value = await fetchSearch({ q: query.value, limit: pageSize, offset })
      total.value =
        results.value.length === pageSize
          ? currentPage.value * pageSize + pageSize
          : offset + results.value.length
    } finally {
      loading.value = false
    }
  }

  const handlePageChange = (page: number) => {
    currentPage.value = page
    load()
  }
</script>

<style scoped lang="scss">
  .page-grid {
    display: grid;
    gap: 16px;
  }

  .toolbar,
  .panel {
    background: var(--art-main-bg-color);
    border: 1px solid var(--art-border-color);
    border-radius: 8px;
  }

  .toolbar {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 120px;
    gap: 12px;
    padding: 16px;
  }

  .panel-header,
  .result-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .result-list {
    display: grid;
    gap: 14px;
    margin-top: 16px;
  }

  .result-item {
    padding-bottom: 14px;
    border-bottom: 1px solid var(--art-border-color);
  }

  .result-item:last-child {
    border-bottom: 0;
    padding-bottom: 0;
  }

  .meta,
  .participants,
  .panel-header span {
    color: var(--art-gray-600);
  }

  .pagination-wrapper {
    display: flex;
    justify-content: center;
    margin-top: 16px;
  }
</style>
