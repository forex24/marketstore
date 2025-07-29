package symbols

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/alpacahq/marketstore/v4/contrib/binancefeeder_v2/api"
	"github.com/alpacahq/marketstore/v4/utils/log"
)

// Manager 符号管理器
type Manager struct {
	symbols        []string
	excludeSymbols []string
	allSymbols     []string
	mutex          sync.RWMutex
	apiClient      *api.Client
}

// NewManager 创建新的符号管理器
func NewManager(symbols, excludeSymbols []string) *Manager {
	return &Manager{
		symbols:        symbols,
		excludeSymbols: excludeSymbols,
	}
}

// SetAPIClient 设置API客户端
func (m *Manager) SetAPIClient(client *api.Client) {
	m.apiClient = client
}

// GetAllSymbols 获取所有符号
func (m *Manager) GetAllSymbols() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 如果用户在配置中显式指定了符号，则优先返回这些符号（排除 excludeSymbols）
	if len(m.symbols) > 0 {
		return m.filterSymbols(m.symbols)
	}

	// 否则返回从交易所加载的全部可交易符号（同样需要排除）
	return m.filterSymbols(m.allSymbols)
}

// GetSymbols 获取符号列表
func (m *Manager) GetSymbols() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.filterSymbols(m.symbols)
}

// GetExcludeSymbols 获取排除的符号列表
func (m *Manager) GetExcludeSymbols() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.excludeSymbols
}

// LoadAllSymbols 从API加载所有可用符号
func (m *Manager) LoadAllSymbols(ctx context.Context) error {
	if m.apiClient == nil {
		return fmt.Errorf("API client not set")
	}

	info, err := m.apiClient.GetExchangeInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get exchange info: %w", err)
	}

	var symbols []string
	for _, symbol := range info.Symbols {
		if symbol.Status == "TRADING" && symbol.IsSpotTradingAllowed {
			symbols = append(symbols, symbol.Symbol)
		}
	}

	m.mutex.Lock()
	m.allSymbols = symbols
	m.mutex.Unlock()

	log.Info("Loaded %d trading symbols from Binance", len(symbols))
	return nil
}

// filterSymbols 过滤符号列表
func (m *Manager) filterSymbols(symbols []string) []string {
	if len(m.excludeSymbols) == 0 {
		return symbols
	}

	var filtered []string
	excludeMap := make(map[string]bool)
	for _, symbol := range m.excludeSymbols {
		excludeMap[strings.ToUpper(symbol)] = true
	}

	for _, symbol := range symbols {
		if !excludeMap[strings.ToUpper(symbol)] {
			filtered = append(filtered, symbol)
		}
	}

	return filtered
}

// IsValidSymbol 检查符号是否有效
func (m *Manager) IsValidSymbol(symbol string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 检查是否在排除列表中
	for _, exclude := range m.excludeSymbols {
		if strings.EqualFold(symbol, exclude) {
			return false
		}
	}

	// 如果指定了具体符号，检查是否在列表中
	if len(m.symbols) > 0 {
		for _, s := range m.symbols {
			if strings.EqualFold(symbol, s) {
				return true
			}
		}
		return false
	}

	// 否则检查是否在所有符号中
	for _, s := range m.allSymbols {
		if strings.EqualFold(symbol, s) {
			return true
		}
	}

	return false
}

// GetSymbolCount 获取符号数量
func (m *Manager) GetSymbolCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.GetAllSymbols())
}

// AddSymbol 添加符号
func (m *Manager) AddSymbol(symbol string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查是否已存在
	for _, s := range m.symbols {
		if strings.EqualFold(s, symbol) {
			return
		}
	}

	m.symbols = append(m.symbols, strings.ToUpper(symbol))
}

// RemoveSymbol 移除符号
func (m *Manager) RemoveSymbol(symbol string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i, s := range m.symbols {
		if strings.EqualFold(s, symbol) {
			m.symbols = append(m.symbols[:i], m.symbols[i+1:]...)
			break
		}
	}
}

// AddExcludeSymbol 添加排除符号
func (m *Manager) AddExcludeSymbol(symbol string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查是否已存在
	for _, s := range m.excludeSymbols {
		if strings.EqualFold(s, symbol) {
			return
		}
	}

	m.excludeSymbols = append(m.excludeSymbols, strings.ToUpper(symbol))
}

// RemoveExcludeSymbol 移除排除符号
func (m *Manager) RemoveExcludeSymbol(symbol string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for i, s := range m.excludeSymbols {
		if strings.EqualFold(s, symbol) {
			m.excludeSymbols = append(m.excludeSymbols[:i], m.excludeSymbols[i+1:]...)
			break
		}
	}
}

// GetSymbolsByPattern 根据模式获取符号
func (m *Manager) GetSymbolsByPattern(pattern string) []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var matched []string
	allSymbols := m.GetAllSymbols()

	for _, symbol := range allSymbols {
		if strings.Contains(strings.ToUpper(symbol), strings.ToUpper(pattern)) {
			matched = append(matched, symbol)
		}
	}

	return matched
}

// GetSymbolsByQuoteAsset 根据计价资产获取符号
func (m *Manager) GetSymbolsByQuoteAsset(quoteAsset string) []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var matched []string
	allSymbols := m.GetAllSymbols()

	for _, symbol := range allSymbols {
		if strings.HasSuffix(strings.ToUpper(symbol), strings.ToUpper(quoteAsset)) {
			matched = append(matched, symbol)
		}
	}

	return matched
} 