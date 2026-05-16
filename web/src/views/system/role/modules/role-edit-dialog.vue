<template>
  <ElDialog
    v-model="visible"
    :title="
      dialogType === 'add'
        ? t('systemPage.role.dialog.addTitle')
        : t('systemPage.role.dialog.editTitle')
    "
    width="30%"
    align-center
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="120px">
      <ElFormItem :label="t('systemPage.role.search.name')" prop="roleName">
        <ElInput
          v-model="form.roleName"
          :placeholder="t('systemPage.role.search.namePlaceholder')"
        />
      </ElFormItem>
      <ElFormItem :label="t('systemPage.role.search.code')" prop="roleCode">
        <ElInput
          v-model="form.roleCode"
          :placeholder="t('systemPage.role.search.codePlaceholder')"
        />
      </ElFormItem>
      <ElFormItem :label="t('systemPage.role.search.description')" prop="description">
        <ElInput
          v-model="form.description"
          type="textarea"
          :rows="3"
          :placeholder="t('systemPage.role.search.descriptionPlaceholder')"
        />
      </ElFormItem>
      <ElFormItem :label="t('systemPage.role.dialog.enabled')">
        <ElSwitch v-model="form.enabled" />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="handleClose">{{ t('common.cancel') }}</ElButton>
      <ElButton type="primary" @click="handleSubmit">{{ t('common.submit') }}</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { useI18n } from 'vue-i18n'

  const { t } = useI18n()

  type RoleListItem = Api.SystemManage.RoleListItem

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit'
    roleData?: RoleListItem
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    roleData: undefined
  })

  const emit = defineEmits<Emits>()

  const formRef = ref<FormInstance>()

  /**
   * 弹窗显示状态双向绑定
   */
  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  /**
   * 表单验证规则
   */
  const rules = reactive<FormRules>({
    roleName: [
      { required: true, message: t('systemPage.role.search.namePlaceholder'), trigger: 'blur' },
      { min: 2, max: 20, message: t('userCenter.validation.length'), trigger: 'blur' }
    ],
    roleCode: [
      { required: true, message: t('systemPage.role.search.codePlaceholder'), trigger: 'blur' },
      { min: 2, max: 50, message: t('userCenter.validation.length'), trigger: 'blur' }
    ],
    description: [
      {
        required: true,
        message: t('systemPage.role.search.descriptionPlaceholder'),
        trigger: 'blur'
      }
    ]
  })

  /**
   * 表单数据
   */
  const form = reactive<RoleListItem>({
    roleId: 0,
    roleName: '',
    roleCode: '',
    description: '',
    createTime: '',
    enabled: true
  })

  /**
   * 监听弹窗打开，初始化表单数据
   */
  watch(
    () => props.modelValue,
    (newVal) => {
      if (newVal) initForm()
    }
  )

  /**
   * 监听角色数据变化，更新表单
   */
  watch(
    () => props.roleData,
    (newData) => {
      if (newData && props.modelValue) initForm()
    },
    { deep: true }
  )

  /**
   * 初始化表单数据
   * 根据弹窗类型填充表单或重置表单
   */
  const initForm = () => {
    if (props.dialogType === 'edit' && props.roleData) {
      Object.assign(form, props.roleData)
    } else {
      Object.assign(form, {
        roleId: 0,
        roleName: '',
        roleCode: '',
        description: '',
        createTime: '',
        enabled: true
      })
    }
  }

  /**
   * 关闭弹窗并重置表单
   */
  const handleClose = () => {
    visible.value = false
    formRef.value?.resetFields()
  }

  /**
   * 提交表单
   * 验证通过后调用接口保存数据
   */
  const handleSubmit = async () => {
    if (!formRef.value) return

    try {
      await formRef.value.validate()
      // TODO: 调用新增/编辑接口
      const message =
        props.dialogType === 'add'
          ? t('systemPage.role.dialog.addSuccess')
          : t('systemPage.role.dialog.editSuccess')
      ElMessage.success(message)
      emit('success')
      handleClose()
    } catch (error) {
      console.log('表单验证失败:', error)
    }
  }
</script>
