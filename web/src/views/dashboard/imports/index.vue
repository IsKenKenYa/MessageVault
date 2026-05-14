<template>
  <div class="page-grid">
    <section class="toolbar">
      <ElUpload :auto-upload="false" :show-file-list="false" :on-change="handleFileChange" accept=".json">
        <ElButton>Choose MsgLayer File</ElButton>
      </ElUpload>
      <span class="file-name">{{ selectedFile?.name || 'No file selected' }}</span>
      <ElButton :disabled="!selectedFile" @click="runValidate">Validate</ElButton>
      <ElButton type="primary" :disabled="!selectedFile" @click="runImport">Import</ElButton>
    </section>

    <section class="panel">
      <div class="panel-header">
        <h3>Import history</h3>
        <span>{{ items.length }} imports</span>
      </div>
      <p v-if="statusMessage" class="status">{{ statusMessage }}</p>
      <ElTable :data="items" size="large" v-loading="loading">
        <ElTableColumn prop="id" label="Import" min-width="190" />
        <ElTableColumn prop="schema_version" label="Version" width="140" />
        <ElTableColumn prop="event_count" label="Events" width="100" />
        <ElTableColumn prop="identity_count" label="Identities" width="110" />
        <ElTableColumn prop="imported_at" label="Imported At" min-width="180" />
        <ElTableColumn prop="source_path" label="Source" min-width="220" />
        <ElTableColumn label="Action" width="120">
          <template #default="{ row }">
            <ElButton link type="primary" @click="download(row.id)">Export</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </section>
  </div>
</template>

<script setup lang="ts">
  import type { UploadFile } from 'element-plus'
  import { downloadImport, fetchImports, uploadImport, validateImport } from '@/api/commory'

  defineOptions({ name: 'Imports' })

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
    statusMessage.value = result.valid ? 'Validation passed.' : 'Validation failed.'
  }

  const runImport = async () => {
    if (!selectedFile.value) return
    const result = await uploadImport(selectedFile.value)
    statusMessage.value = `Imported ${result.import_id} (${result.msglayer_version}).`
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

  .panel {
    padding: 18px;
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
