// Copyright 2019 Google LLC
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package diag

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
)

const (
	// ENDPOINT is the googleapis host.
	ENDPOINT = "googleads.googleapis.com.:80"
)

// SysInfo stores the relevant system information.
type SysInfo struct {
	Host     string
	CPUs     int
	OS       string
	Arch     string
	GOROOT   string
	PageSize int
	Heap     uint64
}

// Init intializes the struct with the runtime system parameters.
func (s *SysInfo) Init() {
	host, err := os.Hostname()
	if err != nil {
		host = "ERROR"
	}

	s.Host = host
	s.CPUs = runtime.NumCPU()
	s.OS = runtime.GOOS
	s.Arch = runtime.GOARCH
	s.GOROOT = runtime.GOROOT()
	s.CPUs = runtime.NumCPU()
	s.PageSize = os.Getpagesize()
	s.Heap = heap()
}

// Print outputs the contents of a Sysinfo structure to stdout.
func (s *SysInfo) Print() {
	fmt.Printf("Host: %s\nCPUs: %d\nOS: %s\nArch: %s\nPageSize: %d bytes\nHeap: %d bytes\n",
		s.Host, s.CPUs, s.OS, s.Arch, s.PageSize, s.Heap)
}

// heap returns the amount of heap in bytes for this runtime.
func heap() uint64 {
	mstats := &runtime.MemStats{}
	runtime.ReadMemStats(mstats)
	return mstats.TotalAlloc
}

// ConnEndpoint opens a tcp connection to the endpoint
func ConnEndpoint() error {
	conn, err := net.Dial("tcp", ENDPOINT)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

// PrintIPv4 prints local non-loopback IPv4 addresses
func PrintIPv4(host string) {
	addrs, err := net.LookupIP(host)
	if err != nil {
		log.Printf("ERROR: PrintIPV4: %v\n", err)
	}

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			fmt.Printf("IPV4:%s\n ", ipv4)

		}
	}
}
