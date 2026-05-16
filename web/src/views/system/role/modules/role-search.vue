<template>
  <ArtSearchBar
    ref="searchBarRef"
    v-model="formData"
    :items="formItems"
    :rules="rules"
    @reset="handleReset"
    @search="handleSearch"
  >
  </ArtSearchBar>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'

  const { t } = useI18n()

  type RoleSearchFormParams = Api.SystemManage.RoleSearchParams & {
    daterange?: string[]
  }

  interface Props {
    modelValue: RoleSearchFormParams
  }

  interface Emits {
    (e: 'update:modelValue', value: RoleSearchFormParams): void
    (e: 'search', params: RoleSearchFormParams): void
    (e: 'reset'): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  const searchBarRef = ref()

  /**
   * 表单数据双向绑定
   */
  const formData = computed({
    get: () => props.modelValue,
    set: (val) => emit('update:modelValue', val)
  })

  /**
   * 表单校验规则
   */
  const rules = {}

  /**
   * 角色状态选项
   */
  const statusOptions = ref([
    { label: t('common.enabled'), value: true },
    { label: t('common.disabled'), value: false }
  ])

  /**
   * 搜索表单配置项
   */
  const formItems = computed(() => [
    {
      label: t('systemPage.role.search.name'),
      key: 'roleName',
      type: 'input',
      placeholder: t('systemPage.role.search.namePlaceholder'),
      clearable: true
    },
    {
      label: t('systemPage.role.search.code'),
      key: 'roleCode',
      type: 'input',
      placeholder: t('systemPage.role.search.codePlaceholder'),
      clearable: true
    },
    {
      label: t('systemPage.role.search.description'),
      key: 'description',
      type: 'input',
      placeholder: t('systemPage.role.search.descriptionPlaceholder'),
      clearable: true
    },
    {
      label: t('systemPage.role.search.status'),
      key: 'enabled',
      type: 'select',
      props: {
        placeholder: t('systemPage.role.search.statusPlaceholder'),
        options: statusOptions.value,
        clearable: true
      }
    },
    {
      label: t('systemPage.role.search.date'),
      key: 'daterange',
      type: 'datetime',
      props: {
        style: { width: '100%' },
        placeholder: t('systemPage.role.search.datePlaceholder'),
        type: 'daterange',
        rangeSeparator: t('systemPage.role.search.rangeSeparator'),
        startPlaceholder: t('systemPage.role.search.startDate'),
        endPlaceholder: t('systemPage.role.search.endDate'),
        valueFormat: 'YYYY-MM-DD',
        shortcuts: [
          { text: t('systemPage.role.search.today'), value: [new Date(), new Date()] },
          {
            text: t('systemPage.role.search.week'),
            value: [new Date(Date.now() - 604800000), new Date()]
          },
          {
            text: t('systemPage.role.search.month'),
            value: [new Date(Date.now() - 2592000000), new Date()]
          }
        ]
      }
    }
  ])

  /**
   * 处理重置事件
   */
  const handleReset = () => {
    emit('reset')
  }

  /**
   * 处理搜索事件
   * 验证表单后触发搜索
   */
  const handleSearch = async (params: RoleSearchFormParams) => {
    await searchBarRef.value.validate()
    emit('search', params)
  }
</script>
