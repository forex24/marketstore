package writer

import (
	"fmt"
	"time"

	"github.com/alpacahq/marketstore/v4/contrib/binancefeeder_v2/api"
	"github.com/alpacahq/marketstore/v4/executor"
	"github.com/alpacahq/marketstore/v4/utils/io"
	"github.com/alpacahq/marketstore/v4/utils/log"
)

// Writer 数据写入器
type Writer struct {
	timeframe string
	timezone  string
}

// NewWriter 创建新的写入器
func NewWriter(timeframe, timezone string) *Writer {
	return &Writer{
		timeframe: timeframe,
		timezone:  timezone,
	}
}

// WriteKlines 写入K线数据
func (w *Writer) WriteKlines(symbol string, klines []api.KlineData) error {
	if len(klines) == 0 {
		return nil
	}

	// 准备数据
	openTime := make([]int64, len(klines))
	open := make([]float64, len(klines))
	high := make([]float64, len(klines))
	low := make([]float64, len(klines))
	close := make([]float64, len(klines))
	volume := make([]float64, len(klines))

	for i, kline := range klines {
		// 转换时间戳
		openTime[i] = kline.OpenTime / 1000 // 转换为秒
		open[i] = kline.Open
		high[i] = kline.High
		low[i] = kline.Low
		close[i] = kline.Close
		volume[i] = kline.Volume
	}

	// 将Binance间隔转换为MarketStore的Timeframe格式
	tf := w.ConvertTimeframe(w.timeframe)
	if !w.ValidateTimeframe(tf) {
		return fmt.Errorf("invalid timeframe %s after conversion", tf)
	}

	// 创建TimeBucketKey
	tbk := io.NewTimeBucketKey(fmt.Sprintf("%s/%s/OHLCV", symbol, tf))

	// 创建ColumnSeries
	cs := io.NewColumnSeries()
	cs.AddColumn("Epoch", openTime)
	cs.AddColumn("Open", open)
	cs.AddColumn("High", high)
	cs.AddColumn("Low", low)
	cs.AddColumn("Close", close)
	cs.AddColumn("Volume", volume)

	// 创建ColumnSeriesMap
	csm := io.NewColumnSeriesMap()
	csm.AddColumnSeries(*tbk, cs)

	// 写入数据
	if err := executor.WriteCSM(csm, false); err != nil {
		return fmt.Errorf("failed to write klines for %s: %w", symbol, err)
	}

	log.Info("Wrote %d klines for %s", len(klines), symbol)
	return nil
}

// WriteTrades 写入交易数据
func (w *Writer) WriteTrades(symbol string, trades []api.TradeData) error {
	if len(trades) == 0 {
		return nil
	}

	// 准备数据
	epoch := make([]int64, len(trades))
	nanosecond := make([]int32, len(trades))
	price := make([]float64, len(trades))
	size := make([]float64, len(trades))
	exchange := make([]string, len(trades))
	tape := make([]string, len(trades))

	for i, trade := range trades {
		// 转换时间戳
		timestamp := time.Unix(trade.Time/1000, (trade.Time%1000)*1000000)
		epoch[i] = timestamp.Unix()
		nanosecond[i] = int32(timestamp.Nanosecond())
		price[i] = trade.Price
		size[i] = trade.Quantity
		exchange[i] = "B" // BINANCE
		tape[i] = "A"     // SPOT
	}

	// 创建TimeBucketKey
	tbk := io.NewTimeBucketKey(fmt.Sprintf("%s/1Min/Trade", symbol))

	// 创建ColumnSeries
	cs := io.NewColumnSeries()
	cs.AddColumn("Epoch", epoch)
	cs.AddColumn("Nanosecond", nanosecond)
	cs.AddColumn("Price", price)
	cs.AddColumn("Size", size)
	cs.AddColumn("Exchange", exchange)
	cs.AddColumn("Tape", tape)

	// 创建ColumnSeriesMap
	csm := io.NewColumnSeriesMap()
	csm.AddColumnSeries(*tbk, cs)

	// 写入数据
	if err := executor.WriteCSM(csm, false); err != nil {
		return fmt.Errorf("failed to write trades for %s: %w", symbol, err)
	}

	log.Info("Wrote %d trades for %s", len(trades), symbol)
	return nil
}

// WriteDepth 写入深度数据
func (w *Writer) WriteDepth(symbol string, depth *api.DepthData) error {
	if depth == nil {
		return nil
	}

	// 创建深度数据模型（自定义格式）
	// 这里需要根据MarketStore的数据格式进行调整
	// 暂时记录日志
	log.Info("Received depth data for %s: %d bids, %d asks", 
		symbol, len(depth.Bids), len(depth.Asks))

	return nil
}

// WriteAggregatedData 写入聚合数据
func (w *Writer) WriteAggregatedData(symbol, timeframe string, data []api.KlineData) error {
	if len(data) == 0 {
		return nil
	}

	// 准备数据
	openTime := make([]int64, len(data))
	open := make([]float64, len(data))
	high := make([]float64, len(data))
	low := make([]float64, len(data))
	close := make([]float64, len(data))
	volume := make([]float64, len(data))

	for i, kline := range data {
		// 转换时间戳
		openTime[i] = kline.OpenTime / 1000 // 转换为秒
		open[i] = kline.Open
		high[i] = kline.High
		low[i] = kline.Low
		close[i] = kline.Close
		volume[i] = kline.Volume
	}

	// 将Binance间隔转换为MarketStore的Timeframe格式
	tf := w.ConvertTimeframe(timeframe)
	if !w.ValidateTimeframe(tf) {
		return fmt.Errorf("invalid timeframe %s after conversion", tf)
	}

	// 创建TimeBucketKey
	tbk := io.NewTimeBucketKey(fmt.Sprintf("%s/%s/OHLCV", symbol, tf))

	// 创建ColumnSeries
	cs := io.NewColumnSeries()
	cs.AddColumn("Epoch", openTime)
	cs.AddColumn("Open", open)
	cs.AddColumn("High", high)
	cs.AddColumn("Low", low)
	cs.AddColumn("Close", close)
	cs.AddColumn("Volume", volume)

	// 创建ColumnSeriesMap
	csm := io.NewColumnSeriesMap()
	csm.AddColumnSeries(*tbk, cs)

	// 写入数据
	if err := executor.WriteCSM(csm, false); err != nil {
		return fmt.Errorf("failed to write aggregated data for %s/%s: %w", symbol, timeframe, err)
	}

	log.Info("Wrote %d aggregated bars for %s/%s", len(data), symbol, timeframe)
	return nil
}

// WriteCustomData 写入自定义数据
func (w *Writer) WriteCustomData(symbol, dataType string, data interface{}) error {
	// 根据数据类型处理
	switch dataType {
	case "kline":
		if klines, ok := data.([]api.KlineData); ok {
			return w.WriteKlines(symbol, klines)
		}
	case "trade":
		if trades, ok := data.([]api.TradeData); ok {
			return w.WriteTrades(symbol, trades)
		}
	case "depth":
		if depth, ok := data.(*api.DepthData); ok {
			return w.WriteDepth(symbol, depth)
		}
	default:
		return fmt.Errorf("unsupported data type: %s", dataType)
	}

	return fmt.Errorf("invalid data type for %s", dataType)
}

// GetTimeframe 获取时间框架
func (w *Writer) GetTimeframe() string {
	return w.timeframe
}

// SetTimeframe 设置时间框架
func (w *Writer) SetTimeframe(timeframe string) {
	w.timeframe = timeframe
}

// GetTimezone 获取时区
func (w *Writer) GetTimezone() string {
	return w.timezone
}

// SetTimezone 设置时区
func (w *Writer) SetTimezone(timezone string) {
	w.timezone = timezone
}

// ConvertTimeframe 转换时间框架格式
func (w *Writer) ConvertTimeframe(binanceInterval string) string {
	// 币安时间间隔到MarketStore时间框架的映射
	timeframeMap := map[string]string{
		"1m":  "1Min",
		"3m":  "3Min",
		"5m":  "5Min",
		"15m": "15Min",
		"30m": "30Min",
		"1h":  "1H",
		"2h":  "2H",
		"4h":  "4H",
		"6h":  "6H",
		"8h":  "8H",
		"12h": "12H",
		"1d":  "1D",
		"3d":  "3D",
		"1w":  "1W",
		"1M":  "1M",
	}

	if timeframe, exists := timeframeMap[binanceInterval]; exists {
		return timeframe
	}

	// 如果找不到映射，返回原值
	return binanceInterval
}

// ValidateTimeframe 验证时间框架
func (w *Writer) ValidateTimeframe(timeframe string) bool {
	validTimeframes := map[string]bool{
		"1Min": true, "3Min": true, "5Min": true, "15Min": true,
		"30Min": true, "1H": true, "2H": true, "4H": true,
		"6H": true, "8H": true, "12H": true, "1D": true,
		"3D": true, "1W": true, "1M": true,
	}

	return validTimeframes[timeframe]
}

// CreateTimeBucketKey 创建时间桶键
func (w *Writer) CreateTimeBucketKey(symbol, timeframe, attributeGroup string) *io.TimeBucketKey {
	key := fmt.Sprintf("%s/%s/%s", symbol, timeframe, attributeGroup)
	return io.NewTimeBucketKey(key, "Symbol/Timeframe/AttributeGroup")
}

// GetDataShapes 获取数据形状
func (w *Writer) GetDataShapes(dataType string) []io.DataShape {
	switch dataType {
	case "OHLCV":
		return []io.DataShape{
			{Name: "Open", Type: io.FLOAT32},
			{Name: "High", Type: io.FLOAT32},
			{Name: "Low", Type: io.FLOAT32},
			{Name: "Close", Type: io.FLOAT32},
			{Name: "Volume", Type: io.FLOAT32},
		}
	case "Trade":
		return []io.DataShape{
			{Name: "Price", Type: io.FLOAT32},
			{Name: "Size", Type: io.FLOAT32},
			{Name: "Exchange", Type: io.STRING},
			{Name: "Tape", Type: io.STRING},
		}
	default:
		return nil
	}
} 