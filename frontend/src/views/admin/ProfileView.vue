<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getProfile, updatePassword, updateProfile, uploadFile } from '@/api'
import type { User } from '@/api'
import { markPageReady } from '@/composables/usePageReady'
import { ElMessage } from 'element-plus'

const user = ref<User | null>(null)
const loading = ref(false)
const saving = ref(false)
const errorMsg = ref('')

// 基本信息表单
const form = ref({ username: '', email: '', phone: '', avatar: '' })
const avatarUploading = ref(false)

// 修改密码表单
const pwdForm = ref({ old_password: '', new_password: '', confirm: '' })
const pwdSaving = ref(false)
const pwdErrorMsg = ref('')

onMounted(async () => {
  markPageReady()
  try {
    user.value = await getProfile()
    form.value = {
      username: user.value.username,
      email: user.value.email ?? '',
      phone: user.value.phone ?? '',
      avatar: user.value.avatar ?? '',
    }
  } catch (e) {
    errorMsg.value = (e as Error).message || '加载个人信息失败'
  }
})

async function saveProfile() {
  errorMsg.value = ''
  if (!form.value.username.trim()) {
    errorMsg.value = '用户名不能为空'
    return
  }
  saving.value = true
  try {
    const updated = await updateProfile({
      username: form.value.username.trim(),
      email: form.value.email.trim(),
      phone: form.value.phone.trim(),
      avatar: form.value.avatar.trim(),
    })
    user.value = updated
    ElMessage.success('个人信息已更新')
  } catch (e) {
    errorMsg.value = (e as Error).message || '保存失败'
  } finally {
    saving.value = false
  }
}

async function handleAvatarUpload(options: { file: File | Blob }) {
  const file = options.file
  if (!(file instanceof File)) return
  if (!file.type.startsWith('image/')) {
    ElMessage.warning('请选择图片文件')
    return
  }
  avatarUploading.value = true
  try {
    const result = await uploadFile(file)
    form.value.avatar = result.url
    ElMessage.success('头像上传成功')
  } catch (e) {
    ElMessage.error((e as Error).message || '上传失败')
  } finally {
    avatarUploading.value = false
  }
}

async function savePassword() {
  pwdErrorMsg.value = ''
  if (!pwdForm.value.old_password || !pwdForm.value.new_password) {
    pwdErrorMsg.value = '原密码和新密码不能为空'
    return
  }
  if (pwdForm.value.new_password.length < 8) {
    pwdErrorMsg.value = '新密码长度不能少于 8 位'
    return
  }
  if (pwdForm.value.new_password !== pwdForm.value.confirm) {
    pwdErrorMsg.value = '两次输入的新密码不一致'
    return
  }
  pwdSaving.value = true
  try {
    await updatePassword({
      old_password: pwdForm.value.old_password,
      new_password: pwdForm.value.new_password,
    })
    ElMessage.success('密码修改成功，请重新登录')
    pwdForm.value = { old_password: '', new_password: '', confirm: '' }
  } catch (e) {
    pwdErrorMsg.value = (e as Error).message || '修改失败'
  } finally {
    pwdSaving.value = false
  }
}
</script>

<template>
  <div v-loading="loading">
    <div class="page-header">
      <h1>个人信息</h1>
    </div>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>基本信息</template>
          <el-alert v-if="errorMsg" :title="errorMsg" type="error" show-icon style="margin-bottom: 16px" />
          <el-form label-position="top">
            <el-form-item label="用户名" required>
              <el-input v-model="form.username" />
            </el-form-item>
            <el-form-item label="邮箱">
              <el-input v-model="form.email" placeholder="用于接收验证码与通知" />
            </el-form-item>
            <el-form-item label="手机号">
              <el-input v-model="form.phone" />
            </el-form-item>
            <el-form-item label="头像 URL">
              <el-input v-model="form.avatar" placeholder="https://…">
                <template #append>
                  <el-upload :show-file-list="false" accept="image/*" :http-request="handleAvatarUpload"
                    :disabled="avatarUploading">
                    <el-button :loading="avatarUploading" :disabled="avatarUploading">上传</el-button>
                  </el-upload>
                </template>
              </el-input>
            </el-form-item>
            <el-button type="primary" :loading="saving" @click="saveProfile">保存</el-button>
          </el-form>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>修改密码</template>
          <el-alert v-if="pwdErrorMsg" :title="pwdErrorMsg" type="error" show-icon style="margin-bottom: 16px" />
          <el-form label-position="top">
            <el-form-item label="原密码" required>
              <el-input v-model="pwdForm.old_password" type="password" show-password autocomplete="current-password" />
            </el-form-item>
            <el-form-item label="新密码" required>
              <el-input v-model="pwdForm.new_password" type="password" show-password autocomplete="new-password"
                placeholder="至少 8 位" />
            </el-form-item>
            <el-form-item label="确认新密码" required>
              <el-input v-model="pwdForm.confirm" type="password" show-password autocomplete="new-password" />
            </el-form-item>
            <el-button type="primary" :loading="pwdSaving" @click="savePassword">修改密码</el-button>
          </el-form>
        </el-card>

        <el-card style="margin-top: 16px">
          <template #header>账号信息</template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="角色">
              {{ user?.role }}<el-tag v-if="user?.is_master === 1" type="warning" style="margin-left: 8px">超级管理员</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="user?.status === 1 ? 'success' : 'danger'">
                {{ user?.status === 1 ? '启用' : '禁用' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="最近登录 IP">{{ user?.last_ip || '-' }}</el-descriptions-item>
            <el-descriptions-item label="最近登录时间">{{ user?.last_login_at || '-' }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ user?.created_at || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
