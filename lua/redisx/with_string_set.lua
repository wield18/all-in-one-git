local client = require("redisx.client")
local baseValidator = require("redisx.validate_base_func")
local funcValidate = require("redisx.validate_not_empty_string")
local errors = require("redisx.errors")


local function Mset(redis_type, args)
    if #args % 2 ~= 0 then -- 数量不对
        return errors.new(errors.CODE.ARGS_COUNT_ERROR, redis_type .. "数量不是 2 的倍数")        
    end
    -- 这里要求key是string就行
    -- 7. 逐个校验每个 (score, member) 对
    local pair_count = #args / 2
    for i = 0, pair_count do
        local key_idx = i * 2 + 1
        local val_idx = i * 2 + 2
        
        local key = args[key_idx]
        local val = args[val_idx]
        if key == nil and val == nil then
            break
        end
        -- 校验 key 必须是 string
        if type(key) ~= "string" or key == "" then
            return errors.new(errors.CODE.ARGS_TYPE_ERROR,
                "参数错误: 第 " .. key_idx .. " 个参数必须是字符串，实际类型为 " .. type(key))
        end
        
        -- 校验 val 必须是 string or number
        if type(val) ~= "number" and (type(val) ~= "string" or val == "") then
            return errors.new(errors.CODE.ARGS_TYPE_ERROR,
                "参数错误: 第 " .. key_idx .. " 个参数必须是 string or number,实际类型为 " .. type(val))
        end
    end

end

local function SetEX(redis_type, args)
    local err = baseValidator.ArgsCountMustBeOne(redis_type, args)
    if err ~= nil then
        return err
    end
    return baseValidator.argsMustBeInterger(redis_type, args[1])
end

local default_type = "SET"
local allowed_type_funcs = {
    ["NX"] = baseValidator.ArgsCountIs0,
    ["EX"] = SetEX,
    ["MSET"] = Mset,
    ["SET"] = baseValidator.ArgsCountIs0,
}
-- 支持 NX EX MSET SET
-- return boolean
local function with_set(key, value, redis_type, ...)
    -- value 简单检查下 string or number
    -- if type(value) -- 这个可以塞 ... 里, 不了吧
    if type(value) ~= "number" and type(value) ~= "string" then
        return nil, errors.new(errors.CODE.ARGS_TYPE_ERROR, redis_type .. " value type 不对" .. type(value))
    end

    -- 1. 一般验证
    local key, redis_type, args, err = funcValidate(key, redis_type, default_type, allowed_type_funcs, ...)
    if err ~= nil then
        return nil, err
    end

    -- client(callback, ...), 这里虽说没有传任何参数给 pcall 但方法里直接写入了传入的参数而已
    -- 2. 执行 SET 命令
    local result, err = client(function(red)
        local result, err
        if redis_type == "SET" then
            result, err = red:set(key, value)
        elseif redis_type == "MSET" then
            if #args == 0 then
                ngx.log(ngx.WARN, "没有传入任何 key-value")
            end
            result, err = red:mset(key, value ,table.unpack(args))
        elseif redis_type == "EX" then
            local ttl = tonumber(args[1])
            result, err = red:set(key, value, redis_type, ttl)
        elseif redis_type == "NX" then
            result, err = red:set(key, value, redis_type)
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
    
    -- 只有 NX 会 --> number
    -- 直接原样返回
    return result, nil
end


return with_set