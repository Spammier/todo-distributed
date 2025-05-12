<template>
  <div class="home-view">
    <h2>我的待办事项</h2>
    
    <div class="todo-container">
      <!-- 添加新待办事项 -->
      <div class="card add-todo">
        <form @submit.prevent="handleAddTodo">
          <div class="form-group">
            <label for="newTodoTitle">标题</label>
            <input 
              id="newTodoTitle"
              v-model="newTodo.title"
              type="text"
              placeholder="待办事项标题"
              required
            />
          </div>
          
          <div class="form-group">
            <label for="newTodoDescription">描述</label>
            <textarea 
              id="newTodoDescription"
              v-model="newTodo.description"
              placeholder="待办事项描述（可选）"
              rows="3"
            ></textarea>
          </div>
          
          <button type="submit" :disabled="isSubmitting">添加</button>
        </form>
      </div>
      
      <!-- 待办事项列表 -->
      <div v-if="isLoading" class="loading">
        加载中...
      </div>
      
      <div v-else-if="todos.length === 0" class="empty-state card">
        <p>暂无待办事项，快来添加吧！</p>
      </div>
      
      <div v-else>
        <!-- 批量操作栏 -->
        <div class="batch-actions card" v-if="todos.length > 0">
          <div class="select-all">
            <input 
              type="checkbox" 
              id="selectAll" 
              :checked="selectedTodos.length === todos.length && todos.length > 0"
              :indeterminate="selectedTodos.length > 0 && selectedTodos.length < todos.length"
              @change="toggleSelectAll"
            />
            <label for="selectAll">全选</label>
          </div>
          <div class="batch-buttons" v-if="selectedTodos.length > 0">
            <span class="selected-count">已选择 {{ selectedTodos.length }} 项</span>
            <button 
              class="batch-complete-btn" 
              @click="batchUpdateStatus('MARK_AS_COMPLETED')"
              :disabled="isBatchUpdating || selectedCompletedCount === selectedTodos.length"
            >
              标记为已完成
            </button>
            <button 
              class="batch-incomplete-btn" 
              @click="batchUpdateStatus('MARK_AS_INCOMPLETE')"
              :disabled="isBatchUpdating || selectedCompletedCount === 0"
            >
              标记为未完成
            </button>
          </div>
        </div>

        <div class="todo-list">
          <div 
            v-for="todo in todos" 
            :key="todo.id" 
            class="todo-item card"
            :class="{ completed: todo.completed, editing: editingTodoId === todo.id, selected: isSelected(todo.id) }"
          >
            <div v-if="editingTodoId === todo.id" class="edit-form">
              <input type="text" v-model="editFormData.title" placeholder="标题" class="edit-input-title" required>
              <textarea v-model="editFormData.description" placeholder="描述" class="edit-textarea-desc" rows="2"></textarea>
              <div class="edit-actions">
                <button @click="saveEdit(todo.id)" class="save-btn" :disabled="isSubmittingEdit">保存</button>
                <button @click="cancelEdit" class="cancel-btn">取消</button>
              </div>
              <div v-if="editError" class="error-message">{{ editError }}</div>
            </div>
            
            <div v-else>
              <div class="todo-header">
                <div class="todo-title">
                  <input 
                    type="checkbox" 
                    :checked="todo.completed"
                    @change="toggleTodoStatus(todo)"
                    class="status-checkbox"
                  />
                  <input 
                    type="checkbox" 
                    :checked="isSelected(todo.id)"
                    @change="toggleSelectTodo(todo.id)"
                    class="select-checkbox"
                  />
                  <h3>{{ todo.title }}</h3>
                </div>
                <div class="todo-actions">
                  <button 
                    class="edit-btn" 
                    @click="startEdit(todo)"
                    :disabled="isDeleting[todo.id]"
                  >
                    编辑
                  </button>
                  <button 
                    class="delete-btn" 
                    @click="deleteTodo(todo.id)"
                    :disabled="isDeleting[todo.id]"
                  >
                    删除
                  </button>
                </div>
              </div>
              <div class="todo-description">
                {{ todo.description }}
              </div>
              <div class="todo-footer">
                <small>创建于: {{ formatDate(todo.created_at) }}</small>
                <small v-if="todo.updated_at !== todo.created_at">
                  更新于: {{ formatDate(todo.updated_at) }}
                </small>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, reactive } from 'vue'
import { useStore } from 'vuex'

export default {
  name: 'HomeView',
  setup() {
    const store = useStore()
    
    const isLoading = ref(true)
    const isSubmitting = ref(false)
    const isDeleting = reactive({})
    const isSubmittingEdit = ref(false)
    const isBatchUpdating = ref(false)
    
    const todos = computed(() => store.getters.todos)
    const selectedTodos = ref([])
    
    const selectedCompletedCount = computed(() => {
      return todos.value.filter(todo => 
        selectedTodos.value.includes(todo.id) && todo.completed
      ).length
    })
    
    const newTodo = ref({
      title: '',
      description: ''
    })
    
    const editingTodoId = ref(null)
    const editFormData = reactive({ title: '', description: '' })
    const editError = ref('')
    
    // 判断待办事项是否被选中
    const isSelected = (todoId) => {
      return selectedTodos.value.includes(todoId)
    }
    
    // 切换单个待办事项的选中状态
    const toggleSelectTodo = (todoId) => {
      const index = selectedTodos.value.indexOf(todoId)
      if (index === -1) {
        selectedTodos.value.push(todoId)
      } else {
        selectedTodos.value.splice(index, 1)
      }
    }
    
    // 全选/取消全选
    const toggleSelectAll = (e) => {
      if (e.target.checked) {
        selectedTodos.value = todos.value.map(todo => todo.id)
      } else {
        selectedTodos.value = []
      }
    }
    
    // 批量更新待办事项状态
    const batchUpdateStatus = async (action) => {
      if (selectedTodos.value.length === 0) return
      
      isBatchUpdating.value = true
      try {
        await store.dispatch('batchUpdateTodos', {
          todoIds: selectedTodos.value,
          action: action
        })
        // 成功后清空选中
        selectedTodos.value = []
      } catch (error) {
        console.error('批量更新待办事项失败:', error)
      } finally {
        isBatchUpdating.value = false
      }
    }
    
    // 获取所有待办事项
    const fetchTodos = async () => {
      isLoading.value = true
      try {
        await store.dispatch('fetchTodos')
      } catch (error) {
        console.error('获取待办事项失败:', error)
      } finally {
        isLoading.value = false
      }
    }
    
    // 添加新待办事项
    const handleAddTodo = async () => {
      isSubmitting.value = true
      try {
        await store.dispatch('createTodo', {
          title: newTodo.value.title,
          description: newTodo.value.description
        })
        // 重置表单
        newTodo.value.title = ''
        newTodo.value.description = ''
      } catch (error) {
        console.error('添加待办事项失败:', error)
      } finally {
        isSubmitting.value = false
      }
    }
    
    // 切换待办事项状态
    const toggleTodoStatus = async (todo) => {
      const newCompletedStatus = !todo.completed
      try {
        await store.dispatch('updateTodo', {
          id: todo.id,
          todoData: { completed: newCompletedStatus, title: todo.title, description: todo.description }
        })
      } catch (error) {
        console.error('更新待办事项状态失败:', error)
      }
    }
    
    // 删除待办事项
    const deleteTodo = async (todoId) => {
      isDeleting[todoId] = true
      try {
        await store.dispatch('deleteTodo', todoId)
        // 如果被删除的待办事项在选中列表中，需要从选中列表中移除
        const selectedIndex = selectedTodos.value.indexOf(todoId)
        if (selectedIndex !== -1) {
          selectedTodos.value.splice(selectedIndex, 1)
        }
      } catch (error) {
        console.error('删除待办事项失败:', error)
      } finally {
        isDeleting[todoId] = false
      }
    }
    
    // 开始编辑
    const startEdit = (todo) => {
      editingTodoId.value = todo.id
      editFormData.title = todo.title
      editFormData.description = todo.description
      editError.value = ''
    }

    // 取消编辑
    const cancelEdit = () => {
      editingTodoId.value = null
    }

    // 保存编辑
    const saveEdit = async (todoId) => {
      if (!editFormData.title) {
        editError.value = '标题不能为空'
        return
      }
      isSubmittingEdit.value = true
      editError.value = ''
      try {
        const currentTodo = todos.value.find(t => t.id === todoId)
        if (!currentTodo) return

        await store.dispatch('updateTodo', {
          id: todoId,
          todoData: { 
            title: editFormData.title, 
            description: editFormData.description, 
            completed: currentTodo.completed
          }
        })
        editingTodoId.value = null
      } catch (error) {
        console.error('保存编辑失败:', error)
        editError.value = error.response?.data?.error || '保存失败，请重试'
      } finally {
        isSubmittingEdit.value = false
      }
    }
    
    // 格式化日期
    const formatDate = (dateString) => {
      const date = new Date(dateString)
      return date.toLocaleString('zh-CN')
    }
    
    onMounted(fetchTodos)
    
    return {
      isLoading,
      isSubmitting,
      isDeleting,
      isSubmittingEdit,
      isBatchUpdating,
      todos,
      selectedTodos,
      selectedCompletedCount,
      newTodo,
      handleAddTodo,
      toggleTodoStatus,
      deleteTodo,
      formatDate,
      editingTodoId,
      editFormData,
      editError,
      startEdit,
      saveEdit,
      cancelEdit,
      isSelected,
      toggleSelectTodo,
      toggleSelectAll,
      batchUpdateStatus
    }
  }
}
</script>

<style scoped>
/* 基本布局和容器 */
.home-view {
  width: 100%;
  padding: 30px 15px; /* 增加上下内边距 */
  /* --- 移除背景色，让全局 body 背景生效 --- */
  /* background-color: var(--dark-bg, #1a1a2e); */ 
  min-height: calc(100vh - 60px); /* 假设 header 高度为 60px */
}

h2 {
  margin-bottom: 30px;
  text-align: center;
  color: var(--dark-text-primary, #e0e0e0);
  font-weight: 600;
  font-size: 1.8rem; /* 稍大字体 */
}

.todo-container {
  width: 100%;
  max-width: 750px; /* 稍微加宽 */
  margin: 0 auto;
}

/* 通用卡片样式 (如果全局没有定义) */
.card {
  border-radius: var(--border-radius, 8px);
  padding: 25px; /* 增加内边距 */
  margin-bottom: 25px; /* 统一底部间距 */
}

/* 添加待办表单 */
.add-todo {
  /* 移除背景色覆盖，继承全局 .card 背景 */
  /* background-color: var(--dark-bg-primary) !important; */
  /* 设置 padding 匹配 login-card */
  padding: 40px !important; 
}

.add-todo .form-group {
  margin-bottom: 20px; /* 增加组间距 */
}

.add-todo label {
  font-weight: 500;
  margin-bottom: 8px; /* 增加标签下方间距 */
  display: block;
  font-size: 0.9rem;
}

/* 通用输入框和文本域样式 */
input[type="text"],
input[type="email"], /* 确保 email 也包含 */
input[type="password"], /* 确保 password 也包含 */
textarea {
  width: 100%;
  padding: 12px 15px; /* 增加内边距 */
  border-radius: var(--border-radius, 6px);
  /* border: 1px solid var(--dark-border-color, #40405c);
  background-color: var(--dark-input-bg, #1e1e30);
  color: var(--dark-text-primary, #e0e0e0); */
  font-size: 1rem;
  transition: border-color 0.3s ease, box-shadow 0.3s ease;
  box-sizing: border-box; /* 确保 padding 不会撑大元素 */
}

input[type="text"]:focus,
input[type="email"]:focus,
input[type="password"]:focus,
textarea:focus {
  outline: none;
  border-color: var(--primary-color, #4CAF50);
  box-shadow: 0 0 0 3px rgba(var(--primary-color-rgb, 76, 175, 80), 0.3); /* 更明显的聚焦效果 */
}

textarea {
  resize: vertical; /* 允许垂直调整大小 */
  min-height: 80px; /* 最小高度 */
}

/* 添加按钮 */
.add-todo button[type="submit"] {
  width: 100%;
  padding: 12px 20px;
  border: none;
  border-radius: var(--border-radius, 6px);
  background-color: var(--primary-color, #4CAF50);
  color: white;
  font-weight: bold;
  font-size: 1rem;
  cursor: pointer;
  transition: background-color 0.3s ease, transform 0.1s ease;
}

.add-todo button[type="submit"]:hover:not(:disabled) {
  background-color: var(--primary-color-dark, #388E3C);
}

.add-todo button[type="submit"]:active:not(:disabled) {
  transform: scale(0.98);
}

.add-todo button[type="submit"]:disabled {
  background-color: var(--dark-disabled-bg, #555);
  cursor: not-allowed;
}

/* 批量操作区域样式 */
.batch-actions {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  padding: 15px 25px !important;
}

.select-all {
  display: flex;
  align-items: center;
}

.select-all label {
  margin-left: 8px;
  font-weight: 500;
  cursor: pointer;
}

.batch-buttons {
  display: flex;
  align-items: center;
  gap: 10px;
}

.selected-count {
  margin-right: 10px;
  font-size: 0.9rem;
  color: var(--dark-text-secondary, #a0a0b3);
}

.batch-complete-btn, .batch-incomplete-btn {
  padding: 6px 12px;
  border-radius: var(--border-radius-small, 4px);
  cursor: pointer;
  font-size: 0.85rem;
  border: none;
  transition: all 0.2s ease;
}

.batch-complete-btn {
  background-color: var(--primary-color, #4CAF50);
  color: white;
}

.batch-incomplete-btn {
  background-color: var(--secondary-color, #FF9800);
  color: white;
}

.batch-complete-btn:hover:not(:disabled), .batch-incomplete-btn:hover:not(:disabled) {
  filter: brightness(1.1);
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
}

.batch-complete-btn:disabled, .batch-incomplete-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 待办事项列表 */
.todo-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.todo-item {
  /* 移除背景色覆盖，继承全局 .card 背景 */
  /* background-color: var(--dark-bg-primary) !important; */
  /* 设置 padding 匹配 login-card */
  padding: 40px !important; 
  transition: all 0.3s ease;
  position: relative; /* 为了完成时的覆盖层 */
}

.todo-item.completed {
  /* 移除背景色覆盖 */
  /* background-color: var(--dark-card-bg-completed, #252535); */ 
  /* 使用继承的背景，仅改变边框和文本样式 */
  border-left: 4px solid var(--primary-color-dark, #388E3C);
  padding-left: 36px; /* 调整左 padding 以适应边框 (40px - 4px) */
}

.todo-item.selected {
  border: 2px solid var(--accent-color, #22c55e);
}

.todo-item.completed .todo-title h3 {
  text-decoration: line-through;
  color: var(--dark-text-disabled, #777);
}
.todo-item.completed .todo-description {
  color: var(--dark-text-disabled, #777);
}


.todo-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  border-bottom: 1px solid var(--dark-border-color, #40405c);
  padding-bottom: 12px;
}

.todo-title {
  display: flex;
  align-items: center;
  gap: 12px; /* 增加 checkbox 和标题间距 */
}

.todo-title h3 {
  font-size: 1.2rem; /* 标题稍大 */
  font-weight: 500;
  margin: 0;
  color: var(--dark-text-primary, #e0e0e0);
}

.todo-description {
  margin-bottom: 15px;
  color: var(--dark-text-secondary, #a0a0b3);
  white-space: pre-line;
  line-height: 1.6; /* 增加行高 */
}

.todo-footer {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem; /* 稍小字体 */
  color: var(--dark-text-secondary, #a0a0b3);
  opacity: 0.7;
  padding-top: 10px;
  border-top: 1px dashed var(--dark-border-color, #40405c); /* 底部虚线 */
}

/* 按钮通用样式 */
.todo-actions button {
  padding: 6px 12px;
  border-radius: var(--border-radius-small, 4px);
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 0.85rem;
  margin-left: 8px; /* 按钮间距 */
  border: none;
}
.todo-actions button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 删除按钮 */
.delete-btn {
  background-color: var(--danger-color, #f44336);
  color: white;
}
.delete-btn:hover:not(:disabled) {
  background-color: var(--danger-color-dark, #d32f2f);
  box-shadow: 0 2px 5px rgba(244, 67, 54, 0.3);
}

/* 编辑按钮 */
.edit-btn {
  background-color: var(--secondary-color, #ff9800);
  color: white;
}
.edit-btn:hover:not(:disabled) {
  background-color: var(--secondary-color-dark, #f57c00);
   box-shadow: 0 2px 5px rgba(255, 152, 0, 0.3);
}

/* 编辑表单样式 */
.todo-item.editing {
  border-color: var(--primary-color, #4CAF50); /* 编辑时边框高亮 */
  box-shadow: 0 0 15px rgba(var(--primary-color-rgb, 76, 175, 80), 0.2);
}

.edit-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.edit-input-title,
.edit-textarea-desc {
  /* 继承通用输入框样式 */
  font-size: 1rem; /* 确保字体大小合适 */
}

.edit-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 10px; /* 与上方文本域间距 */
}

/* 保存按钮 */
.save-btn {
  background-color: var(--primary-color, #4CAF50);
  color: white;
}
.save-btn:hover:not(:disabled) {
  background-color: var(--primary-color-dark, #388E3C);
}

/* 取消按钮 */
.cancel-btn {
  background-color: transparent;
  color: var(--dark-text-secondary, #a0a0b3);
  border: 1px solid var(--dark-border-color, #40405c);
}
.cancel-btn:hover {
  background-color: var(--dark-hover-bg, #3a3a4e);
  border-color: var(--dark-text-secondary, #a0a0b3);
}

/* 复选框样式 - 状态和选择 */
.status-checkbox, 
.select-checkbox {
  appearance: none;
  -webkit-appearance: none;
  background-color: var(--dark-input-bg, #1e1e30);
  border: 1px solid var(--dark-border-color, #40405c);
  padding: 8px; /* 调整大小 */
  display: inline-block;
  position: relative;
  cursor: pointer;
  border-radius: 3px;
  transition: background-color 0.2s ease, border-color 0.2s ease;
  vertical-align: middle; /* 垂直居中 */
}

.status-checkbox {
  margin-right: 5px;
}

.select-checkbox {
  margin-right: 10px;
  border-color: var(--accent-color, #22c55e);
}

.status-checkbox:checked,
.select-checkbox:checked {
  background-color: var(--primary-color, #4CAF50);
  border-color: var(--primary-color, #4CAF50);
}

.select-checkbox:checked {
  background-color: var(--accent-color, #22c55e);
  border-color: var(--accent-color, #22c55e);
}

.status-checkbox:checked::after,
.select-checkbox:checked::after {
  content: '\2714'; /* 勾号 */
  font-size: 12px; /* 勾号大小 */
  color: white;
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
}

.status-checkbox:focus,
.select-checkbox:focus {
  outline: none;
  box-shadow: 0 0 0 3px rgba(var(--primary-color-rgb, 76, 175, 80), 0.3);
}

/* 加载和空状态 */
.loading, .empty-state {
  text-align: center;
  padding: 60px 20px; /* 增加内边距 */
  color: var(--dark-text-secondary, #a0a0b3);
  background-color: var(--dark-card-bg, #2a2a3e); /* 应用卡片背景 */
  border-radius: var(--border-radius, 8px);
  border: 1px dashed var(--dark-border-color, #40405c); /* 虚线边框 */
}
.empty-state p {
  font-size: 1.1rem;
}

/* 错误消息 */
.error-message {
  background-color: var(--danger-color-light, #ffebee);
  color: var(--danger-color-dark, #c62828);
  padding: 12px 15px;
  border-radius: var(--border-radius-small, 4px);
  margin-top: 10px;
  font-size: 0.9rem;
  border-left: 3px solid var(--danger-color-dark, #c62828);
}
</style> 