# MosDNS 功能测试用例

本目录包含针对 MosDNS 主要功能的测试用例，按照功能领域进行分类。

## 测试目录结构

- `server_test/`: 测试各种 DNS 服务器协议的功能，包括 UDP、TCP、DoH、DoQ
- `matcher_test/`: 测试各类 DNS 请求匹配器的功能，如域名匹配、IP 匹配等
- `executable_test/`: 测试各类可执行插件的功能，如 DNS 转发、缓存、重定向等
- `integration_test/`: 集成测试，测试多个组件协同工作的场景

## 测试功能覆盖

### 服务器功能测试

- UDP 服务器：测试基本的 DNS 请求响应功能
- TCP 服务器：测试 TCP 协议的 DNS 请求响应
- DoH 服务器：测试 HTTP 协议的 DNS 请求响应
- DoQ 服务器：测试 QUIC 协议的 DNS 请求响应

### 匹配器功能测试

- 域名匹配器(qname)：测试域名匹配规则
- IP 匹配器(client_ip, resp_ip)：测试客户端 IP 和响应 IP 的匹配
- 查询类型匹配器(qtype)：测试 DNS 查询类型的匹配
- 随机匹配器(random)：测试随机匹配功能

### 可执行插件功能测试

- 转发(forward)：测试 DNS 请求转发功能
- 缓存(cache)：测试 DNS 响应缓存功能
- 重定向(redirect)：测试 DNS 请求重定向功能
- 黑洞(black_hole)：测试 DNS 黑洞功能
- ECS 处理(ecs_handler)：测试 EDNS Client Subnet 功能

### 集成测试

- 简单 DNS 服务器：测试基本 DNS 服务功能
- 转发与缓存：测试 DNS 转发与缓存组合
- 域名过滤：测试域名匹配与黑洞/重定向组合
- 负载均衡：测试多上游服务器的负载均衡功能
