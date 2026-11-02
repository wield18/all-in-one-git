-- 引入
local client = require("redisx.client")
local funcValidate = require("redisx.validate_not_empty_string")
local baseValidator = require("redisx.validate_base_func")
local errors = require("redisx.errors")

-- | 命令 | 成功返回值（类型） | 失败/异常情况 | 错误返回值 |
-- |------|-------------------|---------------|-----------|
-- | `red:sadd(key, m1, ...)` | `新增数量` (number) | key 存在但不是 set | `nil, "ERR operation against a key holding the wrong kind of value"` |
-- | `red:srem(key, m1, ...)` | `移除数量` (number) | key 存在但不是 set | 同上 |
-- | `red:smembers(key)` | `{m1, m2, ...}` (table) | key 不存在 | `{}, nil` |
-- | `red:sismember(key, member)` | `1` (存在) / `0` (不存在) (number) | key 存在但不是 set | `nil, "ERR operation against a key holding the wrong kind of value"` |
-- | `red:scard(key)` | `数量` (number) | key 不存在 | `0, nil` |
-- | `red:spop(key)` | `值` (string) 或 `ngx.null` | key 不存在或 set 为空 | `ngx.null, nil` |
-- | `red:sinter(k1, k2, ...)` | `{m1, m2, ...}` (table) | 某个 key 存在但不是 set | `nil, "ERR operation against a key holding the wrong kind of value"` |
-- | `red:sunion(k1, k2, ...)` | `{m1, m2, ...}` (table) | 同上 | 同上 |
-- | `red:sdiff(k1, k2, ...)` | `{m1, m2, ...}` (table) | 同上 | 同上 |

local function Spop(redis_type, args)
    local err = baseValidator.ArgsCountMustBeOne(redis_type, args)
    if err ~= nil then
        return err
    end
    return baseValidator.argsMustBePositive(redis_type, args[1])
end


local default_type = "SCARD"
local allowed_type_funcs = { 
SADD = baseValidator.ArgsCountNot0AndMustStrings, --> count
SREM = baseValidator.ArgsCountNot0AndMustStrings, --> count
SMEMBERS = baseValidator.ArgsCountIs0, --> table
SISMEMBER = baseValidator.ArgsCountMustBeOne, --> count
SCARD = baseValidator.ArgsCountIs0, --> count
SPOP = Spop, --> table
SINTER = baseValidator.ArgsCountNot0AndMustStrings, --> table 交集
SUNION = baseValidator.ArgsCountNot0AndMustStrings, --> table 并集
SDIFF = baseValidator.ArgsCountNot0AndMustStrings, --> table 差集 属于左边但不右边
}

-- 只支持 LPUSH RPUSH LPOP RPOP LLEN LRANGE LREN
-- return (string or table or number or ngx.null or nil) (errors.new() or nil)
-- (string or table or number or ngx.null or nil) 里 nil 是错误
-- ngx.null 没有拿到
-- errors.new() 错误 
-- string
-- table
local function with_set(key, redis_type, ...)
    -- 1. 一般验证
    local key, redis_type, args, err = funcValidate(key, redis_type, default_type, allowed_type_funcs, ...)
    if err ~= nil then
        return nil, err
    end
    -- 2. 执行命令
    -- SADD, SREM, SMEMBERS, SISMEMBER, SCARD, SPOP, SINTER, SUNION, SDIFF
    -- return (string or ngx.null or table), (string or errors.new() -> table)
    local result, err = client(function(red)
        local result, err
        if redis_type == "SADD" then
            result, err = red:sadd(key, table.unpack(args))
        elseif redis_type == "SREM" then
            result, err = red:srem(key, table.unpack(args))
        elseif redis_type == "SMEMBERS" then
            result, err = red:smembers(key)
        elseif redis_type == "SISMEMBER" then
            result, err = red:sismember(key, args[1])
        elseif redis_type == "SCARD" then
            result, err = red:scard(key)
        elseif redis_type == "SPOP" then
            local count
            if #args == 0 then
                count = 1
            end
            if #args == 1 then
                count = tonumber(args[1])
            end
            result, err = red:spop(key, count)
        elseif redis_type == "SINTER" then
            result, err = red:sinter(key, table.unpack(args))
        elseif redis_type == "SUNION" then
            result, err = red:sunion(key, table.unpack(args))
        elseif redis_type == "SDIFF" then
            result, err = red:sdiff(key, table.unpack(args))
        end
        return result, err
    end)

    -- 3. 检查执行结果
    -- 这里的 err 可能是 table or nil or string 
    if err ~= nil then
        if type(err) == "string" then
            err = errors.new(errors.CODE.REDIS_CMD_FAIL, redis_type .. " :" .. err)
        end
        return nil, err
    end


    -- table or number, spop 默认使用也直接 table
    -- 5.1 MGET 特殊处理：检查是否所有值都是 ngx.null
    if type(result) == "table" then
        local all_null = true
        for i, v in ipairs(result) do
            if v ~= ngx.null then
                all_null = false
                break
            end
        end
        if all_null then
            return ngx.null, nil  -- 全部不存在，返回 ngx.null
        end
    end

    return result, nil
end


return with_set