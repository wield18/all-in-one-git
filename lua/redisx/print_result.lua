
local function print(result, err)
    if err ~= nil then
        if type(err) == "table" then
            ngx.say(err.msg)
        else
            ngx.say("unknown err: " .. (err or "nil"))
        end
        ngx.say(type(err))
    end
    ngx.say(type(result))
    ngx.say(result or "nil")
end

return print