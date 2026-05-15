<template>
  <div class="contacts-layout">
    <section class="list-panel">
      <div class="panel-header">
        <h3>Contacts</h3>
        <span>{{ contacts.length }}</span>
      </div>
      <ElInput v-model="keyword" placeholder="Filter contacts" clearable />
      <div class="contact-list">
        <button
          v-for="item in filteredContacts"
          :key="item.id"
          class="contact-row"
          :class="{ active: selected?.id === item.id }"
          @click="selectContact(item)"
        >
          <strong>{{ item.display_name }}</strong>
          <span>{{ item.phones[0] || item.emails[0] || item.id }}</span>
        </button>
      </div>
    </section>

    <section class="detail-panel">
      <template v-if="selected">
        <div class="panel-header">
          <div>
            <h3>{{ selected.display_name }}</h3>
            <p class="detail-meta">{{ selected.type }} · {{ selected.labels.join(', ') }}</p>
          </div>
          <ElButton link type="primary" @click="loadTimeline">Load Activity</ElButton>
        </div>

        <ElDescriptions :column="1" border>
          <ElDescriptionsItem label="Phones">{{
            selected.phones.join(', ') || '-'
          }}</ElDescriptionsItem>
          <ElDescriptionsItem label="Emails">{{
            selected.emails.join(', ') || '-'
          }}</ElDescriptionsItem>
          <ElDescriptionsItem label="Source">{{ selected.meta?.source || '-' }}</ElDescriptionsItem>
        </ElDescriptions>

        <div class="activity-list">
          <div v-for="event in events" :key="event.event_id" class="activity-row">
            <div>
              <strong>{{ event.content_summary || event.type }}</strong>
              <p class="detail-meta">{{ event.type }} · {{ event.direction }}</p>
            </div>
            <span class="detail-meta">{{ event.timestamp }}</span>
          </div>
        </div>
      </template>
      <div v-else class="empty-state"
        >Select a contact to inspect its MsgLayer identity and recent activity.</div
      >
    </section>
  </div>
</template>

<script setup lang="ts">
  import { fetchIdentities, fetchTimeline } from '@/api/commory'

  defineOptions({ name: 'Contacts' })

  const contacts = ref<Api.Commory.Identity[]>([])
  const selected = ref<Api.Commory.Identity>()
  const events = ref<Api.Commory.TimelineItem[]>([])
  const keyword = ref('')

  const filteredContacts = computed(() =>
    contacts.value.filter((item) => {
      const needle = keyword.value.trim().toLowerCase()
      if (!needle) return true
      return (
        item.display_name.toLowerCase().includes(needle) ||
        item.phones.some((phone) => phone.toLowerCase().includes(needle))
      )
    })
  )

  const selectContact = async (item: Api.Commory.Identity) => {
    selected.value = item
    await loadTimeline()
  }

  const loadTimeline = async () => {
    if (!selected.value) return
    events.value = await fetchTimeline({ participant: selected.value.id, limit: 50 })
  }

  const load = async () => {
    contacts.value = await fetchIdentities()
    selected.value = contacts.value[0]
    await loadTimeline()
  }

  onMounted(load)
</script>

<style scoped lang="scss">
  .contacts-layout {
    display: grid;
    grid-template-columns: 320px minmax(0, 1fr);
    gap: 18px;
  }

  .list-panel,
  .detail-panel {
    background: var(--art-main-bg-color);
    border: 1px solid var(--art-border-color);
    border-radius: 8px;
    padding: 18px;
  }

  .panel-header,
  .activity-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .contact-list,
  .activity-list {
    display: grid;
    gap: 10px;
    margin-top: 16px;
  }

  .contact-row {
    border: 1px solid var(--art-border-color);
    border-radius: 8px;
    background: transparent;
    padding: 12px;
    text-align: left;
  }

  .contact-row.active {
    border-color: var(--el-color-primary);
    background: rgba(64, 158, 255, 0.08);
  }

  .contact-row span,
  .detail-meta,
  .empty-state {
    color: var(--art-gray-600);
  }

  .activity-row {
    padding-bottom: 12px;
    border-bottom: 1px solid var(--art-border-color);
  }

  .activity-row:last-child {
    border-bottom: 0;
    padding-bottom: 0;
  }

  @media (max-width: 960px) {
    .contacts-layout {
      grid-template-columns: 1fr;
    }
  }
</style>
