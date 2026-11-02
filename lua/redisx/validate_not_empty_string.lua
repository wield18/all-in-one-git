local errors = require("redisx.errors")

-- 验证过后的所有输入
-- return key, redis_type, args, (errors.new() or nil)
-- errors.new() -> table, nil -> no-error
local function validate_not_empty_string(key, redis_type, default_type, allowed_type_funcs, ...)
    -- 0. 没有 redis_type 直接 defaultKey
    if redis_type == nil then
        redis_type = tostring(default_type) -- default_type 不做检查内部使用
    end

    -- 0.1 key 验证
    if type(key) ~= "string" or key == "" then
        return "", "", {}, errors.new(errors.CODE.KEY_TYPE_ERROR)
    end

    -- 2. 校验 redis_type（白名单）
    local fun = allowed_type_funcs[redis_type]
    if  type(fun) ~= "function" then --> 或许我写个函数链,不想这样
        return "", "", {}, errors.new(errors.CODE.REDIS_TYPE_INVALID,
         "当前为: " .. (redis_type or "nil"))
    end
    
    local args = { ... } -- Argument
    -- 3. 个性检查
    local err = fun(redis_type, args)
    if err ~= nil then
        return "", "", {}, err
    end

    return key, redis_type, args, nil
end

return validate_not_empty_string