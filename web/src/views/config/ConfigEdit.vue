<template>
  <el-card>
    <div slot="header">编辑配置</div>
    <el-form :model="form" ref="form" label-width="100px" v-loading="loading">
      <el-form-item label="配置名称">
        <el-input v-model="form.name" disabled></el-input>
      </el-form-item>
      <el-form-item label="环境">
        <el-input v-model="form.env" disabled></el-input>
      </el-form-item>
      <el-form-item label="所属应用">
        <el-input :value="form.appNames" disabled></el-input>
      </el-form-item>
      <el-form-item label="版本">
        <el-input v-model="form.version"></el-input>
      </el-form-item>
      <el-form-item label="配置内容">
        <el-input v-model="form.contents" type="textarea" :rows="20"></el-input>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSubmit">保存</el-button>
        <el-button @click="$router.back()">取消</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script>
import { getConfig, updateConfig } from '../../api/config'

export default {
  data() {
    return {
      form: {},
      loading: false
    }
  },
  created() {
    this.loadConfig()
  },
  methods: {
    async loadConfig() {
      this.loading = true
      try {
        const res = await getConfig(this.$route.params.id)
        this.form = res.data
      } finally {
        this.loading = false
      }
    },
    async handleSubmit() {
      try {
        await updateConfig(this.form)
        this.$message.success('保存成功')
        if (this.form.appId) {
          this.$router.push(`/app/${this.form.appId}/config/list`)
        } else {
          this.$router.push('/app/list')
        }
      } catch (e) {}
    }
  }
}
</script>
