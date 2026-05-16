<template>
  <div class="commory-page">
    <ElCard class="overview-card" shadow="never">
      <div class="overview-header">
        <div>
          <p class="eyebrow">Commory</p>
          <h2>{{ t('dashboardPage.console.title') }}</h2>
          <p>{{ t('dashboardPage.console.subtitle') }}</p>
        </div>
        <ElButton type="primary" @click="router.push('/dashboard/imports')" v-ripple>
          {{ t('dashboardPage.console.openImports') }}
        </ElButton>
      </div>
    </ElCard>

    <section class="stats-grid">
      <ElCard v-for="card in cards" :key="card.label" class="stat-card" shadow="never">
        <span>{{ card.label }}</span>
        <strong>{{ card.value }}</strong>
        <p>{{ card.hint }}</p>
      </ElCard>
    </section>

    <section class="content-grid">
      <ElCard class="art-table-card" shadow="never">
        <template #header>
          <div class="panel-header">
            <h3>{{ t('dashboardPage.console.recentImports') }}</h3>
            <span>{{
              t('dashboardPage.console.itemCount', { count: summary?.recentImports.length || 0 })
            }}</span>
          </div>
        </template>
        <ElTable :data="summary?.recentImports || []" size="large" v-loading="loading">
          <ElTableColumn prop="id" :label="t('dashboardPage.columns.import')" min-width="180" />
          <ElTableColumn
            prop="schema_version"
            :label="t('dashboardPage.columns.version')"
            width="140"
          />
          <ElTableColumn
            prop="event_count"
            :label="t('dashboardPage.columns.events')"
            width="100"
          />
          <ElTableColumn
            prop="imported_at"
            :label="t('dashboardPage.columns.importedAt')"
            min-width="180"
          />
        </ElTable>
      </ElCard>

      <ElCard class="events-card" shadow="never">
        <template #header>
          <div class="panel-header">
            <h3>{{ t('dashboardPage.console.recentEvents') }}</h3>
            <ElButton link type="primary" @click="router.push('/dashboard/timeline')">
              {{ t('dashboardPage.console.viewTimeline') }}
            </ElButton>
          </div>
        </template>
        <ElEmpty
          v-if="!loading && !(summary?.recentEvents || []).length"
          :description="t('dashboardPage.console.emptyEvents')"
        />
        <div v-else class="event-list" v-loading="loading">
          <div v-for="event in summary?.recentEvents || []" :key="event.event_id" class="event-row">
            <div>
              <p class="event-title">{{ event.content_summary || event.type }}</p>
              <p class="muted">{{ event.type }} · {{ event.direction }}</p>
            </div>
            <span class="muted">{{ event.timestamp }}</span>
          </div>
        </div>
      </ElCard>
    </section>
  </div>
</template>

<script setup lang="ts">
  import { fetchDashboardSummary } from '@/api/commory'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'Console' })

  const { t } = useI18n()
  const router = useRouter()
  const summary = ref<Api.Commory.DashboardSummary>()
  const loading = ref(false)

  const cards = computed(() => [
    {
      label: t('dashboardPage.console.cards.imports'),
      value: summary.value?.importCount ?? 0,
      hint: t('dashboardPage.console.hints.imports')
    },
    {
      label: t('dashboardPage.console.cards.events'),
      value: summary.value?.eventCount ?? 0,
      hint: t('dashboardPage.console.hints.events')
    },
    {
      label: t('dashboardPage.console.cards.identities'),
      value: summary.value?.identityCount ?? 0,
      hint: t('dashboardPage.console.hints.identities')
    },
    {
      label: t('dashboardPage.console.cards.lastActivity'),
      value: summary.value?.lastActivity || t('dashboardPage.console.noImports'),
      hint: t('dashboardPage.console.hints.lastActivity')
    }
  ])

  const load = async () => {
    loading.value = true
    try {
      summary.value = await fetchDashboardSummary()
    } finally {
      loading.value = false
    }
  }

  onMounted(load)
  onActivated(load)
</script>

<style scoped lang="scss">
  .commory-page {
    display: grid;
    gap: 16px;
  }

  .overview-card,
  .stat-card,
  .events-card {
    border: 1px solid var(--art-border-color);
    border-radius: 8px;
  }

  .overview-header {
    display: flex;
    gap: 16px;
    align-items: flex-start;
    justify-content: space-between;

    h2,
    p {
      margin: 0;
    }

    h2 {
      margin-top: 5px;
      font-size: 24px;
      font-weight: 600;
      color: var(--art-gray-800);
    }

    p:not(.eyebrow) {
      margin-top: 8px;
      color: var(--art-gray-600);
    }
  }

  .eyebrow {
    margin: 0;
    color: var(--main-color);
    font-size: 13px;
    font-weight: 600;
  }

  .stats-grid,
  .content-grid {
    display: grid;
    gap: 16px;
  }

  .stats-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .content-grid {
    grid-template-columns: minmax(0, 1.08fr) minmax(360px, 0.92fr);
  }

  .stat-card :deep(.el-card__body) {
    display: grid;
    gap: 10px;
    min-height: 132px;
  }

  .stat-card {
    span,
    p {
      margin: 0;
      color: var(--art-gray-600);
    }

    strong {
      font-size: 28px;
      line-height: 1.1;
      color: var(--art-gray-900);
    }
  }

  .panel-header,
  .event-row {
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
  }

  .panel-header {
    h3 {
      margin: 0;
      font-size: 16px;
      font-weight: 600;
    }

    span {
      color: var(--art-gray-600);
    }
  }

  .event-list {
    display: grid;
    gap: 14px;
  }

  .event-row {
    padding-bottom: 14px;
    border-bottom: 1px solid var(--art-border-color);

    &:last-child {
      padding-bottom: 0;
      border-bottom: 0;
    }
  }

  .event-title {
    margin: 0 0 4px;
  }

  .muted {
    color: var(--art-gray-600);
  }

  @media (max-width: 1100px) {
    .stats-grid,
    .content-grid {
      grid-template-columns: 1fr;
    }

    .overview-header {
      flex-direction: column;
      align-items: stretch;
    }
  }
</style>
