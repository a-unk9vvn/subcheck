package main

import (
 "bufio"
 "flag"
 "fmt"
 "net"
 "net/http"
 "os"
 "strings"
 "sync"
 "time"
)

var statusColors = map[string]string{
 "2xx":   "\033[32m", // Green
 "3xx":   "\033[34m", // Blue
 "4xx":   "\033[35m", // Purple
 "reset": "\033[0m",  // Reset color
}

// CheckProtocol determines whether a URL uses HTTPS or HTTP
func CheckProtocol(url string) string {
 httpsURL := "https://" + strings.TrimPrefix(url, "http://")
 resp, err := http.Get(httpsURL)
 if err == nil && resp.StatusCode < 400 {
  fmt.Printf("Target %s is using HTTPS.\n", url)
  return httpsURL
 }
 fmt.Printf("Target %s is using HTTP.\n", url)
 return "http://" + strings.TrimPrefix(url, "http://")
}

// CheckDirectory checks the status of a directory on the target URL
func CheckDirectory(url string, directory string, outputChan chan<- string, semaphore chan struct{}) {
 defer func() { <-semaphore }()
 fullURL := fmt.Sprintf("%s/%s", strings.TrimRight(url, "/"), directory)
 client := &http.Client{Timeout: 5 * time.Second}
 resp, err := client.Get(fullURL)
 if err != nil {
  outputChan <- fmt.Sprintf("Error accessing %s: %v", fullURL, err)
  return
 }
 defer resp.Body.Close()
 statusCode := resp.StatusCode
 statusFamily := fmt.Sprintf("%dxx", statusCode/100)
 color := statusColors["reset"]
 if c, ok := statusColors[statusFamily]; ok {
  color = c
 }
 if statusCode == 404 {
  color = "\033[35m"
 }
 outputChan <- fmt.Sprintf("%s%s - Status: %d%s", color, fullURL, statusCode, statusColors["reset"])
}

// LookupDNS retrieves detailed DNS records for a domain
func LookupDNS(domain string) string {
 var builder strings.Builder
 builder.WriteString(fmt.Sprintf("=== DNS Lookup Results for Domain: %s ===\n\n", domain))
 builder.WriteString("[+] A Records:\n")
 ips, err := net.LookupIP(domain)
 if err != nil {
  builder.WriteString(fmt.Sprintf("  [!] Error fetching A records: %v\n", err))
 } else {
  for _, ip := range ips {
   if ip.To4() != nil {
    builder.WriteString(fmt.Sprintf("  - IP Address: %s\n", ip))
   }
  }
 }
 builder.WriteString("\n[+] CNAME Records:\n")
 cname, err := net.LookupCNAME(domain)
 if err != nil {
  builder.WriteString(fmt.Sprintf("  [!] Error fetching CNAME records: %v\n", err))
 } else {
  builder.WriteString(fmt.Sprintf("  - Canonical Name: %s\n", cname))
 }
 builder.WriteString("\n[+] MX Records:\n")
 mxRecords, err := net.LookupMX(domain)
 if err != nil {
  builder.WriteString(fmt.Sprintf("  [!] Error fetching MX records: %v\n", err))
 } else {
  for _, mx := range mxRecords {
   builder.WriteString(fmt.Sprintf("  - Mail Server: %s, Priority: %d\n", mx.Host, mx.Pref))
  }
 }
 builder.WriteString("\n[+] NS Records:\n")
 nsRecords, err := net.LookupNS(domain)
 if err != nil {
  builder.WriteString(fmt.Sprintf("  [!] Error fetching NS records: %v\n", err))
 } else {
  for _, ns := range nsRecords {
   builder.WriteString(fmt.Sprintf("  - Name Server: %s\n", ns.Host))
  }
 }
 builder.WriteString("\n[+] TXT Records:\n")
 txtRecords, err := net.LookupTXT(domain)
 if err != nil {
  builder.WriteString(fmt.Sprintf("  [!] Error fetching TXT records: %v\n", err))
 } else {
  for _, txt := range txtRecords {
   builder.WriteString(fmt.Sprintf("  - %s\n", txt))
  }
 }
 builder.WriteString("\n=== End of Results ===\n")
 return builder.String()
}

func PrintHelp() {
 fmt.Println("Usage: go run main.go [OPTIONS]")
 fmt.Println("Options:")
 fmt.Println("  -u <url>     : Target URL (e.g., example.com)")
 fmt.Println("  -l <file>    : File containing directories to check")
 fmt.Println("  -d <domain>  : Domain to fetch DNS records")
 fmt.Println("  -o <file>    : Output file (optional)")
 fmt.Println("  -t <threads> : Number of threads to use (default: 50)")
 fmt.Println("  -h           : Show this help message")
}

func main() {
 urlPtr := flag.String("u", "", "Target URL (e.g., example.com)")
listPtr := flag.String("l", "", "File containing directories to check")
 domainPtr := flag.String("d", "", "Domain to fetch DNS records")
 outputPtr := flag.String("o", "", "Output file (optional)")
 threadPtr := flag.Int("t", 50, "Number of threads to use (default: 50)")
 helpPtr := flag.Bool("h", false, "Show help message")
 flag.Parse()
 if *helpPtr || len(os.Args) < 3 {
  PrintHelp()
  return
 }
 if *urlPtr != "" && *listPtr != "" {
  finalURL := CheckProtocol(*urlPtr)
  file, err := os.Open(*listPtr)
  if err != nil {
   fmt.Printf("Error opening list file: %v\n", err)
   return
  }
  defer file.Close()
  outputChan := make(chan string, 1000)
  var wg sync.WaitGroup
  semaphore := make(chan struct{}, *threadPtr)
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
   directory := scanner.Text()
   wg.Add(1)
   semaphore <- struct{}{}
   go func(dir string) {
    defer wg.Done()
    CheckDirectory(finalURL, dir, outputChan, semaphore)
   }(directory)
  }
  go func() {
   for result := range outputChan {
    if *outputPtr != "" {
     os.WriteFile(*outputPtr, []byte(result+"\n"), 0644)
    } else {
     fmt.Println(result)
    }
   }
  }()
  wg.Wait()
  close(outputChan)
 }
 if *domainPtr != "" {
  results := LookupDNS(*domainPtr)
  if *outputPtr != "" {
   os.WriteFile(*outputPtr, []byte(results), 0644)
  } else {
   fmt.Println(results)
  }
 }
}
