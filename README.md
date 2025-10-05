# 设计思路
## 0. 思考路径
#### 1.首先接到此任务后我就开始快速使用ai创建了一个go的学习项目来快速学习go的后端开发思路以及语法规则，还有gin的简单使用（此快速学习项目已上传到了github：https://github.com/Charlunn/go_tutorial ）
#### 2.初步学习完后就开始考虑go的后端开发分层架构是如何的然后与ai进行探讨生成了一个go_start项目（已上传至github：https://github.com/Charlunn/go-starter-template ）然后就开始了项目的开发
#### 3.随后就是先看合同思考数据库选型，设计数据库表结构然后使用我的go_start项目作为基础将思路说给了ai然后让ai进行了第一轮的快速原型，直到能够正常启动，就开始修修补补，完善api和功能
#### 4.待到第一次e2e全绿后我就开始打磨项目并设置了actions来实现ci/cd（注意此时我发现了e2e执行过后数据会被保存因为持久化了所以不能每次都全绿，因此我将dockercompose文件分成了dev和prod版本，在make docker-up时候会默认启动dev容器，此状态下并不会挂载卷，也就是说数据不会被持久化储存，若需要持久化存储数据需要使用make docker-up ENV=prod命令来启动生产环境容器），然后从main迁出了dev作为开发分支尝试从那时开始实现git flow工作流
#### 5.最后就是额外写了一个简单的前端页面来进行交互然后不断地修bug完善项目
## 1. 数据库选型和设计

在一开始，本项目就决定选用一款关系型数据库使用，但最初我是选用了mysql使用，因为其口碑以及稳定性还有我个人对其比较熟悉的原因，但在开发过程中我发掘在处理多json数据上postgresql的表现要大大优于mysql所以就选择了postgre作为数据库。
### 表结构

**`movies` 表**
```sql
CREATE TABLE IF NOT EXISTS movies (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL UNIQUE,
    genre TEXT NOT NULL,
    release_date DATE NOT NULL,
    distributor TEXT,
    budget BIGINT,
    mpa_rating TEXT,
    box_office JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```
此表是整个系统的关键，我在设计时就选用了uuid的方式作为主键，避免了自增id可能会产生的一些问题，包括对于电影的存储也不应使用自增（主要我感觉自增维护难）
并且使用了带时区的时间戳，避免了时区不同导致的问题

**`ratings` 表**
```sql
CREATE TABLE IF NOT EXISTS ratings (
    movie_id UUID NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    rater_id TEXT NOT NULL,
    rating DECIMAL(2, 1) NOT NULL CHECK (rating >= 0.5 AND rating <= 5.0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (movie_id, rater_id)
);
```
使用了复合主键这样可以避免单一用户对于同一部电影能够刷分的问题，且为了保证评分值的精确采用了decimal类型进行存储评分



## 2. 后端服务选型和设计

项目专门采用了本人并未学习过的go语言，但在最初的学习过程中我就发现go语言有点像是python对于java的地位，go在我初期的理解中就是一个翻版的c，保留了c的诸多优点并解决了c的一些不足，且整体语法就如python一样简单，并且听说还有并发功能非常强大，所以借此机会多学一门语言，这也就是为什么选用go作为后端语言，gin作为框架的原因，不过选用gin是因为在学了原生go后端开发后察觉gin还要方便不少就选用了gin。
### 架构分层
那么以下的架构分层则是我学习go的过程中了解到的go社区后端开发的最常用的一种框架，并且在此之上对数据库的连接也进行了封装，但大体还是分为这三层。
1.  **Handler (处理层)**: 如其名仅仅包含处理方法，负责接受对应的请求和调用相关的service层逻辑，如springboot的controller层

2.  **Service (服务层)**: 包含整个应用的核心业务逻辑，比如电影创建和调取外部api等操作然后调用repository层进行数据交互

3.  **Repository (仓库层)**: 就是个专门与数据交互的仓库吧，增删改查就在这写供服务层调用

---
在这个小项目的制作过程中没有一个环节不在感叹go+gin的好用！
## 3. 项目优化思考
当然项目目前只是初步完成的状态，只是在e2e全部跑绿之后我额外使用了自己的一台服务器为仓库提供了cicd操作来实现每一次提交自动构建并触发e2e操作，但现在的ci只是简单的构建校验和处于开发环境下的e2e测试验证。
然后就是为项目额外写了一个前端页面来进行简单的交互，这也就是我额外做的一些小事
### 那么我认为未来的项目优化方向有以下几点
#### 1. 完善ci/cd
我认为在未来可以加入更多的单元测试以及lint等操作来验证代码质量。并在cd阶段是当然要专门制定一台生产服务器来部署（因为现在的方案是若要启动开发服务器进行e2e测试就得终结之前的生产服务器。当然可以通过端口来分别部署，但我想最好是分离测试与生产服务器）
#### 2. 增加缓存
可以引入redis等缓存中间件来对高频访问的电影数据进行缓存以降低服务器压力
#### 3. 加入更多的功能来完善系统
比如完善用户系统，然后加入jwt来鉴权，
#### 4. 性能优化
就是防抖节流啥的

## 4. 总结
总的来说这个项目算是一个学习go语言的契机，丰富了个人技术积累的情况下也进一步锻炼了自己的整个项目的开发及规划能力。进一步的理解了后端开发工作流及actions来实现cicd的流程。不管是否被贵司录用，这将都是一段宝贵的经历。