# codo golang SDK

## 背景
许多的业务需求其实有大量的重复工作, 可以使用通用的能力完成

## 目标
- [ ] 基础架构
  - [ ] 通用的 mysql 客户端支持 otel
  - [ ] 通用的 redis 客户端支持 otel
  - [ ] 通用的 http 客户端支持 otel
  - [ ] 通用的 消息队列 otel
  - [ ] 通用的 协程池 支持 otel
  - [ ] 通用的 配置 组件
  - [ ] 通用的 日志 组件
  - [ ] 通用的 otel propagator
- [ ] 工具链
  - [ ] 自动生成 HTTP 代码(proto 转 http)
  - [ ] 自动生成 MYSQL 代码(数据库转结构体)
  - [ ] LINT 检查

## 目录结构
```
.
├── CHANGELOG.md 变更日志
├── Makefile 快捷工具
├── README.md
├── adapter 适配器
│   └── kratos kratos 适配
├── app 应用层
├── client 客户端
│   └── xhttp http 客户端
├── config 配置统一处理
│   ├── config.go
│   ├── config_test.go
│   └── testdata
├── consts 常量定义
│   ├── bytes.go
│   └── consts.go
├── go.mod 
├── go.sum
├── internal # 私有包
│   └── meta # lib元数据
├── logger # 日志组件
│   ├── global.go
│   ├── helper.go
│   ├── level.go
│   ├── logger.go
│   └── std.go
├── middleware # 通用中间件
│   └── xsign.middleware.go
├── mq
├── mysql # mysql 客户端 wrapper
│   └── mysql.go
├── redis # redis 客户端 wrapper
│   └── redis.go
├── tools # 小工具
│   ├── cascmd # cas 
│   └── xsgin # sign 签名
└── xnet # 网络工具
    ├── xip # ip 工具
    └── xtls # tls 工具
```