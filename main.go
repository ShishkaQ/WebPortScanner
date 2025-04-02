package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type ScanResult struct {
	Port    int
	State   string
	Service string
}

var commonPorts = map[int]string{
	21:  "FTP",
	22:  "SSH",
	80:  "HTTP",
	443: "HTTPS",
	3306: "MySQL",
}

func main() {
	host := flag.String("host", "scanme.nmap.org", "Target host")
	startPort := flag.Int("start", 1, "Start port")
	endPort := flag.Int("end", 1024, "End port")
	timeout := flag.Duration("timeout", 500*time.Millisecond, "Connection timeout")
	workers := flag.Int("workers", 100, "Number of workers")
	output := flag.String("output", "", "Output file (optional)")
	flag.Parse()

	if *startPort > *endPort {
		fmt.Println("Invalid port range")
		os.Exit(1)
	}

	ports := make(chan int, *workers)
	results := make(chan ScanResult)
	var wg sync.WaitGroup

	// Launch workers
	for i := 0; i < *workers; i++ {
		go worker(*host, *timeout, ports, results, &wg)
	}

	// Send ports to scan
	go func() {
		for port := *startPort; port <= *endPort; port++ {
			wg.Add(1)
			ports <- port
		}
		close(ports)
	}()

	// Handle results
	var outputFile *os.File
	if *output != "" {
		var err error
		outputFile, err = os.Create(*output)
		if err != nil {
			fmt.Printf("Error creating file: %v\n", err)
			os.Exit(1)
		}
		defer outputFile.Close()
	}

	go func() {
		totalPorts := *endPort - *startPort + 1
		scanned := 0
		for result := range results {
			scanned++
			progress := float64(scanned) / float64(totalPorts) * 100
			fmt.Printf("\rProgress: %.2f%%", progress)

			if result.State == "open" {
				service := getService(result.Port)
				line := fmt.Sprintf("Port %d (%s) is %s\n", result.Port, service, result.State)
				fmt.Print(line)
				
				if outputFile != nil {
					outputFile.WriteString(line)
				}
			}
			wg.Done()
		}
	}()

	wg.Wait()
	close(results)
	fmt.Println("\nScan completed!")
}

func worker(host string, timeout time.Duration, ports <-chan int, results chan<- ScanResult, wg *sync.WaitGroup) {
	for port := range ports {
		address := fmt.Sprintf("%s:%d", host, port)
		conn, err := net.DialTimeout("tcp", address, timeout)
		
		result := ScanResult{Port: port}
		if err == nil {
			conn.Close()
			result.State = "open"
		} else {
			result.State = "closed"
		}
		
		results <- result
	}
}

func getService(port int) string {
	if service, ok := commonPorts[port]; ok {
		return service
	}
	return "unknown"
}