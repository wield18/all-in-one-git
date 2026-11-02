-- 引入
local redis_util = require("redisx.redis_util")

-- 返回 number or nil
-- nil 表示错误
local function redis_del(key, ...)
    if type(key) ~= "string" and key ~= "" then
        ngx.log(ngx.ERR, "Redis type wrong or nil")
        return nil
    end

    -- 1. 先检查所有参数是否都是字符串
    local args = { ... }
    for i, v in ipairs(args) do
        if type(v) ~= "string" then
            ngx.log(ngx.ERR, "参数错误: 第 ", i, " 个参数不是字符串, 类型为 ", type(v))
            return nil
        end
    end

    -- 连接 Redis
    local red, err = redis_util.get_connection()
    if not red then
        ngx.log(ngx.ERR, "Redis 连接失败: ", err)
        return nil
    end
    
    local result, err
    result, err = red:del(key, table.unpack(args)) 
    -- 放回连接池
    redis_util.keepalive(red)

    if err then
        ngx.log(ngx.ERR, "Redis GET失败: ", err)
        return nil
    end
    
    return result
end


return redis_del