<template>
  <div class="change-password-view">
    <div class="card">
      <h2>修改密码</h2>
      <form @submit.prevent="handleChangePassword">
        <div v-if="error" class="error-message">{{ error }}</div>
        <div v-if="success" class="success-message">{{ success }}</div>
        
        <div class="form-group">
          <label for="oldPassword">旧密码</label>
          <input 
            id="oldPassword"
            v-model="oldPassword"
            type="password"
            placeholder="请输入旧密码"
            required
          />
        </div>
        
        <div class="form-group">
          <label for="newPassword">新密码</label>
          <input 
            id="newPassword"
            v-model="newPassword"
            type="password"
            placeholder="请输入新密码"
            required
          />
        </div>
        
        <div class="form-group">
          <label for="confirmPassword">确认新密码</label>
          <input 
            id="confirmPassword"
            v-model="confirmPassword"
            type="password"
            placeholder="请再次输入新密码"
            required
          />
        </div>
        
        <div class="form-actions">
          <button type="submit" :disabled="loading">
            {{ loading ? '提交中...' : '修改密码' }}
          </button>
          <button type="button" class="cancel-btn" @click="$router.push('/')">
            取消
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import { useStore } from 'vuex'
import { useRouter } from 'vue-router'
import axios from 'axios'

export default {
  name: 'ChangePasswordView',
  setup() {
    const store = useStore()
    const router = useRouter()
    
    const oldPassword = ref('')
    const newPassword = ref('')
    const confirmPassword = ref('')
    const loading = ref(false)
    const error = ref('')
    const success = ref('')
    
    const handleChangePassword = async () => {
      // 表单验证
      if (newPassword.value !== confirmPassword.value) {
        error.value = '两次输入的新密码不一致'
        return
      }
      
      loading.value = true
      error.value = ''
      success.value = ''
      
      try {
        const response = await axios.post('/change-password', {
          old_password: oldPassword.value,
          new_password: newPassword.value
        })
        
        success.value = '密码修改成功！'
        // 重置表单
        oldPassword.value = ''
        newPassword.value = ''
        confirmPassword.value = ''
        
        // 3秒后跳转到首页
        setTimeout(() => {
          router.push('/')
        }, 3000)
      } catch (err) {
        if (err.response && err.response.data) {
          error.value = err.response.data.error || '密码修改失败，请重试'
        } else {
          error.value = '服务器错误，请稍后重试'
        }
      } finally {
        loading.value = false
      }
    }
    
    return {
      oldPassword,
      newPassword,
      confirmPassword,
      loading,
      error,
      success,
      handleChangePassword
    }
  }
}
</script>

<style scoped>
.change-password-view {
  max-width: 400px;
  margin: 40px auto;
}

h2 {
  text-align: center;
  margin-bottom: 30px; /* 匹配登录标题间距 */
  color: var(--dark-text-primary, #e0e0e0);
  font-size: 24px; /* 匹配登录标题大小 */
  font-weight: 600;
}

.form-group {
  margin-bottom: 15px;
}

label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}

.form-actions {
  margin-top: 20px;
  display: flex;
  gap: 10px;
}

.cancel-btn {
  background-color: #9e9e9e;
}

.cancel-btn:hover {
  background-color: #757575;
}

.error-message {
  background-color: #ffebee;
  color: #c62828;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
}

.success-message {
  background-color: #e8f5e9;
  color: #2e7d32;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
}
</style> 