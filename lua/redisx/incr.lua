local client = require("redisx.client")

-- return number or nil
-- number: 正常
-- nil: 参数或网络或操作错误
local function with_get(key, increment)
    -- 0 key, increment 验证 
    if increment == nil then
        increment = 1
    end
    local incr = tonumber(increment)
    if not incr or math.floor(incr) ~= incr then
        ngx.log(ngx.ERR, "incr 必须是正负整数，当前为 " .. increment)
        return nil
    end
    if type(key) ~= "string" or key == "" then
        ngx.log(ngx.ERR, "Redis key: nil")
        return nil
    end

    -- 3. 执行 SET 命令
    local result, err = client(function(red)
        local result, err
        result, err = red:incrby(key, incr)
        return result, err
    end)

    -- 4. 检查执行结果
    if err then
        ngx.log(ngx.ERR, "Redis ERR: ", err)
        return nil
    end
    -- incr 不可能为 ngx.null
    return result
end


return with_get