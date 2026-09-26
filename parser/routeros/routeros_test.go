package routeros

import (
	"reflect"
	"testing"
)

func TestRouterOSParser(t *testing.T) {
	sample := `# 2025-01-12 06:20:22 by RouterOS 7.17rc7
# software id = XGWP-9N00
#
/ip address
add address=192.168.88.1/24 comment=defconf interface=bridge1 network=192.168.88.0
/system identity
set name=MikroTik`

	var parsed any
	if err := (&Parser{}).Unmarshal([]byte(sample), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	root, ok := parsed.(map[string]any)
	if !ok {
		t.Fatalf("expected map root, got %T", parsed)
	}

	addr, ok := root["/ip address"].([]any)
	if !ok || len(addr) != 1 {
		t.Fatalf("expected one /ip address entry, got %v", root["/ip address"])
	}

	first := addr[0].(map[string]any)
	if first["command"] != "add" {
		t.Errorf("command = %v, want add", first["command"])
	}
	if first["address"] != "192.168.88.1/24" {
		t.Errorf("address = %v, want 192.168.88.1/24", first["address"])
	}
	if first["interface"] != "bridge1" {
		t.Errorf("interface = %v, want bridge1", first["interface"])
	}

	identity := root["/system identity"].([]any)
	if identity[0].(map[string]any)["name"] != "MikroTik" {
		t.Errorf("identity name = %v, want MikroTik", identity[0])
	}
}

func TestLineContinuation(t *testing.T) {
	// A value wrapped mid-token (no space before the backslash) must rejoin
	// tightly, while a wrap between tokens (space before the backslash) keeps
	// the tokens separated.
	sample := `/interface ethernet
set advertise="10M-baseT-half,100M\
    -baseT-full" arp=enabled \
    name=ether1`

	var parsed any
	if err := (&Parser{}).Unmarshal([]byte(sample), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	e := parsed.(map[string]any)["/interface ethernet"].([]any)[0].(map[string]any)
	if e["advertise"] != "10M-baseT-half,100M-baseT-full" {
		t.Errorf("advertise = %q, want %q", e["advertise"], "10M-baseT-half,100M-baseT-full")
	}
	if e["arp"] != "enabled" {
		t.Errorf("arp = %v, want enabled", e["arp"])
	}
	if e["name"] != "ether1" {
		t.Errorf("name = %v, want ether1", e["name"])
	}
}

func TestSelectorsAndFlags(t *testing.T) {
	sample := `/interface list
set [ find name=all ] comment="contains all interfaces" name=all
/interface ethernet switch port
set 0 !egress-rate !ingress-rate`

	var parsed any
	if err := (&Parser{}).Unmarshal([]byte(sample), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	root := parsed.(map[string]any)

	list := root["/interface list"].([]any)[0].(map[string]any)
	if list["comment"] != "contains all interfaces" {
		t.Errorf("comment = %v, want quoted string", list["comment"])
	}
	args, ok := list["arguments"].([]any)
	if !ok || len(args) != 1 || args[0] != "[ find name=all ]" {
		t.Errorf("arguments = %v, want the find selector kept intact", list["arguments"])
	}

	port := root["/interface ethernet switch port"].([]any)[0].(map[string]any)
	portArgs := port["arguments"].([]any)
	want := []any{"0", "!egress-rate", "!ingress-rate"}
	if !reflect.DeepEqual(portArgs, want) {
		t.Errorf("port arguments = %v, want %v", portArgs, want)
	}
}

func TestMultipleEntriesPreserveOrder(t *testing.T) {
	sample := `/ip firewall filter
add action=accept chain=input comment=first
add action=drop chain=input comment=second`

	var parsed any
	if err := (&Parser{}).Unmarshal([]byte(sample), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	rules := parsed.(map[string]any)["/ip firewall filter"].([]any)
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].(map[string]any)["comment"] != "first" {
		t.Errorf("first rule comment = %v", rules[0])
	}
	if rules[1].(map[string]any)["comment"] != "second" {
		t.Errorf("second rule comment = %v", rules[1])
	}
}

func TestEmptyAndCommentOnly(t *testing.T) {
	var parsed any
	if err := (&Parser{}).Unmarshal([]byte("# just a comment\n\n"), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m, ok := parsed.(map[string]any); !ok || len(m) != 0 {
		t.Errorf("expected empty map, got %v", parsed)
	}
}

func TestCommandBeforeSectionErrors(t *testing.T) {
	err := (&Parser{}).Unmarshal([]byte("add name=oops"), new(any))
	if err == nil {
		t.Fatal("expected an error for a command outside any section")
	}
}

func TestUnquote(t *testing.T) {
	cases := map[string]string{
		`"plain"`:           "plain",
		`""`:                "",
		`unquoted`:          "unquoted",
		`"with \"quotes\""`: `with "quotes"`,
		`"back\\slash"`:     `back\slash`,
	}
	for in, want := range cases {
		if got := unquote(in); got != want {
			t.Errorf("unquote(%s) = %q, want %q", in, got, want)
		}
	}
}
