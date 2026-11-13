package providers

import (
	"strings"
	"sync"

	"github.com/dv-net/dv-merchant/pkg/aml"
	"github.com/dv-net/dv-merchant/pkg/aml/providers/aml_bot"
	"github.com/dv-net/dv-merchant/pkg/aml/providers/bitok"
	"github.com/dv-net/dv-merchant/pkg/aml/providers/coinkyt"
)

var signalLabels = sync.OnceValue(func() map[string]string {
	listers := []aml.SignalCategoryLister{
		(*aml_bot.Client)(nil),
		(*bitok.Client)(nil),
		(*coinkyt.Client)(nil),
	}

	labels := make(map[string]string)
	for _, lister := range listers {
		for _, c := range lister.SignalCategories() {
			key := strings.ToLower(c.Category)
			if _, ok := labels[key]; !ok {
				labels[key] = c.Label
			}
		}
	}

	return labels
})

func SignalCategoryLabel(category string) (string, bool) {
	label, ok := signalLabels()[strings.ToLower(category)]
	return label, ok
}
