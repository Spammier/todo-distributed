# Todo 应用前端

基于 Vue 3, Vuex 4, Vue Router 4 和 Axios 的待办事项管理应用前端

## 项目功能

- **用户认证**
  - 用户注册
  - 用户登录
  - 修改密码
  - 基于 JWT 的身份验证
- **待办事项管理**
  - 查看所有待办事项列表
  - 添加新的待办事项
  - 切换待办事项的完成状态
  - 删除待办事项

## 项目设置

1. **安装依赖**

    ```bash
    npm install
    ```

2. **配置环境变量**
    - 复制 `.env.example` (如果存在) 或创建一个新的 `.env` 文件。
    - 在 `.env` 文件中设置以下变量：

      ```dotenv
      # 后端 API 的完整基础 URL
      VUE_APP_API_BASE_URL=http://localhost:8080/api
      
      # 后端服务器的地址 (用于开发服务器代理)
      VUE_APP_API_PROXY_TARGET=http://localhost:8080
      ```

    - **注意**: `.env` 文件已被添加到 `.gitignore`，不会提交到版本库。

3. **启动开发服务器**

    ```bash
    npm run serve
    ```

    应用将在 `http://localhost:8080` (或其他可用端口) 运行。

4. **构建生产版本**

    ```bash
    npm run build
    ```

## 环境变量

本项目使用 `.env` 文件来管理环境变量。主要变量包括：

- `VUE_APP_API_BASE_URL`: 指定 Axios 请求的基础 API URL。
- `VUE_APP_API_PROXY_TARGET`: 指定开发服务器代理的目标后端服务器地址。

在部署到生产环境时，需要在部署环境中设置这些环境变量，指向生产环境的 API 地址。

## 注意事项

- **后端依赖**: 此前端应用需要一个兼容的后端 API 服务。请确保后端服务已启动并可通过 `VUE_APP_API_BASE_URL` 访问。
- **开发服务器**: 开发服务器使用了代理来解决跨域问题。所有 `/api` 开头的请求会被代理到 `VUE_APP_API_PROXY_TARGET`。
- **浏览器兼容性**: 主要针对现代浏览器开发。
