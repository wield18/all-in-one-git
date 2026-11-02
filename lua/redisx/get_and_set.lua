-- 引入
local redis_util = require("redisx.redis_util")

local function redis_get_and_set(key, new_value)
    if type(key) ~= "string" or type(new_value) ~= "string" then
        ngx.log(ngx.ERR, "Redis type wrong")
        return "", false
    end

    -- 连接 Redis
    local red, err = redis_util.get_connection()
    if not red then
        ngx.log(ngx.ERR, "Redis 连接失败: ", err)
        return "", false
    end
    
    local result, err
    result, err = red:getset(key, new_value) -- 获得旧值
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