local errors = require("redisx.errors")

-- 定义允许的 ZADD 选项
local ZADD_OPTIONS = {
    NX = true,
    XX = true,
    GT = true,
    LT = true,
    CH = true,
    INCR = true
}

-- 互斥选项组
local MUTEX_GROUPS = {
    { NX = true,
     XX = true,},
    { GT = true,
     LT = true,},}

-- 检查是否有互斥选项同时出现
local function check_mutex_options(options)
    for _, group in ipairs(MUTEX_GROUPS) do
        local found = {}
        for opt in pairs(group) do -- 它这里一组不能能俩个
            if options[opt] then
                table.insert(found, opt)
            end
        end
        if #found > 1 then
            return errors.new(errors.CODE.ARGS_TYPE_ERROR,
                "ZADD 选项互斥: " .. table.concat(found, " 和 ") .. " 不能同时使用")
        end
    end
    return nil
end



-- 主校验函数
-- return (nil or errTable)
local function zaddValidate(args)
    if type(args) ~= "table" or #args == 0 then
        return errors.new(errors.CODE.ARGS_COUNT_ERROR, "ZADD 参数不能为空")
    end
    -- 1. 分离选项和实际数据
    local options = {}
    local data_start_index = 1
    local has_incr = false
    
    -- 收集所有选项（从第一个参数开始，直到遇到非选项）
    for i, v in ipairs(args) do
        if type(v) == "string" and ZADD_OPTIONS[v] then
            local opt = v
            options[opt] = true
            if opt == "INCR" then
                has_incr = true
            end
            data_start_index = i + 1
        else
            -- 遇到非选项，退出选项收集
            break
        end
    end

    -- 2. 检查选项互斥
    local err = check_mutex_options(options)
    if err then
        return err
    end

    -- 3. 提取实际数据（从 data_start_index 开始）
    local data_args = {}
    for i = data_start_index, #args do
        table.insert(data_args, args[i])
    end

    -- 4. 检查数据部分不能为空
    if #data_args == 0 then
        return errors.new(errors.CODE.ARGS_COUNT_ERROR, 
            "ZADD 缺少成员-分数对数据")
    end
    
    -- 5. 如果开启了 INCR，检查只能有一个成员-分数对
    if has_incr then
        if #data_args ~= 2 then
            return errors.new(errors.CODE.ARGS_COUNT_ERROR,
                "ZADD INCR 模式只能操作一个成员，需要 2 个参数 (分数 和 成员)，实际传了 " .. #data_args .. " 个")
        end
    end

    -- 6. 验证数据格式：必须是 (number, string) 的配对
    -- 这里就验证了成对就行
    if #data_args % 2 ~= 0 then
        return errors.new(errors.CODE.ARGS_COUNT_ERROR,
            "ZADD 参数必须是成对的 (分数, 成员)，总参数个数必须为偶数，实际传了 " .. #data_args .. " 个")
    end

    
    -- 7. 逐个校验每个 (score, member) 对
    local pair_count = #data_args / 2
    for i = 0, pair_count do
        local score_idx = i * 2 + 1
        local member_idx = i * 2 + 2
        
        local score = data_args[score_idx]
        local member = data_args[member_idx]
        if score == nil and member == nil then
            break
        end
        -- 校验 score 必须是整数
        if type(score) ~= "number" or math.floor(score) ~= score then
            return errors.new(errors.CODE.ARGS_TYPE_ERROR,
                "参数错误: 第 " .. (data_start_index + score_idx - 1) .. " 个参数必须是整数，实际类型为 " .. type(score))
        end
        
        -- 校验 member 必须是字符串且非空
        if type(member) ~= "string" or member == "" then
            return errors.new(errors.CODE.ARGS_TYPE_ERROR,
                "参数错误: 第 " .. (data_start_index + member_idx - 1) .. " 个参数必须是字符串 (成员)，实际类型为 " .. type(member))
        end
    end

    return nil
end

return zaddValidate
