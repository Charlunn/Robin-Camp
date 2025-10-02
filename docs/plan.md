# 项目执行计划（基于 Gin）— 不修改原有测试用例

强制约束
- 不得修改现有测试用例（包括 e2e 测试、接口结构、状态码、错误信息、Header 名称与大小写、Content-Type）。
- 新功能采用增量扩展与向后兼容策略；默认配置保持现有行为以确保现有测试稳定通过。

第一阶段：API 内存实现（基于 Gin）
1) 契约与文档
   - 使用 swag 注释 + gin-swagger 生成文档。
   - 要求：与现有接口保持兼容，不影响现有 e2e 测试；新增接口走新 path（如 /movies）。

2) 基础设施（不破坏现状，增量添加）
   - 中间件：gin.Logger()、gin.Recovery()（如已存在，保持不变）
   - 增量补充（默认安全配置）：
     - 请求 ID：X-Request-ID（透传或自动生成），仅在日志与响应 data 内附带，不改变现有响应字段结构
     - CORS：gin-contrib/cors，根据环境变量开启，默认与现有行为一致
     - 统一响应与错误码：新增在新接口上使用，避免改动旧接口返回体

3) 分层与模型（不影响旧逻辑）
   - model：新增 Movie、Rating（与 DB 解耦）
   - repository：新增内存实现（map + sync.RWMutex），接口稳定，便于后续替换为 DB
   - service：业务（校验、Upsert 规则、平均值计算）
   - handler：只做解析/返回；旧 handler 保持不变，新接口与新 handler 独立

4) API（内存版，新增接口不影响现有）
   - POST /movies：创建电影（title 唯一，201/409）
   - GET /movies：列表（首版可不含分页/搜索）
   - POST /movies/{title}/ratings：上报评分（Upsert 主键：movie_title + X-User-ID）
   - GET /movies/{title}/rating：聚合查询（avg、count）
   - 统一使用 net/http 状态码常量（http.StatusCreated 等）

5) 可观测性（不侵入旧路径）
   - /healthz 健康检查
   - /metrics 暴露 Prometheus 基础指标（HTTP 维度），不改变旧接口

6) 测试策略（不修改既有用例）
   - service 单测：基于内存仓储
   - handler 轻量 e2e（httptest）：仅覆盖新增接口；现有 e2e 不动

第二阶段：持久化与数据库（不破坏接口层）
7) 迁移与 Schema（新增迁移，不改旧迁移）
   - movies(id PK, title UNIQUE, created_at)
   - ratings(id PK, movie_id FK, user_id, score, created_at, updated_at)
   - UNIQUE(movie_id, user_id)
   - 按现有迁移工具新增 000002_xxx.sql，不修改旧迁移

8) 数据库集成（可配置切换，默认内存）
   - 通过环境变量切换 repository：memory / db
   - MySQL（默认）或 PostgreSQL（如后续切换）
   - 连接池与超时设置从环境变量读取
   - Upsert：MySQL 使用 ON DUPLICATE KEY UPDATE

9) 重构业务（对外接口不变）
   - 保持 repository 接口不变，仅替换实现
   - 聚合查询：SELECT AVG(score), COUNT(*) 方案优先
   - 错误映射：将唯一约束等 DB 错误映射为 CONFLICT/NOT_FOUND 等统一错误码

第四阶段：容器化与配置（新增文件，不动旧接口）
10) Dockerfile
    - 多阶段构建（alpine → distroless/非 root）
    - 只拷贝必要文件；避免引入对旧运行方式的破坏

11) docker-compose.yml
    - services：app、db、（可选）migrations
    - 健康检查、网络、卷；可选自动迁移
    - 保持端口与现有 e2e 使用端口不冲突

12) 环境变量
    - 新增 .env.example（APP_PORT、DB_DSN、LOG_LEVEL、ENV、CORS 源等）
    - 程序读取 env；默认值保持与当前行为兼容

交付物清单（新增为主）
- 新接口的 Swagger 文档（swag 注释或 openapi.yml）
- 统一响应结构与错误码枚举（仅用于新增接口）
- Gin 中间件：请求 ID、CORS、错误映射（以新增路由组方式接入）
- handler/service/repository 骨架与内存实现（新增）
- 数据库迁移 SQL（movies、ratings），DB repository 实现（新增）
- Dockerfile、docker-compose.yml、.env.example（新增）
- 测试：新增接口的单测/httptest；不修改既有 e2e

备注
- 如现有 e2e 覆盖 users 相关逻辑，保持 users 代码路径与接口不变。
- 新增 movies 相关代码位于独立文件与路由组，避免影响旧逻辑。
