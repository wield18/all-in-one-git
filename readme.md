这是一个简单的 user-video-comment 服务。

下面只说明部分文件、对应 html 部分的对应关系, server.txt, config.yaml示例

个人页面地址：http://122.51.184.238/ ，个人认为 域名 跟 HTTPS 只是真正上线必备但只是配一下就行，但目前我给路人也只有一个页面能访问 /video/RandomRecommendVideo，可以看视频，其他功能都没了。想闲着没事想试试其他的,问问作者

其实主要还是想求职用,
./pkg/consumer,个人实现感觉不错,虽说我没写它的监控部分,但效果可以,具体使用见./pkg/consumer/readme.md
网关使用openresty,编写部分lua
路由结构统一使用gin的中间件结构,请求写入统一 middleware.Result 写入
前端页面是直接deep生成的模板加简单配置

**目录说明**

- `cmd`：一套直接测试使用的文件
- `config`：所有结构体
- `internal`：所有内部服务
  - `base`：基础初始化示例
    - `constant`：所有常量
  - `consume_router`：二层消费路由，其实就是自己写的异步执行的任务
  - `entity`：所有结构体
  - `local`：基于 base 继续初始化的实例，全局使用
  - `middleware`：中间件
  - `repository`：对内服务交互的部分
  - `router`：路由
  - `service`：主要逻辑部分
- `pkg`：外部服务或自研服务
  - `consumer`：一个异步批量执行任务
  - `storage`：存储服务
  - `ekit`：实用工具部分
    - `captcha`：验证码图片
    - `cipherx`：密码哈希
    - `http_err`：基础的携带 http code 的 err
    - `randx`：随机字符
    - `slogx`：日志使用
    - `validatorx`：验证器
  - `emailqq`：qq 邮箱
  - `es`：前置查询库
  - `mysqlx`：数据库
  - `redisx`：查询及临时数据
- `server.txt`: 示例systemctl服务,启动者确保有权限

**目录结构**

```
all-in-one-git/
├── cmd/
│   ├── http_server/
│   │   ├── main.go
│   │   └── main_test.go
│   ├── online_server/
│   │   └── main.go
│   ├── ram_server/
│   │   ├── constant.go
│   │   ├── copy.go
│   │   ├── log.log
│   │   ├── main.go
│   │   ├── main_test.go
│   │   └── state.log
│   └── test_setver/
│       ├── internal/
│       │   └── router.go
│       ├── test/
│       ├── main.go
│       └── main_test.go
├── config/
│   └── config.go
├── internal/
│   ├── base/
│   │   ├── constant/
│   │   │   └── constant.go
│   │   ├── no_use/
│   │   │   └── abc.go.txt
│   │   ├── test/
│   │   │   └── abc_test.go
│   │   ├── apis.go
│   │   ├── err.go
│   │   ├── func.go
│   │   ├── geoip.go
│   │   ├── keys.go
│   │   ├── re_validate.go
│   │   ├── slice.go
│   │   ├── slog.go
│   │   └── video.go
│   ├── consume_router/
│   │   ├── consume_testing.go
│   │   └── router.go
│   ├── entity/
│   │   ├── migrate/
│   │   │   └── main.go
│   │   ├── validate/
│   │   │   └── validate.go
│   │   ├── common.go
│   │   ├── common_func.go
│   │   ├── copy.go
│   │   ├── func.go
│   │   ├── redis.go
│   │   ├── type.go
│   │   ├── user.go
│   │   ├── user_db_options.go
│   │   ├── user_trasmit.go
│   │   ├── video.go
│   │   ├── video_db_options.go
│   │   └── video_transmit.go
│   ├── local/
│   │   ├── count_ticker/
│   │   │   ├── active.go
│   │   │   └── view.go
│   │   ├── token_storage/
│   │   │   ├── err.go
│   │   │   └── token_storage.go
│   │   ├── video_buf/
│   │   │   ├── es_search.go
│   │   │   ├── option.go
│   │   │   ├── video_buf.go
│   │   │   └── video_buf_test.go
│   │   └── apis.go
│   ├── middleware/
│   │   ├── test/
│   │   │   └── abc.go
│   │   ├── bind.go
│   │   ├── common.go
│   │   └── err.go
│   ├── repository/
│   │   ├── test/
│   │   │   └── test_test.go
│   │   ├── copy.go
│   │   ├── emailqq.go
│   │   ├── es.go
│   │   ├── mysql.go
│   │   ├── redis.go
│   │   ├── struct.go
│   │   ├── user_mysql.go
│   │   └── view_mysql.go
│   ├── router/
│   │   ├── router.go
│   │   └── wraps.go
│   └── service/
│       ├── comment.go
│       ├── common.go
│       ├── copy.go
│       ├── func.go
│       ├── ierr.go
│       ├── other.go
│       ├── user.go
│       └── video.go
├── lua/
│   └── redisx/
│       ├── append.lua
│       ├── client.lua
│       ├── del.lua
│       ├── errors.lua
│       ├── get.lua
│       ├── get_and_set.lua
│       ├── get_or_mget.lua
│       ├── incr.lua
│       ├── incrby.lua
│       ├── mget.lua
│       ├── mset.lua
│       ├── pcall.lua
│       ├── print_result.lua
│       ├── redis_util.lua
│       ├── set.lua
│       ├── set_all.lua
│       ├── set_ex_or_nx.lua
│       ├── set_test.lua
│       ├── setex.lua
│       ├── setnx.lua
│       ├── validate_ZADD.lua
│       ├── validate_base_func.lua
│       ├── validate_not_empty_string.lua
│       ├── with_del.lua
│       ├── with_get.lua
│       ├── with_list.lua
│       ├── with_set.lua
│       ├── with_sorted_set.lua
│       └── with_string_set.lua
├── pkg/
│   ├── consumer/
│   │   ├── constant.go
│   │   ├── consume_data.go
│   │   ├── consume_func.go
│   │   ├── consumer.go
│   │   ├── consumer_builder.go
│   │   ├── consumer_pool.go
│   │   ├── consumer_pool_test.go
│   │   ├── consumer_test.go
│   │   ├── err.go
│   │   ├── readme.md
│   │   ├── signal_pool.go
│   │   ├── signal_pool_builder.go
│   │   └── signal_pool_test.go
│   ├── ekit/
│   │   ├── captcha/
│   │   │   ├── captcha.go
│   │   │   └── test_test.go
│   │   ├── cipherx/
│   │   │   ├── cipher.go
│   │   │   └── cipher_test.go
│   │   ├── http_err/
│   │   │   ├── constant.go
│   │   │   ├── err.go
│   │   │   └── err_test.go
│   │   ├── randx/
│   │   │   └── rand_code.go
│   │   ├── slogx/
│   │   │   └── slog.go
│   │   └── validatorx/
│   │       ├── validator.go
│   │       └── validator_test.go
│   ├── emailqq/
│   │   └── emailqq.go
│   ├── es/
│   │   └── es.go
│   ├── mysqlx/
│   │   ├── mysql.go
│   │   └── slog.go
│   ├── redisx/
│   │   ├── hook.go
│   │   └── redis.go
│   └── storage/
│       ├── abc.go
│       ├── abc_test.go
│       └── storages.go
├── trees/
│   └── main.go
├── LICENSE
├── go.mod
├── go.sum
├── readme.md
├── schema_only.sql
└── server.txt
```
