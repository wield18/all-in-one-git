-- 引入
local client = require("redisx.client")
local funcValidate = require("redisx.validate_not_empty_string")
local baseValidator = require("redisx.validate_base_func")
local zaddValidator = require("redisx.validate_ZADD")
local errors = require("redisx.errors")


local OPTIONS = {
    WITHSCORE = 1001
}

local AMAPOPTS = {
    [OPTIONS.WITHSCORE] = "WITHSCORE",
}
-- 我如果这样写可以是可以但
-- 我先简单写个map完事吧

-- zadd 的可选参数在前边我得怎么限制呢
local function Zadd(_, args)
    return zaddValidator(args)
end

local function Zrank(redis_type, args)
    if #args == 0 or #args > 2 then
        return errors.new(errors.CODE.ARGS_COUNT_ERROR, redis_type .. " args 不可等于 0 或大于 2")
    end
    local val = args[1] -- 一定拿到
    if type(val) ~= "string" or val == "" then
        return errors.new(errors.CODE.ARGS_TYPE_ERROR, redis_type .. " val 必须是string")
    end
    -- 检查一下 withscores 的写法
    if #args == 2 then
        local withscores = args[2]
        if withscores ~= AMAPOPTS[OPTIONS.WITHSCORE] then
            return errors.new(errors.CODE.OPTION_ARGS, redis_type .. " WITHSCORES is " .. (withscores or "nil"))
        end
    end
    return nil
end

-- 就检查前倆必须整数
local function Zrange(redis_type, args)
    if #args < 2 then
        return errors.new(errors.CODE.ARGS_COUNT_ERROR, redis_type .. " 不应该小于2")
    end
    local start = args[1]
    local ed = args[2]
    local err = baseValidator.argsMustBeInterger(redis_type, start)
    if err ~= nil then
        return err
    end
    return baseValidator.argsMustBeInterger(redis_type, ed)
end

local function Zincrby(redis_type, args)
    if #args ~= 2 then
        return errors.new(errors.CODE.ARGS_COUNT_ERROR, redis_type .. " 必须为2")
    end
    local incr = args[1]
    local member = args[2]
    local err = baseValidator.argsMustBeInterger(redis_type, incr)
    if err ~= nil then
        return err
    end
    if type(member) ~= "string" or member == "" then
        return errors.new(errors.CODE.ARGS_TYPE_ERROR, 
         redis_type .. " 必须为非空string, 当前为: " .. type(member) .. " : " .. (member or "nil"))
    end

    return nil
end

-- 只要是 number 就过 这是 lua redis 不好的地 不能用字符串
local function Zremrangebyrank(redis_type, args)
    if #args ~= 2 then
        return errors.new(errors.CODE.ARGS_COUNT_ERROR, redis_type .. " 必须为2")
    end
    local start = args[1]
    local ed = args[2]
    local err = baseValidator.argsMustBeInterger(redis_type, start)
    if err ~= nil then
        return err
    end
    return baseValidator.argsMustBeInterger(redis_type, ed)
end

-- local function Zremrangebyrank(redis_type, args)
--     if #args ~= 2 then
--         return errors.new(errors.CODE.ARGS_COUNT_ERROR, redis_type .. " 必须为2")
--     end
--     for i, v in ipairs(args) do
--         local t = type(v)
--         if t ~= "string" and t ~= "number" then
--             return errors.new(errors.CODE.ARGS_TYPE_ERROR,
--                 "参数错误: 第 " .. i .. " 个参数必须是字符串或数字, 实际类型为 " .. t)
--         end
--         if t == "string" and v == "" then
--             return errors.new(errors.CODE.ARGS_TYPE_ERROR,
--                 "参数错误: 第 " .. i .. " 个参数不能为空字符串")
--         end
--     end
--     return nil
-- end

-- | 操作 | OpenResty 方法 | 说明 |
-- | :--- | :--- | :--- |
-- | 添加或更新成员 | `red:zadd(key, score, member)` | 向有序集合添加一个成员，或更新已有成员的分数。 |
-- | 按排名范围获取 | `red:zrange(key, start, stop)` | 按分数**从小到大**排序，返回指定排名范围的成员。 |
-- | 按排名范围逆序获取 | `red:zrevrange(key, start, stop)` | 按分数**从大到小**排序，返回指定排名范围的成员（排行榜常用）。 |
-- | 按分数范围获取 | `red:zrangebyscore(key, min, max)` | 返回分数在 `min` 和 `max` 之间的成员（从小到大）。 |
-- | 按分数范围逆序获取 | `red:zrevrangebyscore(key, max, min)` | 返回分数在 `min` 和 `max` 之间的成员（从大到小）。 |
-- | 获取成员数量 | `red:zcard(key)` | 返回有序集合中的成员总数。 |
-- | 获取成员排名 | `red:zrank(key, member)` | 返回指定成员的排名（从小到大，0 表示第一名）。 |
-- | 获取成员逆序排名 | `red:zrevrank(key, member)` | 返回指定成员的排名（从大到小，0 表示第一名）。 |
-- | 增加成员分数 | `red:zincrby(key, increment, member)` | 为指定成员的分数增加 `increment`。 |
-- | 移除成员 | `red:zrem(key, member)` | 移除有序集合中的一个或多个成员。 |
-- | 移除按排名范围 | `red:zremrangebyrank(key, start, stop)` | 移除指定排名范围内的所有成员。 |
-- | 移除按分数范围 | `red:zremrangebyscore(key, min, max)` | 移除指定分数范围内的所有成员。 |

-- 可选参数不做判断，但给与对应函数可以简单传入
local default_type = "ZCARD" -- 默认 CARD
local allowed_type_funcs = { 
ZADD = Zadd, --> count or (INCR 增加到的分数)
ZCARD = baseValidator.ArgsCountIs0, --> count
ZREM = baseValidator.ArgsCountNot0AndMustStrings, --> count
ZRANK = Zrank, --> table key value withscore --> table or number
ZREVRANK = Zrank, --> table or number
ZRANGE = Zrange,
ZREVRANGE = Zrange, 
ZRANGEBYSCORE = Zrange,
ZREVRANGEBYSCORE = Zrange,
ZINCRBY = Zincrby,
ZREMRANGEBYRANK = Zremrangebyrank, -- 不能用 "+inf"
ZREMRANGEBYSCORE = Zremrangebyrank, -- 不能用 "+inf"
}

-- 如果直接全部方法原名当然可以，但方法太多，要不然自定义个可选参数，这样的话，就少了三个函数，别了吧
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
    -- ZADD, SCARD, ZREM, ZRANK, ZREVRANK, ZRANGE, ZREVRANGE, ZRANGEBYSCORE, ZREVRANGEBYSCORE, ZINCRBY, ZREMRANGEBYRANK, ZREMRANGEBYSCORE
    -- return (string or ngx.null or table), (string or errors.new() -> table)
    local result, err = client(function(red)
        local result, err
        if redis_type == "ZADD" then
            result, err = red:zadd(key, table.unpack(args))
        elseif redis_type == "ZCARD" then
            result, err = red:zcard(key, table.unpack(args))
        elseif redis_type == "ZREM" then
            result, err = red:zrem(key, table.unpack(args))
        elseif redis_type == "ZRANK" then
            result, err = red:zrank(key, table.unpack(args))
        elseif redis_type == "ZREVRANK" then
            result, err = red:zrevrank(key, table.unpack(args))
        elseif redis_type == "ZRANGE" then
            result, err = red:zrange(key, table.unpack(args))
        elseif redis_type == "ZREVRANGE" then
            result, err = red:zrevrange(key, table.unpack(args))
        elseif redis_type == "ZRANGEBYSCORE" then
            result, err = red:zrangebyscore(key, table.unpack(args))
        elseif redis_type == "ZREVRANGEBYSCORE" then
            result, err = red:zrevrangebyscore(key, table.unpack(args))
        elseif redis_type == "ZINCRBY" then
            result, err = red:zincrby(key, table.unpack(args))
        elseif redis_type == "ZREMRANGEBYRANK" then
            for i, v in ipairs(args) do
                ngx.say(type(v))
                ngx.say(v)
            end
            result, err = red:zremrangebyrank(key, table.unpack(args))
        elseif redis_type == "ZREMRANGEBYSCORE" then
            result, err = red:zremrangebyscore(key, table.unpack(args))
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