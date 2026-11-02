-- set_test

-- 引入
local redis_util = require("redisx.redis_util")

-- 默认 60 秒
local function redis_set(key, value, redis_type, ttl)
    if key == nil or key == "" or value == nil or value == "" then
        ngx.log(ngx.ERR, "Redis key_value: nil")
        return false
    end

    --  -- 2. 校验 set_type（白名单）
    -- local allowed_types = { NX = true, EX = true }
    -- if not allowed_types[redis_type] then
    --     ngx.log(ngx.ERR, "set_type 参数错误: 只支持 NX 或 EX, 当前为 ", redis_type or "nil")
    --     return false
    -- end

    -- 连接 Redis
    local red, err = redis_util.get_connection()
    if not red then
        ngx.log(ngx.ERR, "Redis 连接失败: ", err)
        return false
    end
    
    local result, err
    if ttl == nil then
        result, err = red:set(key, value, redis_type) -- NX
    else 
        result, err = red:set(key, value, redis_type, ttl) -- NX
    end
    -- 放回连接池
    redis_util.keepalive(red)
    -- if err then
    --     ngx.log(ngx.ERR, "Redis set 失败: ", err)
    --     return false
    -- end

    if result == ngx.null then -- 成功时返回 "OK"，失败时返回 ngx.null
        ngx.log(ngx.ERR, "Redis set 返回异常: ", result or "nil")
        return false
    end

    if result ~= "OK" then -- 成功时返回 "OK"，失败时返回 ngx.null
        ngx.log(ngx.ERR, "Redis nil 返回异常: ", result or "nil")
        return false
    end
    
    -- 存储结果到 ngx.ctx（注意：原代码里 redis_value 未定义，修正为 result）
    -- ngx.ctx.redis_value = result
    return true
end


return redis_set
