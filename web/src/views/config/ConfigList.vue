<template>
  <el-card>
    <div slot="header" style="display: flex; justify-content: space-between; align-items: center">
      <span>配置列表</span>
      <div>
        <el-select v-model="query.env" placeholder="环境" clearable style="width: 120px; margin-right: 10px" @change="loadData">
          <el-option v-for="p in profiles" :key="p.value" :label="p.label" :value="p.value"></el-option>
        </el-select>
        <el-button type="primary" @click="$router.push('/config/add')">新建配置</el-button>
      </div>
    </div>
    <el-table :data="list" v-loading="loading" border stripe>
      <el-table-column prop="id" label="ID" width="60"></el-table-column>
      <el-table-column prop="name" label="配置名"></el-table-column>
      <el-table-column prop="env" label="环境" width="80"></el-table-column>
      <el-table-column prop="appNames" label="所属应用"></el-table-column>
      <el-table-column prop="type" label="类型" width="80">
        <template slot-scope="scope">
          <span v-if="scope.row.type === 1">文件</span>
          <span v-else-if="scope.row.type === 2">配置项</span>
          <span v-else-if="scope.row.type === 3">JSON</span>
        </template>
      </el-table-column>
      <el-table-column prop="version" label="版本" width="100"></el-table-column>
      <el-table-column label="操作" width="200">
        <template slot-scope="scope">
          <el-button type="text" @click="$router.push(`/config/edit/${scope.row.id}`)">编辑</el-button>
          <el-button type="text" @click="showHistory(scope.row)">历史</el-button>
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

    <el-dialog title="配置历史" :visible.sync="historyVisible" width="900px">
      <el-table :data="historyList" border stripe>
        <el-table-column prop="id" label="ID" width="60"></el-table-column>
        <el-table-column prop="createdBy" label="操作人" width="100"></el-table-column>
        <el-table-column label="操作时间" width="180">
          <template slot-scope="scope">
            {{ formatTime(scope.row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作">
          <template slot-scope="scope">
            <el-button type="text" @click="viewHistoryContent(scope.row)">查看内容</el-button>
            <el-button type="text" @click="handleRollback(scope.row)">回滚到此版本</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog title="历史版本内容" :visible.sync="contentVisible" width="800px" append-to-body>
      <div style="margin-bottom: 10px">
        <strong>配置名称：</strong>{{ currentHistory.name }} | 
        <strong>版本：</strong>{{ currentHistory.version }} | 
        <strong>类型：</strong>{{ getTypeName(currentHistory.type) }}
      </div>
      <el-input 
        v-model="formattedContent" 
        type="textarea" 
        :rows="25" 
        readonly
        style="font-family: monospace"
      ></el-input>
    </el-dialog>
  </el-card>
</template>

<script>
import { getConfigList, deleteConfig, getConfigHistory, rollbackConfig } from '../../api/config'
import request from '../../api/request'

export default {
  data() {
    return {
      list: [],
      loading: false,
      total: 0,
      query: { pageNo: 1, pageSize: 20, env: '' },
      profiles: [],
      historyVisible: false,
      historyList: [],
      currentConfig: null,
      contentVisible: false,
      currentHistory: {},
      formattedContent: ''
    }
  },
  created() {
    this.loadData()
    this.loadProfiles()
  },
  methods: {
    async loadData() {
      this.loading = true
      try {
        const res = await getConfigList(this.query)
        this.list = res.data || []
        this.total = res.total || 0
      } finally {
        this.loading = false
      }
    },
    async loadProfiles() {
      try {
        const res = await request({ url: '/admin/profile/options', method: 'get' })
        this.profiles = res.data || []
      } catch (e) {}
    },
    handlePageChange(page) {
      this.query.pageNo = page
      this.loadData()
    },
    async handleDelete(row) {
      try {
        await this.$confirm('确定删除此配置？')
        await deleteConfig(row.id)
        this.$message.success('删除成功')
        this.loadData()
      } catch (e) {}
    },
    async showHistory(row) {
      this.currentConfig = row
      try {
        const res = await getConfigHistory(row.id)
        this.historyList = res.data || []
        this.historyVisible = true
      } catch (e) {}
    },
    viewHistoryContent(row) {
      this.currentHistory = row
      const content = row.contents || ''
      if (row.type === 3) {
        try {
          this.formattedContent = JSON.stringify(JSON.parse(content), null, 2)
        } catch (e) {
          this.formattedContent = content
        }
      } else {
        this.formattedContent = content
      }
      this.contentVisible = true
    },
    getTypeName(type) {
      if (type === 1) return '文件'
      if (type === 2) return '配置项'
      if (type === 3) return 'JSON'
      return '未知'
    },
    formatTime(time) {
      if (!time) return ''
      const d = new Date(time)
      const pad = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
    },
    async handleRollback(row) {
      try {
        await this.$confirm('确定回滚到此版本？')
        await rollbackConfig(row.id)
        this.$message.success('回滚成功')
        this.historyVisible = false
        this.loadData()
      } catch (e) {}
    }
  }
}
</script>
