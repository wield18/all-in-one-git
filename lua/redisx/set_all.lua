-- 引入
local redis_util = require("redisx.redis_util")

local function redis_setAll(key, value, redis_type, ...)
    -- 先检查所有参数是否都是字符串
    local args = { ... }
    for i, v in ipairs(args) do
        if type(v) ~= "string" then
            ngx.log(ngx.ERR, "参数错误: 第 ", i, " 个参数不是字符串, 类型为 ", type(v))
            return false
        end
    end


    -- 2. 校验 redis_type（白名单）
    local allowed_types = { NX = true, EX = true, Mset = true, Set = true }
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
    if redis_type == "Set" then
        if #args ~= 0 then
            ngx.log(ngx.ERR, "不应该传入任何多余参数")
            return false
        end
        result, err = red:set(key, value)
    elseif redis_type == "Mset" then
        if #args % 2 ~= 0 then -- 数量不对
            ngx.log(ngx.ERR, "数量不是 2 的倍数")
            return false
        end
        if #args == 0 then
            ngx.log(ngx.WARN, "没有传入任何 key-value")
        end
        result, err = red:set(key, value ,table.unpack(args))
    elseif redis_type == "EX" then
        if #args ~= 1 then
            ngx.log(ngx.ERR, "使用EX的ttl数量必须为 1 ")
            return false
        end
        result, err = red:set(key, value, redis_type, args[1])
    elseif redis_type == "NX" then
        result, err = red:set(key, value, redis_type)
    end

    -- if ttl == nil then
    --     result, err = red:set(key, value, redis_type)
    -- elseif type(ttl) == "number" then
    --     result, err = red:set(key, value, redis_type, ttl)
    -- else
    --     ngx.log(ngx.ERR, "ttl 参数类型错误: 期望 number, 实际为 ", type(ttl))
    --     redis_util.keepalive(red)
    --     return false
    -- end

    -- 放回连接池（即使在错误分支也执行了 keepalive）
    redis_util.keepalive(red)

    -- 4. 检查执行结果
    if err then
        ngx.log(ngx.ERR, "Redis SET 命令执行失败: ", err)
        return false
    end

    -- 这里如果是 Mset 的话 返回不是 "OK", 但 redis 终端看是 “OK”
    
    -- NX 如果要nil检查使用 if result == ngx.null then 因为 NX 会默认先 GET 下
    if result ~= "OK" then 
        ngx.log(ngx.ERR, "Redis SET 返回异常: ", result or "nil")
        return false
    end

    return true
end

return redis_setAll