package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alpacahq/marketstore/v4/contrib/binancefeeder_v2/api"
	"github.com/alpacahq/marketstore/v4/contrib/binancefeeder_v2/configs"
	"github.com/alpacahq/marketstore/v4/contrib/binancefeeder_v2/symbols"
	"github.com/alpacahq/marketstore/v4/contrib/binancefeeder_v2/writer"
	"github.com/alpacahq/marketstore/v4/utils/log"
	"github.com/gorilla/websocket"
)

// Realtime 实时数据处理器
type Realtime struct {
	symbolManager *symbols.Manager
	apiClient     *api.Client
	writer        *writer.Writer
	config        configs.RealtimeConfig
	connections   map[string]*websocket.Conn
	mutex         sync.RWMutex
	aggregators   map[string]*DataAggregator
}

// DataAggregator 数据聚合器
type DataAggregator struct {
	symbol    string
	timeframe string
	buffer    []api.KlineData
	mutex     sync.Mutex
	writer    *writer.Writer
}

// NewRealtime 创建新的实时数据处理器
func NewRealtime(symbolManager *symbols.Manager, apiClient *api.Client, writer *writer.Writer, config configs.RealtimeConfig) *Realtime {
	return &Realtime{
		symbolManager: symbolManager,
		apiClient:     apiClient,
		writer:        writer,
		config:        config,
		connections:   make(map[string]*websocket.Conn),
		aggregators:   make(map[string]*DataAggregator),
	}
}

// Run 运行实时数据处理器
func (r *Realtime) Run(ctx context.Context) {
	log.Info("Starting Binance realtime data feed")

	// 获取所有符号
	symbols := r.symbolManager.GetAllSymbols()
	if len(symbols) == 0 {
		log.Error("No symbols available for realtime feed")
		return
	}

	// 根据流类型启动相应的处理器
	switch r.config.StreamType {
	case "kline":
		r.startKlineStreams(ctx, symbols)
	case "trade":
		r.startTradeStreams(ctx, symbols)
	case "depth":
		r.startDepthStreams(ctx, symbols)
	default:
		log.Error("Unsupported stream type: %s", r.config.StreamType)
	}
}

// startKlineStreams 启动K线数据流
func (r *Realtime) startKlineStreams(ctx context.Context, symbols []string) {
	// 创建WebSocket连接
	wsURL := r.apiClient.GetWebSocketURL()
	
	// 构建流名称
	var streams []string
	for _, symbol := range symbols {
		streams = append(streams, fmt.Sprintf("%s@kline_%s", strings.ToLower(symbol), r.config.UpdateFreq))
	}

	// 连接WebSocket
	u, err := url.Parse(wsURL)
	if err != nil {
		log.Error("Failed to parse WebSocket URL: %v", err)
		return
	}

	// 添加流参数
	q := u.Query()
	q.Set("streams", strings.Join(streams, "/"))
	u.RawQuery = q.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error("Failed to connect to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	log.Info("Connected to Binance WebSocket for %d symbols", len(symbols))

	// 启动消息处理器
	go r.handleKlineMessages(ctx, conn, symbols)

	// 保持连接
	for {
		select {
		case <-ctx.Done():
			log.Info("Realtime feed cancelled")
			return
		default:
			// 发送ping保持连接
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Error("Failed to send ping: %v", err)
				return
			}
			time.Sleep(30 * time.Second)
		}
	}
}

// handleKlineMessages 处理K线消息
func (r *Realtime) handleKlineMessages(ctx context.Context, conn *websocket.Conn, symbols []string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Error("Failed to read WebSocket message: %v", err)
				return
			}

			// 解析消息
			var wsMessage struct {
				Stream string `json:"stream"`
				Data   struct {
					Symbol string `json:"s"`
					Kline  struct {
						OpenTime         int64   `json:"t"`
						Open             string  `json:"o"`
						High             string  `json:"h"`
						Low              string  `json:"l"`
						Close            string  `json:"c"`
						Volume           string  `json:"v"`
						CloseTime        int64   `json:"T"`
						QuoteAssetVolume string  `json:"q"`
						NumberOfTrades   int64   `json:"n"`
						TakerBuyBase     string  `json:"V"`
						TakerBuyQuote    string  `json:"Q"`
						IsFinal          bool    `json:"x"`
					} `json:"k"`
				} `json:"data"`
			}

			if err := json.Unmarshal(message, &wsMessage); err != nil {
				log.Error("Failed to unmarshal WebSocket message: %v", err)
				continue
			}

			// 处理K线数据
			if wsMessage.Data.Kline.IsFinal {
				r.processKlineData(wsMessage.Data.Symbol, wsMessage.Data.Kline)
			}
		}
	}
}

// processKlineData 处理K线数据
func (r *Realtime) processKlineData(symbol string, klineData struct {
	OpenTime         int64   `json:"t"`
	Open             string  `json:"o"`
	High             string  `json:"h"`
	Low              string  `json:"l"`
	Close            string  `json:"c"`
	Volume           string  `json:"v"`
	CloseTime        int64   `json:"T"`
	QuoteAssetVolume string  `json:"q"`
	NumberOfTrades   int64   `json:"n"`
	TakerBuyBase     string  `json:"V"`
	TakerBuyQuote    string  `json:"Q"`
	IsFinal          bool    `json:"x"`
}) {
	// 转换数据
	open, _ := strconv.ParseFloat(klineData.Open, 64)
	high, _ := strconv.ParseFloat(klineData.High, 64)
	low, _ := strconv.ParseFloat(klineData.Low, 64)
	close, _ := strconv.ParseFloat(klineData.Close, 64)
	volume, _ := strconv.ParseFloat(klineData.Volume, 64)
	quoteVolume, _ := strconv.ParseFloat(klineData.QuoteAssetVolume, 64)
	takerBuyBase, _ := strconv.ParseFloat(klineData.TakerBuyBase, 64)
	takerBuyQuote, _ := strconv.ParseFloat(klineData.TakerBuyQuote, 64)

	kline := api.KlineData{
		OpenTime:         klineData.OpenTime,
		Open:             open,
		High:             high,
		Low:              low,
		Close:            close,
		Volume:           volume,
		CloseTime:        klineData.CloseTime,
		QuoteAssetVolume: quoteVolume,
		NumberOfTrades:   klineData.NumberOfTrades,
		TakerBuyBase:     takerBuyBase,
		TakerBuyQuote:    takerBuyQuote,
	}

	// 写入数据
	if err := r.writer.WriteKlines(symbol, []api.KlineData{kline}); err != nil {
		log.Error("Failed to write realtime kline for %s: %v", symbol, err)
	}

	// 聚合数据
	r.aggregateData(symbol, kline)
}

// startTradeStreams 启动交易数据流
func (r *Realtime) startTradeStreams(ctx context.Context, symbols []string) {
	// 创建WebSocket连接
	wsURL := r.apiClient.GetWebSocketURL()
	
	// 构建流名称
	var streams []string
	for _, symbol := range symbols {
		streams = append(streams, fmt.Sprintf("%s@trade", strings.ToLower(symbol)))
	}

	// 连接WebSocket
	u, err := url.Parse(wsURL)
	if err != nil {
		log.Error("Failed to parse WebSocket URL: %v", err)
		return
	}

	// 添加流参数
	q := u.Query()
	q.Set("streams", strings.Join(streams, "/"))
	u.RawQuery = q.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error("Failed to connect to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	log.Info("Connected to Binance WebSocket for trade streams")

	// 启动消息处理器
	go r.handleTradeMessages(ctx, conn)

	// 保持连接
	for {
		select {
		case <-ctx.Done():
			log.Info("Trade stream cancelled")
			return
		default:
			// 发送ping保持连接
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Error("Failed to send ping: %v", err)
				return
			}
			time.Sleep(30 * time.Second)
		}
	}
}

// handleTradeMessages 处理交易消息
func (r *Realtime) handleTradeMessages(ctx context.Context, conn *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Error("Failed to read WebSocket message: %v", err)
				return
			}

			// 解析消息
			var wsMessage struct {
				Stream string `json:"stream"`
				Data   struct {
					Symbol    string  `json:"s"`
					ID        int64   `json:"t"`
					Price     string  `json:"p"`
					Quantity  string  `json:"q"`
					QuoteQty  string  `json:"Q"`
					Time      int64   `json:"T"`
					IsBuyerMaker bool `json:"m"`
				} `json:"data"`
			}

			if err := json.Unmarshal(message, &wsMessage); err != nil {
				log.Error("Failed to unmarshal WebSocket message: %v", err)
				continue
			}

			// 处理交易数据
			r.processTradeData(wsMessage.Data)
		}
	}
}

// processTradeData 处理交易数据
func (r *Realtime) processTradeData(tradeData struct {
	Symbol    string  `json:"s"`
	ID        int64   `json:"t"`
	Price     string  `json:"p"`
	Quantity  string  `json:"q"`
	QuoteQty  string  `json:"Q"`
	Time      int64   `json:"T"`
	IsBuyerMaker bool `json:"m"`
}) {
	// 转换数据
	price, _ := strconv.ParseFloat(tradeData.Price, 64)
	quantity, _ := strconv.ParseFloat(tradeData.Quantity, 64)
	quoteQty, _ := strconv.ParseFloat(tradeData.QuoteQty, 64)

	trade := api.TradeData{
		ID:           tradeData.ID,
		Price:        price,
		Quantity:     quantity,
		QuoteQty:     quoteQty,
		Time:         tradeData.Time,
		IsBuyerMaker: tradeData.IsBuyerMaker,
	}

	// 写入数据
	if err := r.writer.WriteTrades(tradeData.Symbol, []api.TradeData{trade}); err != nil {
		log.Error("Failed to write realtime trade for %s: %v", tradeData.Symbol, err)
	}
}

// startDepthStreams 启动深度数据流
func (r *Realtime) startDepthStreams(ctx context.Context, symbols []string) {
	// 创建WebSocket连接
	wsURL := r.apiClient.GetWebSocketURL()
	
	// 构建流名称
	var streams []string
	for _, symbol := range symbols {
		streams = append(streams, fmt.Sprintf("%s@depth20@100ms", strings.ToLower(symbol)))
	}

	// 连接WebSocket
	u, err := url.Parse(wsURL)
	if err != nil {
		log.Error("Failed to parse WebSocket URL: %v", err)
		return
	}

	// 添加流参数
	q := u.Query()
	q.Set("streams", strings.Join(streams, "/"))
	u.RawQuery = q.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error("Failed to connect to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	log.Info("Connected to Binance WebSocket for depth streams")

	// 启动消息处理器
	go r.handleDepthMessages(ctx, conn)

	// 保持连接
	for {
		select {
		case <-ctx.Done():
			log.Info("Depth stream cancelled")
			return
		default:
			// 发送ping保持连接
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Error("Failed to send ping: %v", err)
				return
			}
			time.Sleep(30 * time.Second)
		}
	}
}

// handleDepthMessages 处理深度消息
func (r *Realtime) handleDepthMessages(ctx context.Context, conn *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Error("Failed to read WebSocket message: %v", err)
				return
			}

			// 解析消息
			var wsMessage struct {
				Stream string `json:"stream"`
				Data   struct {
					Symbol       string     `json:"s"`
					LastUpdateID int64      `json:"u"`
					Bids         [][]string `json:"b"`
					Asks         [][]string `json:"a"`
				} `json:"data"`
			}

			if err := json.Unmarshal(message, &wsMessage); err != nil {
				log.Error("Failed to unmarshal WebSocket message: %v", err)
				continue
			}

			// 处理深度数据
			depth := &api.DepthData{
				LastUpdateID: wsMessage.Data.LastUpdateID,
				Bids:         wsMessage.Data.Bids,
				Asks:         wsMessage.Data.Asks,
			}

			if err := r.writer.WriteDepth(wsMessage.Data.Symbol, depth); err != nil {
				log.Error("Failed to write realtime depth for %s: %v", wsMessage.Data.Symbol, err)
			}
		}
	}
}

// aggregateData 聚合数据
func (r *Realtime) aggregateData(symbol string, kline api.KlineData) {
	// 获取或创建聚合器
	r.mutex.Lock()
	key := fmt.Sprintf("%s_%s", symbol, r.config.UpdateFreq)
	aggregator, exists := r.aggregators[key]
	if !exists {
		aggregator = &DataAggregator{
			symbol:    symbol,
			timeframe: r.config.UpdateFreq,
			buffer:    make([]api.KlineData, 0),
			writer:    r.writer,
		}
		r.aggregators[key] = aggregator
	}
	r.mutex.Unlock()

	// 添加数据到缓冲区
	aggregator.mutex.Lock()
	aggregator.buffer = append(aggregator.buffer, kline)
	aggregator.mutex.Unlock()

	// 检查是否需要聚合
	if len(aggregator.buffer) >= r.config.BufferSize {
		aggregator.flush()
	}
}

// flush 刷新聚合数据
func (a *DataAggregator) flush() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if len(a.buffer) == 0 {
		return
	}

	// 聚合数据
	aggregated := a.aggregateKlines(a.buffer)
	
	// 写入聚合数据
	if err := a.writer.WriteAggregatedData(a.symbol, a.timeframe, aggregated); err != nil {
		log.Error("Failed to write aggregated data for %s/%s: %v", a.symbol, a.timeframe, err)
	}

	// 清空缓冲区
	a.buffer = a.buffer[:0]
}

// aggregateKlines 聚合K线数据
func (a *DataAggregator) aggregateKlines(klines []api.KlineData) []api.KlineData {
	if len(klines) == 0 {
		return nil
	}

	// 简单的聚合逻辑（可以根据需要扩展）
	aggregated := api.KlineData{
		OpenTime: klines[0].OpenTime,
		Open:     klines[0].Open,
		High:     klines[0].High,
		Low:      klines[0].Low,
		Close:    klines[len(klines)-1].Close,
		Volume:   klines[0].Volume,
		CloseTime: klines[len(klines)-1].CloseTime,
	}

	// 计算最高价和最低价
	for _, kline := range klines {
		if kline.High > aggregated.High {
			aggregated.High = kline.High
		}
		if kline.Low < aggregated.Low {
			aggregated.Low = kline.Low
		}
		aggregated.Volume += kline.Volume
		aggregated.QuoteAssetVolume += kline.QuoteAssetVolume
		aggregated.NumberOfTrades += kline.NumberOfTrades
		aggregated.TakerBuyBase += kline.TakerBuyBase
		aggregated.TakerBuyQuote += kline.TakerBuyQuote
	}

	return []api.KlineData{aggregated}
} 