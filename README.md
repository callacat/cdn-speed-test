# 通用 CDN 优选 IP 测试工具 (CDN-Speed-Test)

## 项目简介

本项目旨在提供一个通用的、高效的CDN节点（IP）速度和可用性测试工具。灵感来源于 [XIU2/CloudflareSpeedTest](https://github.com/XIU2/CloudflareSpeedTest)，但目标是支持更广泛的CDN服务商，提供更灵活的配置和更丰富的数据维度。

用户可以通过本工具，快速地从大量的CDN IP中筛选出延迟最低、丢包率最低且连接速度最快的节点，以优化网络访问体验。

##核心功能

* [cite_start]**动态IP获取**：支持从CDN服务商的官方API动态获取IP地址段，并自动展开为独立的IP列表 [cite: 2]。
* **多维度性能探测**：综合评估IP的延迟、丢包率、HTTP连通性及下载速度。
* [cite_start]**高度并发测试**：利用并发机制，在短时间内完成对大量IP的测试，显著提升效率 [cite: 2]。
* [cite_start]**IP地理位置映射**：通过本地离线数据库（如MaxMind GeoLite2或纯真IP库）快速查询IP的地理位置信息，丰富数据维度 [cite: 3]。
* **灵活配置**：支持通过配置文件或命令行参数自定义测试参数，如并发数、延迟上下限、测速文件等。
* [cite_start]**格式化结果输出**：在终端以表格形式清晰展示优选IP结果，并支持导出为CSV文件，便于后续分析 [cite: 3]。

## 设计方案

工具的执行流程分为以下几个核心阶段：

### 阶段〇：IP集获取与准备

1.  **IP来源**：
    * [cite_start]**API动态获取**：当本地不存在IP列表文件时，工具将通过内置的或用户指定的API（例如腾讯云EdgeOne的API）获取最新的IP地址段 [cite: 2]。
    * **本地文件**：支持直接读取用户提供的 `ip.txt` 和 `ipv6.txt` 文件。
2.  [cite_start]**IP地址展开**：从API获取的CIDR格式的IP范围将被完全展开，形成一个包含所有独立IPv4/IPv6地址的初始测试池 [cite: 2]。

### 阶段一：初步筛选（延迟与丢包率）

1.  **并发探测**：
    * [cite_start]针对整个IP池，启动一个高度并发的探测过程。并发数可通过配置调整（例如，上限1000）[cite: 2]。
    * [cite_start]默认采用TCP协议对`443`端口进行多次（例如4次）“ping”测试，以计算平均延迟和丢包率。这种方式比ICMP ping更能反映CDN节点的真实网络情况 [cite: 2]。
2.  **排序与筛选**：
    * [cite_start]对所有成功探测的IP进行排序，主键为**平均延迟**（升序），次要键为**丢包率**（0%优先）[cite: 2]。
    * 根据用户配置的延迟上限（例如 `200ms`）和丢包率上限（例如 `10%`）过滤掉不合格的IP。
    * [cite_start]选取表现最佳的一部分IP（例如，延迟最低的20个）进入下一轮测试，以避免对大量低质量IP进行不必要的高成本测试 [cite: 2]。

### 阶段二：可用性与速度测试

1.  **HTTP连通性验证**：
    * [cite_start]对筛选出的候选IP，并发地发起HTTP请求，以验证其作为代理访问目标网站的实际可用性 [cite: 2]。
    * [cite_start]请求成功并返回有效的HTTP状态码（如2xx或3xx）的IP被视为“可用”[cite: 2]。
2.  **下载速度测试**（可选但建议）：
    * 对通过连通性验证的IP，并发下载一个指定大小的测速文件（URL可在配置中指定）。
    * 记录每个IP的下载速度（例如 `MB/s`），作为衡量其带宽性能的核心指标。

### 阶段三：数据丰富化与最终分析

1.  **地理位置映射**：
    * [cite_start]使用本地的IP数据库（如 `.mmdb` 格式的GeoLite2或 `ipdb` 格式的纯真IP库）为每个通过测试的IP匹配地理位置和运营商信息 [cite: 3]。
    * [cite_start]采用本地查询能有效避免在线API的网络延迟和速率限制问题 [cite: 3]。
2.  **最终排序与呈现**：
    * 将包含IP地址、平均延迟、丢包率、下载速度、地理位置等信息的最终数据集进行综合排序。
    * 排序逻辑可配置，例如：**下载速度**（降序） > **平均延迟**（升序） > **丢包率**（升序）。
    * 在命令行终端以格式化表格的形式清晰地呈现最终结果。
    * [cite_start]提供选项将完整结果保存为CSV文件，以供归档和进一步分析 [cite: 3]。

## 查询模块设计方案 (Go语言实现)

为了实现高效、可靠的IP地理位置查询，查询模块将采用以下设计：

* [cite_start]**数据源**：支持单源纯真IP库的 `qqwry.ipdb` 标准版，同时覆盖IPv4和IPv6 [cite: 3]。
* [cite_start]**优先级查询链**：优先查询本地IPDB文件。未来可扩展，当本地查询失败时，调用在线API作为兜底方案 [cite: 3]。
* [cite_start]**热更新机制**：工具可配置为定时从指定URL（如GitHub Proxy或镜像）拉取最新的IP数据库文件，并热加载到内存中，无需重启程序即可使用最新数据 [cite: 3]。
* [cite_start]**统一接口**：提供一个上层透明的统一查询接口 `Lookup(ip string) (*IPInfo, error)` [cite: 3]。

**优势**：

* [cite_start]**无外部依赖**：绝大多数查询在本地完成，性能高，不受网络和QPS限制 [cite: 3]。
* [cite_start]**双栈覆盖**：同时为IPv4和IPv6提供高质量的地理位置数据 [cite: 3]。
* [cite_start]**稳定可控**：数据源和更新策略完全可由用户自主配置和控制 [cite: 3]。

## 配置文件 (config.yaml) 示例

```yaml
# 测试相关配置
test:
  concurrency: 1000      # 最大并发数
  retries: 4             # 每个IP的延迟测试次数
  timeout: 5s            # TCP连接超时时间
  latency_max: 400       # 最大延迟 (ms)
  latency_min: 0         # 最小延迟 (ms)
  top_n: 20              # 进入第二阶段测试的IP数量

# HTTP连通性与测速配置
http:
  target_url: "[https://www.google.com/generate_204](https://www.google.com/generate_204)"  # HTTP连通性验证URL
  speed_test_url: "[https://cachefly.cachefly.net/100mb.test](https://cachefly.cachefly.net/100mb.test)" # 测速文件URL
  speed_test_timeout: 30s # 测速超时时间

# IP来源配置
ip_source:
  api_url: "[https://api.example.com/ips](https://api.example.com/ips)" # CDN服务商的IP API地址
  local_files:
    ipv4: "./ip.txt"
    ipv6: "./ipv6.txt"

# 结果输出配置
output:
  format: "table"        # 输出格式: table, csv
  csv_path: "./result.csv" # CSV文件保存路径
