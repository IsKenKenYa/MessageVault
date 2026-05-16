<template>
  <ElDialog
    :title="dialogTitle"
    :model-value="visible"
    @update:model-value="handleCancel"
    width="860px"
    align-center
    class="menu-dialog"
    @closed="handleClosed"
  >
    <ArtForm
      ref="formRef"
      v-model="form"
      :items="formItems"
      :rules="rules"
      :span="width > 640 ? 12 : 24"
      :gutter="20"
      label-width="100px"
      :show-reset="false"
      :show-submit="false"
    >
      <template #menuType>
        <ElRadioGroup v-model="form.menuType" :disabled="disableMenuType">
          <ElRadioButton value="menu" label="menu">{{
            t('systemPage.menu.dialog.menu')
          }}</ElRadioButton>
          <ElRadioButton value="button" label="button">{{
            t('systemPage.menu.dialog.button')
          }}</ElRadioButton>
        </ElRadioGroup>
      </template>
    </ArtForm>

    <template #footer>
      <span class="dialog-footer">
        <ElButton @click="handleCancel">{{ t('common.cancel') }}</ElButton>
        <ElButton type="primary" @click="handleSubmit">{{ t('common.confirm') }}</ElButton>
      </span>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormRules } from 'element-plus'
  import { ElIcon, ElTooltip } from 'element-plus'
  import { QuestionFilled } from '@element-plus/icons-vue'
  import { formatMenuTitle } from '@/utils/router'
  import type { AppRouteRecord } from '@/types/router'
  import type { FormItem } from '@/components/core/forms/art-form/index.vue'
  import ArtForm from '@/components/core/forms/art-form/index.vue'
  import { useWindowSize } from '@vueuse/core'
  import { useI18n } from 'vue-i18n'

  const { width } = useWindowSize()
  const { t } = useI18n()

  /**
   * 创建带 tooltip 的表单标签
   * @param label 标签文本
   * @param tooltip 提示文本
   * @returns 渲染函数
   */
  const createLabelTooltip = (label: string, tooltip: string) => {
    return () =>
      h('span', { class: 'flex items-center' }, [
        h('span', label),
        h(
          ElTooltip,
          {
            content: tooltip,
            placement: 'top'
          },
          () => h(ElIcon, { class: 'ml-0.5 cursor-help' }, () => h(QuestionFilled))
        )
      ])
  }

  interface MenuFormData {
    id: number
    name: string
    path: string
    label: string
    component: string
    icon: string
    isEnable: boolean
    sort: number
    isMenu: boolean
    keepAlive: boolean
    isHide: boolean
    isHideTab: boolean
    link: string
    isIframe: boolean
    showBadge: boolean
    showTextBadge: string
    fixedTab: boolean
    activePath: string
    roles: string[]
    isFullPage: boolean
    authName: string
    authLabel: string
    authIcon: string
    authSort: number
  }

  interface Props {
    visible: boolean
    editData?: AppRouteRecord | any
    type?: 'menu' | 'button'
    lockType?: boolean
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'submit', data: MenuFormData): void
  }

  const props = withDefaults(defineProps<Props>(), {
    visible: false,
    type: 'menu',
    lockType: false
  })

  const emit = defineEmits<Emits>()

  const formRef = ref()
  const isEdit = ref(false)

  const form = reactive<MenuFormData & { menuType: 'menu' | 'button' }>({
    menuType: 'menu',
    id: 0,
    name: '',
    path: '',
    label: '',
    component: '',
    icon: '',
    isEnable: true,
    sort: 1,
    isMenu: true,
    keepAlive: true,
    isHide: false,
    isHideTab: false,
    link: '',
    isIframe: false,
    showBadge: false,
    showTextBadge: '',
    fixedTab: false,
    activePath: '',
    roles: [],
    isFullPage: false,
    authName: '',
    authLabel: '',
    authIcon: '',
    authSort: 1
  })

  const rules = reactive<FormRules>({
    name: [
      { required: true, message: t('systemPage.menu.dialog.namePlaceholder'), trigger: 'blur' },
      { min: 2, max: 20, message: t('userCenter.validation.length'), trigger: 'blur' }
    ],
    path: [
      { required: true, message: t('systemPage.menu.dialog.pathPlaceholder'), trigger: 'blur' }
    ],
    label: [
      { required: true, message: t('systemPage.menu.dialog.labelPlaceholder'), trigger: 'blur' }
    ],
    authName: [
      { required: true, message: t('systemPage.menu.dialog.authNamePlaceholder'), trigger: 'blur' }
    ],
    authLabel: [
      { required: true, message: t('systemPage.menu.dialog.authLabelPlaceholder'), trigger: 'blur' }
    ]
  })

  /**
   * 表单项配置
   */
  const formItems = computed<FormItem[]>(() => {
    const baseItems: FormItem[] = [
      { label: t('systemPage.menu.dialog.menuType'), key: 'menuType', span: 24 }
    ]

    // Switch 组件的 span：小屏幕 12，大屏幕 6
    const switchSpan = width.value < 640 ? 12 : 6

    if (form.menuType === 'menu') {
      return [
        ...baseItems,
        {
          label: t('systemPage.menu.dialog.name'),
          key: 'name',
          type: 'input',
          props: { placeholder: t('systemPage.menu.dialog.namePlaceholder') }
        },
        {
          label: createLabelTooltip(
            t('systemPage.menu.dialog.path'),
            t('systemPage.menu.dialog.pathTip')
          ),
          key: 'path',
          type: 'input',
          props: { placeholder: t('systemPage.menu.dialog.pathPlaceholder') }
        },
        {
          label: t('systemPage.menu.dialog.label'),
          key: 'label',
          type: 'input',
          props: { placeholder: t('systemPage.menu.dialog.labelPlaceholder') }
        },
        {
          label: createLabelTooltip(
            t('systemPage.menu.dialog.component'),
            t('systemPage.menu.dialog.componentTip')
          ),
          key: 'component',
          type: 'input',
          props: { placeholder: t('systemPage.menu.dialog.componentPlaceholder') }
        },
        {
          label: t('systemPage.menu.dialog.icon'),
          key: 'icon',
          type: 'input',
          props: { placeholder: t('systemPage.menu.dialog.iconPlaceholder') }
        },
        {
          label: createLabelTooltip(
            t('systemPage.menu.dialog.roles'),
            t('systemPage.menu.dialog.rolesTip')
          ),
          key: 'roles',
          type: 'inputtag',
          props: { placeholder: t('systemPage.menu.dialog.rolesPlaceholder') }
        },
        {
          label: t('systemPage.menu.dialog.sort'),
          key: 'sort',
          type: 'number',
          props: { min: 1, controlsPosition: 'right', style: { width: '100%' } }
        },
        {
          label: t('systemPage.menu.dialog.link'),
          key: 'link',
          type: 'input',
          props: { placeholder: t('systemPage.menu.dialog.linkPlaceholder') }
        },
        {
          label: t('systemPage.menu.dialog.badge'),
          key: 'showTextBadge',
          type: 'input',
          props: { placeholder: t('systemPage.menu.dialog.badgePlaceholder') }
        },
        {
          label: createLabelTooltip(
            t('systemPage.menu.dialog.activePath'),
            t('systemPage.menu.dialog.activePathTip')
          ),
          key: 'activePath',
          type: 'input',
          props: { placeholder: t('systemPage.menu.dialog.activePathPlaceholder') }
        },
        {
          label: t('systemPage.menu.dialog.isEnable'),
          key: 'isEnable',
          type: 'switch',
          span: switchSpan
        },
        {
          label: t('systemPage.menu.dialog.keepAlive'),
          key: 'keepAlive',
          type: 'switch',
          span: switchSpan
        },
        {
          label: t('systemPage.menu.dialog.isHide'),
          key: 'isHide',
          type: 'switch',
          span: switchSpan
        },
        {
          label: t('systemPage.menu.dialog.isIframe'),
          key: 'isIframe',
          type: 'switch',
          span: switchSpan
        },
        {
          label: t('systemPage.menu.dialog.showBadge'),
          key: 'showBadge',
          type: 'switch',
          span: switchSpan
        },
        {
          label: t('systemPage.menu.dialog.fixedTab'),
          key: 'fixedTab',
          type: 'switch',
          span: switchSpan
        },
        {
          label: t('systemPage.menu.dialog.isHideTab'),
          key: 'isHideTab',
          type: 'switch',
          span: switchSpan
        },
        {
          label: t('systemPage.menu.dialog.isFullPage'),
          key: 'isFullPage',
          type: 'switch',
          span: switchSpan
        }
      ]
    } else {
      return [
        ...baseItems,
        {
          label: t('systemPage.menu.dialog.authName'),
          key: 'authName',
          type: 'input',
          props: { placeholder: t('systemPage.menu.dialog.authNamePlaceholder') }
        },
        {
          label: t('systemPage.menu.dialog.authLabel'),
          key: 'authLabel',
          type: 'input',
          props: { placeholder: t('systemPage.menu.dialog.authLabelPlaceholder') }
        },
        {
          label: t('systemPage.menu.dialog.authSort'),
          key: 'authSort',
          type: 'number',
          props: { min: 1, controlsPosition: 'right', style: { width: '100%' } }
        }
      ]
    }
  })

  const dialogTitle = computed(() => {
    if (form.menuType === 'menu') {
      return isEdit.value
        ? t('systemPage.menu.dialog.editMenu')
        : t('systemPage.menu.dialog.newMenu')
    }
    return isEdit.value
      ? t('systemPage.menu.dialog.editButton')
      : t('systemPage.menu.dialog.newButton')
  })

  /**
   * 是否禁用菜单类型切换
   */
  const disableMenuType = computed(() => {
    if (isEdit.value) return true
    if (!isEdit.value && form.menuType === 'menu' && props.lockType) return true
    return false
  })

  /**
   * 重置表单数据
   */
  const resetForm = (): void => {
    formRef.value?.reset()
    form.menuType = 'menu'
  }

  /**
   * 加载表单数据（编辑模式）
   */
  const loadFormData = (): void => {
    if (!props.editData) return

    isEdit.value = true

    if (form.menuType === 'menu') {
      const row = props.editData
      form.id = row.id || 0
      form.name = formatMenuTitle(row.meta?.title || '')
      form.path = row.path || ''
      form.label = row.name || ''
      form.component = row.component || ''
      form.icon = row.meta?.icon || ''
      form.sort = row.meta?.sort || 1
      form.isMenu = row.meta?.isMenu ?? true
      form.keepAlive = row.meta?.keepAlive ?? false
      form.isHide = row.meta?.isHide ?? false
      form.isHideTab = row.meta?.isHideTab ?? false
      form.isEnable = row.meta?.isEnable ?? true
      form.link = row.meta?.link || ''
      form.isIframe = row.meta?.isIframe ?? false
      form.showBadge = row.meta?.showBadge ?? false
      form.showTextBadge = row.meta?.showTextBadge || ''
      form.fixedTab = row.meta?.fixedTab ?? false
      form.activePath = row.meta?.activePath || ''
      form.roles = row.meta?.roles || []
      form.isFullPage = row.meta?.isFullPage ?? false
    } else {
      const row = props.editData
      form.authName = row.title || ''
      form.authLabel = row.authMark || ''
      form.authIcon = row.icon || ''
      form.authSort = row.sort || 1
    }
  }

  /**
   * 提交表单
   */
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    try {
      await formRef.value.validate()
      emit('submit', { ...form })
      ElMessage.success(
        t('systemPage.menu.dialog.submitSuccess', {
          action: isEdit.value ? t('common.edit') : t('common.add')
        })
      )
      handleCancel()
    } catch {
      ElMessage.error(t('systemPage.menu.dialog.validateFailed'))
    }
  }

  /**
   * 取消操作
   */
  const handleCancel = (): void => {
    emit('update:visible', false)
  }

  /**
   * 对话框关闭后的回调
   */
  const handleClosed = (): void => {
    resetForm()
    isEdit.value = false
  }

  /**
   * 监听对话框显示状态
   */
  watch(
    () => props.visible,
    (newVal) => {
      if (newVal) {
        form.menuType = props.type
        nextTick(() => {
          if (props.editData) {
            loadFormData()
          }
        })
      }
    }
  )

  /**
   * 监听菜单类型变化
   */
  watch(
    () => props.type,
    (newType) => {
      if (props.visible) {
        form.menuType = newType
      }
    }
  )
</script>
