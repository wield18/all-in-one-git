-- redisx/err.lua
-- 统一错误码和错误信息

local M = {}

-- 错误码定义
M.CODE = {
    SUCCESS = 0,
    PARAM_ERROR = 1001,
    KEY_TYPE_ERROR = 1002,
    VALUE_EMPTY = 1003,
    REDIS_TYPE_INVALID = 1004,
    REDIS_CONN_FAIL = 1005,
    REDIS_PCALL_FAIL = 1006,
    POSITIVE_INTERGER = 1007,
    ARGS_COUNT_ERROR = 1008,
    ARGS_TYPE_ERROR = 1009,
    KEY_NOT_EXIST = 1010,
    PERMISSION_DENIED = 1011,
    INTERGER = 1012,
    REDIS_CMD_FAIL = 1013,
    OPTION_ARGS = 1014,
    SET_NOT_OK = 1015,
}

-- 错误信息映射
M.MSG = {
    [M.CODE.SUCCESS] = "success",
    [M.CODE.PARAM_ERROR] = "parameter wrong | 参数错误",
    [M.CODE.KEY_TYPE_ERROR] = "key type wrong or empty | key 类型错误或空字符串",
    [M.CODE.VALUE_EMPTY] = "value type wrong or empty | value 类型错误或空字符串",
    [M.CODE.REDIS_TYPE_INVALID] = "redis_type invaild | redis_type 参数不合法",
    [M.CODE.REDIS_CONN_FAIL] = "Redis connect failed | Redis 连接失败",
    [M.CODE.REDIS_PCALL_FAIL] = "Redis pcall err | Redis Pcall 执行失败",
    [M.CODE.REDIS_CMD_FAIL] = "Redis cmd err | Redis Cmd 执行失败",
    [M.CODE.POSITIVE_INTERGER] = "must be positive interger | 必须是正整数",
    [M.CODE.INTERGER] = "must be interger | 必须是整数",
    [M.CODE.ARGS_COUNT_ERROR] = "args quantity wrong | 参数数量错误",
    [M.CODE.ARGS_TYPE_ERROR] = "args type wrong | 参数类型错误",
    [M.CODE.KEY_NOT_EXIST] = "key not exist | key 不存在",
    [M.CODE.PERMISSION_DENIED] = "permission denied | 权限不足",
    [M.CODE.OPTION_ARGS] = "option args error | 可选参数有误",
    [M.CODE.SET_NOT_OK] = "set not ok | Set指令没有OK",
}

-- 生成错误对象
-- parameters code, extra
-- code 必传 extra(string) 可传
function M.new(code, extra)
    local msg = M.MSG[code] or "unknown err | 未知错误"
    if extra then
        msg = msg .. ": " .. tostring(extra)
    end
    return {
        code = code,
        msg = msg,
    } 
end

-- 判断是否成功
function M.is_ok(err)
    return err.code == M.CODE.SUCCESS
end

return M