module(...,package.seeall)

local ffi = require("ffi")
local lib = require("core.lib")
local transmit, receive = link.transmit, link.receive

local datagram = require("lib.protocol.datagram")
local ethernet = require("lib.protocol.ethernet")
local ipv6 = require("lib.protocol.ipv6")
local ipv4 = require("lib.protocol.ipv4")
local udp = require("lib.protocol.udp")
local dns = require("org.conman.dns")
-- dns test.com a 
A = {}

function A:new()
    local d1 = lib.hexundump ([[
      52:54:00:02:02:02 52:54:00:01:01:01 08 00 45 00
	  00 41 60 f6 00 00 40 11 71 54 c0 a8 03 af df 05
	  05 05 11 6e 00 35 00 2d da bd 92 42 01 20 00 01
	  00 00 00 00 00 01 04 74 65 73 74 03 63 6f 6d 00
	  00 01 00 01 00 00 29 10 00 00 00 00 00 00 00
   ]], 79)    
   local p = packet.from_string(d1) 
   return setmetatable({packet=p, packet_counter = 1}, {__index=A})
end

function A:pull ()
   for _, o in ipairs(self.output) do
      for i = 1, link.nwritable(o) do
	  --while not link.empty(i) and not link.full(o) do
         transmit(o, packet.clone(self.packet))
		 self.packet_counter = self.packet_counter + 1
      end
   end
end

B = {}

function B:new()
   return setmetatable({}, {__index=B})
end

function B:pull ()
   local dg_tx = datagram:new()
   local src = ethernet:pton("02:00:00:00:00:01")
   local dst = ethernet:pton("02:00:00:00:00:02")
   local localhost = ipv4:pton("127.0.0.1")
   local req,err = dns.encode {
	id       = 1234,
	query    = true,
	rd       = true,
	opcode   = 'query',
	question = {
		name  = 'test.com.',
		type  = 'a',
		class = 'in'
		},

	}
	assert(err == nil, err)
	dg_tx:push_raw(req, string.len(req))
   dg_tx:push(udp:new({src_port = 4462,
					   dst_port = 53}))
   dg_tx:push(ipv4:new({src = localhost,
                        dst = localhost,
						protocol = 17, --udp 0x11
                        ttl = 64}))
   dg_tx:push(ethernet:new({src = src,
                            dst = dst,
                            type = 0x0800}))
   for _, o in ipairs(self.output) do
	  while not link.full(o) do
         transmit(o, packet.clone(dg_tx:packet()))
      end
   end
end

