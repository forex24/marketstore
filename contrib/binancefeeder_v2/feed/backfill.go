package feed

import (
	"context"
	"sync"
	"time"

	"github.com/alpacahq/marketstore/v4/contrib/binancefeeder_v2/api"
	"github.com/alpacahq/marketstore/v4/contrib/binancefeeder_v2/configs"
	"github.com/alpacahq/marketstore/v4/contrib/binancefeeder_v2/symbols"
	"github.com/alpacahq/marketstore/v4/contrib/binancefeeder_v2/writer"
	"github.com/alpacahq/marketstore/v4/utils/log"
)

// Backfill 历史数据回填
type Backfill struct {
	symbolManager *symbols.Manager
	apiClient     *api.Client
	writer        *writer.Writer
	config        configs.BackfillConfig
	workerPool    chan struct{}
}

// NewBackfill 创建新的backfill实例
func NewBackfill(symbolManager *symbols.Manager, apiClient *api.Client, writer *writer.Writer, config configs.BackfillConfig) *Backfill {
	// 创建工作池
	workerPool := make(chan struct{}, config.Parallelism)
	for i := 0; i < config.Parallelism; i++ {
		workerPool <- struct{}{}
	}

	return &Backfill{
		symbolManager: symbolManager,
		apiClient:     apiClient,
		writer:        writer,
		config:        config,
		workerPool:    workerPool,
	}
}

// Run 运行backfill
func (b *Backfill) Run(ctx context.Context) {
	log.Info("Starting Binance backfill from %s to %s", 
		b.config.StartTime.Format("2006-01-02 15:04:05"),
		b.config.EndTime.Format("2006-01-02 15:04:05"))

	// 获取所有符号
	symbols := b.symbolManager.GetAllSymbols()
	if len(symbols) == 0 {
		log.Error("No symbols available for backfill")
		return
	}

	log.Info("Backfilling data for %d symbols", len(symbols))

	// 创建等待组
	var wg sync.WaitGroup

	// 为每个符号启动backfill
	for _, symbol := range symbols {
		select {
		case <-ctx.Done():
			log.Info("Backfill cancelled")
			return
		case <-b.workerPool:
			wg.Add(1)
			go func(sym string) {
				defer wg.Done()
				defer func() { b.workerPool <- struct{}{} }()
				b.backfillSymbol(ctx, sym)
			}(symbol)
		}
	}

	// 等待所有goroutine完成
	wg.Wait()
	log.Info("Backfill completed")
}

// backfillSymbol 为单个符号执行backfill
func (b *Backfill) backfillSymbol(ctx context.Context, symbol string) {
	log.Info("Starting backfill for symbol: %s", symbol)

	// 计算时间范围
	startTime := b.config.StartTime
	endTime := b.config.EndTime

	// 分批处理数据
	for currentTime := startTime; currentTime.Before(endTime); {
		select {
		case <-ctx.Done():
			log.Info("Backfill cancelled for symbol: %s", symbol)
			return
		default:
		}

		// 计算批次结束时间
		batchEndTime := currentTime.Add(time.Duration(b.config.BatchSize) * getIntervalDuration(b.config.Interval))
		if batchEndTime.After(endTime) {
			batchEndTime = endTime
		}

		// 获取K线数据
		klines, err := b.apiClient.GetKlines(ctx, symbol, b.config.Interval, currentTime, batchEndTime, b.config.BatchSize)
		if err != nil {
			log.Error("Failed to get klines for %s: %v", symbol, err)
			// 继续下一个批次
			currentTime = batchEndTime
			continue
		}

		if len(klines) > 0 {
			// 写入数据
			if err := b.writer.WriteKlines(symbol, klines); err != nil {
				log.Error("Failed to write klines for %s: %v", symbol, err)
			} else {
				log.Info("Backfilled %d klines for %s from %s to %s", 
					len(klines), symbol,
					time.Unix(klines[0].OpenTime/1000, 0).Format("2006-01-02 15:04:05"),
					time.Unix(klines[len(klines)-1].OpenTime/1000, 0).Format("2006-01-02 15:04:05"))
			}
		}

		// 更新当前时间
		currentTime = batchEndTime

		// 添加延迟避免API限制 - 基于币安官方限制
		// 币安限制：1200 requests/minute，我们使用1000 requests/minute
		// 每个请求间隔：60秒/1000 = 60毫秒
		// 为了安全起见，我们使用100毫秒的间隔
		time.Sleep(100 * time.Millisecond)
	}

	log.Info("Completed backfill for symbol: %s", symbol)
}

// backfillAggregatedData 回填聚合数据
func (b *Backfill) backfillAggregatedData(ctx context.Context, symbol string, timeframes []string) {
	for _, timeframe := range timeframes {
		select {
		case <-ctx.Done():
			return
		default:
		}

		log.Info("Backfilling aggregated data for %s/%s", symbol, timeframe)

		// 获取聚合数据
		klines, err := b.apiClient.GetKlines(ctx, symbol, timeframe, b.config.StartTime, b.config.EndTime, b.config.BatchSize)
		if err != nil {
			log.Error("Failed to get aggregated klines for %s/%s: %v", symbol, timeframe, err)
			continue
		}

		if len(klines) > 0 {
			// 写入聚合数据
			if err := b.writer.WriteAggregatedData(symbol, timeframe, klines); err != nil {
				log.Error("Failed to write aggregated data for %s/%s: %v", symbol, timeframe, err)
			}
		}

		// 添加延迟 - 聚合数据需要更长的延迟
		// 币安限制：1200 requests/minute，我们使用1000 requests/minute
		time.Sleep(200 * time.Millisecond)
	}
}

// backfillTrades 回填交易数据
func (b *Backfill) backfillTrades(ctx context.Context, symbol string) {
	log.Info("Backfilling trades for symbol: %s", symbol)

	// 分批获取交易数据
	startTime := b.config.StartTime
	endTime := b.config.EndTime

	for currentTime := startTime; currentTime.Before(endTime); {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 获取交易数据
		trades, err := b.apiClient.GetRecentTrades(ctx, symbol, b.config.BatchSize)
		if err != nil {
			log.Error("Failed to get trades for %s: %v", symbol, err)
			break
		}

		if len(trades) > 0 {
			// 写入交易数据
			if err := b.writer.WriteTrades(symbol, trades); err != nil {
				log.Error("Failed to write trades for %s: %v", symbol, err)
			}
		}

		// 更新时间和延迟 - 交易数据需要更长的延迟
		// 币安限制：1200 requests/minute，我们使用1000 requests/minute
		currentTime = currentTime.Add(time.Hour) // 每小时获取一次
		time.Sleep(500 * time.Millisecond)
	}
}

// getIntervalDuration 获取时间间隔的持续时间
func getIntervalDuration(interval string) time.Duration {
	switch interval {
	case "1m":
		return time.Minute
	case "3m":
		return 3 * time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "30m":
		return 30 * time.Minute
	case "1h":
		return time.Hour
	case "2h":
		return 2 * time.Hour
	case "4h":
		return 4 * time.Hour
	case "6h":
		return 6 * time.Hour
	case "8h":
		return 8 * time.Hour
	case "12h":
		return 12 * time.Hour
	case "1d":
		return 24 * time.Hour
	case "3d":
		return 3 * 24 * time.Hour
	case "1w":
		return 7 * 24 * time.Hour
	case "1M":
		return 30 * 24 * time.Hour
	default:
		return time.Minute
	}
}

// GetProgress 获取进度信息
func (b *Backfill) GetProgress() map[string]interface{} {
	return map[string]interface{}{
		"start_time": b.config.StartTime,
		"end_time":   b.config.EndTime,
		"interval":   b.config.Interval,
		"batch_size": b.config.BatchSize,
		"parallelism": b.config.Parallelism,
	}
} 