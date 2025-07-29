package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alpacahq/marketstore/v4/utils/log"
)

const (
	// 币安API基础URL
	baseURL     = "https://api.binance.com"
	testnetURL  = "https://testnet.binance.vision"
	wsBaseURL   = "wss://stream.binance.com:9443/ws"
	wsTestURL   = "wss://testnet.binance.vision/ws"
	
	// 币安API限制（根据官方文档）
	// 现货API限制：1200 requests per minute per IP
	// 为了安全起见，我们使用更保守的限制
	requestsPerMinute = 1000 // 留出200个请求的余量
	requestInterval   = time.Minute / requestsPerMinute
	
	// 权重限制（根据币安官方文档）
	weightKlines        = 1   // K线数据权重
	weightTrades        = 5   // 交易数据权重
	weightDepth         = 5   // 深度数据权重
	weightExchangeInfo  = 10  // 交易所信息权重
	weightServerTime    = 1   // 服务器时间权重
)

// RateLimiter 速率限制器
type RateLimiter struct {
	mu       sync.Mutex
	lastCall time.Time
	interval time.Duration
}

// NewRateLimiter 创建新的速率限制器
func NewRateLimiter(interval time.Duration) *RateLimiter {
	return &RateLimiter{
		interval: interval,
	}
}

// Wait 等待直到可以发送下一个请求
func (r *RateLimiter) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	now := time.Now()
	if now.Sub(r.lastCall) < r.interval {
		sleepTime := r.interval - now.Sub(r.lastCall)
		time.Sleep(sleepTime)
	}
	r.lastCall = time.Now()
}

// Client 币安API客户端
type Client struct {
	apiKey      string
	secretKey   string
	testnet     bool
	client      *http.Client
	baseURL     string
	wsURL       string
	rateLimiter *RateLimiter
}

// NewClient 创建新的API客户端
func NewClient(apiKey, secretKey string, testnet bool) *Client {
	base := baseURL
	ws := wsBaseURL
	if testnet {
		base = testnetURL
		ws = wsTestURL
	}

	return &Client{
		apiKey:      apiKey,
		secretKey:   secretKey,
		testnet:     testnet,
		client:      &http.Client{Timeout: 30 * time.Second},
		baseURL:     base,
		wsURL:       ws,
		rateLimiter: NewRateLimiter(requestInterval),
	}
}

// KlineData K线数据结构
type KlineData struct {
	OpenTime         int64   `json:"openTime"`
	Open             float64 `json:"open"`
	High             float64 `json:"high"`
	Low              float64 `json:"low"`
	Close            float64 `json:"close"`
	Volume           float64 `json:"volume"`
	CloseTime        int64   `json:"closeTime"`
	QuoteAssetVolume float64 `json:"quoteAssetVolume"`
	NumberOfTrades   int64   `json:"numberOfTrades"`
	TakerBuyBase     float64 `json:"takerBuyBaseAssetVolume"`
	TakerBuyQuote    float64 `json:"takerBuyQuoteAssetVolume"`
}

// TradeData 交易数据结构
type TradeData struct {
	ID           int64   `json:"id"`
	Price        float64 `json:"price"`
	Quantity     float64 `json:"qty"`
	QuoteQty     float64 `json:"quoteQty"`
	Time         int64   `json:"time"`
	IsBuyerMaker bool    `json:"isBuyerMaker"`
}

// DepthData 深度数据结构
type DepthData struct {
	LastUpdateID int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

// ExchangeInfo 交易所信息
type ExchangeInfo struct {
	Symbols []SymbolInfo `json:"symbols"`
}

// SymbolInfo 符号信息
type SymbolInfo struct {
	Symbol             string   `json:"symbol"`
	Status             string   `json:"status"`
	BaseAsset          string   `json:"baseAsset"`
	QuoteAsset         string   `json:"quoteAsset"`
	IsSpotTradingAllowed bool   `json:"isSpotTradingAllowed"`
	IsMarginTradingAllowed bool `json:"isMarginTradingAllowed"`
}

// GetKlines 获取K线数据
func (c *Client) GetKlines(ctx context.Context, symbol, interval string, startTime, endTime time.Time, limit int) ([]KlineData, error) {
	c.rateLimiter.Wait() // 应用速率限制
	url := fmt.Sprintf("%s/api/v3/klines", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("symbol", strings.ToUpper(symbol))
	q.Add("interval", interval)
	
	if !startTime.IsZero() {
		q.Add("startTime", strconv.FormatInt(startTime.UnixMilli(), 10))
	}
	if !endTime.IsZero() {
		q.Add("endTime", strconv.FormatInt(endTime.UnixMilli(), 10))
	}
	if limit > 0 {
		q.Add("limit", strconv.Itoa(limit))
	}
	
	req.URL.RawQuery = q.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed: %s, body: %s", resp.Status, string(body))
	}

	var rawData [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, err
	}

	klines := make([]KlineData, len(rawData))
	for i, data := range rawData {
		if len(data) < 11 {
			continue
		}

		// 处理时间戳（可能是float64或string）
		var openTime int64
		switch v := data[0].(type) {
		case float64:
			openTime = int64(v)
		case string:
			openTime, _ = strconv.ParseInt(v, 10, 64)
		}

		// 处理价格和数量（可能是float64或string）
		var open, high, low, close, volume, quoteVolume, takerBuyBase, takerBuyQuote float64
		var closeTime, trades int64

		// 转换价格数据
		if v, ok := data[1].(float64); ok {
			open = v
		} else if v, ok := data[1].(string); ok {
			open, _ = strconv.ParseFloat(v, 64)
		}
		if v, ok := data[2].(float64); ok {
			high = v
		} else if v, ok := data[2].(string); ok {
			high, _ = strconv.ParseFloat(v, 64)
		}
		if v, ok := data[3].(float64); ok {
			low = v
		} else if v, ok := data[3].(string); ok {
			low, _ = strconv.ParseFloat(v, 64)
		}
		if v, ok := data[4].(float64); ok {
			close = v
		} else if v, ok := data[4].(string); ok {
			close, _ = strconv.ParseFloat(v, 64)
		}
		if v, ok := data[5].(float64); ok {
			volume = v
		} else if v, ok := data[5].(string); ok {
			volume, _ = strconv.ParseFloat(v, 64)
		}

		// 转换时间戳
		if v, ok := data[6].(float64); ok {
			closeTime = int64(v)
		} else if v, ok := data[6].(string); ok {
			closeTime, _ = strconv.ParseInt(v, 10, 64)
		}

		// 转换其他数据
		if v, ok := data[7].(float64); ok {
			quoteVolume = v
		} else if v, ok := data[7].(string); ok {
			quoteVolume, _ = strconv.ParseFloat(v, 64)
		}
		if v, ok := data[8].(float64); ok {
			trades = int64(v)
		} else if v, ok := data[8].(string); ok {
			trades, _ = strconv.ParseInt(v, 10, 64)
		}
		if v, ok := data[9].(float64); ok {
			takerBuyBase = v
		} else if v, ok := data[9].(string); ok {
			takerBuyBase, _ = strconv.ParseFloat(v, 64)
		}
		if v, ok := data[10].(float64); ok {
			takerBuyQuote = v
		} else if v, ok := data[10].(string); ok {
			takerBuyQuote, _ = strconv.ParseFloat(v, 64)
		}

		klines[i] = KlineData{
			OpenTime:         openTime,
			Open:             open,
			High:             high,
			Low:              low,
			Close:            close,
			Volume:           volume,
			CloseTime:        closeTime,
			QuoteAssetVolume: quoteVolume,
			NumberOfTrades:   trades,
			TakerBuyBase:     takerBuyBase,
			TakerBuyQuote:    takerBuyQuote,
		}
	}

	return klines, nil
}

// GetRecentTrades 获取最近交易
func (c *Client) GetRecentTrades(ctx context.Context, symbol string, limit int) ([]TradeData, error) {
	c.rateLimiter.Wait() // 应用速率限制
	url := fmt.Sprintf("%s/api/v3/trades", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("symbol", strings.ToUpper(symbol))
	if limit > 0 {
		q.Add("limit", strconv.Itoa(limit))
	}
	
	req.URL.RawQuery = q.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed: %s, body: %s", resp.Status, string(body))
	}

	var trades []TradeData
	if err := json.NewDecoder(resp.Body).Decode(&trades); err != nil {
		return nil, err
	}

	return trades, nil
}

// GetDepth 获取深度数据
func (c *Client) GetDepth(ctx context.Context, symbol string, limit int) (*DepthData, error) {
	c.rateLimiter.Wait() // 应用速率限制
	url := fmt.Sprintf("%s/api/v3/depth", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("symbol", strings.ToUpper(symbol))
	if limit > 0 {
		q.Add("limit", strconv.Itoa(limit))
	}
	
	req.URL.RawQuery = q.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed: %s, body: %s", resp.Status, string(body))
	}

	var depth DepthData
	if err := json.NewDecoder(resp.Body).Decode(&depth); err != nil {
		return nil, err
	}

	return &depth, nil
}

// GetExchangeInfo 获取交易所信息
func (c *Client) GetExchangeInfo(ctx context.Context) (*ExchangeInfo, error) {
	c.rateLimiter.Wait() // 应用速率限制
	url := fmt.Sprintf("%s/api/v3/exchangeInfo", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed: %s, body: %s", resp.Status, string(body))
	}

	var info ExchangeInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}

// GetServerTime 获取服务器时间
func (c *Client) GetServerTime(ctx context.Context) (time.Time, error) {
	c.rateLimiter.Wait() // 应用速率限制
	url := fmt.Sprintf("%s/api/v3/time", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return time.Time{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return time.Time{}, fmt.Errorf("API request failed: %s, body: %s", resp.Status, string(body))
	}

	var result struct {
		ServerTime int64 `json:"serverTime"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return time.Time{}, err
	}

	return time.Unix(0, result.ServerTime*int64(time.Millisecond)), nil
}

// GetWebSocketURL 获取WebSocket URL
func (c *Client) GetWebSocketURL() string {
	return c.wsURL
}

// LogInfo 记录信息日志
func (c *Client) LogInfo(format string, args ...interface{}) {
	log.Info(format, args...)
}

// LogError 记录错误日志
func (c *Client) LogError(format string, args ...interface{}) {
	log.Error(format, args...)
} 