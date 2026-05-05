package gitproviders

import "testing"

func TestKonfluxPACWebhookSecretDataKey_exampleFromUser(t *testing.T) {
	const u = "https://gitlab.com/rhtap-test-organization-jk/testrepo"
	const want = "https___gitlab.com_rhtap-test-organization-jk_testrepo"
	if g := KonfluxPACWebhookSecretDataKey(u); g != want {
		t.Fatalf("got %q want %q", g, want)
	}
	if g := KonfluxPACWebhookSecretDataKey(u + ".git"); g != want {
		t.Fatalf(".git suffix: got %q want %q", g, want)
	}
}
