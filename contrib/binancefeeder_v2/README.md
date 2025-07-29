# BinanceFeederV2 Plugin for MarketStore

一个功能完整的币安数据源插件，使用最新的API，支持实时数据流和历史数据回填。

## 🚀 特性

### ✅ 核心功能
- **实时数据流** - 通过WebSocket获取实时K线、交易和深度数据
- **历史数据回填** - 支持批量下载历史数据
- **多时间框架** - 支持1m/3m/5m/15m/30m/1h/4h/1d/1w等时间框架
- **数据聚合** - 自动聚合tick数据到更高时间框架
- **并行处理** - 支持多符号并行数据获取
- **错误恢复** - 自动重试和错误处理机制

### 📊 支持的数据类型
- **K线数据 (OHLCV)** - 开盘价、最高价、最低价、收盘价、成交量
- **交易数据** - 实时交易记录
- **深度数据** - 订单簿深度信息
- **聚合数据** - 自动聚合到不同时间框架

### 🔧 技术特点
- **最新API** - 使用币安最新的v3 API
- **WebSocket支持** - 实时数据流
- **REST API** - 历史数据获取
- **插件架构** - 符合MarketStore插件规范
- **配置灵活** - 丰富的配置选项

## 📦 安装

### 1. 编译插件
```bash
cd contrib/binancefeeder_v2
make build
```

### 2. 安装到GOPATH
```bash
make install
```

### 3. 配置MarketStore
复制示例配置文件并修改：
```bash
cp mkts.example.yml /path/to/your/mkts.yml
```

## ⚙️ 配置

### 基础配置
```yaml
bgworkers:
  - module: binancefeeder_v2.so
    name: BinanceFeederV2
    config:
      # API配置
      api_key: "your_binance_api_key_here"
      secret_key: "your_binance_secret_key_here"
      testnet: false

      # 符号配置
      symbols:
        - "BTCUSDT"
        - "ETHUSDT"
        - "BNBUSDT"
      
      # 时间框架
      timeframe: "1m"
```

### Backfill配置
```yaml
      backfill:
        enabled: true
        start_time: "2024-01-01T00:00:00Z"
        end_time: "2024-01-31T23:59:59Z"
        batch_size: 1000
        parallelism: 5
        interval: "1m"
```

### 实时数据配置
```yaml
      realtime:
        enabled: true
        stream_type: "kline"  # kline, trade, depth
        update_freq: "1m"
        buffer_size: 1000
        max_retries: 3
        retry_delay: "5s"
```

## 🎯 使用方法

### 1. 启动MarketStore
```bash
marketstore start --config mkts.yml
```

### 2. 查看日志
```bash
tail -f marketstore.log
```

### 3. 查询数据
```python
import pymarketstore as pymkts

client = pymkts.Client()
data = client.query('BTCUSDT/1m/OHLCV', start='2024-01-01', end='2024-01-02')
print(data)
```

## 📈 数据格式

### K线数据 (OHLCV)
```python
{
    'Epoch': [1640995200, 1640995260, ...],
    'Open': [46200.0, 46250.0, ...],
    'High': [46300.0, 46350.0, ...],
    'Low': [46100.0, 46150.0, ...],
    'Close': [46250.0, 46300.0, ...],
    'Volume': [100.5, 150.2, ...]
}
```

### 交易数据
```python
{
    'Epoch': [1640995200, 1640995201, ...],
    'Price': [46200.0, 46201.0, ...],
    'Size': [0.1, 0.05, ...],
    'Exchange': ['BINANCE', 'BINANCE', ...],
    'Tape': ['SPOT', 'SPOT', ...]
}
```

## 🔄 时间框架映射

| 币安间隔 | MarketStore时间框架 | 描述 |
|---------|-------------------|------|
| 1m      | 1Min              | 1分钟 |
| 3m      | 3Min              | 3分钟 |
| 5m      | 5Min              | 5分钟 |
| 15m     | 15Min             | 15分钟 |
| 30m     | 30Min             | 30分钟 |
| 1h      | 1H                | 1小时 |
| 4h      | 4H                | 4小时 |
| 1d      | 1D                | 1天 |
| 1w      | 1W                | 1周 |

## 🛠️ 开发

### 项目结构
```
binancefeeder_v2/
├── api/           # API客户端
├── configs/       # 配置管理
├── feed/          # 数据获取模块
├── symbols/       # 符号管理
├── writer/        # 数据写入
├── mkts.example.yml
├── Makefile
└── README.md
```

### 编译
```bash
make build
```

### 测试
```bash
make test
```

### 代码格式化
```bash
make fmt
```

## 📋 API限制

### 币安API限制
- **REST API**: 1200 requests/minute
- **WebSocket**: 5 connections per IP
- **K线数据**: 1000条/请求
- **交易数据**: 1000条/请求

### 插件优化
- 自动限流控制
- 连接池管理
- 批量处理
- 错误重试

## 🚨 注意事项

### 1. API密钥
- 请妥善保管API密钥
- 建议使用只读权限的API密钥
- 生产环境请使用环境变量

### 2. 数据存储
- 确保有足够的磁盘空间
- 定期清理旧数据
- 监控磁盘使用情况

### 3. 网络连接
- 确保网络连接稳定
- 配置合适的超时时间
- 监控连接状态

## 🔧 故障排除

### 常见问题

#### 1. 连接失败
```
Error: Failed to connect to WebSocket
```
**解决方案**: 检查网络连接和防火墙设置

#### 2. API限制
```
Error: API request failed: 429
```
**解决方案**: 减少并发请求数量或增加延迟

#### 3. 数据写入失败
```
Error: Failed to write klines
```
**解决方案**: 检查磁盘空间和权限

### 日志级别
```yaml
log_level: info  # debug, info, warn, error
```

## 📞 支持

### 问题报告
请在GitHub Issues中报告问题，并提供：
- 错误日志
- 配置信息
- 复现步骤

### 贡献
欢迎提交Pull Request来改进插件功能。

## 📄 许可证

本项目遵循MarketStore的许可证条款。

## 🔗 相关链接

- [MarketStore文档](https://github.com/alpacahq/marketstore)
- [币安API文档](https://binance-docs.github.io/apidocs/spot/en/)
- [币安WebSocket文档](https://binance-docs.github.io/apidocs/spot/en/#websocket-market-streams) 