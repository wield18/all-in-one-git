local client = require("redisx.client")
local baseValidator = require("redisx.validate_base_func")
local funcValidate = require("redisx.validate_not_empty_string")
local errors = require("redisx.errors")

-- 返回 number or nil
-- nil 表示错误
local function with_del(key, ...)
    -- 1. 一般验证
    local key, redis_type, args, err = funcValidate(key, "DEL", "DEL", {["DEL"] = baseValidator.ArgsCountNot0AndMustStrings}, ...)
    if err ~= nil then
        return nil, err
    end
    
    -- 2. 执行 DEl 命令
    local result, err = client(function(red)
        return red:del(key, table.unpack(args))
    end)

    -- 3. 检查执行结果
    -- 这里的 err 可能是 table or nil or string 
    if err ~= nil then
        if type(err) == "string" then
            err = errors.new(errors.CODE.REDIS_CMD_FAIL, redis_type .. " :" .. err)
        end
        return nil, err
    end
    
    return result, nil
end


return with_del