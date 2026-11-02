-- redisx/client.lua（统一入口）
local errors = require("redisx.errors")
local redis_util = require("redisx.redis_util")

-- client 使用时可以传参数但记得callback接收就行
local function client(callback, ...)
    local red, err = redis_util.get_connection()
    if not red then
        return nil, errors.new(errors.CODE.REDIS_CONN_FAIL, err)
    end
    
    local ok, result, err = pcall(callback, red, ...)
    
    redis_util.keepalive(red)
    
    -- 是否崩溃
    if not ok then
        return nil, errors.new(errors.CODE.REDIS_PCALL_FAIL, (err or "nil"))
    end

    return result, err
end

return client