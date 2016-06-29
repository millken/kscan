module(..., package.seeall)
local ffi = require("ffi")
local packet = require("core.packet")
local utils = require("program.kscan.utils")
Processer = {}

function Processer:new ()
   local o = { packet_counter = 1 }
   return setmetatable(o, {__index = Processer})
end

function Processer:push()
   local i = assert(self.input.input, "input port not found")

   for _ = 1, link.nreadable(i) do
      self:process_packet(i)
      self.packet_counter = self.packet_counter + 1
   end

end

function Processer:process_packet(i)
   local p = link.receive(i)	
	local s = ffi.string(p.data, p.length)
	--utils.hexdump(s)
   --link.transmit(o, p)
	packet.free(p)
end
