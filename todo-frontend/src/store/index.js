import { createStore } from 'vuex'
import axios from 'axios'

// 不再从 localStorage 加载 user，只加载 token
// const savedUser = localStorage.getItem('user')
const savedToken = localStorage.getItem('token')

export default createStore({
  state: {
    // user: savedUser ? JSON.parse(savedUser) : null, // 移除 user 加载
    user: null, // 初始 user 为 null
    token: savedToken || null,
    todos: []
  },
  getters: {
    isAuthenticated: state => !!state.token,
    // 如果 state.user 为 null，返回默认游客信息
    currentUser: state => state.user || { username: '游客' },
    todos: state => state.todos
  },
  mutations: {
    SET_USER(state, userPayload) {
      state.user = userPayload
      // 不再将 user 保存到 localStorage
    },
    SET_TOKEN(state, token) {
      state.token = token
      localStorage.setItem('token', token)
    },
    CLEAR_AUTH(state) {
      state.user = null // 确保登出时清空 user
      state.token = null
      localStorage.removeItem('token')
    },
    SET_TODOS(state, todos) {
      state.todos = todos
    },
    ADD_TODO(state, todo) {
      state.todos.push(todo)
    },
    UPDATE_TODO(state, updatedTodo) {
      const index = state.todos.findIndex(todo => todo.id === updatedTodo.id)
      if (index !== -1) {
        state.todos.splice(index, 1, updatedTodo)
      }
    },
    REMOVE_TODO(state, todoId) {
      state.todos = state.todos.filter(todo => todo.id !== todoId)
    }
  },
  actions: {
    // 用户登录
    async login({ commit }, credentials) {
      try {
        // 请求 API 网关的 /api/login
        const response = await axios.post('/login', credentials)
        // 响应现在包含 { token: '...', username: '...' }
        commit('SET_TOKEN', response.data.token)
        // 提交 SET_USER，payload 为 { username: response.data.username }
        commit('SET_USER', { username: response.data.username })
        return response
      } catch (error) {
        throw error
      }
    },

    // 用户注册
    async register({ commit }, userData) { // userData 现在应包含 email
      try {
        // 请求 API 网关的 /api/register
        const response = await axios.post('/register', userData)
        // 注册成功后，API 网关不返回 token 或 user
        // 不需要 commit 任何东西
        return response
      } catch (error) {
        throw error
      }
    },

    // 退出登录
    logout({ commit }) {
      commit('CLEAR_AUTH')
      // 可选：可以在这里清除其他状态，例如 todos
      // commit('SET_TODOS', [])
    },
    
    // 获取所有待办事项
    async fetchTodos({ commit }) {
      try {
        const response = await axios.get('/todos')
        commit('SET_TODOS', response.data)
        return response
      } catch (error) {
        throw error
      }
    },
    
    // 创建待办事项
    async createTodo({ commit }, todoData) {
      try {
        // 移除或注释掉单个/批量的判断逻辑，因为 HomeView 只发送单个对象
        let response
        // if (Array.isArray(todoData)) {
          // response = await axios.post('/todos', { todos: todoData })
          // // 更新状态
          // if (response.data.todos) {
          //   response.data.todos.forEach(todo => {
          //     commit('ADD_TODO', todo)
          //   })
          // }
        // } else {
        // 直接发送 todoData 作为请求体
        response = await axios.post('/todos', todoData)
        // 更新状态
        commit('ADD_TODO', response.data)
        // }
        return response
      } catch (error) {
        throw error
      }
    },
    
    // 更新待办事项
    async updateTodo({ commit, state }, { id, todoData }) {
      // 先找到当前待办项
      const existingTodo = state.todos.find(todo => todo.id === id)
      if (!existingTodo) return Promise.reject(new Error('待办事项未找到'))
      
      try {
        const response = await axios.put(`/todos/${id}`, todoData)
        // 成功后用服务器返回的数据更新
        commit('UPDATE_TODO', response.data)
        return response
      } catch (error) {
        // 出错时，不进行任何更新，让组件自己处理视图更新的回滚
        throw error
      }
    },
    
    // 删除待办事项
    async deleteTodo({ commit }, todoId) {
      try {
        await axios.delete(`/todos/${todoId}`)
        commit('REMOVE_TODO', todoId)
      } catch (error) {
        throw error
      }
    }
  }
}) 