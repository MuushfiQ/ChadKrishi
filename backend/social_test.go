package main

import "testing"

func TestOrderStatePermissions(t *testing.T) {
	cases := []struct {
		state, action string
		buyer, seller bool
		want          string
		ok            bool
	}{
		{"waiting_payment", "paid", true, false, "buyer_paid", true},
		{"waiting_payment", "paid", false, true, "", false},
		{"waiting_payment", "received", false, true, "", false},
		{"buyer_paid", "received", true, false, "", false},
		{"buyer_paid", "received", false, true, "completed", true},
		{"buyer_paid", "cancel", true, false, "", false},
		{"buyer_paid", "not_received", false, true, "waiting_payment", true},
		{"completed", "paid", true, false, "", false},
		{"completed", "cancel", false, true, "", false},
		{"cancelled", "received", false, true, "", false},
		{"waiting_payment", "cancel", true, false, "cancelled", true},
		{"waiting_payment", "cancel", false, false, "", false},
	}
	for _, c := range cases {
		got, ok := nextOrderState(c.state, c.action, c.buyer, c.seller)
		if got != c.want || ok != c.ok {
			t.Errorf("%+v got %s %v", c, got, ok)
		}
	}
}
