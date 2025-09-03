package exrate

// Currency represents a currency with its stablecoin flag
type Currency struct {
	Code     string
	IsStable bool
}

// CurrencyList is a list of currencies
type CurrencyList []Currency

// Currency pair for exchange rate
type CurrencyPair struct {
	From, To string
}

func (cp *CurrencyPair) Flip() {
	cp.From, cp.To = cp.To, cp.From
}

// Iter iterate over currency pairs where at least one currency is a stablecoin
func (cl CurrencyList) Iter() func() (CurrencyPair, bool) {
	var i, j int
	list := []Currency(cl)
	return func() (v CurrencyPair, ok bool) {
		for {
			if i >= len(list) || j >= len(list) {
				return v, false
			}

			j++
			if j >= len(list) {
				i++
				j = i + 1
			}

			if i >= len(list) || j >= len(list) {
				return v, false
			}

			// Only generate pairs where at least one currency is a stablecoin
			if list[i].IsStable || list[j].IsStable {
				v = CurrencyPair{From: list[i].Code, To: list[j].Code}
				ok = true
				return
			}
		}
	}
}

func NewCurrencyFilter(list CurrencyList) CurrencyFilter {
	filter := CurrencyFilter{
		symbols: make(map[string]CurrencyPair),
	}

	next := list.Iter()
	for pair, ok := next(); ok; pair, ok = next() {
		filter.symbols[pair.From+pair.To] = pair
		pair.Flip()
		filter.symbols[pair.From+pair.To] = pair
	}
	return filter
}

type CurrencyFilter struct {
	symbols map[string]CurrencyPair
}

func (cf CurrencyFilter) HasPair(p CurrencyPair) bool {
	_, ok := cf.FindSymbol(p.From + p.To)
	return ok
}

func (cf CurrencyFilter) FindSymbol(s string) (CurrencyPair, bool) {
	v, ok := cf.symbols[s]
	return v, ok
}
