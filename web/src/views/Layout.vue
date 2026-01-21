<template>
  <el-container style="height: 100vh">
    <el-aside width="200px" style="background: #1a237e">
      <div class="logo">Kenaito配置中心</div>
      <el-menu
        :default-active="activeMenu"
        background-color="#1a237e"
        text-color="rgba(255, 255, 255, 0.7)"
        active-text-color="#fff"
        router
      >
        <template v-for="menu in menus">
          <el-submenu v-if="menu.children && menu.children.length > 1" :index="menu.title" :key="menu.title">
            <template slot="title">
              <i :class="menu.icon"></i>
              <span>{{ menu.title }}</span>
            </template>
            <el-menu-item v-for="child in menu.children" :key="child.url" :index="child.url">
              {{ child.title }}
            </el-menu-item>
          </el-submenu>
          <el-menu-item v-else-if="menu.children && menu.children.length === 1" :index="menu.children[0].url" :key="menu.children[0].url">
            <i :class="menu.icon"></i>
            <span>{{ menu.children[0].title }}</span>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header style="background: #fff; display: flex; align-items: center; justify-content: space-between; box-shadow: 0 1px 4px rgba(0,21,41,.08)">
        <span style="font-size: 16px; font-weight: bold">欢迎使用 Kenaito配置中心</span>
        <div>
          <span style="margin-right: 15px">{{ userInfo ? userInfo.name : '' }}</span>
          <el-button type="text" @click="handleLogout">退出</el-button>
        </div>
      </el-header>
      <el-main style="background: #f0f2f5; padding: 20px">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script>
import { getMenus, getUserInfo, logout } from '../api/auth'
import { mapState } from 'vuex'

export default {
  computed: {
    ...mapState(['userInfo', 'menus']),
    activeMenu() {
      return this.$route.path
    }
  },
  created() {
    this.loadMenus()
  },
  methods: {
    async loadMenus() {
      if (this.menus.length === 0) {
        try {
          const res = await getMenus()
          this.$store.dispatch('setMenus', res.data)
        } catch (e) {}
      }
      if (!this.userInfo) {
        try {
          const res = await getUserInfo()
          this.$store.dispatch('setUserInfo', res.data)
        } catch (e) {}
      }
    },
    async handleLogout() {
      try {
        await logout()
      } catch (e) {}
      localStorage.removeItem('token')
      this.$store.dispatch('setUserInfo', null)
      this.$store.dispatch('setMenus', [])
      this.$router.push('/login')
    }
  }
}
</script>

<style scoped>
.logo {
  height: 60px;
  line-height: 60px;
  text-align: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  background: rgba(0, 0, 0, 0.2);
}
.el-menu {
  border-right: none;
}
.el-menu-item:hover,
.el-submenu__title:hover {
  background-color: rgba(255, 255, 255, 0.1) !important;
}
.el-menu-item.is-active {
  background-color: rgba(255, 255, 255, 0.15) !important;
}
</style>
