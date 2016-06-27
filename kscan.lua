module(..., package.seeall)

local Processer = require("program.kscan.processer")
local pcap_filter = require("apps.packet_filter.pcap_filter")
local raw = require("apps.socket.raw")

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
   config.app(c, "rr", raw.RawSocket, interface)
   config.app(c,"pcap_filter", pcap_filter.PcapFilter,
              {filter=filter_rules, state_table = false})
   config.app(c, "pr", Processer.Processer)

   config.link(c, "rr.tx -> pcap_filter.input")
   config.link(c, "pcap_filter.output -> pr.input")
   config.link(c, "pr.output -> rr.rx")

   --[[
	timer.activate(timer.new("report", function ()
            engine.report_apps()
         end, 1e9, "repeating"))
	]]--
   engine.configure(c)
   engine.main()
end
