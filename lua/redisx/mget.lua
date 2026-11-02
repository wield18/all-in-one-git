-- 引入
local redis_util = require("redisx.redis_util")

local function redis_get_and_set(key, ...)
    if type(key) ~= "string" and key ~= "" then
        ngx.log(ngx.ERR, "Redis type wrong or nil")
        return "", false
    end

    -- 1. 先检查所有参数是否都是字符串
    local args = { ... }
    for i, v in ipairs(args) do
        if type(v) ~= "string" then
            ngx.log(ngx.ERR, "参数错误: 第 ", i, " 个参数不是字符串, 类型为 ", type(v))
            return false
        end
    end

    -- 连接 Redis
    local red, err = redis_util.get_connection()
    if not red then
        ngx.log(ngx.ERR, "Redis 连接失败: ", err)
        return "", false
    end
    
    local result, err
    result, err = red:mget(key, table.unpack(args)) -- table
    -- 放回连接池
    redis_util.keepalive(red)

    if err then
        ngx.log(ngx.ERR, "Redis GET失败: ", err)
        return "", false
    end

    if result == ngx.null then
        ngx.log(ngx.ERR, "Redis 没有查询到: ", key)
        return "", false
    end
    
    return result, true
end


return redis_get_and_set