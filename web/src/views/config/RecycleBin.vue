<template>
  <div>
    <el-card style="margin-bottom: 16px">
      <el-form :inline="true" @submit.native.prevent="handleSearch">
        <el-form-item label="关键字">
          <el-input v-model="query.keyword" placeholder="配置名或内容" clearable style="width: 200px"></el-input>
        </el-form-item>
        <el-form-item label="环境">
          <el-select v-model="query.env" placeholder="请选择" clearable style="width: 160px">
            <el-option v-for="p in profiles" :key="p.value" :label="p.label" :value="p.value"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="应用">
          <el-select v-model="query.appId" placeholder="请选择" clearable filterable style="width: 200px">
            <el-option v-for="app in apps" :key="app.value" :label="app.label" :value="parseInt(app.value)"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card>
      <div slot="header">回收站（仅显示已删除配置）</div>
      <el-table :data="list" v-loading="loading" border stripe>
        <el-table-column prop="id" label="ID" width="60"></el-table-column>
        <el-table-column prop="name" label="配置名"></el-table-column>
        <el-table-column prop="env" label="环境" width="80"></el-table-column>
        <el-table-column prop="appNames" label="所属应用"></el-table-column>
        <el-table-column prop="version" label="版本" width="100"></el-table-column>
        <el-table-column prop="updatedBy" label="删除人" width="120"></el-table-column>
        <el-table-column label="删除时间" width="180">
          <template slot-scope="scope">
            {{ formatTime(scope.row.updatedAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template slot-scope="scope">
            <el-button type="text" @click="handleRestore(scope.row)">还原</el-button>
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
    </el-card>
  </div>
</template>

<script>
import { getConfigList, restoreConfig } from '../../api/config'
import request from '../../api/request'

export default {
  data() {
    return {
      list: [],
      loading: false,
      total: 0,
      query: { pageNo: 1, pageSize: 10, env: '', appId: null, keyword: '' },
      profiles: [],
      apps: []
    }
  },
  created() {
    this.loadData()
    this.loadProfiles()
    this.loadApps()
  },
  methods: {
    handleSearch() {
      this.query.pageNo = 1
      this.loadData()
    },
    handleReset() {
      this.query = { pageNo: 1, pageSize: 10, env: '', appId: null, keyword: '' }
      this.loadData()
    },
    async loadData() {
      this.loading = true
      try {
        const params = {
          pageNo: this.query.pageNo,
          pageSize: this.query.pageSize,
          deleted: 1
        }
        if (this.query.env) {
          params.env = this.query.env
        }
        if (this.query.appId) {
          params.appId = this.query.appId
        }
        if (this.query.keyword) {
          params.keyword = this.query.keyword
        }
        const res = await getConfigList(params)
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
    async loadApps() {
      try {
        const res = await request({ url: '/admin/app/options', method: 'get' })
        this.apps = res.data || []
      } catch (e) {}
    },
    handlePageChange(page) {
      this.query.pageNo = page
      this.loadData()
    },
    async handleRestore(row) {
      try {
        await this.$confirm(`确定还原配置「${row.name}」？`)
        await restoreConfig(row.id)
        this.$message.success('还原成功')
        this.loadData()
      } catch (e) {}
    },
    formatTime(time) {
      if (!time) return ''
      const d = new Date(time)
      const pad = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
    }
  }
}
</script>
