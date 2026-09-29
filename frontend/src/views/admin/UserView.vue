<script setup lang="ts">
import { onMounted, ref } from 'vue'
import {
  createUser,
  deleteUser,
  listAdminUsers,
  resetUserPassword,
  updateUser,
} from '@/api'
import type { User } from '@/api'
import { usePagination } from '@/composables/usePagination'
import { ElMessage, ElMessageBox } from 'element-plus'

const { result, loading, load, changePage, reload } = usePagination<User>((p) =>
  listAdminUsers({ ...p, keyword: keyword.value.trim() }),
)

const keyword = ref('')
const searchLoading = ref(false)

// 新建/编辑对话框
const dialogVisible = ref(false)
const editing = ref<User | null>(null)
const saving = ref(false)
const form = ref({
  username: '',
  password: '',
  email: '',
  phone: '',
  avatar: '',
  role: 'admin',
  status: 1,
  is_master: 0,
})
const formError = ref('')

// 重置密码对话框
const resetVisible = ref(false)
const resetTarget = ref<User | null>(null)
const resetForm = ref({ new_password: '', confirm: '' })
const resetSaving = ref(false)
const resetError = ref('')

function openNew() {
  editing.value = null
  form.value = {
    username: '',
    password: '',
    email: '',
    phone: '',
    avatar: '',
    role: 'admin',
    status: 1,
    is_master: 0,
  }
  formError.value = ''
  dialogVisible.value = true
}

function openEdit(u: User) {
  editing.value = u
  form.value = {
    username: u.username,
    password: '',
    email: u.email ?? '',
    phone: u.phone ?? '',
    avatar: u.avatar ?? '',
    role: u.role,
    status: u.status,
    is_master: u.is_master,
  }
  formError.value = ''
  dialogVisible.value = true
}

async function save() {
  formError.value = ''
  if (!form.value.username.trim()) {
    formError.value = '用户名不能为空'
    return
  }
  if (!editing.value && form.value.password.length < 8) {
    formError.value = '初始密码长度不能少于 8 位'
    return
  }
  if (form.value.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.value.email)) {
    formError.value = '邮箱格式不正确'
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      const params = {
        username: form.value.username.trim(),
        email: form.value.email.trim(),
        phone: form.value.phone.trim(),
        avatar: form.value.avatar.trim(),
        role: form.value.role,
        status: form.value.status,
        is_master: form.value.is_master,
      }
      await updateUser(editing.value.id, params)
      ElMessage.success('用户已更新')
    } else {
      const params = {
        username: form.value.username.trim(),
        password: form.value.password,
        email: form.value.email.trim(),
        phone: form.value.phone.trim(),
        avatar: form.value.avatar.trim(),
        role: form.value.role,
        status: form.value.status,
        is_master: form.value.is_master,
      }
      await createUser(params)
      ElMessage.success('用户已创建')
    }
    dialogVisible.value = false
    await reload()
  } catch (e) {
    formError.value = (e as Error).message || '保存失败'
  } finally {
    saving.value = false
  }
}

async function search() {
  searchLoading.value = true
  try {
    await reload()
  } finally {
    searchLoading.value = false
  }
}

function openReset(u: User) {
  resetTarget.value = u
  resetForm.value = { new_password: '', confirm: '' }
  resetError.value = ''
  resetVisible.value = true
}

async function submitReset() {
  if (!resetTarget.value) return
  resetError.value = ''
  if (resetForm.value.new_password.length < 8) {
    resetError.value = '新密码长度不能少于 8 位'
    return
  }
  if (resetForm.value.new_password !== resetForm.value.confirm) {
    resetError.value = '两次输入的新密码不一致'
    return
  }
  resetSaving.value = true
  try {
    await resetUserPassword(resetTarget.value.id, { new_password: resetForm.value.new_password })
    ElMessage.success('密码已重置')
    resetVisible.value = false
  } catch (e) {
    resetError.value = (e as Error).message || '重置失败'
  } finally {
    resetSaving.value = false
  }
}

async function toggleStatus(u: User) {
  const target = u.status === 1 ? 0 : 1
  const action = target === 1 ? '启用' : '禁用'
  try {
    await ElMessageBox.confirm(`确定${action}用户「${u.username}」？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await updateUser(u.id, { status: target })
    ElMessage.success(`已${action}`)
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message || `${action}失败`)
  }
}

async function remove(u: User) {
  try {
    await ElMessageBox.confirm(`确定删除用户「${u.username}」？删除后不可恢复。`, '提示', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await deleteUser(u.id)
    ElMessage.success('删除成功')
    await load()
  } catch (e) {
    ElMessage.error((e as Error).message || '删除失败')
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="page-header">
      <h1>用户管理</h1>
      <el-button type="primary" @click="openNew">新建用户</el-button>
    </div>

    <el-card>
      <div style="display: flex; gap: 12px; margin-bottom: 16px">
        <el-input
          v-model="keyword"
          placeholder="按用户名 / 邮箱搜索"
          clearable
          style="max-width: 320px"
          @keyup.enter="search"
          @clear="search"
        />
        <el-button :loading="searchLoading" @click="search">搜索</el-button>
      </div>

      <el-table v-loading="loading" :data="result.list">
        <el-table-column prop="username" label="用户名" min-width="120" />
        <el-table-column prop="email" label="邮箱" min-width="180">
          <template #default="{ row }">{{ row.email || '-' }}</template>
        </el-table-column>
        <el-table-column prop="role" label="角色" width="100" />
        <el-table-column label="超级管理员" width="110">
          <template #default="{ row }">
            <el-tag v-if="row.is_master === 1" type="warning">是</el-tag>
            <el-tag v-else type="info">否</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_login_at" label="最近登录" width="170">
          <template #default="{ row }">{{ row.last_login_at || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button link type="warning" @click="openReset(row)">重置密码</el-button>
            <el-button link :type="row.status === 1 ? 'danger' : 'success'" @click="toggleStatus(row)">
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button link type="danger" @click="remove(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="result.total > result.pageSize"
        style="margin-top: 16px; justify-content: flex-end"
        background
        layout="prev, pager, next, total"
        :total="result.total"
        :page-size="result.pageSize"
        :current-page="result.page"
        @current-change="changePage"
      />
    </el-card>

    <!-- 新建/编辑用户 -->
    <el-dialog v-model="dialogVisible" :title="editing ? '编辑用户' : '新建用户'" width="520px">
      <el-alert v-if="formError" :title="formError" type="error" show-icon style="margin-bottom: 16px" />
      <el-form label-position="top">
        <el-form-item label="用户名" required>
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item v-if="!editing" label="初始密码" required>
          <el-input v-model="form.password" type="password" show-password placeholder="至少 8 位" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="角色">
              <el-input v-model="form.role" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-select v-model="form.status" style="width: 100%">
                <el-option label="启用" :value="1" />
                <el-option label="禁用" :value="0" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="超级管理员">
          <el-switch v-model="form.is_master" :active-value="1" :inactive-value="0" />
          <span style="margin-left: 8px; color: var(--el-text-color-secondary); font-size: 12px">
            仅超级管理员可访问用户管理；不能取消自己账号的超级管理员权限
          </span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>

    <!-- 重置密码 -->
    <el-dialog v-model="resetVisible" title="重置密码" width="440px">
      <el-alert
        v-if="resetError"
        :title="resetError"
        type="error"
        show-icon
        style="margin-bottom: 16px"
      />
      <el-form label-position="top">
        <el-form-item label="目标用户">
          <el-input :model-value="resetTarget?.username" disabled />
        </el-form-item>
        <el-form-item label="新密码" required>
          <el-input
            v-model="resetForm.new_password"
            type="password"
            show-password
            autocomplete="new-password"
            placeholder="至少 8 位，无需原密码"
          />
        </el-form-item>
        <el-form-item label="确认新密码" required>
          <el-input v-model="resetForm.confirm" type="password" show-password autocomplete="new-password" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetVisible = false">取消</el-button>
        <el-button type="primary" :loading="resetSaving" @click="submitReset">重置</el-button>
      </template>
    </el-dialog>
  </div>
</template>
