-- 引入
local client = require("redisx.client")
local funcValidate = require("redisx.validate_not_empty_string")
local baseValidator = require("redisx.validate_base_func")
local errors = require("redisx.errors")

-- | 命令 | 成功返回值（类型） | 失败/异常情况 | 错误返回值 |
-- |------|-------------------|---------------|-----------|
-- | `red:lpush(key, val1, ...)` | `新长度` (number) | key 存在但不是 list | `nil, "ERR operation against a key holding the wrong kind of value"` |
-- | `red:rpush(key, val1, ...)` | `新长度` (number) | 同上 | 同上 |
-- | `red:lpop(key)` | `值` (string) 或 `ngx.null` | key 不存在或 list 为空 | `ngx.null, nil` |
-- | `red:rpop(key)` | `值` (string) 或 `ngx.null` | 同上 | 同上 |
-- | `red:llen(key)` | `长度` (number) | key 不存在 | `0, nil` |
-- | `red:lrange(key, start, stop)` | `{val1, val2, ...}` (table) | key 不存在 | `{}, nil`（空 table） |
-- | `red:lrem(key, count, val)` | `移除数量` (number) | key 存在但不是 list | `nil, "ERR operation against a key holding the wrong kind of value"` |

-- LPOP RPOP
local function ArgsPOP(redis_type, args) 
    if #args > 1 then
        return errors.new(errors.CODE.ARGS_COUNT_ERROR, redis_type .. " args不得大于1")
    end
    -- 如果是一个必须是 number
    if #args == 1 then
        -- 直接写上
        local count = args[1]
        if type(args[1]) ~= "number" then
            count = tonumber(args[1]) -- 不是 number，尝试改
        end
        -- 必须正整数
        if count == nil or count <= 0 or math.floor(count) ~= count then
            -- 这里未必string
            return errors.new(errors.CODE.POSITIVE_INTERGER, 
             redis_type .. " args[1]: " .. (tostring(args[1]) or "nil"))  
        end
    end
    return nil
end

-- LRANGE
local function ArgsLRANGE(redis_type, args)
    local err = baseValidator.argsCountMustBe(redis_type, args, 2)
    if err ~= nil then
        return err
    end
    local start, ed = tonumber(args[1]), tonumber(args[2])
    if start == nil or math.floor(start) ~= start or
        ed == nil or math.floor(ed) ~= ed then
        -- 能进来的都可能为 nil
        return errors.new(errors.CODE.INTERGER, redis_type .. ": " .. (start or "nil start") .. " and " .. (ed or "nil ed")) 
    end
    return nil
end

-- LREM
local function ArgsLREM(redis_type, args)
    if #args ~= 2 then
        return errors.new(errors.CODE.ARGS_COUNT_ERROR, redis_type .. " args必须为2")
    end
    local count, element = tonumber(args[1]), args[2]
    if not count or math.floor(count) ~= count then
        return errors.new(errors.CODE.INTERGER, 
         redis_type .. " args[1] is not INTERGER") 
    end
    if type(element) ~= "string" then
        return errors.new(errors.CODE.ARGS_TYPE_ERROR,
         redis_type .. " args[2]'s type is " .. type(args[2])) 
    end
    return nil
end

local default_type = "LLEN"
local allowed_type_funcs = { 
 LPUSH = baseValidator.ArgsCountNot0AndMustStrings,
 RPUSH = baseValidator.ArgsCountNot0AndMustStrings,
 LPOP = ArgsPOP,
 RPOP = ArgsPOP,
 LLEN = baseValidator.ArgsCountIs0,
 LRANGE = ArgsLRANGE,
 LREM = ArgsLREM,
}

-- 只支持 LPUSH RPUSH LPOP RPOP LLEN LRANGE LREN
-- return (string or table or number or ngx.null or nil) (errors.new() or nil)
-- (string or table or number or ngx.null or nil) 里 nil 是错误
-- ngx.null 没有拿到
-- errors.new() 错误 
-- string
-- table
local function with_list(key, redis_type, ...)
    -- 1. 一般验证
    local key, redis_type, args, err = funcValidate(key, redis_type, default_type, allowed_type_funcs, ...)
    if err ~= nil then
        return nil, err
    end
    -- 2. 执行命令
    -- return (string or ngx.null or table), (string or errors.new() -> table)
    local result, err = client(function(red)
        local result, err
        if redis_type == "LPUSH" then
            result, err = red:lpush(key, table.unpack(args))
        elseif redis_type == "RPUSH" then
            result, err = red:rpush(key, table.unpack(args))
        elseif redis_type == "LPOP" then
            local count
            if #args == 0 then
                count = 1
            end
            if #args == 1 then
                count = tonumber(args[1])
            end
            result, err = red:lpop(key, count)
        elseif redis_type == "RPOP" then
            local count
            if #args == 0 then
                count = 1
            end
            if #args == 1 then
                count = tonumber(args[1])
            end
            result, err = red:lpop(key, count)
        elseif redis_type == "LLEN" then
            result, err = red:llen(key)
        elseif redis_type == "LRANGE" then
            local start, ed = tonumber(args[1]), tonumber(args[2])
            result, err = red:lrange(key, start, ed)
        elseif redis_type == "LREM" then
            local count, element = tonumber(args[1]), args[2]
            result, err = red:lrem(key, count, element)
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


    -- OK or string_val or table or number or ngx.null
    -- 我只需要对获得 table 的如果全为空 使用 ngx.null 就行
    if result == ngx.null then -- 没有拿到
        return result, nil
    end

    -- 5.1 MGET 特殊处理：检查是否所有值都是 ngx.null
    -- OK or string_val or table or number
    -- lpop rpop lrange
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


return with_list