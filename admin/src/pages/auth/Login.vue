<template>
  <section>
    <h1>后台登录</h1>
    <p class="hint">使用管理员账号登录后可进入数据源、审核和内容管理。</p>

    <div class="form-grid">
      <label>
        用户名
        <input v-model="username" data-test="username" />
      </label>
      <label>
        密码
        <input v-model="password" type="password" data-test="password" />
      </label>
    </div>

    <button data-test="login" @click="onLogin">登录</button>
    <p v-if="error" data-test="error" class="status-error">{{ error }}</p>
    <p v-if="token" data-test="success" class="status-success">登录成功</p>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { login } from '../../api/auth'

const username = ref('admin')
const password = ref('')
const token = ref('')
const error = ref('')
const emit = defineEmits<{
  (e: 'login-success', token: string): void
}>()

async function onLogin() {
  error.value = ''
  token.value = ''
  try {
    token.value = await login(username.value, password.value)
    emit('login-success', token.value)
  } catch (err) {
    error.value = (err as Error).message || '登录失败'
  }
}
</script>
