package problems

import "testing"

func TestCatalogOwnershipAndAllowlist(t *testing.T) {
	list := List()
	if len(list) != 10 || list[0].ID != "a-plus-b" {
		t.Fatal("unexpected catalog order or length")
	}
	seen := map[string]bool{}
	for _, p := range list {
		if seen[p.ID] || p.Title == "" || p.Summary == "" || p.Description == "" || p.InputDescription == "" || p.OutputDescription == "" || p.StarterCode == "" || len(p.Samples) == 0 {
			t.Fatalf("incomplete or duplicated definition: %s", p.ID)
		}
		seen[p.ID] = true
		if p.Type != "batch" && p.Type != "interactive" && p.Type != "communication" {
			t.Fatalf("unsupported type: %s", p.ID)
		}
		if _, ok := Get(p.ID); !ok || !Exists(p.ID) {
			t.Fatalf("registered problem not found: %s", p.ID)
		}
	}
	for _, id := range []string{"", "missing", "../a-plus-b", "a-plus-b/", `a-plus-b\sample.in`, "%2e%2e", "A-PLUS-B"} {
		if Exists(id) {
			t.Fatalf("unknown or path-like problem accepted: %q", id)
		}
	}
	list[0].ID = "injected"
	list[0].Samples[0].Output = "changed"
	p, _ := Get("a-plus-b")
	if Exists("injected") || p.Samples[0].Output != "3\n" {
		t.Fatal("catalog mutated through List")
	}
	p.Samples[0].Input = "changed"
	p, _ = Get("a-plus-b")
	if p.Samples[0].Input != "1 2\n" {
		t.Fatal("catalog mutated through Get")
	}
	communication, _ := Get("communication-sum")
	if len(communication.StarterCodes) != 2 || communication.StarterCode != communication.StarterCodes["sender"] {
		t.Fatal("communication templates absent")
	}
	communication.StarterCodes["receiver"] = "changed"
	communication, _ = Get("communication-sum")
	if communication.StarterCodes["receiver"] == "changed" {
		t.Fatal("communication templates not owned")
	}
}
