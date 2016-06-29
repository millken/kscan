module(..., package.seeall)

local Processer = require("program.kscan.processer")
local testdns = require("program.kscan.test.dns")
local pcap_filter = require("apps.packet_filter.pcap_filter")
local raw = require("apps.socket.raw")
local Intel82599 = require("apps.intel.intel_app").Intel82599
local C = require("ffi").C

function run (parameters)
   if not (#parameters == 1) then
      print("Usage: kscan <interface>")
      main.exit(1)
   end
   local interface = parameters[1]

   local filter_rules = 
   [[
	(tcp and dst port 80)
   ]]

   local c = config.new()
   --[[
   config.app(c, "rr", raw.RawSocket, interface)
   config.app(c,"pcap_filter", pcap_filter.PcapFilter,
              {filter=filter_rules, state_table = false})
   config.app(c, "pr", Processer.Processer)

   config.link(c, "rr.tx -> pcap_filter.input")
   config.link(c, "pcap_filter.output -> pr.input")
   config.link(c, "pr.output -> rr.rx")
	--]]

   config.app(c, "rr", testdns.A)
   config.app(c, "pr", Processer.Processer)
   --config.app(c, "pr", Intel82599, {pciaddr="0000:0c:00.0"})
   config.link(c, "rr.tx -> pr.input")
   engine.configure(c)
   engine.main({duration = 1, report = {showapps = true, showlinks = true, showload= true}})
   local source = engine.app_table.rr.output.tx
   assert(source, "no source?")
   local s= link.stats(source)
   print("source:      txpackets= ", s.txpackets, "  rxpackets= ", s.rxpackets, "  txdrop= ", s.txdrop, "packet_count= ", testdns.A.packet_counter)
end
