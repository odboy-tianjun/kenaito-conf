<template>
  <div class="login-container">
    <el-card class="login-card">
      <h2 class="login-title">配置中心</h2>
      <el-form :model="form" :rules="rules" ref="form" label-width="0">
        <el-form-item prop="userName">
          <el-input v-model="form.userName" placeholder="用户名" prefix-icon="el-icon-user"></el-input>
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="el-icon-lock" @keyup.enter.native="handleLogin"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleLogin" :loading="loading" style="width: 100%">登录</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import { login, getUserInfo, getMenus } from '../api/auth'

export default {
  data() {
    return {
      form: {
        userName: '',
        password: ''
      },
      rules: {
        userName: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
        password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
      },
      loading: false
    }
  },
  methods: {
    async handleLogin() {
      this.$refs.form.validate(async (valid) => {
        if (!valid) return
        this.loading = true
        try {
          const res = await login(this.form)
          localStorage.setItem('token', res.data.token)
          this.$store.dispatch('setUserInfo', res.data.userInfo)
          const menuRes = await getMenus()
          this.$store.dispatch('setMenus', menuRes.data)
          this.$router.push('/')
        } catch (e) {
          // error handled by interceptor
        } finally {
          this.loading = false
        }
      })
    }
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: #2d3a4b;
}
.login-card {
  width: 400px;
  padding: 20px;
}
.login-title {
  text-align: center;
  color: #303133;
  margin-bottom: 30px;
}
</style>
