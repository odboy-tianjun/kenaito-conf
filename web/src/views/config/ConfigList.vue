<template>
  <div>
    <el-card style="margin-bottom: 16px">
      <el-form :inline="true" @submit.native.prevent="handleSearch">
        <el-form-item label="环境">
          <el-select v-model="query.env" placeholder="请选择" clearable style="width: 160px">
            <el-option v-for="p in profiles" :key="p.value" :label="p.label" :value="p.value"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card>
      <div slot="header" style="display: flex; justify-content: flex-end">
        <el-button type="primary" @click="$router.push('/config/add')">新建配置</el-button>
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
      <el-table-column prop="clientCount" label="在线客户端" width="120">
        <template slot-scope="scope">
          <el-button v-if="scope.row.clientCount > 0" type="text" @click="showClients(scope.row)">
            {{ scope.row.clientCount }} 个
          </el-button>
          <span v-else style="color: #909399">无</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template slot-scope="scope">
          <el-button type="text" @click="showHistory(scope.row)">变更历史</el-button>
          <el-button type="text" @click="$router.push(`/config/edit/${scope.row.id}`)">编辑</el-button>
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

    <el-dialog title="变更历史" :visible.sync="historyVisible" width="900px">
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

    <el-dialog title="历史版本对比" :visible.sync="contentVisible" width="1200px" append-to-body>
      <div style="margin-bottom: 10px">
        <strong>配置名称：</strong>{{ currentHistory.name }} |
        <strong>版本：</strong>{{ currentHistory.version }} |
        <strong>类型：</strong>{{ getTypeName(currentHistory.type) }}
      </div>
      <div style="margin-bottom: 8px; font-size: 13px; color: #666;">
        <span style="background: #ffeef0; padding: 2px 6px; margin-right: 10px;">红色 = 已删除</span>
        <span style="background: #e6ffed; padding: 2px 6px; margin-right: 10px;">绿色 = 新增</span>
        <span style="background: #fff8e1; padding: 2px 6px;">黄色 = 已修改</span>
      </div>
      <div class="diff-container">
        <div class="diff-panel">
          <div class="diff-panel-header">当前生效配置</div>
          <div class="diff-panel-body">
            <div v-for="(line, index) in leftLines" :key="'l'+index" :class="['diff-line', line.type]">
              <span class="line-number">{{ line.lineNum }}</span>
              <span class="line-content">{{ line.text }}</span>
            </div>
          </div>
        </div>
        <div class="diff-panel">
          <div class="diff-panel-header">历史版本</div>
          <div class="diff-panel-body">
            <div v-for="(line, index) in rightLines" :key="'r'+index" :class="['diff-line', line.type]">
              <span class="line-number">{{ line.lineNum }}</span>
              <span class="line-content">{{ line.text }}</span>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>

    <el-dialog title="在线客户端" :visible.sync="clientVisible" width="700px">
      <el-table :data="clientList" border stripe>
        <el-table-column prop="ip" label="IP 地址" width="160"></el-table-column>
        <el-table-column prop="lastFetchType" label="拉取方式" width="120">
          <template slot-scope="scope">
            <span v-if="scope.row.lastFetchType === 'fetchAll'">全量拉取</span>
            <span v-else>增量拉取</span>
          </template>
        </el-table-column>
        <el-table-column label="最后拉取时间">
          <template slot-scope="scope">
            {{ formatTime(scope.row.lastFetchAt) }}
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </el-card>
  </div>
</template>

<script>
import { getConfigList, deleteConfig, getConfigHistory, rollbackConfig, getConfig } from '../../api/config'
import { getClientListByConfig } from '../../api/client'
import request from '../../api/request'

function computeDiff(oldText, newText) {
  const oldLines = oldText.split('\n')
  const newLines = newText.split('\n')
  const m = oldLines.length
  const n = newLines.length

  const dp = Array.from({ length: m + 1 }, () => new Array(n + 1).fill(0))
  for (let i = 1; i <= m; i++) {
    for (let j = 1; j <= n; j++) {
      if (oldLines[i - 1] === newLines[j - 1]) {
        dp[i][j] = dp[i - 1][j - 1] + 1
      } else {
        dp[i][j] = Math.max(dp[i - 1][j], dp[i][j - 1])
      }
    }
  }

  const ops = []
  let i = m, j = n
  while (i > 0 || j > 0) {
    if (i > 0 && j > 0 && oldLines[i - 1] === newLines[j - 1]) {
      ops.push({ type: 'equal', oldIdx: i - 1, newIdx: j - 1 })
      i--; j--
    } else if (j > 0 && (i === 0 || dp[i][j - 1] >= dp[i - 1][j])) {
      ops.push({ type: 'added', newIdx: j - 1 })
      j--
    } else {
      ops.push({ type: 'removed', oldIdx: i - 1 })
      i--
    }
  }
  ops.reverse()

  // 配对相邻 removed+added 为 modified
  const paired = []
  let k = 0
  while (k < ops.length) {
    if (ops[k].type === 'removed') {
      const removedGroup = []
      while (k < ops.length && ops[k].type === 'removed') {
        removedGroup.push(ops[k])
        k++
      }
      const addedGroup = []
      while (k < ops.length && ops[k].type === 'added') {
        addedGroup.push(ops[k])
        k++
      }
      const pairCount = Math.min(removedGroup.length, addedGroup.length)
      for (let p = 0; p < pairCount; p++) {
        paired.push({ type: 'modified', oldIdx: removedGroup[p].oldIdx, newIdx: addedGroup[p].newIdx })
      }
      for (let p = pairCount; p < removedGroup.length; p++) {
        paired.push(removedGroup[p])
      }
      for (let p = pairCount; p < addedGroup.length; p++) {
        paired.push(addedGroup[p])
      }
    } else {
      paired.push(ops[k])
      k++
    }
  }

  // 生成左右面板行
  const left = []
  const right = []
  let leftNum = 0, rightNum = 0
  for (const op of paired) {
    if (op.type === 'equal') {
      leftNum++; rightNum++
      left.push({ text: oldLines[op.oldIdx], type: 'unchanged', lineNum: leftNum })
      right.push({ text: newLines[op.newIdx], type: 'unchanged', lineNum: rightNum })
    } else if (op.type === 'modified') {
      leftNum++; rightNum++
      left.push({ text: oldLines[op.oldIdx], type: 'modified', lineNum: leftNum })
      right.push({ text: newLines[op.newIdx], type: 'modified', lineNum: rightNum })
    } else if (op.type === 'removed') {
      leftNum++
      left.push({ text: oldLines[op.oldIdx], type: 'removed', lineNum: leftNum })
      right.push({ text: '', type: 'empty', lineNum: '' })
    } else if (op.type === 'added') {
      rightNum++
      left.push({ text: '', type: 'empty', lineNum: '' })
      right.push({ text: newLines[op.newIdx], type: 'added', lineNum: rightNum })
    }
  }
  return { left, right }
}

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
      leftLines: [],
      rightLines: [],
      clientVisible: false,
      clientList: []
    }
  },
  created() {
    this.loadData()
    this.loadProfiles()
  },
  methods: {
    handleSearch() {
      this.query.pageNo = 1
      this.loadData()
    },
    handleReset() {
      this.query.env = ''
      this.query.pageNo = 1
      this.loadData()
    },
    async loadData() {
      this.loading = true
      try {
        const params = {
          pageNo: this.query.pageNo,
          pageSize: this.query.pageSize,
          appId: this.$route.params.appId
        }
        if (this.query.env) {
          params.env = this.query.env
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
    async viewHistoryContent(row) {
      this.currentHistory = row
      const historyContent = row.contents || ''
      const formatted = row.type === 3 ? (() => { try { return JSON.stringify(JSON.parse(historyContent), null, 2) } catch (e) { return historyContent } })() : historyContent

      this.leftLines = []
      this.rightLines = []
      try {
        const res = await getConfig(row.originId)
        if (res && res.data && res.data.contents) {
          const currentRaw = res.data.type === 3 ? (() => { try { return JSON.stringify(JSON.parse(res.data.contents), null, 2) } catch (e) { return res.data.contents } })() : res.data.contents
          const result = computeDiff(formatted, currentRaw)
          this.leftLines = result.right
          this.rightLines = result.left
        }
      } catch (e) {
        this.rightLines = formatted.split('\n').map((line, i) => ({ lineNum: i + 1, text: line, type: 'unchanged' }))
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
    },
    async showClients(row) {
      try {
        const res = await getClientListByConfig(row.id)
        this.clientList = res.data || []
        this.clientVisible = true
      } catch (e) {}
    }
  }
}
</script>

<style scoped>
::v-deep .diff-container {
  display: flex;
  gap: 2px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
}

::v-deep .diff-panel {
  flex: 1;
  min-width: 0;
}

::v-deep .diff-panel-header {
  background: #f5f7fa;
  padding: 8px 12px;
  font-size: 13px;
  font-weight: bold;
  color: #606266;
  border-bottom: 1px solid #dcdfe6;
}

::v-deep .diff-panel-body {
  max-height: 500px;
  overflow-y: auto;
  font-family: monospace;
  font-size: 13px;
}

::v-deep .diff-line {
  display: flex;
  padding: 0 10px;
  line-height: 22px;
  white-space: pre;
  min-height: 22px;
}

::v-deep .diff-line .line-number {
  flex-shrink: 0;
  width: 40px;
  text-align: right;
  padding-right: 10px;
  color: #909399;
  border-right: 1px solid #ebeef5;
  margin-right: 10px;
  user-select: none;
}

::v-deep .diff-line .line-content {
  flex: 1;
  overflow-x: auto;
}

::v-deep .diff-line.added {
  background: #e6ffed;
}

::v-deep .diff-line.removed {
  background: #ffeef0;
}

::v-deep .diff-line.modified {
  background: #fff8e1;
}

::v-deep .diff-line.unchanged {
  background: #fff;
}

::v-deep .diff-line.empty {
  background: #fafafa;
}
</style>
