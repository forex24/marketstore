# MarketStore客户端查询Symbol列表指南

## 1. 概述

MarketStore提供了`ListSymbols` API来查询数据库中支持的所有symbol列表。这个API可以通过多种客户端方式调用，包括Go客户端、Python客户端和直接的HTTP请求。

## 2. API接口定义

### 2.1 请求结构
```go
type ListSymbolsRequest struct {
    // "symbol", 或 "tbk"
    Format string `msgpack:"format,omitempty"`
}
```

### 2.2 响应结构
```go
type ListSymbolsResponse struct {
    Results []string
}
```

### 2.3 支持的格式

#### 2.3.1 Symbol格式 (`format: "symbol"`)
返回所有可用的symbol名称，例如：
```json
{
    "Results": ["AAPL", "GOOGL", "MSFT", "BTCUSDT", "ETHUSDT"]
}
```

#### 2.3.2 TBK格式 (`format: "tbk"`)
返回完整的TimeBucketKey格式，包含symbol、时间框架和属性组，例如：
```json
{
    "Results": [
        "AAPL/1Min/OHLCV",
        "AAPL/1H/OHLCV", 
        "GOOGL/1Min/OHLCV",
        "BTCUSDT/1Min/OHLCV",
        "BTCUSDT/1H/OHLCV"
    ]
}
```

## 3. 客户端调用方式

### 3.1 Go客户端

#### 3.1.1 创建客户端
```go
package main

import (
    "fmt"
    "github.com/alpacahq/marketstore/v4/frontend/client"
)

func main() {
    // 创建客户端连接
    cl, err := client.NewClient("http://localhost:5993/rpc")
    if err != nil {
        panic(err)
    }
    
    // 查询symbol列表
    symbols, err := cl.DoRPC("ListSymbols", &frontend.ListSymbolsRequest{
        Format: "symbol",
    })
    if err != nil {
        panic(err)
    }
    
    // 类型断言
    symbolList := symbols.([]string)
    fmt.Printf("Available symbols: %v\n", symbolList)
}
```

#### 3.1.2 查询TBK格式
```go
// 查询完整的TBK格式
tbks, err := cl.DoRPC("ListSymbols", &frontend.ListSymbolsRequest{
    Format: "tbk",
})
if err != nil {
    panic(err)
}

tbkList := tbks.([]string)
for _, tbk := range tbkList {
    fmt.Printf("TBK: %s\n", tbk)
}
```

### 3.2 Python客户端 (pymarketstore)

#### 3.2.1 安装客户端
```bash
pip install pymarketstore
```

#### 3.2.2 基本使用
```python
import pymarketstore as pymkts

# 创建客户端
client = pymkts.Client("http://localhost:5993/rpc")

# 查询symbol列表 (需要直接调用RPC)
# 注意：pymarketstore可能没有直接暴露ListSymbols方法
# 需要使用底层RPC调用
```

#### 3.2.3 使用底层RPC调用
```python
import requests
import msgpack

# 创建ListSymbols请求
request_data = {
    "method": "DataService.ListSymbols",
    "params": [{"format": "symbol"}]
}

# 发送请求
response = requests.post(
    "http://localhost:5993/rpc",
    data=msgpack.packb(request_data),
    headers={"Content-Type": "application/x-msgpack"}
)

# 解析响应
if response.status_code == 200:
    result = msgpack.unpackb(response.content)
    symbols = result.get("result", {}).get("Results", [])
    print(f"Available symbols: {symbols}")
```

### 3.3 直接HTTP请求

#### 3.3.1 使用curl (需要msgpack工具)
```bash
# 安装msgpack工具
pip install msgpack-python

# 创建请求数据
python3 -c "
import msgpack
import json
data = {'method':'DataService.ListSymbols','params':[{'format':'symbol'}]}
packed = msgpack.packb(data)
print(''.join([f'{b:02x}' for b in packed]))
" > request.hex

# 发送请求
curl -X POST http://localhost:5993/rpc \
  -H "Content-Type: application/x-msgpack" \
  --data-binary @request.hex
```

#### 3.3.2 使用Go程序发送请求
```go
package main

import (
    "bytes"
    "fmt"
    "io"
    "net/http"
    "github.com/vmihailenco/msgpack/v5"
)

func main() {
    // 创建请求数据
    request := map[string]interface{}{
        "method": "DataService.ListSymbols",
        "params": []interface{}{
            map[string]interface{}{
                "format": "symbol",
            },
        },
    }
    
    // 编码为msgpack
    data, err := msgpack.Marshal(request)
    if err != nil {
        panic(err)
    }
    
    // 发送HTTP请求
    resp, err := http.Post(
        "http://localhost:5993/rpc",
        "application/x-msgpack",
        bytes.NewReader(data),
    )
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    // 读取响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }
    
    // 解码响应
    var result map[string]interface{}
    err = msgpack.Unmarshal(body, &result)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Response: %+v\n", result)
}
```

## 4. 实际使用示例

### 4.1 查询所有可用的symbol
```go
func getAllSymbols(client *client.Client) ([]string, error) {
    result, err := client.DoRPC("ListSymbols", &frontend.ListSymbolsRequest{
        Format: "symbol",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to list symbols: %w", err)
    }
    
    symbols, ok := result.([]string)
    if !ok {
        return nil, fmt.Errorf("unexpected response type")
    }
    
    return symbols, nil
}
```

### 4.2 查询特定symbol的所有时间框架
```go
func getTimeframesForSymbol(client *client.Client, symbol string) ([]string, error) {
    result, err := client.DoRPC("ListSymbols", &frontend.ListSymbolsRequest{
        Format: "tbk",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to list TBKs: %w", err)
    }
    
    tbks, ok := result.([]string)
    if !ok {
        return nil, fmt.Errorf("unexpected response type")
    }
    
    var timeframes []string
    for _, tbk := range tbks {
        // 解析TBK格式: symbol/timeframe/attributegroup
        parts := strings.Split(tbk, "/")
        if len(parts) == 3 && parts[0] == symbol {
            timeframes = append(timeframes, parts[1])
        }
    }
    
    return timeframes, nil
}
```

### 4.3 查询特定symbol的所有属性组
```go
func getAttributeGroupsForSymbol(client *client.Client, symbol string) ([]string, error) {
    result, err := client.DoRPC("ListSymbols", &frontend.ListSymbolsRequest{
        Format: "tbk",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to list TBKs: %w", err)
    }
    
    tbks, ok := result.([]string)
    if !ok {
        return nil, fmt.Errorf("unexpected response type")
    }
    
    var attributeGroups []string
    for _, tbk := range tbks {
        parts := strings.Split(tbk, "/")
        if len(parts) == 3 && parts[0] == symbol {
            attributeGroups = append(attributeGroups, parts[2])
        }
    }
    
    return attributeGroups, nil
}
```

## 5. 错误处理

### 5.1 常见错误
```go
func handleListSymbolsError(err error) {
    if err != nil {
        switch {
        case strings.Contains(err.Error(), "connection refused"):
            fmt.Println("MarketStore服务未运行")
        case strings.Contains(err.Error(), "timeout"):
            fmt.Println("请求超时")
        case strings.Contains(err.Error(), "not queryable"):
            fmt.Println("服务器不可查询")
        default:
            fmt.Printf("未知错误: %v\n", err)
        }
    }
}
```

### 5.2 重试机制
```go
func listSymbolsWithRetry(client *client.Client, maxRetries int) ([]string, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        result, err := client.DoRPC("ListSymbols", &frontend.ListSymbolsRequest{
            Format: "symbol",
        })
        if err == nil {
            symbols, ok := result.([]string)
            if ok {
                return symbols, nil
            }
        }
        
        lastErr = err
        time.Sleep(time.Duration(i+1) * time.Second) // 指数退避
    }
    
    return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
```

## 6. 性能优化

### 6.1 缓存symbol列表
```go
type SymbolCache struct {
    symbols    []string
    lastUpdate time.Time
    mutex      sync.RWMutex
    ttl        time.Duration
}

func (sc *SymbolCache) GetSymbols(client *client.Client) ([]string, error) {
    sc.mutex.RLock()
    if time.Since(sc.lastUpdate) < sc.ttl {
        symbols := make([]string, len(sc.symbols))
        copy(symbols, sc.symbols)
        sc.mutex.RUnlock()
        return symbols, nil
    }
    sc.mutex.RUnlock()
    
    // 更新缓存
    sc.mutex.Lock()
    defer sc.mutex.Unlock()
    
    result, err := client.DoRPC("ListSymbols", &frontend.ListSymbolsRequest{
        Format: "symbol",
    })
    if err != nil {
        return nil, err
    }
    
    symbols, ok := result.([]string)
    if !ok {
        return nil, fmt.Errorf("unexpected response type")
    }
    
    sc.symbols = symbols
    sc.lastUpdate = time.Now()
    
    return symbols, nil
}
```

## 7. 总结

MarketStore客户端可以通过以下方式查询支持的symbol列表：

1. **API方法**: `DataService.ListSymbols`
2. **支持的格式**: 
   - `"symbol"`: 返回symbol名称列表
   - `"tbk"`: 返回完整的TimeBucketKey列表
3. **客户端支持**:
   - Go客户端: 直接支持
   - Python客户端: 需要底层RPC调用
   - HTTP客户端: 支持msgpack协议

这个API对于以下场景非常有用：
- 动态发现可用的交易对
- 验证symbol是否存在
- 获取symbol的所有可用时间框架和属性组
- 构建动态查询界面 