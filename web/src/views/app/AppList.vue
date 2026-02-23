<template>
  <el-card>
    <div slot="header" style="display: flex; justify-content: space-between; align-items: center">
      <span>应用管理</span>
      <el-button type="primary" @click="showAdd">新建应用</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="60"></el-table-column>
      <el-table-column prop="name" label="应用名称"></el-table-column>
      <el-table-column prop="code" label="应用编码"></el-table-column>
      <el-table-column prop="groupName" label="所属业务组" width="120"></el-table-column>
      <el-table-column prop="ownerName" label="负责人" width="100"></el-table-column>
      <el-table-column prop="token" label="Token" width="200">
        <template slot-scope="scope">
          <template v-if="scope.row.token">
            <span style="font-family: monospace; color: #606266">{{ maskToken(scope.row.token) }}</span>
            <el-button type="text" size="mini" style="margin-left: 8px" @click="copyToken(scope.row.token)">复制</el-button>
          </template>
          <span v-else style="color: #909399">未设置</span>
        </template>
      </el-table-column>
      <el-table-column prop="enabled" label="状态" width="100">
        <template slot-scope="scope">
          <el-switch v-model="scope.row.enabled" @change="handleSwitch(scope.row)"></el-switch>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template slot-scope="scope">
          <el-button type="text" @click="showEdit(scope.row)">编辑</el-button>
          <el-button type="text" @click="handleToken(scope.row)">
            {{ scope.row.token ? '删除Token' : '生成Token' }}
          </el-button>
          <el-button type="text" style="color: #F56C6C" @click="handleDelete(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      style="margin-top: 15px; text-align: right"
      @current-change="handlePageChange"
      :current-page="query.pageNo"
      :page-size="query.pageSize"
      :total="total"
      layout="total, prev, pager, next"
    ></el-pagination>

    <el-dialog :title="dialogTitle" :visible.sync="dialogVisible" width="500px">
      <el-form :model="form" :rules="rules" ref="form" label-width="100px">
        <el-form-item label="应用名称" prop="name">
          <el-input v-model="form.name"></el-input>
        </el-form-item>
        <el-form-item label="应用编码" prop="code">
          <el-input v-model="form.code" :disabled="!!form.id"></el-input>
        </el-form-item>
        <el-form-item label="所属业务组">
          <el-select v-model="form.groupId" placeholder="请选择">
            <el-option v-for="g in groups" :key="g.value" :label="g.label" :value="parseInt(g.value)"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="负责人">
          <el-select v-model="form.ownerId" placeholder="请选择">
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
import { getAppList, addApp, updateApp, deleteApp, generateToken, deleteToken } from '../../api/app'
import request from '../../api/request'

export default {
  data() {
    return {
      list: [],
      loading: false,
      total: 0,
      query: { pageNo: 1, pageSize: 20 },
      dialogVisible: false,
      dialogTitle: '',
      form: {},
      rules: {
        name: [{ required: true, message: '请输入应用名称', trigger: 'blur' }],
        code: [{ required: true, message: '请输入应用编码', trigger: 'blur' }]
      },
      groups: [],
      users: []
    }
  },
  created() {
    this.loadData()
    this.loadOptions()
  },
  methods: {
    async loadData() {
      this.loading = true
      try {
        const res = await getAppList(this.query)
        this.list = res.data || []
        this.total = res.total || 0
      } finally {
        this.loading = false
      }
    },
    async loadOptions() {
      try {
        const [groupRes, userRes] = await Promise.all([
          request({ url: '/admin/group/options', method: 'get' }),
          request({ url: '/admin/user/options', method: 'get' })
        ])
        this.groups = groupRes.data || []
        this.users = userRes.data || []
      } catch (e) {}
    },
    showAdd() {
      this.dialogTitle = '新建应用'
      this.form = { name: '', code: '', groupId: null, ownerId: null, remarks: '' }
      this.dialogVisible = true
    },
    showEdit(row) {
      this.dialogTitle = '编辑应用'
      this.form = { ...row }
      this.dialogVisible = true
    },
    handleSubmit() {
      this.$refs.form.validate(async (valid) => {
        if (!valid) return
        try {
          if (this.form.id) {
            await updateApp(this.form)
          } else {
            await addApp(this.form)
          }
          this.$message.success('保存成功')
          this.dialogVisible = false
          this.loadData()
        } catch (e) {}
      })
    },
    async handleDelete(row) {
      try {
        await this.$confirm('确定删除此应用？')
        await deleteApp(row.id)
        this.$message.success('删除成功')
        this.loadData()
      } catch (e) {}
    },
    async handleSwitch(row) {
      try {
        await request({ url: `/admin/app/switch/${row.id}`, method: 'post' })
      } catch (e) {
        row.enabled = !row.enabled
      }
    },
    async handleToken(row) {
      if (row.token) {
        try {
          await this.$confirm('确定删除此应用的 Token？删除后客户端将无法通过 Token 校验。')
          await deleteToken(row.id)
          this.$message.success('Token 已删除')
          this.loadData()
        } catch (e) {}
      } else {
        try {
          await this.$confirm('确定为此应用生成 Token？生成后客户端需要配置此 Token 才能拉取配置。')
          const res = await generateToken(row.id)
          this.$message.success('Token 已生成: ' + res.data)
          this.loadData()
        } catch (e) {}
      }
    },
    handlePageChange(page) {
      this.query.pageNo = page
      this.loadData()
    },
    maskToken(token) {
      if (!token || token.length <= 8) return '****'
      return token.substring(0, 4) + '****' + token.substring(token.length - 4)
    },
    copyToken(token) {
      navigator.clipboard.writeText(token).then(() => {
        this.$message.success('Token 已复制到剪贴板')
      }).catch(() => {
        this.$message.error('复制失败')
      })
    }
  }
}
</script>
