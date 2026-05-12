<template>
  <el-card>
    <div slot="header">新建配置</div>
    <el-form :model="form" :rules="rules" ref="form" label-width="100px">
      <el-form-item label="配置类型">
        <el-radio-group v-model="form.global">
          <el-radio :label="true">全局配置</el-radio>
          <el-radio :label="false">应用配置</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="所属业务组" v-if="form.global" prop="groupId">
        <el-select v-model="form.groupId" placeholder="请选择">
          <el-option v-for="g in groups" :key="g.value" :label="g.label" :value="parseInt(g.value)"></el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="所属应用" v-if="!form.global" prop="appId">
        <el-select v-model="form.appId" placeholder="请选择">
          <el-option v-for="a in apps" :key="a.value" :label="a.label" :value="parseInt(a.value)"></el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="环境" prop="env">
        <el-select v-model="form.env" placeholder="请选择">
          <el-option v-for="p in profiles" :key="p.value" :label="p.label" :value="p.value"></el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="版本">
        <el-input v-model="form.version" placeholder="latest"></el-input>
      </el-form-item>
      <el-form-item label="配置格式">
        <el-select v-model="form.type" placeholder="请选择" style="width: 240px">
          <el-option :value="2" label="配置项 (key=value)"></el-option>
          <el-option :value="1" label="文件 (properties/xml/yaml)"></el-option>
          <el-option :value="3" label="JSON"></el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="配置名称" v-if="form.type === 2 || form.type === 1" prop="name">
        <el-input v-model="form.name" :placeholder="form.type === 2 ? '如: db.host' : '如: application.properties'"></el-input>
      </el-form-item>
      <el-form-item label="配置内容" prop="contents">
        <el-input v-model="form.contents" type="textarea" :rows="15" placeholder="请输入配置内容"></el-input>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="handleSubmit">保存</el-button>
        <el-button @click="$router.back()">取消</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<script>
import { addConfig } from '../../api/config'
import request from '../../api/request'

export default {
  data() {
    return {
      form: {
        global: false,
        groupId: null,
        appId: null,
        env: '',
        version: 'latest',
        type: 2,
        name: '',
        contents: ''
      },
      rules: {
        env: [{ required: true, message: '请选择环境', trigger: 'change' }],
        contents: [{ required: true, message: '请输入配置内容', trigger: 'blur' }]
      },
      groups: [],
      apps: [],
      profiles: []
    }
  },
  created() {
    this.loadOptions()
  },
  methods: {
    async loadOptions() {
      try {
        const [groupRes, appRes, profileRes] = await Promise.all([
          request({ url: '/admin/group/options', method: 'get' }),
          request({ url: '/admin/app/options', method: 'get' }),
          request({ url: '/admin/profile/options', method: 'get' })
        ])
        this.groups = groupRes.data || []
        this.apps = appRes.data || []
        this.profiles = profileRes.data || []
      } catch (e) {}
    },
    handleSubmit() {
      this.$refs.form.validate(async (valid) => {
        if (!valid) return
        try {
          await addConfig(this.form)
          this.$message.success('保存成功')
          if (this.form.appId) {
            this.$router.push(`/app/${this.form.appId}/config/list`)
          } else {
            this.$router.push('/app/list')
          }
        } catch (e) {}
      })
    }
  }
}
</script>
