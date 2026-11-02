-- 引入
local redis_set_func = require("redisx.set_ex_or_nx")

local function redis_setex(key, value, ttl)
    -- 默认 60 秒
    if type(ttl) ~= "number" then
        ttl = 60
    end
    return redis_set_func(key, value, "EX", ttl)
end


return redis_setex
