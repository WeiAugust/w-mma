<template>
  <section>
    <h1>后台登录</h1>
    <label>
      用户名
      <input v-model="username" data-test="username" />
    </label>
    <label>
      密码
      <input v-model="password" type="password" data-test="password" />
    </label>
    <button data-test="login" @click="onLogin">登录</button>
    <p v-if="error" data-test="error">{{ error }}</p>
    <p v-if="token" data-test="success">登录成功</p>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { login } from '../../api/auth'

const username = ref('admin')
const password = ref('')
const token = ref('')
const error = ref('')

async function onLogin() {
  error.value = ''
  token.value = ''
  try {
    token.value = await login(username.value, password.value)
  } catch (err) {
    error.value = (err as Error).message || '登录失败'
  }
}
</script>
