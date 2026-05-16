<template>
  <ElDialog
    v-model="dialogVisible"
    :title="
      dialogType === 'add'
        ? t('systemPage.user.dialog.addTitle')
        : t('systemPage.user.dialog.editTitle')
    "
    width="30%"
    align-center
  >
    <ElForm ref="formRef" :model="formData" :rules="rules" label-width="80px">
      <ElFormItem :label="t('systemPage.user.search.userName')" prop="username">
        <ElInput
          v-model="formData.username"
          :placeholder="t('systemPage.user.search.userNamePlaceholder')"
        />
      </ElFormItem>
      <ElFormItem :label="t('systemPage.user.search.phone')" prop="phone">
        <ElInput
          v-model="formData.phone"
          :placeholder="t('systemPage.user.search.phonePlaceholder')"
        />
      </ElFormItem>
      <ElFormItem :label="t('systemPage.user.search.gender')" prop="gender">
        <ElSelect v-model="formData.gender">
          <ElOption :label="t('userCenter.male')" value="男" />
          <ElOption :label="t('userCenter.female')" value="女" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem :label="t('systemPage.user.dialog.role')" prop="role">
        <ElSelect v-model="formData.role" multiple>
          <ElOption
            v-for="role in roleList"
            :key="role.roleCode"
            :value="role.roleCode"
            :label="role.roleName"
          />
        </ElSelect>
      </ElFormItem>
    </ElForm>
    <template #footer>
      <div class="dialog-footer">
        <ElButton @click="dialogVisible = false">{{ t('common.cancel') }}</ElButton>
        <ElButton type="primary" @click="handleSubmit">{{ t('common.submit') }}</ElButton>
      </div>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ROLE_LIST_DATA } from '@/mock/temp/formData'
  import type { FormInstance, FormRules } from 'element-plus'
  import { useI18n } from 'vue-i18n'

  const { t } = useI18n()

  interface Props {
    visible: boolean
    type: string
    userData?: Partial<Api.SystemManage.UserListItem>
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'submit'): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  // 角色列表数据
  const roleList = ref(ROLE_LIST_DATA)

  // 对话框显示控制
  const dialogVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value)
  })

  const dialogType = computed(() => props.type)

  // 表单实例
  const formRef = ref<FormInstance>()

  // 表单数据
  const formData = reactive({
    username: '',
    phone: '',
    gender: '男',
    role: [] as string[]
  })

  // 表单验证规则
  const rules: FormRules = {
    username: [
      { required: true, message: t('systemPage.user.search.userNamePlaceholder'), trigger: 'blur' },
      { min: 2, max: 20, message: t('userCenter.validation.length'), trigger: 'blur' }
    ],
    phone: [
      { required: true, message: t('systemPage.user.search.phonePlaceholder'), trigger: 'blur' },
      {
        pattern: /^1[3-9]\d{9}$/,
        message: t('systemPage.user.dialog.phoneInvalid'),
        trigger: 'blur'
      }
    ],
    gender: [{ required: true, message: t('userCenter.validation.gender'), trigger: 'blur' }],
    role: [{ required: true, message: t('systemPage.user.dialog.role'), trigger: 'blur' }]
  }

  /**
   * 初始化表单数据
   * 根据对话框类型（新增/编辑）填充表单
   */
  const initFormData = () => {
    const isEdit = props.type === 'edit' && props.userData
    const row = props.userData

    Object.assign(formData, {
      username: isEdit && row ? row.userName || '' : '',
      phone: isEdit && row ? row.userPhone || '' : '',
      gender: isEdit && row ? row.userGender || '男' : '男',
      role: isEdit && row ? (Array.isArray(row.userRoles) ? row.userRoles : []) : []
    })
  }

  /**
   * 监听对话框状态变化
   * 当对话框打开时初始化表单数据并清除验证状态
   */
  watch(
    () => [props.visible, props.type, props.userData],
    ([visible]) => {
      if (visible) {
        initFormData()
        nextTick(() => {
          formRef.value?.clearValidate()
        })
      }
    },
    { immediate: true }
  )

  /**
   * 提交表单
   * 验证通过后触发提交事件
   */
  const handleSubmit = async () => {
    if (!formRef.value) return

    await formRef.value.validate((valid) => {
      if (valid) {
        ElMessage.success(
          dialogType.value === 'add'
            ? t('systemPage.user.dialog.addSuccess')
            : t('systemPage.user.dialog.updateSuccess')
        )
        dialogVisible.value = false
        emit('submit')
      }
    })
  }
</script>
