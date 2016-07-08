--https://github.com/plajjan/snabbswitch/blob/snabbddos/src/apps/ddos/ddos.lua
module(..., package.seeall)
local ffi = require("ffi")
local packet = require("core.packet")
local utils = require("program.kscan.utils")
local counter = require("core.counter")
local C = ffi.C
Processer = {}

function Processer:new ()
   local o = { counters = {} }
   self = setmetatable(o, {__index = Processer})


   -- schedule periodic task every second
   timer.activate(timer.new(
      "periodic",
      function () self:periodic() end,
      1e9, -- every second
      'repeating'
   ))
   self.counters["push_packets"] = counter.open("kscan/push_packets")
   return self
end

function Processer:periodic()
end

function Processer:get_stats_snapshot()
   return {
      rxpackets = link.stats(self.input.input).txpackets,
      rxbytes = link.stats(self.input.input).txbytes,
      --txpackets = link.stats(self.output.output).txpackets,
      --txbytes = link.stats(self.output.output).txbytes,
      --txdrop = link.stats(self.output.output).txdrop,
      time = tonumber(C.get_time_ns()),
   }
end

function Processer:push()
   local i = assert(self.input.input, "input port not found")

   for _ = 1, link.nreadable(i) do
      self:process_packet(i)
   end

end

function Processer:pull()
end

function Processer:process_packet(i)
   local p = link.receive(i)	
	local counters = self.counters
   counter.add(counters["push_packets"])
	local s = ffi.string(p.data, p.length)
	--utils.hexdump(s)
   --link.transmit(o, p)
	packet.free(p)
end

function Processer:report()
   if self.last_stats == nil then
      self.last_stats = self:get_stats_snapshot()
      return
   end
   last = self.last_stats
   cur = self:get_stats_snapshot()

   print("\n-- Processer report --" .. string.format("%d", tonumber(counter.read(self.counters["push_packets"]))))
   counter.set(self.counters["push_packets"], 0)
   print("Rx: " .. utils.num_prefix((cur.rxpackets - last.rxpackets) / ((cur.time - last.time) / 1e9)) .. "pps / " .. (cur.rxpackets - last.rxpackets) .. " packets / " .. cur.rxbytes .. " bytes")
  -- print("Tx: " .. utils.num_prefix((cur.txpackets - last.txpackets) / ((cur.time - last.time) / 1e9)) .. "pps / " .. cur.txpackets .. " packets / " .. cur.txbytes .. " bytes / " .. cur.txdrop .. " packet drops")

   self.last_stats = cur
end


Sink = {}

function Sink:new ()
   return setmetatable({}, {__index=Sink})
end

function Sink:push ()
   for _, i in ipairs(self.input) do
      for _ = 1, link.nreadable(i) do
        local p = link.receive(i)
		local s = ffi.string(p.data, p.length)
	    --utils.hexdump(s)
        packet.free(p)
      end
   end
end

function Sink:get_stats_snapshot()
   return {
      rxpackets = link.stats(self.input.in1).rxpackets,
      rxbytes = link.stats(self.input.in1).rxbytes,
      --txpackets = link.stats(self.output.output).txpackets,
      --txbytes = link.stats(self.output.output).txbytes,
      --txdrop = link.stats(self.output.output).txdrop,
      time = tonumber(C.get_time_ns()),
   }
end
function Sink:report()
   if self.last_stats == nil then
      self.last_stats = self:get_stats_snapshot()
      return
   end
   last = self.last_stats
   cur = self:get_stats_snapshot()

   print("\n-- Processer report --")
   print("Rx: " .. utils.num_prefix((cur.rxpackets - last.rxpackets) / ((cur.time - last.time) / 1e9)) .. "pps / " .. (cur.rxpackets - last.rxpackets) .. " packets / " .. cur.rxbytes .. " bytes")

   self.last_stats = cur
end
