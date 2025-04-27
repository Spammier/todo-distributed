import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import axios from 'axios'

// 根据环境设置 Axios 基础 URL
// 生产环境使用 VUE_APP_API_BASE_URL，开发环境使用相对路径 '/api' 以便代理
axios.defaults.baseURL = process.env.VUE_APP_API_BASE_URL || '/api';

// 请求拦截器 - 添加token到请求头
axios.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器 - 处理401错误（token过期）
axios.interceptors.response.use(
  response => response,
  error => {
    if (error.response && error.response.status === 401) {
      // 检查是否在登录页，避免无限循环
      if (router.currentRoute.value.name !== 'login') {
        store.dispatch('logout')
        router.push('/login')
      }
    }
    return Promise.reject(error)
  }
)

// 创建Vue应用实例
const app = createApp(App)

// 注册全局axios实例
app.config.globalProperties.$axios = axios

// 注册Vuex和Router
app.use(store)
app.use(router)

// 挂载应用
app.mount('#app') 