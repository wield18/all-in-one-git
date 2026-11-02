-- Redis 连接工具模块（带密码、连接池）
local redis = require("resty.redis")

local _M = {}

-- 配置常量（建议改成你的实际配置）
local REDIS_HOST = "127.0.0.1"      -- Redis IP
local REDIS_PORT = 6379             -- Redis 端口
local REDIS_PASSWORD = "xxx"  -- 你的密码
local REDIS_DB_INDEX = 0            -- 数据库索引（默认0）
local CONN_TIMEOUT = 1000           -- 连接超时(毫秒)
local POOL_SIZE = 3               -- 连接池大小
local KEEPALIVE_TIMEOUT = 60000     -- 空闲连接存活时间(毫秒)

-- 获取一个 Redis 连接（自动认证）
function _M.get_connection()
    local red = redis:new()
    
    -- 1. 设置超时
    red:set_timeouts(CONN_TIMEOUT, CONN_TIMEOUT, CONN_TIMEOUT)
    
    -- 2. 连接
    local ok, err = red:connect(REDIS_HOST, REDIS_PORT)
    if not ok then
        return nil, "连接失败: " .. (err or "unknown")
    end
    
    -- 3. 密码认证（关键步骤）
    if REDIS_PASSWORD and REDIS_PASSWORD ~= "" then
        local res, auth_err = red:auth(REDIS_PASSWORD)
        if not res then
            return nil, "认证失败: " .. (auth_err or "unknown")
        end
    end
    
    -- 4. 选择数据库
    if REDIS_DB_INDEX and REDIS_DB_INDEX > 0 then
        local res, select_err = red:select(REDIS_DB_INDEX)
        if not res then
            return nil, "选择数据库失败: " .. (select_err or "unknown")
        end
    end
    
    -- 5. 把连接和 red 对象一起返回，方便后续放回连接池
    return red, nil
end

-- 将连接放回连接池（重要：必须调用，否则连接不会复用）
function _M.keepalive(red)
    if not red then
        return
    end
    
    -- 直接放回连接池，让 resty.redis 内部处理
    local ok, err = red:set_keepalive(KEEPALIVE_TIMEOUT, POOL_SIZE)
    if not ok then
        -- 放回失败，直接关闭连接
        red:close()
        ngx.log(ngx.ERR, "放回 Redis 连接池失败: ", err or "unknown")
    end
end

-- 关闭连接（一般不用，除非特殊需求）
function _M.close(red)
    if red then
        red:close()
    end
end

return _M

-- local res, err = red:get("key")
-- if res == ngx.null then  -- ⚠️ 注意：Redis 返回空是 ngx.null，不是 nil
--     -- 处理空值
-- end