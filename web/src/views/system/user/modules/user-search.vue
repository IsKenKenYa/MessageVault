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

  interface Props {
    modelValue: Api.SystemManage.UserSearchParams
  }
  interface Emits {
    (e: 'update:modelValue', value: Api.SystemManage.UserSearchParams): void
    (e: 'search', params: Api.SystemManage.UserSearchParams): void
    (e: 'reset'): void
  }
  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  // 表单数据双向绑定
  const searchBarRef = ref()
  const formData = computed({
    get: () => props.modelValue,
    set: (val) => emit('update:modelValue', val)
  })

  // 校验规则
  const rules = {
    // userName: [{ required: true, message: '请输入用户名', trigger: 'blur' }]
  }

  // 动态 options
  const statusOptions = ref<{ label: string; value: string; disabled?: boolean }[]>([])

  // 模拟接口返回状态数据
  function fetchStatusOptions(): Promise<typeof statusOptions.value> {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve([
          { label: t('systemPage.user.status.online'), value: '1' },
          { label: t('systemPage.user.status.offline'), value: '2' },
          { label: t('systemPage.user.status.abnormal'), value: '3' },
          { label: t('systemPage.user.status.cancelled'), value: '4' }
        ])
      }, 1000)
    })
  }

  onMounted(async () => {
    statusOptions.value = await fetchStatusOptions()
  })

  // 表单配置
  const formItems = computed(() => [
    {
      label: t('systemPage.user.search.userName'),
      key: 'userName',
      type: 'input',
      placeholder: t('systemPage.user.search.userNamePlaceholder'),
      clearable: true
    },
    {
      label: t('systemPage.user.search.phone'),
      key: 'userPhone',
      type: 'input',
      props: { placeholder: t('systemPage.user.search.phonePlaceholder'), maxlength: '11' }
    },
    {
      label: t('systemPage.user.search.email'),
      key: 'userEmail',
      type: 'input',
      props: { placeholder: t('systemPage.user.search.emailPlaceholder') }
    },
    {
      label: t('systemPage.user.search.status'),
      key: 'status',
      type: 'select',
      props: {
        placeholder: t('systemPage.user.search.statusPlaceholder'),
        options: statusOptions.value
      }
    },
    {
      label: t('systemPage.user.search.gender'),
      key: 'userGender',
      type: 'radiogroup',
      props: {
        options: [
          { label: t('userCenter.male'), value: '1' },
          { label: t('userCenter.female'), value: '2' }
        ]
      }
    }
  ])

  // 事件
  function handleReset() {
    emit('reset')
  }

  async function handleSearch(params: Api.SystemManage.UserSearchParams) {
    await searchBarRef.value.validate()
    emit('search', params)
  }
</script>
