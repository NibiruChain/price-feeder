package priceprovider

import (
	"sync"
	"time"

	"github.com/NibiruChain/nibiru/x/common"
	"github.com/NibiruChain/price-feeder/feeder/priceprovider/sources"
	"github.com/NibiruChain/price-feeder/types"
	"github.com/rs/zerolog"
)

var _ types.PriceProvider = (*PriceProvider)(nil)

// PriceProvider implements the types.PriceProvider interface.
// it wraps a Source and handles conversions between
// nibiru asset pair to exchange symbols.
type PriceProvider struct {
	logger              zerolog.Logger
	stopSignal          chan struct{}
	done                chan struct{}
	source              types.Source
	sourceName          string
	pairToSymbolMapping map[common.AssetPair]types.Symbol
	lastPricesMutex     sync.Mutex
	lastPrices          map[types.Symbol]types.RawPrice
}

// NewPriceProvider returns a types.PriceProvider given the price source we want to gather prices from,
// the mapping between nibiru common.AssetPair and the source's symbols, and a zerolog.Logger instance.
func NewPriceProvider(sourceName string, pairToSymbolMap map[common.AssetPair]types.Symbol, logger zerolog.Logger) types.PriceProvider {
	var source types.Source
	switch sourceName {
	case sources.Bitfinex:
		source = sources.NewTickSource(symbolsFromPairToSymbolMapping(pairToSymbolMap), sources.BitfinexPriceUpdate, logger)
	case sources.Binance:
		source = sources.NewTickSource(symbolsFromPairToSymbolMapping(pairToSymbolMap), sources.BinancePriceUpdate, logger)
	default:
		panic("unknown price provider: " + sourceName)
	}
	return newPriceProvider(source, sourceName, pairToSymbolMap, logger)
}

// newPriceProvider returns a raw *PriceProvider given a Source implementer, the source name and the
// map of nibiru common.AssetPair to Source's symbols, plus the zerolog.Logger instance.
// Exists for testing purposes.
func newPriceProvider(source types.Source, sourceName string, pairToSymbolsMap map[common.AssetPair]types.Symbol, logger zerolog.Logger) *PriceProvider {
	pp := &PriceProvider{
		logger:              logger.With().Str("component", "price-provider").Str("source", sourceName).Logger(),
		stopSignal:          make(chan struct{}),
		done:                make(chan struct{}),
		source:              source,
		sourceName:          sourceName,
		pairToSymbolMapping: pairToSymbolsMap,
		lastPricesMutex:     sync.Mutex{},
		lastPrices:          map[types.Symbol]types.RawPrice{},
	}

	go pp.loop()

	return pp
}

func (p *PriceProvider) loop() {
	defer close(p.done)
	defer p.source.Close()

	for {
		select {
		case <-p.stopSignal:
			return
		case updates := <-p.source.PriceUpdates():
			p.lastPricesMutex.Lock()
			for symbol, price := range updates {
				p.lastPrices[symbol] = price
			}
			p.lastPricesMutex.Unlock()
		}
	}
}

// GetPrice returns the types.Price for the given common.AssetPair
// in case price has expired, or for some reason it's impossible to
// get the last available price, then an invalid types.Price is returned.
func (p *PriceProvider) GetPrice(pair common.AssetPair) types.Price {
	symbol, symbolExists := p.pairToSymbolMapping[pair]
	// in case this is an unknown symbol, which might happen
	// when for example we have a param update, then we return
	// an abstain vote on the provided asset pair.
	if !symbolExists {
		p.logger.Warn().Str("pair", pair.String()).Msg("unknown nibiru pair")
		return types.Price{
			Pair:       pair,
			Price:      -1, // abstain
			SourceName: p.sourceName,
			Valid:      false,
		}
	}

	p.lastPricesMutex.Lock()
	price, priceExists := p.lastPrices[symbol]
	p.lastPricesMutex.Unlock()

	return types.Price{
		Pair:       pair,
		Price:      price.Price,
		SourceName: p.sourceName,
		Valid:      isValid(price, priceExists),
	}
}

func (p *PriceProvider) Close() {
	close(p.stopSignal)
	<-p.done
}

// symbolsFromPairToSymbolMapping returns the symbols list
// given the map which maps nibiru chain pairs to exchange symbols.
func symbolsFromPairToSymbolMapping(m map[common.AssetPair]types.Symbol) types.Symbols {
	s := make(types.Symbols, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}
	return s
}

// isValid is a helper function which asserts if a price is valid given
// if it was found and the time at which it was last updated.
func isValid(price types.RawPrice, found bool) bool {
	return found && time.Since(price.UpdateTime) < types.PriceTimeout
}
