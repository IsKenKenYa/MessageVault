<template>
  <div class="setup-container">
    <div class="setup-card">
      <div class="setup-header">
        <h2>Commory 系统初始化</h2>
        <p>欢迎使用，请完成以下设置以开始使用系统</p>
      </div>

      <div class="setup-steps">
        <div
          v-for="(step, index) in steps"
          :key="index"
          class="step-item"
          :class="{ active: currentStep === index, done: currentStep > index }"
        >
          <div class="step-number">{{ currentStep > index ? '✓' : index + 1 }}</div>
          <span class="step-label">{{ step }}</span>
        </div>
      </div>

      <div class="setup-content">
        <div v-if="currentStep === 0">
          <h3>数据库检查</h3>
          <p class="step-desc">验证数据库连接状态</p>
          <div class="db-status">
            <ElIcon class="status-icon success"><CircleCheckFilled /></ElIcon>
            <div>
              <p class="db-type">SQLite 数据库</p>
              <p class="db-hint">轻量级文件数据库，适合个人使用和小规模部署</p>
            </div>
          </div>
        </div>

        <div v-if="currentStep === 1">
          <h3>管理员账号</h3>
          <p class="step-desc">设置管理员登录信息</p>
          <div v-if="setupStatus?.root_init" class="admin-exists">
            <ElAlert title="管理员账号已存在" description="系统中已有管理员用户，可直接进入下一步" type="info" :closable="false" show-icon />
          </div>
          <ElForm v-else ref="adminFormRef" :model="adminForm" :rules="adminRules" label-position="top">
            <ElFormItem label="用户名" prop="userName">
              <ElInput v-model="adminForm.userName" placeholder="请输入管理员用户名" />
            </ElFormItem>
            <ElFormItem label="密码" prop="password">
              <ElInput v-model="adminForm.password" type="password" placeholder="至少 8 个字符" show-password />
            </ElFormItem>
            <ElFormItem label="确认密码" prop="confirmPassword">
              <ElInput v-model="adminForm.confirmPassword" type="password" placeholder="再次输入密码" show-password />
            </ElFormItem>
          </ElForm>
        </div>

        <div v-if="currentStep === 2">
          <h3>使用模式</h3>
          <p class="step-desc">选择系统运行模式</p>
          <ElRadioGroup v-model="usageMode" class="mode-group">
            <div class="mode-option" :class="{ selected: usageMode === 'personal' }" @click="usageMode = 'personal'">
              <ElRadio value="personal">个人模式</ElRadio>
              <p class="mode-desc">适合个人使用，简化界面和功能</p>
            </div>
            <div class="mode-option" :class="{ selected: usageMode === 'family' }" @click="usageMode = 'family'">
              <ElRadio value="family">家庭模式</ElRadio>
              <p class="mode-desc">适合家庭成员共享，支持多用户管理</p>
            </div>
          </ElRadioGroup>
        </div>

        <div v-if="currentStep === 3">
          <h3>完成初始化</h3>
          <p class="step-desc">确认设置并完成初始化</p>
          <div class="summary">
            <div class="summary-item">
              <span class="summary-label">数据库</span>
              <span class="summary-value">SQLite</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">管理员</span>
              <span class="summary-value">{{ setupStatus?.root_init ? '已存在' : adminForm.userName || '未设置' }}</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">使用模式</span>
              <span class="summary-value">{{ usageMode === 'personal' ? '个人模式' : '家庭模式' }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="setup-actions">
        <ElButton v-if="currentStep > 0" @click="currentStep--">上一步</ElButton>
        <ElButton v-if="currentStep < 3" type="primary" @click="nextStep">下一步</ElButton>
        <ElButton v-if="currentStep === 3" type="primary" :loading="submitting" @click="handleInitialize">完成初始化</ElButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { CircleCheckFilled } from '@element-plus/icons-vue'
import { fetchSetupStatus, postSetup } from '@/api/setup'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'

defineOptions({ name: 'SetupWizard' })

const router = useRouter()
const currentStep = ref(0)
const submitting = ref(false)
const setupStatus = ref<Api.Setup.SetupStatus>()

const steps = ['数据库检查', '管理员账号', '使用模式', '完成初始化']

const adminFormRef = ref<FormInstance>()
const adminForm = reactive({
  userName: '',
  password: '',
  confirmPassword: ''
})

const adminRules = computed<FormRules>(() => ({
  userName: [{ required: true, message: '请输入管理员用户名', trigger: 'blur' }],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, message: '密码至少 8 个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (_rule: any, value: string, callback: (err?: Error) => void) => {
        if (value !== adminForm.password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}))

const usageMode = ref('personal')

const loadStatus = async () => {
  try {
    setupStatus.value = await fetchSetupStatus()
    if (setupStatus.value?.status) {
      router.replace('/')
    }
  } catch {
    // API not available, allow setup to proceed
  }
}

const nextStep = async () => {
  if (currentStep.value === 1 && !setupStatus.value?.root_init) {
    if (!adminFormRef.value) return
    const valid = await adminFormRef.value.validate().catch(() => false)
    if (!valid) return
  }
  currentStep.value++
}

const handleInitialize = async () => {
  submitting.value = true
  try {
    await postSetup({
      userName: adminForm.userName,
      password: adminForm.password,
      confirmPassword: adminForm.confirmPassword,
      usageMode: usageMode.value
    })
    ElMessage.success('系统初始化成功')
    setTimeout(() => {
      router.replace('/auth/login')
    }, 1200)
  } catch (error) {
    // Error handled by http interceptor
  } finally {
    submitting.value = false
  }
}

onMounted(loadStatus)
</script>

<style scoped lang="scss">
.setup-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: var(--art-main-bg-color);
  padding: 20px;
}

.setup-card {
  width: 100%;
  max-width: 600px;
  background: var(--default-box-color);
  border: 1px solid var(--art-border-color);
  border-radius: 12px;
  padding: 40px;
}

.setup-header {
  text-align: center;
  margin-bottom: 32px;

  h2 {
    margin: 0 0 8px;
    font-size: 24px;
  }

  p {
    color: var(--art-gray-600);
    margin: 0;
  }
}

.setup-steps {
  display: flex;
  justify-content: space-between;
  margin-bottom: 32px;
  padding: 0 20px;
}

.step-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  opacity: 0.5;

  &.active,
  &.done {
    opacity: 1;
  }

  .step-number {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    font-weight: 600;
    background: var(--art-gray-200);
    color: var(--art-gray-600);
  }

  &.active .step-number {
    background: var(--main-color);
    color: #fff;
  }

  &.done .step-number {
    background: #67c23a;
    color: #fff;
  }

  .step-label {
    font-size: 12px;
    color: var(--art-gray-600);
  }
}

.setup-content {
  min-height: 200px;
  margin-bottom: 24px;

  h3 {
    margin: 0 0 8px;
    font-size: 18px;
  }

  .step-desc {
    color: var(--art-gray-600);
    margin: 0 0 20px;
  }
}

.db-status {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px;
  background: var(--art-gray-100);
  border-radius: 8px;

  .status-icon {
    font-size: 24px;

    &.success {
      color: #67c23a;
    }
  }

  .db-type {
    font-weight: 600;
    margin: 0 0 4px;
  }

  .db-hint {
    color: var(--art-gray-600);
    font-size: 13px;
    margin: 0;
  }
}

.admin-exists {
  margin-top: 12px;
}

.mode-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

.mode-option {
  padding: 16px;
  border: 2px solid var(--art-border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: border-color 0.2s;

  &.selected {
    border-color: var(--main-color);
  }

  .mode-desc {
    color: var(--art-gray-600);
    font-size: 13px;
    margin: 4px 0 0;
  }
}

.summary {
  display: grid;
  gap: 12px;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--art-gray-100);
  border-radius: 6px;

  .summary-label {
    color: var(--art-gray-600);
  }

  .summary-value {
    font-weight: 600;
  }
}

.setup-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid var(--art-border-color);
}
</style>
