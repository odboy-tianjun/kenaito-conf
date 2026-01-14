<template>
  <el-card>
    <div slot="header" style="display: flex; justify-content: space-between; align-items: center">
      <span>用户管理</span>
      <el-button type="primary" @click="showAdd">新建用户</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="60"></el-table-column>
      <el-table-column prop="name" label="用户名"></el-table-column>
      <el-table-column prop="mobile" label="手机号"></el-table-column>
      <el-table-column prop="email" label="邮箱"></el-table-column>
      <el-table-column prop="type" label="类型" width="100">
        <template slot-scope="scope">
          <el-tag v-if="scope.row.type === 'superAdmin'" type="danger">超级管理员</el-tag>
          <el-tag v-else-if="scope.row.type === 'groupAdmin'" type="warning">组管理员</el-tag>
          <el-tag v-else>普通用户</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="groupName" label="所属组"></el-table-column>
      <el-table-column prop="enabled" label="状态" width="100">
        <template slot-scope="scope">
          <el-switch 
            v-model="scope.row.enabled" 
            @change="handleSwitch(scope.row)"
            :disabled="scope.row.name === 'admin'">
          </el-switch>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template slot-scope="scope">
          <el-button type="text" @click="showEdit(scope.row)">编辑</el-button>
          <el-button type="text" @click="handleResetPwd(scope.row)">重置密码</el-button>
          <el-button type="text" style="color: #F56C6C" @click="handleDelete(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog :title="dialogTitle" :visible.sync="dialogVisible" width="500px">
      <el-form :model="form" :rules="rules" ref="form" label-width="80px">
        <el-form-item label="用户名" prop="name">
          <el-input v-model="form.name" :disabled="!!form.id"></el-input>
        </el-form-item>
        <el-form-item label="手机号" prop="mobile">
          <el-input v-model="form.mobile"></el-input>
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email"></el-input>
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" :disabled="form.id && form.type === 'superAdmin'">
            <el-option value="user" label="普通用户"></el-option>
            <el-option value="groupAdmin" label="组管理员"></el-option>
            <el-option value="superAdmin" label="超级管理员"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="所属组" v-if="form.type !== 'superAdmin'">
          <el-input v-model="form.groupName" placeholder="输入组名（不存在会自动创建）" :disabled="form.id && form.isSuperAdmin"></el-input>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </span>
    </el-dialog>
  </el-card>
</template>

<script>
import request from '../../api/request'

export default {
  data() {
    return {
      list: [],
      loading: false,
      dialogVisible: false,
      dialogTitle: '',
      form: {},
      rules: {
        name: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
        mobile: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
        type: [{ required: true, message: '请选择类型', trigger: 'change' }]
      }
    }
  },
  created() {
    this.loadData()
  },
  methods: {
    async loadData() {
      this.loading = true
      try {
        const res = await request({ url: '/admin/user/list', method: 'get' })
        this.list = res.data || []
      } finally {
        this.loading = false
      }
    },
    showAdd() {
      this.dialogTitle = '新建用户'
      this.form = { name: '', mobile: '', email: '', type: 'user', groupName: '' }
      this.dialogVisible = true
    },
    showEdit(row) {
      this.dialogTitle = '编辑用户'
      this.form = { 
        id: row.id,
        name: row.name,
        mobile: row.mobile, 
        email: row.email, 
        type: row.type, 
        groupName: row.groupName || '',
        isSuperAdmin: row.type === 'superAdmin'
      }
      this.dialogVisible = true
    },
    handleSubmit() {
      this.$refs.form.validate(async (valid) => {
        if (!valid) return
        try {
          const url = this.form.id ? '/admin/user/update' : '/admin/user/add'
          await request({ url, method: 'post', data: this.form })
          if (this.form.id) {
            this.$message.success('编辑成功')
          } else {
            this.$message.success('创建成功，默认密码：admin123')
          }
          this.dialogVisible = false
          this.loadData()
        } catch (e) {}
      })
    },
    async handleSwitch(row) {
      try {
        await request({ url: `/admin/user/switch/${row.id}`, method: 'post' })
        this.$message.success('操作成功')
        this.loadData()
      } catch (e) {}
    },
    async handleResetPwd(row) {
      try {
        await this.$confirm('确定重置密码？')
        const res = await request({ url: `/admin/user/resetPassword/${row.id}`, method: 'post' })
        this.$message.success(`新密码: ${res.data}`)
      } catch (e) {}
    },
    async handleDelete(row) {
      try {
        await this.$confirm('确定删除此用户？')
        await request({ url: `/admin/user/delete/${row.id}`, method: 'post' })
        this.$message.success('删除成功')
        this.loadData()
      } catch (e) {}
    }
  }
}
</script>
