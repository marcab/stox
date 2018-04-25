package stox

import (
	"context"
	"fmt"
	"math"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/jirwin/quadlek/quadlek"
)

var client *av.Client = nil
var regexp *Regexp = nil

func GetQuote(sym string) (string, error) {
	res, err := client.StockTimeSeries(av.TimeSeriesDaily, sym)
	if err != nil {
		return "", err
	}

	// Results are returned oldest first
	// For `today` grab the last item
	qq := res[len(res)-1]

	// Previous close can default to current open
	pc := qq.Open
	if len(res) > 1 {
		pc = res[len(res)-2].Close
	}

	// Find pct gain/loss from open
	cng := qq.Close - pc
	cngs := fmt.Sprintf("⬇️$%.2f", math.Abs(cng))
	if cng >= 0 {
		cngs = fmt.Sprintf("⬆️$%.2f", cng)
	}

	pct := (cng / pc) * 100
	pcts := fmt.Sprintf("%.2f%%", math.Abs(pct))

	return fmt.Sprintf(
		"https://robinhood.com/stocks/%s\nNow: $%.2f: %s (%s)\nOpen: $%.2f, $%.2f ↔️ $%.2f,",
		sym, qq.Close, cngs, pcts, qq.Open, qq.Low, qq.High), nil

}

func stoxHook(ctx context.Context, hookChan <-chan *quadlek.HookMsg) {
	for {
		select {
		case hook := <-hookChan:
			tokens := strings.Split(hook.Msg.Text, " ")

			for _, t := range regexp.FindAllString(hook.Msg.Text, -1) {
				symbol := t[1:]
				log.Info(fmt.Sprintf("Symobl lookup triggered: %s", symbol))
				quote, err := GetQuote(symbol)
				if err != nil {
					hook.Bot.Say(hook.Msg.Channel, fmt.Sprintf("Could not fetch quote for: %s", symbol))
				} else {
					hook.Bot.Say(hook.Msg.Channel, quote)
				}
			}

		case <-ctx.Done():
			log.Info("Exiting Stox")
			return
		}
	}
}

func Register(apiKey string) quadlek.Plugin {
	client = av.NewClient(apiKey)
	regexp := regexp.MustCompile(`\b$[A-Z]{1,5}\b`)

	return quadlek.MakePlugin(
		"stox",
		nil,
		[]quadlek.Hook{
			quadlek.MakeHook(stoxHook),
		},
		nil,
		nil,
		nil,
	)
}
