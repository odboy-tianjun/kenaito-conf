<template>
  <el-card>
    <div slot="header" style="display: flex; justify-content: space-between; align-items: center">
      <span>环境列表</span>
      <el-button type="primary" @click="showAdd">新建环境</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="60"></el-table-column>
      <el-table-column prop="name" label="环境标识"></el-table-column>
      <el-table-column prop="alias" label="环境名称"></el-table-column>
      <el-table-column prop="isDefault" label="默认" width="80">
        <template slot-scope="scope">
          <el-tag v-if="scope.row.isDefault" type="success">是</el-tag>
          <span v-else>否</span>
        </template>
      </el-table-column>
      <el-table-column prop="enabled" label="状态" width="100">
        <template slot-scope="scope">
          <el-switch v-model="scope.row.enabled" @change="handleSwitch(scope.row)"></el-switch>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180">
        <template slot-scope="scope">
          <el-button type="text" @click="showEdit(scope.row)">编辑</el-button>
          <el-button type="text" style="color: #F56C6C" @click="handleDelete(scope.row)" v-if="!scope.row.isDefault">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog :title="dialogTitle" :visible.sync="dialogVisible" width="500px">
      <el-form :model="form" :rules="rules" ref="form" label-width="80px">
        <el-form-item label="环境标识" prop="name">
          <el-input v-model="form.name" placeholder="如: daily, stage, online" :disabled="!!form.id"></el-input>
        </el-form-item>
        <el-form-item label="环境名称">
          <el-input v-model="form.alias" placeholder="如: 测试环境"></el-input>
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
        name: [{ required: true, message: '请输入环境标识', trigger: 'blur' }]
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
        const res = await request({ url: '/admin/profile/list', method: 'get' })
        this.list = res.data || []
      } finally {
        this.loading = false
      }
    },
    showAdd() {
      this.dialogTitle = '新建环境'
      this.form = { name: '', alias: '' }
      this.dialogVisible = true
    },
    showEdit(row) {
      this.dialogTitle = '编辑环境'
      this.form = { id: row.id, name: row.name, alias: row.alias }
      this.dialogVisible = true
    },
    handleSubmit() {
      this.$refs.form.validate(async (valid) => {
        if (!valid) return
        try {
          const url = this.form.id ? '/admin/profile/update' : '/admin/profile/add'
          await request({ url, method: 'post', data: this.form })
          this.$message.success('保存成功')
          this.dialogVisible = false
          this.loadData()
        } catch (e) {}
      })
    },
    async handleSwitch(row) {
      try {
        await request({ url: `/admin/profile/switch/${row.id}`, method: 'post' })
      } catch (e) {
        row.enabled = !row.enabled
      }
    },
    async handleDelete(row) {
      try {
        await this.$confirm('确定删除此环境？')
        await request({ url: `/admin/profile/delete/${row.id}`, method: 'post' })
        this.$message.success('删除成功')
        this.loadData()
      } catch (e) {}
    }
  }
}
</script>
