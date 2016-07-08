module(..., package.seeall)

local Processer = require("program.kscan.processer")
local testdns = require("program.kscan.test.dns")
local pcap_filter = require("apps.packet_filter.pcap_filter")
local raw = require("apps.socket.raw")
local Intel82599 = require("apps.intel.intel_app").Intel82599
local basic_apps = require("apps.basic.basic_apps")
local pci      = require("lib.hardware.pci")
local lib      = require("core.lib")
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

   local pcideva = lib.getenv("SNABB_PCI_INTEL0") or lib.getenv("SNABB_PCI0")
   local pcidevb = lib.getenv("SNABB_PCI_INTEL1") or lib.getenv("SNABB_PCI1")
    if not pcideva
      or pci.device_info(pcideva).driver ~= 'apps.intel.intel_app'
      or not pcidevb
      or pci.device_info(pcidevb).driver ~= 'apps.intel.intel_app'
   then
      print("SNABB_PCI_INTEL[0|1]/SNABB_PCI[0|1] not set or not suitable.")
      main.exit(2)
   end  
   local device_info_a = pci.device_info(pcideva)
   local device_info_b = pci.device_info(pcidevb)  
   engine.configure(config.new())
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

   --config.app(c, "rr", testdns.B)
  -- config.app(c, "pr", Processer.Processer)
   --config.app(c, 'source1', basic_apps.Source)
   config.app(c, 'sink', Processer.Sink)
   --config.app(c, "i80", Intel82599, {pciaddr=pcideva})
   --config.app(c, "i81", Intel82599, {pciaddr=pcidevb})
   --config.link(c, "source1.out -> i80.rx")
   --config.link(c, "i80.tx -> sink.input") 
   config.app(c, 'source1', basic_apps.Source)
   --config.app(c, 'source2', basic_apps.Source)
   config.app(c, 'nicA', Intel82599, {pciaddr=pcideva})
   config.app(c, 'nicB', Intel82599, {pciaddr=pcidevb})
   --config.app(c, 'sink', basic_apps.Sink)
   config.link(c, 'nicB.tx -> sink.in1')
   config.link(c, 'nicA.tx -> nicB.rx')
   config.link(c, 'source1.out -> nicA.rx')
   --config.link(c, 'source2.out -> nicB.rx')
   --config.link(c, 'nicB.tx -> sink.in2')  
   --config.app(c, "pr", Intel82599, {pciaddr="0000:0c:00.0"})
	timer.activate(timer.new(
      "report",
      function()
          engine.app_table.sink:report()
      end,
      1e9,
      'repeating'
   ))
   engine.configure(c)
    if device_info_a.model == pci.model["82599_T3"] or
         device_info_b.model == pci.model["82599_T3"] then
      C.usleep(2e6)
   end  
   engine.main({duration = 10, report = {showapps = true, showlinks = true, showload= false}})
end
