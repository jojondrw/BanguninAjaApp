package access

import "testing"

func TestModuleForPathLeavesThePublicPayLinkOpen(t *testing.T) {
	for _, path := range []string{"/api/pay/:token", "/api/pay/:token/confirm", "/api/pay/AbCdEfGhIjKlMnOpQrStUvWxYz012345"} {
		if _, guarded := ModuleForPath(path); guarded {
			t.Errorf("%s should stay open for buyers without an account", path)
		}
	}
	for _, path := range []string{"/api/billing/onchain/config", "/api/billing/invoices/:id/onchain-requests", "/api/payables", "/api/payx"} {
		if _, guarded := ModuleForPath(path); !guarded {
			t.Errorf("%s must stay guarded", path)
		}
	}
}
