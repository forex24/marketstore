# 插件开发

MarketStore提供了允许第三方Go插件模块集成的接口。这些接口有两种类型：`trigger`插件和`bgworker`插件。

第三方插件可以使用Go的`build`命令通过`-buildmode=plugin`标志构建为`.so`包，并放置在$GOPATH/bin目录中。放置在那里后，可以通过`triggers`或`bgworkers`标志在提供给`marketstore`启动命令的MarketStore YAML配置文件中引用它们。

当在MarketStore YAML配置中包含和配置插件时，插件会在`marketstore`命令启动时启动。包含的`mkts.yml`文件显示了一些配置的注释示例。

## 触发器（Trigger）
触发器是在向数据库写入匹配特定参数的数据时执行操作的小型应用程序。触发器接口必须实现以下函数：
```go
NewTrigger(config map[string]interface{}) (Trigger, error)
```

此函数返回的触发器实例将在Fire()时被调用，参数为文件路径（相对于根目录）和已写入（追加或更新）的索引。当调用Fire()时，保证新内容已写入磁盘，因此从磁盘读取是安全的。请记住，由于WAL恢复，触发器可能在启动时被调用。

### 配置示例
```
triggers:
  - module: xxxTrigger.so
    on: "*/1Min/OHLCV"
    config: <根据插件配置>
```
"on"值与文件路径匹配以决定是否触发触发器。它可以包含通配符"*"。目前，触发器仅在运行状态下触发。WAL重放时的触发器可能在以后添加。

### 包含的插件
* [磁盘聚合](https://github.com/alpacahq/marketstore/tree/master/contrib/ondiskagg) - 在底层时间框架写入时更新下采样数据。
* [流式传输](https://github.com/alpacahq/marketstore/tree/master/contrib/stream) - 通过MarketStore的流式接口推送数据。

## 后台工作器（BgWorker）
作为独立进程在后台运行的小型应用程序，可以对数据库执行tick数据事务。bgworker接口必须实现以下函数：
```go
NewBgWorker(config map[string]interface{}) (BgWorker, error)
```

后台工作器通过实现接口在MarketStore服务器下运行，在查询接口启动之前的服务器生命周期开始时启动。MarketStore服务器不处理插件内发生的panic。插件可以从panic中恢复，但如果接触内部API，应该小心不要搞乱MarketStore服务器状态。通常最好让它继续运行。

### 配置示例
```
bgworkers:
  - module: xxxWorker.so
    name: datafeed
    config: <根据插件配置>
```

### 包含的插件
* [GDAXFeeder](https://github.com/alpacahq/marketstore/tree/master/contrib/gdaxfeeder) - 从GDAX公共API获取加密货币的历史价格数据。
* [Polygon](https://github.com/alpacahq/marketstore/tree/master/contrib/polygon) - 从[Polygon的API](https://polygon.io/)获取美国股票的历史价格数据。 