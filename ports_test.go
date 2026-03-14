package main

import (
	"testing"
)

func TestParseAddrPort_IPv4(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantAddr    string
		wantPort    uint16
		wantSuccess bool
	}{
		{"localhost", "127.0.0.1:8080", "127.0.0.1", 8080, true},
		{"wildcard", "*:3000", "*", 3000, true},
		{"all interfaces", "0.0.0.0:80", "0.0.0.0", 80, true},
		{"high port", "192.168.1.1:65535", "192.168.1.1", 65535, true},
		{"low port", "10.0.0.1:22", "10.0.0.1", 22, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, port, ok := parseAddrPort(tt.input)
			if ok != tt.wantSuccess {
				t.Errorf("parseAddrPort(%q) success = %v, want %v", tt.input, ok, tt.wantSuccess)
				return
			}
			if ok && addr != tt.wantAddr {
				t.Errorf("parseAddrPort(%q) addr = %q, want %q", tt.input, addr, tt.wantAddr)
			}
			if ok && port != tt.wantPort {
				t.Errorf("parseAddrPort(%q) port = %d, want %d", tt.input, port, tt.wantPort)
			}
		})
	}
}

func TestParseAddrPort_IPv6(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantAddr    string
		wantPort    uint16
		wantSuccess bool
	}{
		{"localhost", "[::1]:8080", "::1", 8080, true},
		{"all interfaces", "[::]:443", "::", 443, true},
		{"full address", "[2001:db8::1]:9090", "2001:db8::1", 9090, true},
		{"link local", "[fe80::1]:3000", "fe80::1", 3000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, port, ok := parseAddrPort(tt.input)
			if ok != tt.wantSuccess {
				t.Errorf("parseAddrPort(%q) success = %v, want %v", tt.input, ok, tt.wantSuccess)
				return
			}
			if ok && addr != tt.wantAddr {
				t.Errorf("parseAddrPort(%q) addr = %q, want %q", tt.input, addr, tt.wantAddr)
			}
			if ok && port != tt.wantPort {
				t.Errorf("parseAddrPort(%q) port = %d, want %d", tt.input, port, tt.wantPort)
			}
		})
	}
}

func TestParseAddrPort_Invalid(t *testing.T) {
	tests := []string{
		"invalid",
		"no-colon",
		":8080",           // No address
		"127.0.0.1:",      // No port
		"127.0.0.1:abc",   // Invalid port
		"[::1:8080",       // Missing closing bracket
		"[::1]",           // No port
		"127.0.0.1:99999", // Port too large
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, _, ok := parseAddrPort(input)
			if ok {
				t.Errorf("parseAddrPort(%q) should have failed but succeeded", input)
			}
		})
	}
}

func TestParseLsofLine_Valid(t *testing.T) {
	// Typical lsof output line
	line := "node      1234 user   21u  IPv4 0x12345      0t0  TCP 127.0.0.1:3000 (LISTEN)"

	entry, ok := parseLsofLine(line)
	if !ok {
		t.Fatal("parseLsofLine() failed, expected success")
	}

	if entry.ProcessName != "node" {
		t.Errorf("ProcessName = %q, want %q", entry.ProcessName, "node")
	}
	if entry.PID != 1234 {
		t.Errorf("PID = %d, want %d", entry.PID, 1234)
	}
	if entry.Address != "127.0.0.1" {
		t.Errorf("Address = %q, want %q", entry.Address, "127.0.0.1")
	}
	if entry.Port != 3000 {
		t.Errorf("Port = %d, want %d", entry.Port, 3000)
	}
}

func TestParseLsofLine_IPv6(t *testing.T) {
	line := "python3   5678 user   4u  IPv6 0x67890      0t0  TCP [::1]:8080 (LISTEN)"

	entry, ok := parseLsofLine(line)
	if !ok {
		t.Fatal("parseLsofLine() failed, expected success")
	}

	if entry.ProcessName != "python3" {
		t.Errorf("ProcessName = %q, want %q", entry.ProcessName, "python3")
	}
	if entry.PID != 5678 {
		t.Errorf("PID = %d, want %d", entry.PID, 5678)
	}
	if entry.Address != "::1" {
		t.Errorf("Address = %q, want %q", entry.Address, "::1")
	}
	if entry.Port != 8080 {
		t.Errorf("Port = %d, want %d", entry.Port, 8080)
	}
}

func TestParseLsofLine_Wildcard(t *testing.T) {
	line := "nginx     9999 root   6u  IPv4 0xabcdef      0t0  TCP *:80 (LISTEN)"

	entry, ok := parseLsofLine(line)
	if !ok {
		t.Fatal("parseLsofLine() failed, expected success")
	}

	if entry.ProcessName != "nginx" {
		t.Errorf("ProcessName = %q, want %q", entry.ProcessName, "nginx")
	}
	if entry.PID != 9999 {
		t.Errorf("PID = %d, want %d", entry.PID, 9999)
	}
	if entry.Address != "*" {
		t.Errorf("Address = %q, want %q", entry.Address, "*")
	}
	if entry.Port != 80 {
		t.Errorf("Port = %d, want %d", entry.Port, 80)
	}
}

func TestParseLsofLine_Invalid(t *testing.T) {
	tests := []string{
		"",
		"too few fields",
		"a b c d e f g h i",            // Not enough fields
		"cmd abc user x x x x x x x",   // Invalid PID
		"cmd 123 user x x x x x xyz x", // Invalid address:port
	}

	for _, line := range tests {
		t.Run(line, func(t *testing.T) {
			_, ok := parseLsofLine(line)
			if ok {
				t.Errorf("parseLsofLine(%q) should have failed but succeeded", line)
			}
		})
	}
}

func TestSortEntries(t *testing.T) {
	entries := []PortEntry{
		{PID: 200, ProcessName: "zebra", Port: 8080},
		{PID: 100, ProcessName: "alpha", Port: 3000},
		{PID: 300, ProcessName: "beta", Port: 80},
	}

	// Test sort by PID ascending
	t.Run("PID ascending", func(t *testing.T) {
		sorted := make([]PortEntry, len(entries))
		copy(sorted, entries)
		SortEntries(sorted, SortByPID, true)

		if sorted[0].PID != 100 || sorted[1].PID != 200 || sorted[2].PID != 300 {
			t.Errorf("Sort by PID ascending failed: got %v", sorted)
		}
	})

	// Test sort by PID descending
	t.Run("PID descending", func(t *testing.T) {
		sorted := make([]PortEntry, len(entries))
		copy(sorted, entries)
		SortEntries(sorted, SortByPID, false)

		if sorted[0].PID != 300 || sorted[1].PID != 200 || sorted[2].PID != 100 {
			t.Errorf("Sort by PID descending failed: got %v", sorted)
		}
	})

	// Test sort by Port ascending
	t.Run("Port ascending", func(t *testing.T) {
		sorted := make([]PortEntry, len(entries))
		copy(sorted, entries)
		SortEntries(sorted, SortByPort, true)

		if sorted[0].Port != 80 || sorted[1].Port != 3000 || sorted[2].Port != 8080 {
			t.Errorf("Sort by Port ascending failed: got %v", sorted)
		}
	})

	// Test sort by Name ascending
	t.Run("Name ascending", func(t *testing.T) {
		sorted := make([]PortEntry, len(entries))
		copy(sorted, entries)
		SortEntries(sorted, SortByName, true)

		if sorted[0].ProcessName != "alpha" || sorted[1].ProcessName != "beta" || sorted[2].ProcessName != "zebra" {
			t.Errorf("Sort by Name ascending failed: got %v", sorted)
		}
	})

	// Test sort by Name descending
	t.Run("Name descending", func(t *testing.T) {
		sorted := make([]PortEntry, len(entries))
		copy(sorted, entries)
		SortEntries(sorted, SortByName, false)

		if sorted[0].ProcessName != "zebra" || sorted[1].ProcessName != "beta" || sorted[2].ProcessName != "alpha" {
			t.Errorf("Sort by Name descending failed: got %v", sorted)
		}
	})
}
