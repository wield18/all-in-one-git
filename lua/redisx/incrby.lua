-- 引入
local redis_util = require("redisx.redis_util")

local function redis_search(key, increment)
    if key == nil or key == "" then
        ngx.log(ngx.ERR, "Redis key: nil")
        return false
    end
    if increment == nil then
        increment = 1
    end
    -- 连接 Redis
    local red, err = redis_util.get_connection()
    if not red then
        ngx.log(ngx.ERR, "Redis 连接失败: ", err)
        return nil
    end
    
    -- 根据 increment 参数动态选择操作
    local result, err
    
    result, err = red:incrby(key, increment)
    -- 放回连接池
    redis_util.keepalive(red)

    if err then
        ngx.log(ngx.ERR, "Redis incr/incrby 失败: ", err)
        return nil
    end
    
    -- 存储结果到 ngx.ctx（注意：原代码里 redis_value 未定义，修正为 result）
    -- ngx.ctx.redis_value = result
    return result
end


return redis_search