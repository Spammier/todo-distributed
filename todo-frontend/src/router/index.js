import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'
import ChangePasswordView from '../views/ChangePasswordView.vue'
import TestLogin from '../views/TestLogin.vue'
import store from '../store'

const routes = [
  {
    path: '/',
    name: 'home',
    component: HomeView,
    meta: { requiresAuth: true }
  },
  {
    path: '/login',
    name: 'login',
    component: LoginView
  },
  {
    path: '/register',
    name: 'register',
    component: RegisterView
  },
  {
    path: '/change-password',
    name: 'changePassword',
    component: ChangePasswordView,
    meta: { requiresAuth: true }
  },
  {
    path: '/test-login',
    name: 'testLogin',
    component: TestLogin
  }
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes
})

// 导航守卫
router.beforeEach((to, from, next) => {
  console.log('Navigation to:', to.path)
  console.log('Auth state:', store.getters.isAuthenticated)
  
  const isAuthenticated = store.getters.isAuthenticated
  
  if (to.meta.requiresAuth && !isAuthenticated) {
    console.log('Redirecting to login due to auth requirement')
    next('/login')
  } else if ((to.name === 'login' || to.name === 'register') && isAuthenticated) {
    console.log('Redirecting to home because already authenticated')
    next('/')
  } else {
    console.log('Proceeding with navigation')
    next()
  }
})

export default router 