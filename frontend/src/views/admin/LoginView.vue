<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { forgotPasswordReset, forgotPasswordSendCode, login, setTokens } from '@/api'
import { useSiteStore } from '@/stores/site'
import { markPageReady } from '@/composables/usePageReady'
import { ElMessage } from 'element-plus'

const router = useRouter()

const form = ref({ username: '', password: '' })
const loading = ref(false)
const errorMsg = ref('')
const siteStore = useSiteStore()

onMounted(async () => {
  markPageReady()
  siteStore.fetchConfig()
})

async function submit() {
  errorMsg.value = ''
  if (!form.value.username.trim() || !form.value.password) {
    errorMsg.value = '请输入用户名和密码'
    return
  }
  loading.value = true
  try {
    const res = await login({
      username: form.value.username.trim(),
      password: form.value.password,
    })
    setTokens(res.token, res.refresh_token)
    ElMessage.success('登录成功')
    router.push('/admin/dashboard')
  } catch (e) {
    errorMsg.value = (e as Error).message || '登录失败'
  } finally {
    loading.value = false
  }
}

// ---------- 忘记密码（邮箱验证码两步流程） ----------
const forgotVisible = ref(false)
const forgotStep = ref<'send' | 'reset'>('send')
const forgotEmail = ref('')
const forgotToken = ref('')
const forgotCode = ref('')
const forgotNewPassword = ref('')
const forgotConfirm = ref('')
const forgotSending = ref(false)
const forgotResetting = ref(false)
const forgotError = ref('')
const countdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

function clearCountdown() {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
}

function openForgot() {
  clearCountdown()
  forgotStep.value = 'send'
  forgotEmail.value = ''
  forgotToken.value = ''
  forgotCode.value = ''
  forgotNewPassword.value = ''
  forgotConfirm.value = ''
  forgotError.value = ''
  countdown.value = 0
  forgotVisible.value = true
}

/** 回到发送步骤（保留邮箱），可修改邮箱或重新获取验证码 */
function backToSend() {
  forgotStep.value = 'send'
  forgotCode.value = ''
  forgotNewPassword.value = ''
  forgotConfirm.value = ''
  forgotError.value = ''
}

async function sendCode() {
  forgotError.value = ''
  if (!forgotEmail.value.trim() || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(forgotEmail.value)) {
    forgotError.value = '请输入有效的邮箱地址'
    return
  }
  forgotSending.value = true
  try {
    const res = await forgotPasswordSendCode({ email: forgotEmail.value.trim() })
    // 邮箱未注册/账号已注销时后端也返回成功（防用户枚举），token 为空且不发邮件
    if (res.token) {
      forgotToken.value = res.token
      forgotError.value = ''
      ElMessage.success('验证码已发送，请查收邮件')
      forgotStep.value = 'reset'
    } else {
      // 不泄露注册状态，仅提示用户确认邮箱；停留本步可重试/修改
      ElMessage.warning('未收到验证码？请确认邮箱已注册且未被注销')
    }
    // 发送冷却提示：60 秒内不可重复发送（与后端一致）
    countdown.value = 60
    clearCountdown()
    countdownTimer = setInterval(() => {
      countdown.value -= 1
      if (countdown.value <= 0) clearCountdown()
    }, 1000)
  } catch (e) {
    forgotError.value = (e as Error).message || '发送失败'
  } finally {
    forgotSending.value = false
  }
}

async function submitReset() {
  forgotError.value = ''
  if (!forgotToken.value) {
    forgotError.value = '验证码已失效，请重新获取'
    return
  }
  if (!forgotCode.value.trim()) {
    forgotError.value = '请输入验证码'
    return
  }
  if (forgotNewPassword.value.length < 8) {
    forgotError.value = '新密码长度不能少于 8 位'
    return
  }
  if (forgotNewPassword.value !== forgotConfirm.value) {
    forgotError.value = '两次输入的新密码不一致'
    return
  }
  forgotResetting.value = true
  try {
    await forgotPasswordReset({
      email: forgotEmail.value.trim(),
      token: forgotToken.value,
      code: forgotCode.value.trim(),
      new_password: forgotNewPassword.value,
    })
    ElMessage.success('密码重置成功，请使用新密码登录')
    clearCountdown()
    forgotVisible.value = false
    form.value.username = ''
    form.value.password = ''
  } catch (e) {
    forgotError.value = (e as Error).message || '重置失败'
  } finally {
    forgotResetting.value = false
  }
}

onBeforeUnmount(clearCountdown)
</script>

<template>
  <div class="admin-login">
    <el-card class="admin-login-card">
      <h1>{{ siteStore.title }} 后台登录</h1>
      <el-alert v-if="errorMsg" :title="errorMsg" type="error" show-icon style="margin-bottom: 16px" />
      <el-form label-position="top" @submit.prevent="submit">
        <el-form-item label="用户名">
          <el-input v-model="form.username" autocomplete="username" size="large" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            autocomplete="current-password"
            size="large"
            @keyup.enter="submit"
          />
        </el-form-item>
        <el-button type="primary" size="large" style="width: 100%" :loading="loading" @click="submit">
          登录
        </el-button>
      </el-form>
      <div style="display: flex; justify-content: space-between; margin-top: 16px">
        <a href="/">返回首页</a>
        <a href="javascript:void(0)" @click="openForgot">忘记密码？</a>
      </div>
    </el-card>

    <!-- 忘记密码：邮箱验证码两步流程 -->
    <el-dialog v-model="forgotVisible" title="忘记密码" width="420px" :close-on-click-modal="false"
      @closed="clearCountdown">
      <el-alert v-if="forgotError" :title="forgotError" type="error" show-icon style="margin-bottom: 16px" />
      <el-form label-position="top" @submit.prevent>
        <el-form-item label="注册邮箱">
          <el-input v-model="forgotEmail" placeholder="用于接收验证码" :disabled="forgotStep === 'reset'" />
        </el-form-item>

        <template v-if="forgotStep === 'send'">
          <el-button
            type="primary"
            style="width: 100%"
            :loading="forgotSending"
            :disabled="countdown > 0"
            @click="sendCode"
          >
            {{ countdown > 0 ? `重新发送（${countdown}s）` : '发送验证码' }}
          </el-button>
          <p style="color: var(--el-text-color-secondary); font-size: 12px; margin-bottom: 0">
            验证码将发送至该邮箱，有效期约 3 分钟；若该邮箱未注册或账号已注销，将不会收到邮件。
          </p>
        </template>

        <template v-else>
          <el-form-item label="邮箱验证码">
            <el-input v-model="forgotCode" placeholder="请输入邮件中的验证码" />
          </el-form-item>
          <el-form-item label="新密码" required>
            <el-input
              v-model="forgotNewPassword"
              type="password"
              show-password
              autocomplete="new-password"
              placeholder="至少 8 位"
            />
          </el-form-item>
          <el-form-item label="确认新密码" required>
            <el-input v-model="forgotConfirm" type="password" show-password autocomplete="new-password" />
          </el-form-item>
          <el-button type="primary" style="width: 100%" :loading="forgotResetting" @click="submitReset">
            重置密码
          </el-button>
          <div style="display: flex; justify-content: space-between; margin-top: 12px">
            <a href="javascript:void(0)" @click="backToSend">重新获取验证码</a>
            <a href="javascript:void(0)" @click="forgotVisible = false">返回登录</a>
          </div>
        </template>
      </el-form>
    </el-dialog>
  </div>
</template>

<style scoped>
.admin-login {
  align-items: center;
  background: var(--color-background);
  display: flex;
  justify-content: center;
  min-height: 100vh;
  padding: 20px;
}

.admin-login-card {
  width: 100%;
  max-width: 400px;
}

.admin-login-card h1 {
  font-size: 22px;
  margin: 0 0 20px;
  text-align: center;
}
</style>
