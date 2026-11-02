local function PCALL(func)
    return pcall(func)
end

local function testPcall()

    local v1, v2
    local ok = PCALL(function ()
        v1 = "gaojian"
        v2 = "gaojian的v2"
    end)
    -- 是否崩溃
    if not ok then
        return nil, error("操作异常")
    end
    return v1, v2
end

return testPcall