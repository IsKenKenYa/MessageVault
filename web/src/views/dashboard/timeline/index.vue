<template>
  <div class="page-grid">
    <section class="toolbar">
      <ElInput v-model="filters.q" placeholder="Search timeline text" clearable />
      <ElSelect v-model="filters.type" clearable placeholder="Event type">
        <ElOption label="All" value="" />
        <ElOption label="SMS" value="sms" />
        <ElOption label="Call" value="call" />
        <ElOption label="Voice" value="voice" />
        <ElOption label="Contact Snapshot" value="contact_snapshot" />
      </ElSelect>
      <ElButton type="primary" @click="load">Apply</ElButton>
    </section>

    <section class="panel">
      <ElTable :data="items" size="large" v-loading="loading">
        <ElTableColumn prop="timestamp" label="Timestamp" min-width="190" />
        <ElTableColumn prop="type" label="Type" width="130" />
        <ElTableColumn prop="direction" label="Direction" width="120" />
        <ElTableColumn prop="content_summary" label="Summary" min-width="320" />
        <ElTableColumn label="Participants" min-width="240">
          <template #default="{ row }">{{ row.participants.join(', ') }}</template>
        </ElTableColumn>
      </ElTable>
      <div class="pagination-wrapper" v-if="items.length > 0">
        <ElPagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
  import { fetchTimeline } from '@/api/commory'

  defineOptions({ name: 'Timeline' })

  const loading = ref(false)
  const items = ref<Api.Commory.TimelineItem[]>([])
  const currentPage = ref(1)
  const pageSize = 20
  const total = ref(0)
  const filters = reactive<Api.Commory.SearchParams>({
    q: '',
    type: '',
    limit: 20
  })

  const load = async () => {
    loading.value = true
    try {
      const offset = (currentPage.value - 1) * pageSize
      filters.offset = offset
      filters.limit = pageSize
      items.value = await fetchTimeline(filters)
      total.value =
        items.value.length === pageSize
          ? currentPage.value * pageSize + pageSize
          : offset + items.value.length
    } finally {
      loading.value = false
    }
  }

  const handlePageChange = (page: number) => {
    currentPage.value = page
    load()
  }

  onMounted(load)
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
    grid-template-columns: minmax(0, 1.4fr) 180px 120px;
    gap: 12px;
    padding: 16px;
  }

  .panel {
    padding: 18px;
  }

  .pagination-wrapper {
    display: flex;
    justify-content: center;
    margin-top: 16px;
  }

  @media (max-width: 900px) {
    .toolbar {
      grid-template-columns: 1fr;
    }
  }
</style>
