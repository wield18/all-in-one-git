local client = require("redisx.client")
local baseValidator = require("redisx.validate_base_func")
local funcValidate = require("redisx.validate_not_empty_string")
local errors = require("redisx.errors")


local function Getset(redis_type, args)
    local err = baseValidator.ArgsCountMustBeOne(redis_type, args)
    if err ~= nil then
        return err
    end
    return baseValidator.ArgsCountNot0AndMustStrings(redis_type, args)
end

local default_type = "GET"
local allowed_type_funcs = {
    ["GET"] = baseValidator.ArgsCountIs0,
    ["MGET"] = baseValidator.ArgsCountNot0AndMustStrings,
    ["GETSET"] = Getset,
    ["TTL"] = baseValidator.ArgsCountIs0,
}

-- 只支持 GET, MGET 或 GETSET
-- return (string or table or nil or ngx.null) boolean
local function with_get(key, redis_type, ...)
    -- 1. 一般验证
    local key, redis_type, args, err = funcValidate(key, redis_type, default_type, allowed_type_funcs, ...)
    if err ~= nil then
        return nil, err
    end

    -- 2. 执行 SET 命令
    local result, err = client(function(red)
        local result, err
        if redis_type == "GET" then
            result, err = red:get(key)
        elseif redis_type == "MGET" then
            result, err = red:mget(key, table.unpack(args))
        elseif redis_type == "GETSET" then
            result, err = red:getset(key, args[1])
        elseif redis_type == "TTL" then
            result, err = red:ttl(key)
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
    
    -- 4 MGET 特殊处理：检查是否所有值都是 ngx.null
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
        -- 否则返回原始的 table，保留部分命中的结果
        return result, nil
    end

    return result, nil
end


return with_get