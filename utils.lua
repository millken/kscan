module(..., package.seeall)
local ffi = require("ffi")
local lib = require("core.lib")

function dump (p)
   return lib.hexdump(ffi.string(packet.data(p), packet.length(p)))
end

function hexdump(buf)
	if buf == nil then return nil end
	for byte=1, #buf, 16 do
		local chunk = buf:sub(byte, byte+15)
		io.write(string.format('%08X  ',byte-1))
		chunk:gsub('.', function (c) io.write(string.format('%02X ',string.byte(c))) end)
		io.write(string.rep(' ',3*(16-#chunk)))
		io.write(' ',chunk:gsub('%c','.'),"\n") 
	end
end
