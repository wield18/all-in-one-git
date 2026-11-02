-- 引入
local errors = require("redisx.errors")

local M = {}

function M.argsMustString(args) --> 这里没有检测
    for i, v in ipairs(args) do
        if type(v) ~= "string" or v == "" then -- 不能为空字符
            return errors.new(errors.CODE.ARGS_TYPE_ERROR,
             "参数错误: 第 " .. i .. " 个参数不是字符串, 类型为 " .. type(v))
        end
    end
    return nil
end

-- LPUSH RPUSH
-- return (errors.new() -> table or nil)
function M.ArgsCountNot0(redis_type, args)
    if #args == 0 then
        return errors.new(errors.CODE.ARGS_COUNT_ERROR, redis_type .. " args必须有值")
    end
    return nil
end

function M.ArgsCountIs0(redis_type, args)
    if #args ~= 0 then
        return errors.new(errors.CODE.ARGS_COUNT_ERROR, redis_type .. " args必须无值")
    end
    return nil
end

function M.argsCountMustBe(redis_type, args, count)
    if type(count) ~= "number" then
        return errors.new(errors.CODE.ARGS_COUNT_ERROR, redis_type .. " count is not type number")
    end
    if #args ~= count then
        return errors.new(errors.CODE.ARGS_COUNT_ERROR, redis_type .. " args必须为" .. count)
    end
    return nil
end

function M.ArgsCountMustBeOne(redis_type, args)
    return M.argsCountMustBe(redis_type, args, 1)
end

function M.argsMustBePositive(redis_type, numb)
    if numb == nil or numb <= 0 or math.floor(numb) ~= numb then
        -- 这里未必string
        return errors.new(errors.CODE.POSITIVE_INTERGER, 
            redis_type .. " : " .. (numb or "nil"))  
    end
    return nil
end


function M.argsMustBeInterger(redis_type, numb)
    if numb == nil or math.floor(numb) ~= numb then
        -- 这里未必string
        return errors.new(errors.CODE.INTERGER, 
            redis_type .. " : " .. (numb or "nil"))  
    end
    return nil
end

function M.ArgsCountNot0AndMustStrings(redis_type, args)
    local err = M.argsMustString(args)
    if err ~= nil then
        return err
    end
    err = M.ArgsCountNot0(redis_type, args)
    return err
end


return M