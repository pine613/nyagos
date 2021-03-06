nyagos.alias("open",function(args)
    local count=0
    for i=1,#args do
        local list=nyagos.glob(args[i])
        if list and #list >= 1 then
            for i=1,#list do
                assert(nyagos.shellexecute("open",list[i]))
            end
        else
            if nyagos.access(args[i]) then
                assert(nyagos.shellexecute("open",args[i]))
            else
                print(args[i] .. ": can not get status")
            end
        end
        count = count +1
    end
    if count <= 0 then
        if nyagos.access(".\\open.cmd",0) then
            nyagos.exec("open.cmd")
        else
            assert(nyagos.shellexecute("open","."))
        end
    end
end)
