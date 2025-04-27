<template>
  <div class="register-view">
    <div class="card register-card">
      <h2>注册</h2>
      <form @submit.prevent="handleRegister">
        <div v-if="error" class="error-message">{{ error }}</div>
        <div v-if="success" class="success-message">{{ success }}</div>
        
        <div class="form-group">
          <label for="username">用户名</label>
          <input 
            id="username"
            v-model="username"
            type="text"
            placeholder="请输入用户名"
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
        
        <div class="form-group">
          <label for="confirmPassword">确认密码</label>
          <input 
            id="confirmPassword"
            v-model="confirmPassword"
            type="password"
            placeholder="请再次输入密码"
            required
          />
        </div>
        
        <div class="form-group">
          <label for="email">邮箱</label>
          <input 
            id="email"
            v-model="email"
            type="email"
            placeholder="请输入邮箱地址"
            required
          />
        </div>
        
        <div class="form-actions">
          <button type="submit" :disabled="loading">
            {{ loading ? '注册中...' : '注册' }}
          </button>
          <div class="login-link">
            已有账号？ <router-link to="/login">登录</router-link>
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
  name: 'RegisterView',
  setup() {
    const store = useStore()
    const router = useRouter()
    
    const username = ref('')
    const password = ref('')
    const confirmPassword = ref('')
    const email = ref('')
    const loading = ref(false)
    const error = ref('')
    const success = ref('')
    
    const handleRegister = async () => {
      // 表单验证
      if (password.value !== confirmPassword.value) {
        error.value = '两次输入的密码不一致'
        return
      }
      if (!email.value || !/.+@.+\..+/.test(email.value)) {
        error.value = '请输入有效的邮箱地址'
        return
      }
      
      loading.value = true
      error.value = ''
      success.value = ''
      
      try {
        await store.dispatch('register', {
          username: username.value,
          password: password.value,
          email: email.value
        })
        
        success.value = '注册成功！请登录'
        // 重置表单
        username.value = ''
        password.value = ''
        confirmPassword.value = ''
        email.value = ''
        
        // 3秒后跳转到登录页
        setTimeout(() => {
          router.push('/login')
        }, 3000)
      } catch (err) {
        if (err.response && err.response.data) {
          error.value = err.response.data.error || '注册失败，请重试'
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
      confirmPassword,
      email,
      loading,
      error,
      success,
      handleRegister
    }
  }
}
</script>

<style scoped>
/* 外部容器只负责居中 */
.register-view {
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

/* --- 核心：让注册卡片完全匹配登录卡片 --- */
.register-card {
  /* 保留尺寸和内边距控制 */
  width: 100%;
  max-width: 420px;
  padding: 40px;
  box-sizing: border-box;

  /* --- 移除覆盖的基础样式，让全局 .card 类生效 --- */
  /* background-color: #1E1E2F; */ /* 移除 */
  /* border-radius: var(--border-radius, 8px); */ /* 移除，应由 .card 提供 */
  /* box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2); */ /* 移除，应由 .card 提供 */
  /* border: 1px solid var(--dark-border-color, #40405c); */ /* 移除，应由 .card 提供 */
  /* --- 结束移除 --- */
}
/* --- 结束卡片匹配 --- */


/* --- 调整内部元素以匹配登录页视觉 --- */
h2 {
  text-align: center;
  margin-bottom: 30px; /* 匹配登录标题间距 */
  color: var(--dark-text-primary, #e0e0e0);
  font-size: 24px; /* 匹配登录标题大小 */
  font-weight: 600;
}

.form-group {
  margin-bottom: 20px; /* 匹配登录表单组间距 */
}

label {
  display: block;
  margin-bottom: 8px; /* 匹配登录标签间距 */
  font-weight: 500;
  font-size: 14px; /* 匹配登录标签大小 */
  color: var(--dark-text-secondary, #a0a0b3);
}

/* 输入框样式 (保持统一) */
.form-group input[type="text"],
.form-group input[type="password"],
.form-group input[type="email"] {
  width: 100%;
  box-sizing: border-box;
  padding: 10px 12px;
  border-radius: 6px;
  border: 1px solid var(--dark-border-color, #40405c);
  background-color: var(--dark-bg-tertiary, #2a2a3e); /* 使用 tertiary 背景 */
  color: var(--dark-text-primary, #e0e0e0);
  font-size: 1rem;
}

.form-group input[type="text"]:focus,
.form-group input[type="password"]:focus,
.form-group input[type="email"]:focus {
  outline: none;
  border-color: var(--primary-color, #4CAF50);
  box-shadow: 0 0 0 3px rgba(var(--primary-color-rgb, 76, 175, 80), 0.3);
}


.form-actions {
  margin-top: 30px; /* 匹配登录操作区间距 */
  /* --- 核心修改：使用 Flexbox 布局 --- */
  display: flex; 
  justify-content: space-between;
  align-items: center;
  /* --- 结束修改 --- */
}

/* 注册按钮样式 (使其与登录按钮类似 - 登录按钮不是100%宽度) */
button[type="submit"] {
  /* 移除 width: 100%; 让按钮宽度自适应内容，或设定固定宽度 */
   width: auto; /* 或者移除这行 */
  padding: 10px 30px; /* 匹配登录页按钮内边距 */
  /* 保留其他基本按钮样式 */
  border: none;
  border-radius: var(--border-radius, 6px);
  background-color: var(--primary-color, #4CAF50);
  color: white;
  font-weight: bold;
  font-size: 1rem;
  cursor: pointer;
  transition: background-color 0.3s ease, transform 0.1s ease;
}
button[type="submit"]:hover:not(:disabled) {
  background-color: var(--primary-color-dark, #388E3C);
}
button[type="submit"]:active:not(:disabled) {
  transform: scale(0.98);
}
button[type="submit"]:disabled {
  background-color: var(--dark-disabled-bg, #555);
  cursor: not-allowed;
  opacity: 0.7;
}


.login-link {
  /* --- 核心修改：移除不必要的样式 --- */
  /* margin-top: 15px; */ /* 移除，由 flex align-items 控制垂直对齐 */
  /* text-align: center; */ /* 移除，由 flex justify-content 控制水平对齐 */
  /* --- 结束修改 --- */
}

.login-link a {
  color: var(--primary-color, #4CAF50);
  text-decoration: none;
}

.login-link a:hover {
  text-decoration: underline;
}

/* 错误和成功消息样式 */
.error-message {
  background-color: rgba(244, 67, 54, 0.1);
  color: var(--danger-color, #f44336);
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
  border-left: 3px solid var(--danger-color, #f44336);
}

.success-message {
  background-color: rgba(76, 175, 80, 0.1);
  color: var(--primary-color-dark, #388E3C);
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
  border-left: 3px solid var(--primary-color-dark, #388E3C);
}

</style> 