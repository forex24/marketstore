package configs

import (
	"fmt"
	"strings"
	"time"
)

// toStringMap 将 map[interface{}]interface{} 或 map[string]interface{} 统一转换为 map[string]interface{}
func toStringMap(in interface{}) map[string]interface{} {
	if in == nil {
		return nil
	}
	res := make(map[string]interface{})
	switch m := in.(type) {
	case map[string]interface{}:
		return m
	case map[interface{}]interface{}:
		for k, v := range m {
			res[fmt.Sprint(k)] = v
		}
		return res
	default:
		return nil
	}
}

// Config 币安插件配置
type Config struct {
	// API配置
	APIKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
	Testnet   bool   `json:"testnet"`

	// 符号配置
	Symbols        []string `json:"symbols"`         // 要获取的符号列表
	ExcludeSymbols []string `json:"exclude_symbols"` // 要排除的符号列表

	// 时间框架配置
	Timeframe string `json:"timeframe"` // 基础时间框架 (1m, 3m, 5m, 15m, 30m, 1h, 4h, 1d, 1w)

	// Backfill配置
	Backfill BackfillConfig `json:"backfill"`

	// 实时数据配置
	Realtime RealtimeConfig `json:"realtime"`
}

// BackfillConfig Backfill配置
type BackfillConfig struct {
	Enabled     bool      `json:"enabled"`      // 是否启用backfill
	StartTime   time.Time `json:"start_time"`   // 开始时间
	EndTime     time.Time `json:"end_time"`     // 结束时间
	BatchSize   int       `json:"batch_size"`   // 批处理大小
	Parallelism int       `json:"parallelism"`  // 并行度
	Interval    string    `json:"interval"`     // 数据间隔 (1m, 3m, 5m, 15m, 30m, 1h, 4h, 1d, 1w)
}

// RealtimeConfig 实时数据配置
type RealtimeConfig struct {
	Enabled     bool   `json:"enabled"`      // 是否启用实时数据
	StreamType  string `json:"stream_type"`  // 流类型 (kline, trade, depth)
	UpdateFreq  string `json:"update_freq"`  // 更新频率
	BufferSize  int    `json:"buffer_size"`  // 缓冲区大小
	MaxRetries  int    `json:"max_retries"`  // 最大重试次数
	RetryDelay  string `json:"retry_delay"`  // 重试延迟
}

// ParseConfig 解析配置
func ParseConfig(conf map[string]interface{}) (*Config, error) {
	config := &Config{}

	// 解析API配置
	if apiKey, ok := conf["api_key"].(string); ok {
		config.APIKey = apiKey
	}
	if secretKey, ok := conf["secret_key"].(string); ok {
		config.SecretKey = secretKey
	}
	if testnet, ok := conf["testnet"].(bool); ok {
		config.Testnet = testnet
	}

	// 解析符号配置
	if symbols, ok := conf["symbols"].([]interface{}); ok {
		config.Symbols = make([]string, len(symbols))
		for i, symbol := range symbols {
			if s, ok := symbol.(string); ok {
				config.Symbols[i] = s
			}
		}
	}
	if excludeSymbols, ok := conf["exclude_symbols"].([]interface{}); ok {
		config.ExcludeSymbols = make([]string, len(excludeSymbols))
		for i, symbol := range excludeSymbols {
			if s, ok := symbol.(string); ok {
				config.ExcludeSymbols[i] = s
			}
		}
	}

	// 解析时间框架
	if timeframe, ok := conf["timeframe"].(string); ok {
		config.Timeframe = timeframe
	} else {
		config.Timeframe = "1m" // 默认1分钟
	}

	// 解析Backfill配置
	if raw, exists := conf["backfill"]; exists {
		backfillConf := toStringMap(raw)
		if backfillConf != nil {
			// 解析enabled字段，兼容bool或字符串
			if enabledVal, exists := backfillConf["enabled"]; exists {
				fmt.Printf("DEBUG: enabled value type: %T, value: %v\n", enabledVal, enabledVal)
				switch v := enabledVal.(type) {
				case bool:
					config.Backfill.Enabled = v
					fmt.Printf("DEBUG: Set enabled to bool: %v\n", v)
				case string:
					lower := strings.ToLower(v)
					config.Backfill.Enabled = lower == "true" || lower == "yes" || lower == "1"
					fmt.Printf("DEBUG: Set enabled to string: %v -> %v\n", v, config.Backfill.Enabled)
				default:
					fmt.Printf("DEBUG: Unknown enabled type: %T, value: %v\n", v, v)
					config.Backfill.Enabled = false
				}
			} else {
				fmt.Printf("DEBUG: enabled field not found in backfill config\n")
			}
			if startTime, ok := backfillConf["start_time"].(string); ok {
				if t, err := time.Parse(time.RFC3339, startTime); err == nil {
					config.Backfill.StartTime = t
				}
			}
			if endTime, ok := backfillConf["end_time"].(string); ok {
				if t, err := time.Parse(time.RFC3339, endTime); err == nil {
					config.Backfill.EndTime = t
				}
			}
			if batchSize, ok := backfillConf["batch_size"].(float64); ok {
				config.Backfill.BatchSize = int(batchSize)
			} else {
				config.Backfill.BatchSize = 1000 // 默认批处理大小
			}
			if parallelism, ok := backfillConf["parallelism"].(float64); ok {
				config.Backfill.Parallelism = int(parallelism)
			} else {
				config.Backfill.Parallelism = 5 // 默认并行度
			}
			if interval, ok := backfillConf["interval"].(string); ok {
				config.Backfill.Interval = interval
			} else {
				config.Backfill.Interval = "1m" // 默认1分钟
			}
		}
	}

	// 解析实时数据配置
	if raw, ok := conf["realtime"]; ok {
		realtimeConf := toStringMap(raw)
		if realtimeConf != nil {
			if enabled, ok := realtimeConf["enabled"].(bool); ok {
				config.Realtime.Enabled = enabled
			}
			if streamType, ok := realtimeConf["stream_type"].(string); ok {
				config.Realtime.StreamType = streamType
			} else {
				config.Realtime.StreamType = "kline" // 默认K线数据
			}
			if updateFreq, ok := realtimeConf["update_freq"].(string); ok {
				config.Realtime.UpdateFreq = updateFreq
			} else {
				config.Realtime.UpdateFreq = "1s" // 默认1秒更新
			}
			if bufferSize, ok := realtimeConf["buffer_size"].(float64); ok {
				config.Realtime.BufferSize = int(bufferSize)
			} else {
				config.Realtime.BufferSize = 1000 // 默认缓冲区大小
			}
			if maxRetries, ok := realtimeConf["max_retries"].(float64); ok {
				config.Realtime.MaxRetries = int(maxRetries)
			} else {
				config.Realtime.MaxRetries = 3 // 默认最大重试次数
			}
			if retryDelay, ok := realtimeConf["retry_delay"].(string); ok {
				config.Realtime.RetryDelay = retryDelay
			} else {
				config.Realtime.RetryDelay = "5s" // 默认重试延迟
			}
		}
	}

	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	// 验证时间框架
	validTimeframes := map[string]bool{
		"1m": true, "3m": true, "5m": true, "15m": true,
		"30m": true, "1h": true, "4h": true, "1d": true, "1w": true,
	}
	if !validTimeframes[config.Timeframe] {
		return fmt.Errorf("invalid timeframe: %s", config.Timeframe)
	}

	// 验证符号列表
	if len(config.Symbols) == 0 {
		return fmt.Errorf("at least one symbol must be specified")
	}

	// 验证Backfill配置
	if config.Backfill.Enabled {
		if config.Backfill.StartTime.IsZero() {
			return fmt.Errorf("backfill start_time is required when enabled")
		}
		if config.Backfill.EndTime.IsZero() {
			config.Backfill.EndTime = time.Now() // 默认到当前时间
		}
		if config.Backfill.StartTime.After(config.Backfill.EndTime) {
			return fmt.Errorf("backfill start_time must be before end_time")
		}
		if !validTimeframes[config.Backfill.Interval] {
			return fmt.Errorf("invalid backfill interval: %s", config.Backfill.Interval)
		}
	}

	// 验证实时数据配置
	if config.Realtime.Enabled {
		validStreamTypes := map[string]bool{
			"kline": true, "trade": true, "depth": true,
		}
		if !validStreamTypes[config.Realtime.StreamType] {
			return fmt.Errorf("invalid stream_type: %s", config.Realtime.StreamType)
		}
	}

	return nil
} 