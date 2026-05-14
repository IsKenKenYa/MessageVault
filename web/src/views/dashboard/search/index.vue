<template>
  <div class="page-grid">
    <section class="toolbar">
      <ElInput v-model="query" placeholder="Search SMS text, transcript, or summary" clearable @keyup.enter="load" />
      <ElButton type="primary" @click="load">Search</ElButton>
    </section>

    <section class="panel">
      <div class="panel-header">
        <h3>Results</h3>
        <span>{{ results.length }} matches</span>
      </div>
      <div class="result-list" v-loading="loading">
        <article v-for="item in results" :key="item.event_id" class="result-item">
          <div class="result-top">
            <strong>{{ item.content_summary || item.type }}</strong>
            <ElTag size="small">{{ item.type }}</ElTag>
          </div>
          <p class="meta">{{ item.timestamp }} · {{ item.direction }}</p>
          <p class="participants">{{ item.participants.join(', ') }}</p>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
  import { fetchSearch } from '@/api/commory'

  defineOptions({ name: 'SearchRecords' })

  const query = ref('')
  const loading = ref(false)
  const results = ref<Api.Commory.TimelineItem[]>([])

  const load = async () => {
    loading.value = true
    try {
      results.value = await fetchSearch({ q: query.value, limit: 100 })
    } finally {
      loading.value = false
    }
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

  .panel {
    padding: 18px;
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
</style>
