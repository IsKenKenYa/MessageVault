<template>
  <div class="setup-page">
    <section class="setup-brand">
      <p class="setup-eyebrow">{{ t('setup.eyebrow') }}</p>
      <h1>{{ t('setup.title') }}</h1>
      <p>{{ t('setup.subtitle') }}</p>

      <div class="brand-status">
        <div class="brand-status__icon">
          <ElIcon><CircleCheckFilled /></ElIcon>
        </div>
        <div>
          <strong>{{ t('setup.database.ready') }}</strong>
          <span>{{ t('setup.database.hint') }}</span>
        </div>
      </div>
    </section>

    <ElCard class="setup-card" shadow="never" v-loading="loadingStatus">
      <template #header>
        <div class="card-header">
          <div>
            <h2>{{ t('setup.panelTitle') }}</h2>
            <p>{{ t('setup.panelSubtitle') }}</p>
          </div>
          <ElTag type="success" effect="light">Commory</ElTag>
        </div>
      </template>

      <ElAlert
        v-if="statusWarning"
        class="setup-alert"
        :title="t('setup.apiUnavailable')"
        type="warning"
        :closable="false"
        show-icon
      />

      <ElSteps :active="currentStep" finish-status="success" align-center>
        <ElStep v-for="step in steps" :key="step" :title="step" />
      </ElSteps>

      <div class="setup-content">
        <section v-show="currentStep === 0" class="step-panel">
          <div class="step-heading">
            <ElIcon><CircleCheckFilled /></ElIcon>
            <div>
              <h3>{{ t('setup.database.title') }}</h3>
              <p>{{ t('setup.database.desc') }}</p>
            </div>
          </div>

          <div class="info-row">
            <span>{{ t('setup.finish.database') }}</span>
            <strong>{{ setupStatus?.database_type || 'SQLite' }}</strong>
          </div>
          <p class="muted">{{ t('setup.database.hint') }}</p>
        </section>

        <section v-show="currentStep === 1" class="step-panel">
          <div class="step-heading">
            <ElIcon><UserFilled /></ElIcon>
            <div>
              <h3>{{ t('setup.admin.title') }}</h3>
              <p>{{ t('setup.admin.desc') }}</p>
            </div>
          </div>

          <ElAlert
            v-if="setupStatus?.root_init"
            :title="t('setup.admin.existsTitle')"
            :description="t('setup.admin.existsDesc')"
            type="info"
            :closable="false"
            show-icon
          />

          <ElForm
            v-else
            ref="adminFormRef"
            :model="adminForm"
            :rules="adminRules"
            label-position="top"
          >
            <ElFormItem :label="t('setup.admin.userName')" prop="userName">
              <ElInput
                v-model="adminForm.userName"
                :placeholder="t('setup.admin.userNamePlaceholder')"
              />
            </ElFormItem>
            <ElFormItem :label="t('setup.admin.password')" prop="password">
              <ElInput
                v-model="adminForm.password"
                type="password"
                :placeholder="t('setup.admin.passwordPlaceholder')"
                show-password
              />
            </ElFormItem>
            <ElFormItem :label="t('setup.admin.confirmPassword')" prop="confirmPassword">
              <ElInput
                v-model="adminForm.confirmPassword"
                type="password"
                :placeholder="t('setup.admin.confirmPasswordPlaceholder')"
                show-password
              />
            </ElFormItem>
          </ElForm>
        </section>

        <section v-show="currentStep === 2" class="step-panel">
          <div class="step-heading">
            <ElIcon><Operation /></ElIcon>
            <div>
              <h3>{{ t('setup.mode.title') }}</h3>
              <p>{{ t('setup.mode.desc') }}</p>
            </div>
          </div>

          <ElRadioGroup v-model="usageMode" class="mode-grid">
            <label class="mode-option" :class="{ selected: usageMode === 'personal' }">
              <ElRadio value="personal">{{ t('setup.mode.personal') }}</ElRadio>
              <span>{{ t('setup.mode.personalDesc') }}</span>
            </label>
            <label class="mode-option" :class="{ selected: usageMode === 'family' }">
              <ElRadio value="family">{{ t('setup.mode.family') }}</ElRadio>
              <span>{{ t('setup.mode.familyDesc') }}</span>
            </label>
          </ElRadioGroup>
        </section>

        <section v-show="currentStep === 3" class="step-panel">
          <div class="step-heading">
            <ElIcon><Select /></ElIcon>
            <div>
              <h3>{{ t('setup.finish.title') }}</h3>
              <p>{{ t('setup.finish.desc') }}</p>
            </div>
          </div>

          <div class="summary-list">
            <div class="info-row">
              <span>{{ t('setup.finish.database') }}</span>
              <strong>{{ setupStatus?.database_type || 'SQLite' }}</strong>
            </div>
            <div class="info-row">
              <span>{{ t('setup.finish.admin') }}</span>
              <strong>{{ adminSummary }}</strong>
            </div>
            <div class="info-row">
              <span>{{ t('setup.finish.usageMode') }}</span>
              <strong>{{ usageModeLabel }}</strong>
            </div>
          </div>
        </section>
      </div>

      <div class="setup-actions">
        <ElButton v-if="currentStep > 0" @click="currentStep--">
          {{ t('setup.actions.prev') }}
        </ElButton>
        <ElButton v-if="currentStep < lastStep" type="primary" @click="nextStep">
          {{ t('setup.actions.next') }}
        </ElButton>
        <ElButton v-else type="primary" :loading="submitting" @click="handleInitialize">
          {{ t('setup.actions.finish') }}
        </ElButton>
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { CircleCheckFilled, Operation, Select, UserFilled } from '@element-plus/icons-vue'
  import { fetchSetupStatus, postSetup } from '@/api/setup'
  import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'SetupWizard' })

  const { t } = useI18n()
  const router = useRouter()
  const currentStep = ref(0)
  const loadingStatus = ref(true)
  const submitting = ref(false)
  const statusWarning = ref(false)
  const setupStatus = ref<Api.Setup.SetupStatus>()
  const lastStep = 3

  const steps = computed(() => [
    t('setup.steps.database'),
    t('setup.steps.admin'),
    t('setup.steps.mode'),
    t('setup.steps.finish')
  ])

  const adminFormRef = ref<FormInstance>()
  const adminForm = reactive({
    userName: '',
    password: '',
    confirmPassword: ''
  })

  const adminRules = computed<FormRules>(() => ({
    userName: [
      { required: true, message: t('setup.validation.userNameRequired'), trigger: 'blur' }
    ],
    password: [
      { required: true, message: t('setup.validation.passwordRequired'), trigger: 'blur' },
      { min: 8, message: t('setup.validation.passwordLength'), trigger: 'blur' }
    ],
    confirmPassword: [
      { required: true, message: t('setup.validation.confirmPasswordRequired'), trigger: 'blur' },
      {
        validator: (_rule: unknown, value: string, callback: (err?: Error) => void) => {
          if (value !== adminForm.password) {
            callback(new Error(t('setup.validation.passwordMismatch')))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ]
  }))

  const usageMode = ref<'personal' | 'family'>('personal')
  const usageModeLabel = computed(() =>
    usageMode.value === 'personal' ? t('setup.mode.personal') : t('setup.mode.family')
  )
  const adminSummary = computed(() => {
    if (setupStatus.value?.root_init) return t('setup.admin.existing')
    return adminForm.userName || t('setup.admin.notSet')
  })

  const loadStatus = async () => {
    loadingStatus.value = true
    statusWarning.value = false
    try {
      setupStatus.value = await fetchSetupStatus()
      if (setupStatus.value?.status) {
        router.replace('/')
      }
    } catch {
      statusWarning.value = true
    } finally {
      loadingStatus.value = false
    }
  }

  const nextStep = async () => {
    if (currentStep.value === 1 && !setupStatus.value?.root_init) {
      if (!adminFormRef.value) return
      const valid = await adminFormRef.value.validate().catch(() => false)
      if (!valid) return
    }
    currentStep.value = Math.min(currentStep.value + 1, lastStep)
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
      ElMessage.success(t('setup.message.success'))
      window.setTimeout(() => {
        router.replace('/auth/login')
      }, 1200)
    } catch {
      // Error handled by http interceptor
    } finally {
      submitting.value = false
    }
  }

  onMounted(loadStatus)
</script>

<style scoped lang="scss">
  .setup-page {
    display: grid;
    grid-template-columns: minmax(280px, 0.8fr) minmax(420px, 560px);
    gap: 28px;
    align-items: center;
    min-height: 100vh;
    padding: 40px;
    background: var(--art-main-bg-color);
  }

  .setup-brand {
    display: grid;
    gap: 18px;
    max-width: 560px;

    h1 {
      margin: 0;
      font-size: 38px;
      line-height: 1.16;
      color: var(--art-gray-800);
    }

    p {
      max-width: 460px;
      margin: 0;
      color: var(--art-gray-600);
      font-size: 16px;
      line-height: 1.8;
    }
  }

  .setup-eyebrow {
    color: var(--main-color) !important;
    font-size: 13px !important;
    font-weight: 600;
  }

  .brand-status {
    display: flex;
    gap: 14px;
    align-items: center;
    width: min(100%, 420px);
    padding: 16px;
    margin-top: 12px;
    background: var(--default-box-color);
    border: 1px solid var(--art-border-color);
    border-radius: 8px;

    strong,
    span {
      display: block;
    }

    span {
      margin-top: 3px;
      color: var(--art-gray-600);
      font-size: 13px;
      line-height: 1.5;
    }
  }

  .brand-status__icon,
  .step-heading .el-icon {
    display: inline-flex;
    flex: 0 0 auto;
    align-items: center;
    justify-content: center;
    width: 38px;
    height: 38px;
    color: var(--main-color);
    background: color-mix(in srgb, var(--main-color) 10%, transparent);
    border-radius: 8px;
  }

  .setup-card {
    border: 1px solid var(--art-border-color);
    border-radius: 8px;

    :deep(.el-card__body) {
      padding: 24px;
    }
  }

  .card-header {
    display: flex;
    gap: 16px;
    align-items: flex-start;
    justify-content: space-between;

    h2,
    p {
      margin: 0;
    }

    h2 {
      font-size: 20px;
      font-weight: 600;
    }

    p {
      margin-top: 6px;
      color: var(--art-gray-600);
      font-size: 13px;
    }
  }

  .setup-alert {
    margin-bottom: 18px;
  }

  .setup-content {
    min-height: 310px;
    padding: 26px 0 8px;
  }

  .step-panel {
    display: grid;
    gap: 18px;
  }

  .step-heading {
    display: flex;
    gap: 12px;
    align-items: flex-start;

    h3,
    p {
      margin: 0;
    }

    h3 {
      font-size: 18px;
      font-weight: 600;
    }

    p {
      margin-top: 5px;
      color: var(--art-gray-600);
      font-size: 13px;
    }
  }

  .info-row {
    display: flex;
    gap: 16px;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px;
    background: var(--art-main-bg-color);
    border: 1px solid var(--art-border-color);
    border-radius: 8px;

    span {
      color: var(--art-gray-600);
    }
  }

  .muted {
    margin: 0;
    color: var(--art-gray-600);
    font-size: 13px;
  }

  .mode-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14px;
    width: 100%;
  }

  .mode-option {
    display: grid;
    gap: 8px;
    min-height: 118px;
    padding: 16px;
    cursor: pointer;
    border: 1px solid var(--art-border-color);
    border-radius: 8px;
    transition:
      border-color 0.2s ease,
      background 0.2s ease;

    &.selected {
      background: color-mix(in srgb, var(--main-color) 8%, transparent);
      border-color: var(--main-color);
    }

    span {
      color: var(--art-gray-600);
      font-size: 13px;
      line-height: 1.6;
    }
  }

  .summary-list {
    display: grid;
    gap: 12px;
  }

  .setup-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
    padding-top: 18px;
    border-top: 1px solid var(--art-border-color);
  }

  @media (max-width: 980px) {
    .setup-page {
      grid-template-columns: 1fr;
      align-items: stretch;
      padding: 24px;
    }

    .setup-brand {
      max-width: none;

      h1 {
        font-size: 30px;
      }
    }
  }

  @media (max-width: 640px) {
    .setup-page {
      padding: 14px;
    }

    .mode-grid {
      grid-template-columns: 1fr;
    }

    .setup-actions {
      flex-direction: column-reverse;

      .el-button {
        width: 100%;
        margin-left: 0;
      }
    }
  }
</style>
