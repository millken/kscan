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
   local src = ethernet:pton("00:1B:21:99:2A:04")
   local dst = ethernet:pton("00:1B:21:99:2A:05")
   local ip_src = ipv4:pton("192.168.55.100")
   local ip_dst = ipv4:pton("192.168.55.101")
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
   dg_tx:push(ipv4:new({src = ip_src,
                        dst = ip_dst,
						protocol = 17, --udp 0x11
                        ttl = 64}))
   dg_tx:push(ethernet:new({src = src,
                            dst = dst,
                            type = 0x0800}))
   for _, o in ipairs(self.output) do
	for i = 1, link.nwritable(o) do
	  --while not link.full(o) do
         transmit(o, packet.clone(dg_tx:packet()))
      end
   end
end

C = {}

function C:new()
   return setmetatable({}, {__index=C})
end

function C:pull ()
   local dg_tx = datagram:new()
   local src = ethernet:pton("00:1B:21:99:2A:04")
   local dst = ethernet:pton("00:1B:21:99:2A:05")
   local ip_src = ipv4:pton("192.168.55.100")
   local ip_dst = ipv4:pton("192.168.55.101")
   dg_tx:push(udp:new({src_port = 4462,
					   dst_port = 53}))
   dg_tx:push(ipv4:new({src = ip_src,
                        dst = ip_dst,
						protocol = 17, --udp 0x11
                        ttl = 64}))
   dg_tx:push(ethernet:new({src = src,
                            dst = dst,
                            type = 0x0800}))
   for _, o in ipairs(self.output) do
	for i = 1, link.nwritable(o) do
         transmit(o, packet.clone(dg_tx:packet()))
      end
   end
end

D = {}

function D:new () 
     local self = {} 
     self.p = packet.allocate() 
	 eth_size = ethernet:sizeof() 
	 ipv4_size = ipv4:sizeof()
	 udp_size = udp:sizeof()

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
     self.p.length = eth_size + ipv4_size + udp_size
     self.eth = ethernet:new_from_mem(self.p.data, eth_size)
	 self.ipv4 = ipv4:new_from_mem(self.p.data + eth_size, ipv4_size)
	 self.udp = udp:new_from_mem(self.p.data + eth_size + ipv4_size, udp_size)
	 self.req = req
	 packet.append(self.p, self.req, #self.req)
     return setmetatable(self, {__index=D}) 
  end 
  
  function D:pull() 
	self.eth:src(ethernet:pton("00:1B:21:99:2A:04"))
	self.eth:dst(ethernet:pton("00:1B:21:99:2A:05"))
	self.eth:type(0x0800)
	self.ipv4:ihl(ipv4:sizeof() / 4)
	self.ipv4:dscp(0)
	self.ipv4:ecn(0)
	self.ipv4:total_length(ipv4:sizeof() + udp:sizeof() + #self.req)
	self.ipv4:id(0)
	self.ipv4:flags(0)
	self.ipv4:frag_off(0)
	self.ipv4:src(ipv4:pton("192.168.55.100"))
	self.ipv4:dst(ipv4:pton("192.168.55.101"))
	self.ipv4:protocol(17)
	self.ipv4:ttl(64)
	self.ipv4:checksum()
	self.udp:src_port(4462)
	self.udp:dst_port(53)
	self.udp:length(udp:sizeof() + #self.req)

    for _, o in ipairs(self.output) do 
       for i = 1, link.nwritable(o) do 
          link.transmit(o, packet.clone(self.p)) 
       end 
    end 
  end 
  
  function D:stop () 
     packet.free(self.p) 
  end 
