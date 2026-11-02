-- 引入
local redis_util = require("redisx.redis_util")

local function redis_setExOrNx(key, value, redis_type, ttl)
    -- 1. 校验 key 和 value（必须是非空字符串）
    if type(key) ~= "string" or type(value) ~= "string" or
     key == "" or value == "" then
        ngx.log(ngx.ERR, "Redis key 或 value 无效, key=", tostring(key), ", value=", tostring(value))
        return false
    end

    -- 2. 校验 redis_type（白名单）
    local allowed_types = { NX = true, EX = true }
    if not allowed_types[redis_type] then
        ngx.log(ngx.ERR, "redis_type 参数错误: 只支持 NX 或 EX, 当前为 ", redis_type or "nil")
        return false
    end

    -- 连接 Redis
    local red, err = redis_util.get_connection()
    if not red then
        ngx.log(ngx.ERR, "Redis 连接失败: ", err)
        return false
    end

    -- 3. 执行 SET 命令
    local result, err
    if ttl == nil then
        result, err = red:set(key, value, redis_type)
    elseif type(ttl) == "number" then
        result, err = red:set(key, value, redis_type, tostring(ttl)) -- 可以使用 string
    else
        ngx.log(ngx.ERR, "ttl 参数类型错误: 期望 number, 实际为 ", type(ttl))
        redis_util.keepalive(red)
        return false
    end

    -- 放回连接池（即使在错误分支也执行了 keepalive）
    redis_util.keepalive(red)

    -- 4. 检查执行结果
    if err then
        ngx.log(ngx.ERR, "Redis SET 命令执行失败: ", err)
        return false
    end

    -- NX 如果要nil检查使用 if result == ngx.null then 因为 NX 会默认先 GET 下
    if result ~= "OK" then 
        ngx.log(ngx.ERR, "Redis SET 返回异常: ", result or "nil")
        return false
    end

    return true
end

return redis_setExOrNx