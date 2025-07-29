package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/alpacahq/marketstore/v4/contrib/binancefeeder_v2/api"
	"github.com/alpacahq/marketstore/v4/contrib/binancefeeder_v2/configs"
	"github.com/alpacahq/marketstore/v4/contrib/binancefeeder_v2/feed"
	"github.com/alpacahq/marketstore/v4/contrib/binancefeeder_v2/symbols"
	"github.com/alpacahq/marketstore/v4/contrib/binancefeeder_v2/writer"
	"github.com/alpacahq/marketstore/v4/plugins/bgworker"
	"github.com/alpacahq/marketstore/v4/utils/log"
)

// BinanceFeederV2 币安数据源插件
type BinanceFeederV2 struct {
	config         *configs.Config
	symbolManager  *symbols.Manager
	apiClient      *api.Client
	writer         *writer.Writer
	backfill       *feed.Backfill
	realtime       *feed.Realtime
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// NewBgWorker 创建新的后台工作器
func NewBgWorker(conf map[string]interface{}) (bgworker.BgWorker, error) {
	log.Info("DEBUG: NewBgWorker called with config: %+v", conf)
	
	// 解析配置
	config, err := configs.ParseConfig(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	
	log.Info("DEBUG: Parsed config - Backfill.Enabled: %v", config.Backfill.Enabled)

	// 创建API客户端
	apiClient := api.NewClient(config.APIKey, config.SecretKey, config.Testnet)

	// 创建符号管理器
	symbolManager := symbols.NewManager(config.Symbols, config.ExcludeSymbols)
	symbolManager.SetAPIClient(apiClient) // 设置API客户端
	if err := symbolManager.LoadAllSymbols(context.Background()); err != nil {
		log.Error("Failed to load all symbols: %v", err)
	}

	// 创建数据写入器
	writer := writer.NewWriter(config.Timeframe, "UTC")

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 创建backfill实例
	backfill := feed.NewBackfill(symbolManager, apiClient, writer, config.Backfill)

	// 创建实时数据实例
	realtime := feed.NewRealtime(symbolManager, apiClient, writer, config.Realtime)

	return &BinanceFeederV2{
		config:        config,
		symbolManager: symbolManager,
		apiClient:     apiClient,
		writer:        writer,
		backfill:      backfill,
		realtime:      realtime,
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Run 运行插件
func (bf *BinanceFeederV2) Run() {
	log.Info("Starting BinanceFeederV2 plugin...")
	log.Info("Backfill enabled: %v", bf.config.Backfill.Enabled)

	// 启动backfill（如果启用）
	if bf.config.Backfill.Enabled {
		bf.wg.Add(1)
		go func() {
			defer bf.wg.Done()
			bf.backfill.Run(bf.ctx)
		}()
		log.Info("Backfill enabled and started")
	}

	// 启动实时数据（如果启用）
	if bf.config.Realtime.Enabled {
		bf.wg.Add(1)
		go func() {
			defer bf.wg.Done()
			bf.realtime.Run(bf.ctx)
		}()
		log.Info("Realtime data enabled and started")
	}

	// 等待上下文取消
	<-bf.ctx.Done()
	log.Info("BinanceFeederV2 plugin stopping...")

	// 等待所有goroutine完成
	bf.wg.Wait()
	log.Info("BinanceFeederV2 plugin stopped")
}

// Shutdown 关闭插件
func (bf *BinanceFeederV2) Shutdown() {
	bf.cancel()
}

func main() {} 