M = {}

M._job_id = nil
M._errors = {}

M.start = function()
    if M._job_id == nil then
        M._job_id = vim.fn.jobstart({ "gosee", vim.fn.expand("%:p") }, {
            on_stderr = function(_, msg)
                for _, data in ipairs(msg) do table.insert(M._errors, data) end
            end,
            on_exit = function()
                M._job_id = nil
                if #M._errors > 0 then
                    vim.notify(table.concat(M._errors), vim.log.levels.ERROR)
                    M._errors = {}
                end
            end
        })
    else
        vim.notify("gosee already running")
    end
end

M.stop = function()
    if M._job_id ~= nil then
        vim.fn.jobstop(M._job_id)
    end
end

M.setup = function()
    vim.api.nvim_create_user_command("Gosee", M.start, {})
    vim.api.nvim_create_user_command("GoseeStart", M.start, {})
    vim.api.nvim_create_user_command("GoseeStop", M.stop, {})
end

return M
