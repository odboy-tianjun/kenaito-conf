import Vue from 'vue'
import VueRouter from 'vue-router'
import store from '../store'

Vue.use(VueRouter)

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue')
  },
  {
    path: '/',
    component: () => import('../views/Layout.vue'),
    redirect: '/profile/list',
    children: [
      {
        path: '/app/config/list',
        name: 'AllConfigList',
        component: () => import('../views/config/AllConfigList.vue')
      },
      {
        path: '/app/:appId/config/list',
        name: 'ConfigList',
        component: () => import('../views/config/ConfigList.vue')
      },
      {
        path: '/config/add',
        name: 'ConfigAdd',
        component: () => import('../views/config/ConfigAdd.vue')
      },
      {
        path: '/config/edit/:id',
        name: 'ConfigEdit',
        component: () => import('../views/config/ConfigEdit.vue')
      },
      {
        path: '/config/recycle',
        name: 'RecycleBin',
        component: () => import('../views/config/RecycleBin.vue')
      },
      {
        path: '/app/list',
        name: 'AppList',
        component: () => import('../views/app/AppList.vue')
      },
      {
        path: '/user/list',
        name: 'UserList',
        component: () => import('../views/user/UserList.vue')
      },
      {
        path: '/group/list',
        name: 'GroupList',
        component: () => import('../views/group/GroupList.vue')
      },
      {
        path: '/profile/list',
        name: 'ProfileList',
        component: () => import('../views/profile/ProfileList.vue')
      }
    ]
  }
]

const router = new VueRouter({
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.path !== '/login' && !token) {
    next('/login')
  } else {
    next()
  }
})

export default router
