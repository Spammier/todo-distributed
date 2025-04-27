<template>
  <div class="login-container">
    <div class="login-card card">
      <h2 class="title">登录</h2>
      
      <form @submit.prevent="handleLogin">
        <div v-if="error" class="error-message">{{ error }}</div>
        
        <div class="form-group">
          <label for="username">账号</label>
          <input 
            id="username"
            v-model="username"
            type="text"
            placeholder="请输入账号"
            required
          />
        </div>
        
        <div class="form-group">
          <label for="password">密码</label>
          <input 
            id="password"
            v-model="password"
            type="password"
            placeholder="请输入密码"
            required
          />
        </div>
        
        <div class="form-actions">
          <button type="submit" class="login-button" :disabled="loading">
            {{ loading ? '登录中...' : '登录' }}
          </button>
          <div class="register-prompt">
            还没有账号？ <router-link to="/register">注册</router-link>
          </div>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import { useStore } from 'vuex'
import { useRouter } from 'vue-router'

export default {
  name: 'LoginView',
  setup() {
    const store = useStore()
    const router = useRouter()
    
    const username = ref('')
    const password = ref('')
    const loading = ref(false)
    const error = ref('')
    
    const handleLogin = async () => {
      loading.value = true
      error.value = ''
      
      try {
        await store.dispatch('login', {
          username: username.value,
          password: password.value
        })
        router.push('/')
      } catch (err) {
        if (err.response && err.response.data) {
          error.value = err.response.data.error || '登录失败，请重试'
        } else {
          error.value = '服务器错误，请稍后重试'
        }
      } finally {
        loading.value = false
      }
    }
    
    return {
      username,
      password,
      loading,
      error,
      handleLogin
    }
  }
}
</script>

<style scoped>
.login-container {
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.login-card {
  width: 100%;
  max-width: 420px; /* 调整卡片最大宽度 */
  padding: 40px; /* 增加内边距 */
}

.title {
  text-align: center;
  margin-bottom: 30px; /* 增加标题下方间距 */
  color: var(--dark-text-primary);
  font-size: 24px; /* 调整字体大小 */
  font-weight: 600;
}

.form-group {
  margin-bottom: 20px; /* 增加表单组间距 */
}

label {
  display: block;
  margin-bottom: 8px; /* 增加标签下方间距 */
  font-weight: 500;
  font-size: 14px; /* 调整标签字体大小 */
  color: var(--dark-text-secondary);
}

input {
  /* 使用新颜色 */
  background-color: var(--dark-bg-tertiary); 
  border-color: var(--dark-border-color);
}

.form-actions {
  margin-top: 30px; 
  display: flex; /* 改回 flex */
  justify-content: space-between; /* 两端对齐 */
  align-items: center; /* 垂直居中 */
}

.login-button {
  /* 移除 width: 100% */
  padding: 10px 30px; /* 根据需要调整按钮大小 */
}

.register-prompt {
  margin-top: 0; /* 移除上方间距 */
  text-align: right; /* 右对齐 */
}

.register-prompt a {
  color: var(--accent-color); /* 确保链接是绿色 */
}

.error-message {
  background-color: var(--error-bg);
  color: var(--error-text);
  border: 1px solid var(--error-border);
}
</style> 