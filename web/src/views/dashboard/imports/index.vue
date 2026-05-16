<template>
  <div class="page-grid">
    <section class="toolbar">
      <ElUpload
        :auto-upload="false"
        :show-file-list="false"
        :on-change="handleFileChange"
        accept=".json"
      >
        <ElButton>{{ t('dashboardPage.imports.chooseFile') }}</ElButton>
      </ElUpload>
      <span class="file-name">{{ selectedFile?.name || t('dashboardPage.imports.noFile') }}</span>
      <ElButton :disabled="!selectedFile" @click="runValidate">{{
        t('dashboardPage.imports.validate')
      }}</ElButton>
      <ElButton type="primary" :disabled="!selectedFile" @click="runImport">{{
        t('dashboardPage.imports.import')
      }}</ElButton>
    </section>

    <ElCard class="panel art-table-card" shadow="never">
      <div class="panel-header">
        <h3>{{ t('dashboardPage.imports.history') }}</h3>
        <span>{{ t('dashboardPage.imports.count', { count: items.length }) }}</span>
      </div>
      <p v-if="statusMessage" class="status">{{ statusMessage }}</p>
      <ElTable :data="items" size="large" v-loading="loading">
        <ElTableColumn prop="id" :label="t('dashboardPage.columns.import')" min-width="190" />
        <ElTableColumn
          prop="schema_version"
          :label="t('dashboardPage.columns.version')"
          width="140"
        />
        <ElTableColumn prop="event_count" :label="t('dashboardPage.columns.events')" width="100" />
        <ElTableColumn
          prop="identity_count"
          :label="t('dashboardPage.columns.identities')"
          width="110"
        />
        <ElTableColumn
          prop="imported_at"
          :label="t('dashboardPage.columns.importedAt')"
          min-width="180"
        />
        <ElTableColumn
          prop="source_path"
          :label="t('dashboardPage.columns.source')"
          min-width="220"
        />
        <ElTableColumn :label="t('dashboardPage.columns.action')" width="120">
          <template #default="{ row }">
            <ElButton link type="primary" @click="download(row.id)">{{
              t('dashboardPage.imports.export')
            }}</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import type { UploadFile } from 'element-plus'
  import { downloadImport, fetchImports, uploadImport, validateImport } from '@/api/commory'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'Imports' })

  const { t } = useI18n()
  const loading = ref(false)
  const items = ref<Api.Commory.ImportSummary[]>([])
  const selectedFile = ref<File | null>(null)
  const statusMessage = ref('')

  const load = async () => {
    loading.value = true
    try {
      items.value = await fetchImports()
    } finally {
      loading.value = false
    }
  }

  const handleFileChange = (uploadFile: UploadFile) => {
    selectedFile.value = uploadFile.raw || null
    statusMessage.value = ''
  }

  const runValidate = async () => {
    if (!selectedFile.value) return
    const result = await validateImport(selectedFile.value)
    statusMessage.value = result.valid
      ? t('dashboardPage.imports.validationPassed')
      : t('dashboardPage.imports.validationFailed')
  }

  const runImport = async () => {
    if (!selectedFile.value) return
    const result = await uploadImport(selectedFile.value)
    statusMessage.value = t('dashboardPage.imports.imported', {
      id: result.import_id,
      version: result.msglayer_version
    })
    await load()
  }

  const download = async (importId: string) => {
    await downloadImport(importId)
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
    grid-template-columns: auto minmax(0, 1fr) auto auto;
    gap: 12px;
    align-items: center;
    padding: 16px;
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .file-name,
  .status,
  .panel-header span {
    color: var(--art-gray-600);
  }

  .status {
    margin: 12px 0 0;
  }

  @media (max-width: 960px) {
    .toolbar {
      grid-template-columns: 1fr;
    }
  }
</style>
