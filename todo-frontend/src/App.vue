<template>
  <div id="app" :class="{ 'dark-theme': true }">
    <header v-if="!isLoggedIn" class="main-header">
      <img src="/logo.png" alt="DoIt Logo" class="logo-image">
    </header>
    <nav v-if="isLoggedIn && !isAuthRoute" class="navbar">
      <div class="container">
        <div class="user-greeting">欢迎, {{ currentUser.username }}!</div>
        <div class="nav-links">
          <router-link to="/">我的待办</router-link> | 
          <router-link to="/change-password">修改密码</router-link> | 
          <a @click="logout" class="logout-link">退出登录</a>
        </div>
      </div>
    </nav>
    <main class="main-content" :class="{ 'logged-in-padding': isLoggedIn }">
      <router-view />
    </main>
  </div>
</template>

<script>
import { computed } from 'vue'
import { useStore } from 'vuex'
import { useRouter, useRoute } from 'vue-router'

export default {
  name: 'App',
  setup() {
    const store = useStore()
    const router = useRouter()
    const route = useRoute()
    
    const isLoggedIn = computed(() => store.getters.isAuthenticated)
    const currentUser = computed(() => store.getters.currentUser)
    
    // 判断当前是否在登录或注册路由
    const isAuthRoute = computed(() => route.name === 'login' || route.name === 'register')
    
    const logout = () => {
      store.dispatch('logout')
      router.push('/login')
    }
    
    return {
      isLoggedIn,
      currentUser,
      isAuthRoute,
      logout
    }
  }
}
</script>

<style>
:root {
  --dark-bg-primary: #111827; /* 类 Tailwind Slate 900 - 深灰蓝 */
  --dark-bg-secondary: #1f2937; /* 类 Tailwind Slate 800 - 登录框/输入框背景 */
  --dark-bg-tertiary: #374151; /* 类 Tailwind Slate 700 - 输入框背景 alternative */
  --dark-text-primary: #f3f4f6; /* 类 Tailwind Gray 100 - 主要文字 */
  --dark-text-secondary: #9ca3af; /* 类 Tailwind Gray 400 - 次要文字/占位符 */
  --dark-border-color: #4b5563; /* 类 Tailwind Slate 600 - 边框 */
  --accent-color: #22c55e; /* 类 Tailwind Green 500 - 更亮的绿色 */
  --accent-color-hover: #16a34a; /* 类 Tailwind Green 600 */
  --error-bg: rgba(239, 68, 68, 0.1); /* 半透明错误背景 */
  --error-text: #ef4444; /* 类 Tailwind Red 500 */
  --error-border: rgba(239, 68, 68, 0.3);
  --success-bg: rgba(34, 197, 94, 0.1); /* 半透明成功背景 */
  --success-text: #22c55e; /* 类 Tailwind Green 500 */
  --success-border: rgba(34, 197, 94, 0.3);
}

* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

html, body {
  height: 100%;
}

body {
  font-family: 'Microsoft YaHei', 'Arial', sans-serif;
  line-height: 1.6;
  /* 改回径向渐变，中心稍亮，向外变暗 */
  background: radial-gradient(ellipse at center, #1f2937 0%, #111827 100%);
  color: var(--dark-text-primary);
}

#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  position: relative;
  /* 移除 #app 上的背景，让 body 的背景生效 */
}

.main-header {
  padding: 25px 40px; /* 增加上下左右 padding 以适应更大的 logo */
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  z-index: 10;
  box-sizing: border-box;
}

.logo-image {
  height: 100px; /* 将高度增加到 100px */
  display: block;
}

.navbar {
  background-color: var(--dark-bg-secondary); /* 使用次要背景色 */
  color: var(--dark-text-primary);
  padding: 15px 0;
  box-shadow: 0 2px 5px rgba(0,0,0,0.4);
}

.navbar .container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.user-greeting {
  font-weight: bold;
}

.nav-links a, .nav-links span {
  color: var(--dark-text-primary);
  text-decoration: none;
  margin-left: 15px;
}

.nav-links a:hover, .logout-link:hover {
  color: var(--accent-color);
  cursor: pointer;
  text-decoration: none;
}

.main-content {
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: 80px 20px 20px 20px;
  box-sizing: border-box;
  transition: padding-top 0.3s ease;
}

.main-content.logged-in-padding {
  padding-top: 20px;
  justify-content: flex-start;
}

/* 全局按钮样式使用新颜色 */
button {
  background-color: var(--accent-color);
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 500;
  transition: background-color 0.2s ease;
}

button:hover {
  background-color: var(--accent-color-hover);
}

button:disabled {
  background-color: #4b5563; /* 使用新的边框/禁用色 */
  color: var(--dark-text-secondary);
  cursor: not-allowed;
}

/* 全局输入框样式使用新颜色 */
input[type="text"],
input[type="password"],
textarea {
  width: 100%;
  padding: 12px 15px;
  margin-bottom: 10px;
  border: 1px solid var(--dark-border-color);
  border-radius: 6px;
  background-color: var(--dark-bg-tertiary); /* 使用更深的第三背景色 */
  color: var(--dark-text-primary);
  font-size: 16px;
}

input::placeholder,
textarea::placeholder {
  color: var(--dark-text-secondary);
}

/* 全局卡片样式使用新颜色 */
.card {
  background: var(--dark-bg-secondary);
  border-radius: 12px;
  box-shadow: 0 8px 25px rgba(0,0,0,0.5);
  padding: 30px 40px;
  border: 1px solid var(--dark-border-color); /* 使用新边框色 */
}

/* 消息提示样式使用新颜色 */
.error-message {
  background-color: var(--error-bg);
  color: var(--error-text);
  padding: 10px 15px;
  border-radius: 6px;
  margin-bottom: 20px;
  border: 1px solid var(--error-border);
  text-align: center;
  font-size: 14px;
}

.success-message {
  background-color: var(--success-bg);
  color: var(--success-text);
  padding: 10px 15px;
  border-radius: 6px;
  margin-bottom: 20px;
  border: 1px solid var(--success-border);
  text-align: center;
  font-size: 14px;
}

/* Todo Item 样式使用新颜色 */
.todo-item {
  border: 1px solid var(--dark-border-color);
}

.todo-item.completed {
  background-color: var(--dark-bg-tertiary);
  border-left: 5px solid var(--accent-color);
}

.todo-item.completed h3 {
  color: var(--dark-text-secondary);
}

.delete-btn {
  background-color: #ef4444; /* Red 500 */
  padding: 5px 10px;
  font-size: 14px;
}

.delete-btn:hover {
  background-color: #dc2626; /* Red 600 */
}

input[type="checkbox"] {
  width: 18px;
  height: 18px;
  margin-right: 10px;
  accent-color: var(--accent-color);
  background-color: var(--dark-bg-tertiary);
  border: 1px solid var(--dark-border-color);
}

</style> 