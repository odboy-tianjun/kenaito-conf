<template>
  <el-card>
    <div slot="header" style="display: flex; justify-content: space-between; align-items: center">
      <span>业务组管理</span>
      <div>
        <el-select v-model="filterName" placeholder="组名称" clearable filterable style="width: 200px; margin-right: 10px">
          <el-option v-for="g in groupOptions" :key="g" :label="g" :value="g"></el-option>
        </el-select>
        <el-button type="primary" @click="showAdd">新建业务组</el-button>
      </div>
    </div>
    <el-table :data="filteredList" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="60"></el-table-column>
      <el-table-column prop="name" label="组名称"></el-table-column>
      <el-table-column prop="master" label="负责人"></el-table-column>
      <el-table-column prop="remarks" label="备注"></el-table-column>
      <el-table-column prop="enabled" label="状态" width="100">
        <template slot-scope="scope">
          <el-switch v-model="scope.row.enabled" @change="handleSwitch(scope.row)"></el-switch>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template slot-scope="scope">
          <el-button type="text" @click="showEdit(scope.row)">编辑</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog :title="dialogTitle" :visible.sync="dialogVisible" width="500px">
      <el-form :model="form" :rules="rules" ref="form" label-width="80px">
        <el-form-item label="组名称" prop="name">
          <el-input v-model="form.name"></el-input>
        </el-form-item>
        <el-form-item label="负责人" prop="masterUid">
          <el-select v-model="form.masterUid" placeholder="请选择">
            <el-option v-for="u in users" :key="u.value" :label="u.label" :value="parseInt(u.value)"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remarks" type="textarea"></el-input>
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
      filterName: '',
      dialogVisible: false,
      dialogTitle: '',
      form: {},
      rules: {
        name: [{ required: true, message: '请输入组名称', trigger: 'blur' }],
        masterUid: [{ required: true, message: '请选择负责人', trigger: 'change' }]
      },
      users: []
    }
  },
  computed: {
    groupOptions() {
      return [...new Set(this.list.map(g => g.name))].filter(name => name && name.trim())
    },
    filteredList() {
      if (!this.filterName) return this.list
      return this.list.filter(g => g.name === this.filterName)
    }
  },
  created() {
    this.loadData()
    this.loadUsers()
  },
  methods: {
    async loadData() {
      this.loading = true
      try {
        const res = await request({ url: '/admin/group/list', method: 'get' })
        this.list = res.data || []
      } finally {
        this.loading = false
      }
    },
    async loadUsers() {
      try {
        const res = await request({ url: '/admin/user/options', method: 'get' })
        this.users = res.data || []
      } catch (e) {}
    },
    async handleSwitch(row) {
      try {
        await request({ url: `/admin/group/switch/${row.id}`, method: 'post' })
      } catch (e) {
        row.enabled = !row.enabled
      }
    },
    showAdd() {
      this.dialogTitle = '新建业务组'
      this.form = { name: '', masterUid: null, remarks: '' }
      this.dialogVisible = true
    },
    showEdit(row) {
      this.dialogTitle = '编辑业务组'
      this.form = { ...row }
      this.dialogVisible = true
    },
    handleSubmit() {
      this.$refs.form.validate(async (valid) => {
        if (!valid) return
        try {
          const url = this.form.id ? '/admin/group/update' : '/admin/group/add'
          await request({ url, method: 'post', data: this.form })
          this.$message.success('保存成功')
          this.dialogVisible = false
          this.loadData()
        } catch (e) {}
      })
    }
  }
}
</script>
