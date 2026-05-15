<template>
  <div class="commory-dashboard">
    <section class="hero">
      <div>
        <p class="eyebrow">Commory</p>
        <h2>MsgLayer archive overview</h2>
        <p class="subtext"
          >Imports, events, identities, and the latest activity for the signed-in archive.</p
        >
      </div>
      <ElButton type="primary" @click="router.push('/dashboard/imports')">Open Imports</ElButton>
    </section>

    <section class="stats-grid">
      <article v-for="card in cards" :key="card.label" class="stat-panel">
        <span class="label">{{ card.label }}</span>
        <strong>{{ card.value }}</strong>
        <span class="hint">{{ card.hint }}</span>
      </article>
    </section>

    <section class="content-grid">
      <div class="panel">
        <div class="panel-header">
          <h3>Recent imports</h3>
          <span>{{ summary?.recentImports.length || 0 }} items</span>
        </div>
        <ElTable :data="summary?.recentImports || []" size="large">
          <ElTableColumn prop="id" label="Import" min-width="180" />
          <ElTableColumn prop="schema_version" label="Version" width="140" />
          <ElTableColumn prop="event_count" label="Events" width="100" />
          <ElTableColumn prop="imported_at" label="Imported At" min-width="180" />
        </ElTable>
      </div>

      <div class="panel">
        <div class="panel-header">
          <h3>Recent events</h3>
          <ElButton link type="primary" @click="router.push('/dashboard/timeline')"
            >View Timeline</ElButton
          >
        </div>
        <div class="event-list">
          <div v-for="event in summary?.recentEvents || []" :key="event.event_id" class="event-row">
            <div>
              <p class="event-title">{{ event.content_summary || event.type }}</p>
              <p class="event-meta">{{ event.type }} · {{ event.direction }}</p>
            </div>
            <span class="event-time">{{ event.timestamp }}</span>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
  import { fetchDashboardSummary } from '@/api/commory'

  defineOptions({ name: 'Console' })

  const router = useRouter()
  const summary = ref<Api.Commory.DashboardSummary>()
  const loading = ref(false)

  const cards = computed(() => [
    {
      label: 'Imports',
      value: summary.value?.importCount ?? 0,
      hint: 'Validated MsgLayer payloads'
    },
    {
      label: 'Events',
      value: summary.value?.eventCount ?? 0,
      hint: 'Timeline records indexed'
    },
    {
      label: 'Identities',
      value: summary.value?.identityCount ?? 0,
      hint: 'Contacts and devices discovered'
    },
    {
      label: 'Last activity',
      value: summary.value?.lastActivity || 'No imports yet',
      hint: 'Most recent event timestamp'
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
  .commory-dashboard {
    display: grid;
    gap: 20px;
  }

  .hero,
  .panel,
  .stat-panel {
    background: var(--art-main-bg-color);
    border: 1px solid var(--art-border-color);
    border-radius: 8px;
  }

  .hero {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    padding: 24px;

    h2 {
      margin: 6px 0 10px;
      font-size: 28px;
      line-height: 1.2;
    }
  }

  .eyebrow,
  .subtext,
  .label,
  .hint,
  .event-meta,
  .event-time,
  .panel-header span {
    color: var(--art-gray-600);
  }

  .stats-grid,
  .content-grid {
    display: grid;
    gap: 20px;
  }

  .stats-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .content-grid {
    grid-template-columns: minmax(0, 1.1fr) minmax(0, 0.9fr);
  }

  .stat-panel,
  .panel {
    padding: 20px;
  }

  .stat-panel {
    display: grid;
    gap: 10px;

    strong {
      font-size: 28px;
      line-height: 1.1;
    }
  }

  .panel-header,
  .event-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .event-list {
    display: grid;
    gap: 14px;
    margin-top: 16px;
  }

  .event-row {
    padding-bottom: 14px;
    border-bottom: 1px solid var(--art-border-color);
  }

  .event-row:last-child {
    padding-bottom: 0;
    border-bottom: 0;
  }

  .event-title {
    margin: 0 0 4px;
  }

  @media (max-width: 1080px) {
    .stats-grid,
    .content-grid {
      grid-template-columns: 1fr;
    }

    .hero {
      flex-direction: column;
      align-items: stretch;
    }
  }
</style>
