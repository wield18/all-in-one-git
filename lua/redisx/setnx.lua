-- 引入
local redis_set_func = require("redisx.set_ex_or_nx")

local function redis_setnx(key, value)
    return redis_set_func(key, value, "NX")
end


return redis_setnx
