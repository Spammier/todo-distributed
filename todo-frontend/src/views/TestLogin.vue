<template>
  <div class="test-login">
    <h1>测试登录页</h1>
    <p>如果您可以看到这个页面，说明Vue应用已正确加载。</p>
    
    <div class="card">
      <h2>快速测试登录</h2>
      <form @submit.prevent="handleTestLogin">
        <button type="submit">模拟登录</button>
      </form>
    </div>
  </div>
</template>

<script>
import { useStore } from 'vuex'
import { useRouter } from 'vue-router'

export default {
  name: 'TestLogin',
  setup() {
    const store = useStore()
    const router = useRouter()
    
    const handleTestLogin = () => {
      // 模拟一个用户登录
      const testUser = {
        id: 1,
        username: '测试用户',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
      
      // 设置用户和令牌
      store.commit('SET_USER', testUser)
      store.commit('SET_TOKEN', 'test-token-123')
      
      // 导航到首页
      router.push('/')
    }
    
    return {
      handleTestLogin
    }
  }
}
</script>

<style scoped>
.test-login {
  max-width: 600px;
  margin: 40px auto;
  text-align: center;
}

h1 {
  margin-bottom: 20px;
}

p {
  margin-bottom: 30px;
}
</style> 