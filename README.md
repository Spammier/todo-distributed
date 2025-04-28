# Todo Distributed - 分布式 Todo 应用

这是一个使用 Go 微服务和 Vue.js 构建的分布式 Todo 列表应用程序。

## 功能特性

* 用户注册、登录和密码修改
* 基于 JWT 的身份认证
* 待办事项 (Todo) 的增、删、改、查 (CRUD)
* 用户注册成功后发送欢迎邮件 (通过 RabbitMQ 异步处理)
* 使用 Docker Compose 进行容器编排

## 技术栈

* **后端微服务:** Go
  * API 网关: Gin (HTTP)
  * 用户服务: gRPC, GORM, bcrypt, JWT
  * 待办事项服务: gRPC, GORM, Redis (可选缓存)
  * 邮件服务: RabbitMQ Consumer, net/smtp
* **前端:** Vue.js 3 (Composition API), Vuex, Vue Router
* **数据库:** MySQL / MariaDB (通过 GORM 连接)
* **消息队列:** RabbitMQ
* **缓存:** Redis (可选, 用于 Todo 服务)
* **容器化:** Docker, Docker Compose
* **协议定义:** Protocol Buffers (protobuf)

## 架构概览

本项目采用微服务架构：

* `todo-frontend`: Vue.js 前端应用，通过 Nginx 提供服务。
* `api-gateway`: 作为后端服务的统一入口，处理 HTTP 请求，验证 JWT，并将请求路由到相应的 gRPC 微服务。
* `user-service`: 处理用户注册、登录、密码修改，生成和验证 JWT，并在注册成功后向 RabbitMQ 发布事件。
* `todo-service`: 处理待办事项的 CRUD 操作。
* `email-service`: 监听 RabbitMQ 上的用户注册事件，并发送欢迎邮件。
* `rabbitmq`: 消息代理，用于服务间的异步通信。
* `redis_cache`: (可选) 缓存服务。
* `db`: (外部或 Docker化) 数据库服务。

## 先决条件

* [Git](https://git-scm.com/)
* [Docker](https://www.docker.com/)
* [Docker Compose](https://docs.docker.com/compose/install/)

## 快速开始

1. **克隆仓库:**

    ```bash
    git clone <your-repository-url>
    cd todo-distributed
    ```

2. **配置环境:**
    * 复制示例环境文件：

        ```bash
        cp .env.example .env
        ```

    * 编辑 `.env` 文件，填入你自己的配置，特别是：
        * `JWT_SECRET_KEY`: 用于 JWT 签名的密钥。
        * 数据库凭证 (`DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME`)。
        * SMTP 凭证 (`SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_SENDER`)，用于发送邮件。**注意：对于 QQ 邮箱等，可能需要使用应用专用授权码，而不是登录密码。**
        * 如果使用 Docker 化的数据库，还需要配置 `MYSQL_ROOT_PASSWORD`。

3. **启动服务:**
    * 使用 Docker Compose 构建并启动所有服务：

        ```bash
        docker-compose up -d --build
        ```

    * 服务启动可能需要一些时间，特别是第一次构建镜像时。

4. **访问应用:**
    * 前端应用: 默认访问 `http://localhost` (或你配置的宿主机端口)。
    * API 网关: 默认监听 `http://localhost:8080`。
    * RabbitMQ 管理界面: 访问 `http://localhost:15672` (默认用户名: `guest`, 密码: `guest`)。

## 环境变量

项目运行依赖于根目录下的 `.env` 文件中定义的环境变量。请参考 `.env.example` 文件了解所有必需的变量及其用途。`.gitignore` 文件已配置为忽略 `.env` 文件，以确保敏感信息不会提交到版本控制中。

## 项目结构

```text
.
├── api-gateway/       # API 网关 (Gin HTTP -> gRPC)
├── docker-compose.yml # Docker Compose 配置文件
├── email-service/     # 邮件服务 (RabbitMQ Consumer)
├── .env.example       # 环境变量示例文件
├── .gitignore         # Git 忽略文件配置
├── password-hasher/   # (辅助工具) 生成 bcrypt 密码哈希
├── proto-definitions/ # Protocol Buffers 定义文件
├── README.md          # 本文件
├── todo-frontend/     # 前端 Vue.js 应用
├── todo-service/      # 待办事项微服务 (gRPC)
└── user-service/      # 用户微服务 (gRPC)
```

## 注意事项

* **数据库迁移:** 后端服务 (`user-service`, `todo-service`) 在启动时会使用 GORM 的 `AutoMigrate` 功能尝试自动创建或更新数据库表结构。
* **密码安全:** 项目使用 bcrypt 进行密码哈希，无法直接解密。如果忘记测试密码，请使用 `password-hasher` 工具生成新密码的哈希，并直接更新数据库。
