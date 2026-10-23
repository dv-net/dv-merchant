package kraken

// Kraken alt names that differ from common currency codes.
var (
	altNameToCode = map[string]string{
		"XBT": "BTC",
		"XDG": "DOGE",
	}
	codeToAltName = map[string]string{
		"BTC":  "XBT",
		"DOGE": "XDG",
	}
)

// CurrencyCode converts a Kraken alt name ("XBT", "XDG") to a common currency code ("BTC", "DOGE").
func CurrencyCode(altName string) string {
	if code, ok := altNameToCode[altName]; ok {
		return code
	}
	return altName
}

// AssetName converts a common currency code ("BTC") to the name accepted by Kraken API ("XBT").
func AssetName(code string) string {
	if name, ok := codeToAltName[code]; ok {
		return name
	}
	return code
}
