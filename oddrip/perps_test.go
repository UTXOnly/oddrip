package oddrip

import "testing"

func TestNew_WiresPerpsServices(t *testing.T) {
	c := New()
	p := c.Perps
	if p == nil {
		t.Fatal("Perps is nil")
	}
	services := map[string]bool{
		"Exchange":     p.Exchange != nil && p.Exchange.client == c,
		"Account":      p.Account != nil && p.Account.client == c,
		"Markets":      p.Markets != nil && p.Markets.client == c,
		"Orders":       p.Orders != nil && p.Orders.client == c,
		"OrderGroups":  p.OrderGroups != nil && p.OrderGroups.client == c,
		"Portfolio":    p.Portfolio != nil && p.Portfolio.client == c,
		"Risk":         p.Risk != nil && p.Risk.client == c,
		"Fees":         p.Fees != nil && p.Fees.client == c,
		"Funding":      p.Funding != nil && p.Funding.client == c,
		"ExitTriggers": p.ExitTriggers != nil && p.ExitTriggers.client == c,
		"FCM":          p.FCM != nil && p.FCM.client == c,
	}
	for name, ok := range services {
		if !ok {
			t.Errorf("Perps.%s is not wired to the client", name)
		}
	}
}
