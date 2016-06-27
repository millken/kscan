module(..., package.seeall)
local ffi = require("ffi")
local packet = require("core.packet")
Processer = {}

function Processer:new ()
   local o = { packet_counter = 1 }
   return setmetatable(o, {__index = Processer})
end

function Processer:push()
   local i = assert(self.input.input, "input port not found")
   local o = assert(self.output.output, "output port not found")

   for _ = 1, link.nreadable(i) do
      self:process_packet(i, o)
      self.packet_counter = self.packet_counter + 1
   end

end


function Processer:process_packet(i, o)
   local p = link.receive(i)	
	local s = ffi.string(p.data, p.length)
	print("FIB: ".. s)
   --link.transmit(o, p)
	packet.free(p)
end
