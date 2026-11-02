-- 引入
local redis_util = require("redisx.redis_util")

local function redis_get(key)
    if type(key) ~= "string" then
        ngx.log(ngx.ERR, "Redis key: nil")
        return "", false
    end

    -- 连接 Redis
    local red, err = redis_util.get_connection()
    if not red then
        ngx.log(ngx.ERR, "Redis 连接失败: ", err)
        return "", false
    end
    
    local result, err
    result, err = red:get(key)
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


return redis_get