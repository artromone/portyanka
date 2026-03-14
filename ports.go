package main

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// PortEntry represents a listening TCP port with associated process information
type PortEntry struct {
	PID             uint32
	ProcessName     string
	Port            uint16
	Address         string
	Command         string
	ConnectionCount int
}

// ListListeningPorts discovers all listening TCP ports on the system
func ListListeningPorts() []PortEntry {
	cmd := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-P", "-n")
	output, err := cmd.Output()
	if err != nil {
		return []PortEntry{}
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 {
		return []PortEntry{}
	}

	var entries []PortEntry
	// Skip header line
	for _, line := range lines[1:] {
		if entry, ok := parseLsofLine(line); ok {
			entries = append(entries, entry)
		}
	}

	// Sort by (PID, Port) for deduplication
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].PID != entries[j].PID {
			return entries[i].PID < entries[j].PID
		}
		return entries[i].Port < entries[j].Port
	})

	// Deduplicate by (PID, Port)
	deduped := make([]PortEntry, 0, len(entries))
	for i, entry := range entries {
		if i == 0 || entry.PID != entries[i-1].PID || entry.Port != entries[i-1].Port {
			deduped = append(deduped, entry)
		}
	}

	// Fetch full command lines
	deduped = fetchCommands(deduped)

	// Get connection counts
	connCounts := GetConnectionCounts()
	for i := range deduped {
		if count, ok := connCounts[deduped[i].Port]; ok {
			deduped[i].ConnectionCount = count
		}
	}

	// Final sort by port
	sort.Slice(deduped, func(i, j int) bool {
		return deduped[i].Port < deduped[j].Port
	})

	return deduped
}

// parseLsofLine parses a single line of lsof output
func parseLsofLine(line string) (PortEntry, bool) {
	fields := strings.Fields(line)
	if len(fields) < 10 {
		return PortEntry{}, false
	}

	processName := fields[0]

	pid, err := strconv.ParseUint(fields[1], 10, 32)
	if err != nil {
		return PortEntry{}, false
	}

	// Address:port is in the second-to-last field
	addrPort := fields[len(fields)-2]
	address, port, ok := parseAddrPort(addrPort)
	if !ok {
		return PortEntry{}, false
	}

	return PortEntry{
		PID:         uint32(pid),
		ProcessName: processName,
		Port:        port,
		Address:     address,
		Command:     "",
	}, true
}

// parseAddrPort parses an address:port string, handling IPv4, IPv6, and wildcard
func parseAddrPort(s string) (string, uint16, bool) {
	// Handle IPv6 with brackets: [::1]:8080
	if strings.HasPrefix(s, "[") {
		closeBracket := strings.Index(s, "]")
		if closeBracket == -1 {
			return "", 0, false
		}
		address := s[1:closeBracket]
		if len(s) <= closeBracket+2 || s[closeBracket+1] != ':' {
			return "", 0, false
		}
		portStr := s[closeBracket+2:]
		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return "", 0, false
		}
		return address, uint16(port), true
	}

	// Handle IPv4 or wildcard: 127.0.0.1:8080 or *:3000
	lastColon := strings.LastIndex(s, ":")
	if lastColon == -1 {
		return "", 0, false
	}

	address := s[:lastColon]
	portStr := s[lastColon+1:]

	// Reject empty address or port
	if address == "" || portStr == "" {
		return "", 0, false
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return "", 0, false
	}

	return address, uint16(port), true
}

// fetchCommands retrieves full command lines for the given port entries
func fetchCommands(entries []PortEntry) []PortEntry {
	if len(entries) == 0 {
		return entries
	}

	// Build comma-separated PID list
	pids := make(map[uint32]bool)
	for _, entry := range entries {
		pids[entry.PID] = true
	}

	pidList := make([]string, 0, len(pids))
	for pid := range pids {
		pidList = append(pidList, fmt.Sprintf("%d", pid))
	}

	// Execute ps command
	cmd := exec.Command("ps", "-ww", "-p", strings.Join(pidList, ","), "-o", "pid=,command=")
	output, err := cmd.Output()
	if err != nil {
		return entries
	}

	// Parse ps output into a map
	cmdMap := make(map[uint32]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split on first whitespace
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		pid, err := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 32)
		if err != nil {
			continue
		}

		cmdMap[uint32(pid)] = strings.TrimSpace(parts[1])
	}

	// Update entries with commands
	for i := range entries {
		if cmd, ok := cmdMap[entries[i].PID]; ok {
			entries[i].Command = cmd
		}
	}

	return entries
}

// GetConnectionCounts counts established connections per port
func GetConnectionCounts() map[uint16]int {
	counts := make(map[uint16]int)

	// Try netstat first (works on both macOS and Linux)
	cmd := exec.Command("netstat", "-an", "-p", "tcp")
	output, err := cmd.Output()
	if err != nil {
		// Try ss on Linux
		cmd = exec.Command("ss", "-tn", "state", "established")
		output, err = cmd.Output()
		if err != nil {
			return counts
		}
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if !strings.Contains(line, "ESTABLISHED") {
			continue
		}

		// Parse the line to extract local port
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// Local address is typically in field 3 or 4 depending on format
		for _, field := range fields {
			if strings.Contains(field, ":") || strings.Contains(field, ".") {
				// Check if this looks like an address:port
				if _, port, ok := parseAddrPort(field); ok {
					counts[port]++
					break
				}
			}
		}
	}

	return counts
}

// KillProcess kills a process with the given PID
func KillProcess(pid uint32, force bool) error {
	signal := "-TERM"
	if force {
		signal = "-KILL"
	}

	cmd := exec.Command("kill", signal, fmt.Sprintf("%d", pid))
	return cmd.Run()
}

// SortBy defines the field to sort by
type SortBy int

const (
	SortByPID SortBy = iota
	SortByPort
	SortByName
)

// SortEntries sorts port entries by the specified field
func SortEntries(entries []PortEntry, sortBy SortBy, ascending bool) {
	sort.SliceStable(entries, func(i, j int) bool {
		var less bool
		switch sortBy {
		case SortByPID:
			less = entries[i].PID < entries[j].PID
		case SortByPort:
			less = entries[i].Port < entries[j].Port
		case SortByName:
			less = strings.ToLower(entries[i].ProcessName) <
				strings.ToLower(entries[j].ProcessName)
		}

		if !ascending {
			return !less
		}
		return less
	})
}
