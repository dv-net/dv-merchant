package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Currency struct {
	ID         string
	Blockchain string
	Code       string
	Identifier string
}

type BlockchainData struct {
	Blockchain     string
	ConstName      string
	NativeCurrency string
	Tokens         []string
}

type TemplateData struct {
	PackageName string
	Blockchains []BlockchainData
}

func main() {
	var seedFile string
	var outputFile string
	var packageName string

	flag.StringVar(&seedFile, "seed", "sql/postgres/seeds/currencies.up.sql", "Path to the SQL seed file")
	flag.StringVar(&outputFile, "output", "internal/models/blockchain_gen.go", "Path to the output Go file")
	flag.StringVar(&packageName, "package", "models", "Package name for the generated file")
	flag.Parse()

	content, err := os.ReadFile(seedFile)
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
	}

	currencies := parseCurrencies(string(content))
	if len(currencies) == 0 {
		log.Fatal("No currencies found in seed file")
	}

	blockchains := extractBlockchains(currencies)
	if len(blockchains) == 0 {
		log.Fatal("No blockchains found in currencies")
	}

	code, err := generateGoCode(packageName, blockchains)
	if err != nil {
		log.Fatalf("Failed to generate Go code: %v", err)
	}

	err = os.WriteFile(outputFile, code, 0600)
	if err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	fmt.Printf("Successfully generated %s with %d blockchains\n", outputFile, len(blockchains))
}

func parseCurrencies(sql string) []Currency {
	valuesRe := regexp.MustCompile(`\((?:'[^']*'|[^,)]+?)(?:, *(?:'[^']*'|[^,)]+?)){16}\)`)
	valuesMatch := valuesRe.FindAllStringSubmatch(sql, -1)
	if len(valuesMatch) < 2 {
		return []Currency{}
	}

	// Remove the first match being VALUES
	valuesMatch = valuesMatch[1:]
	currencies := make([]Currency, 0, len(valuesMatch))

	records := make([][]string, 0, len(valuesMatch))
	for _, match := range valuesMatch {
		records = append(records, strings.Split(match[0], ","))
	}

	for _, record := range records {
		currency := Currency{}

		// Skip fiat
		if parseBool(record[4]) {
			continue
		}

		// Parse currency ID
		{
			currencyID := cleanString(record[0])
			if currencyID == "" {
				continue
			}
			currency.ID = currencyID
		}

		// Parse currency code
		{
			currencyCode := cleanString(record[1])
			if currencyCode == "" {
				continue
			}
			currency.Code = currencyCode
		}

		// Parse currency blockchain
		{
			currencyBlockchain := cleanString(record[5])
			if currencyBlockchain == "" {
				continue
			}
			currency.Blockchain = currencyBlockchain
		}

		// Parse currency contract address (being ticker for native, and address for token)
		{
			currencyContractAddress := cleanString(record[6])
			if currencyContractAddress == "" {
				continue
			}
			currency.Identifier = currencyContractAddress
		}

		currencies = append(currencies, currency)
	}

	return currencies
}

func cleanString(s string) string {
	s = strings.Trim(s, "(")
	s = strings.Trim(s, ")")
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "'")
	if s == "null" {
		return ""
	}
	return s
}

func parseBool(s string) bool {
	s = strings.TrimSpace(s)
	return s == "true"
}

func extractBlockchains(currencies []Currency) []BlockchainData {
	blockchainMap := make(map[string]*BlockchainData)

	for _, c := range currencies {
		if _, exists := blockchainMap[c.Blockchain]; !exists {
			blockchainMap[c.Blockchain] = &BlockchainData{
				Blockchain: c.Blockchain,
				ConstName:  generateConstName(c.Blockchain),
			}
		}

		// Check if this is the native currency
		// Native currency has contract address as lowercase of its code
		if c.Identifier == strings.ToLower(c.Code) {
			blockchainMap[c.Blockchain].NativeCurrency = c.ID
		} else {
			// Add token to the blockchain data
			blockchainMap[c.Blockchain].Tokens = append(blockchainMap[c.Blockchain].Tokens, c.ID)
		}
	}

	blockchains := make([]BlockchainData, 0, len(blockchainMap))
	for _, bd := range blockchainMap {
		blockchains = append(blockchains, *bd)
	}

	sort.Slice(blockchains, func(i, j int) bool {
		return blockchains[i].Blockchain < blockchains[j].Blockchain
	})

	return blockchains
}

func generateConstName(blockchain string) string {
	// Convert blockchain name to a valid Go constant name
	// e.g., "ethereum" -> "BlockchainEthereum", "bsc" -> "BlockchainBSC"
	if blockchain == "bsc" {
		return "BlockchainBinanceSmartChain"
	}
	if blockchain == "bitcoincash" {
		return "BlockchainBitcoinCash"
	}
	caser := cases.Title(language.English)
	return "Blockchain" + caser.String(blockchain)
}

func generateGoCode(packageName string, blockchains []BlockchainData) ([]byte, error) {
	data := TemplateData{
		PackageName: packageName,
		Blockchains: blockchains,
	}

	// Parse and execute template
	tmpl, err := template.New("blockchains").Parse(goTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Printf("failed to format generated code: %v\n", err)
		return buf.Bytes(), nil
	}

	return formatted, nil
}

const goTemplate = `
// Code generated. DO NOT EDIT.

package {{.PackageName}}

import "errors"

type Blockchain string // @name Blockchain

const ({{range .Blockchains}}
	{{.ConstName}} Blockchain = "{{.Blockchain}}"{{end}}
)

func (o Blockchain) String() string {return string(o)}

func (o Blockchain) Valid() error {
	switch o { {{range .Blockchains}}
	case {{.ConstName}}:
		return nil{{end}}
	}
	return errors.New("invalid blockchain: " + string(o))
}

func (o Blockchain) Tokens() ([]string, error) {
	switch o {
	{{range .Blockchains}}case {{.ConstName}}:
		return []string{ {{range .Tokens}}"{{.}}", {{end}} }, nil
	{{end}}
	}
	return nil, errors.New("invalid blockchain: " + string(o))
}

func (o Blockchain) NativeCurrency() (string, error) {
	switch o {
	{{range .Blockchains}}case {{.ConstName}}:
		return "{{.NativeCurrency}}", nil
	{{end}}
	}
	return "", errors.New("invalid blockchain: " + string(o))
}

func AllBlockchain() []Blockchain {
	return []Blockchain{ {{range .Blockchains}}
		{{.ConstName}},{{end}}
	}
}
`
