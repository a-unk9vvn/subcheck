package main

import (
 "bufio"
 "flag"
 "fmt"
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
 // Try HTTPS first
 httpsURL := "https://" + strings.TrimPrefix(url, "http://")
 resp, err := http.Get(httpsURL)
 if err == nil && resp.StatusCode < 400 {
  // If HTTPS works, return it
  fmt.Printf("Target %s is using HTTPS.\n", url)
  return httpsURL
 }
 // Fall back to HTTP if HTTPS fails
 fmt.Printf("Target %s is using HTTP.\n", url)
 return "http://" + strings.TrimPrefix(url, "http://")
}

// CheckDirectory checks the status of a directory on the target URL
func CheckDirectory(url string, directory string, outputChan chan<- string, semaphore chan struct{}) {
 defer func() { <-semaphore }()

 // Construct the full URL for the directory
 fullURL := fmt.Sprintf("%s/%s", strings.TrimRight(url, "/"), directory)
 client := &http.Client{Timeout: 5 * time.Second}

 // Send GET request to the directory
 resp, err := client.Get(fullURL)
 if err != nil {
  outputChan <- fmt.Sprintf("Error accessing %s: %v", fullURL, err)
  return
 }
 defer resp.Body.Close()

 // Determine the status code family and assign colors
 statusCode := resp.StatusCode
 statusFamily := fmt.Sprintf("%dxx", statusCode/100)
 color := statusColors["reset"]

 // Apply color based on status family
 if c, ok := statusColors[statusFamily]; ok {
  color = c
 }

 // Adjust color for 404 specifically (directory not found)
 if statusCode == 404 {
  color = "\033[35m" // Purple for 404
 }

 // Send the result to the channel
 outputChan <- fmt.Sprintf("%s%s - Status: %d%s", color, fullURL, statusCode, statusColors["reset"])
}

// PrintHelp displays the help message
func PrintHelp() {
 fmt.Println("Usage: go run main.go [OPTIONS]")
 fmt.Println("Options:")
 fmt.Println("  -u <url>     : Target URL (e.g., example.com)")
 fmt.Println("  -l <file>    : File containing directories to check")
 fmt.Println("  -o <file>    : Output file (optional)")
 fmt.Println("  -t <threads> : Number of threads to use (default: 50)")
 fmt.Println("  -h           : Show this help message")
}

func main() {
 // Define flags
 urlPtr := flag.String("u", "", "Target URL (e.g., example.com)")
 listPtr := flag.String("l", "", "File containing directories to check")
 outputPtr := flag.String("o", "", "Output file (optional)")
 threadPtr := flag.Int("t", 50, "Number of threads to use (default: 50)")
 helpPtr := flag.Bool("h", false, "Show help message")
 flag.Parse()

 // Show help if -h is specified or insufficient arguments are provided
 if *helpPtr || len(os.Args) < 3 {
  PrintHelp()
  return
 }

 if *urlPtr == "" || *listPtr == "" {
  fmt.Println("Error: -u (URL) and -l (directory list) are required.")
  PrintHelp()
  return
 }

 // Determine protocol if not specified
 finalURL := CheckProtocol(*urlPtr)

 // Open the directory list file
 file, err := os.Open(*listPtr)
 if err != nil {
  fmt.Printf("Error opening list file: %v\n", err)
  return
 }
 defer file.Close()

 // Prepare for output
 var outputFile *os.File
 if *outputPtr != "" {
  outputFile, err = os.Create(*outputPtr)
  if err != nil {
   fmt.Printf("Error creating output file: %v\n", err)
   return
  }
  defer outputFile.Close()
 }

 // Channel for collecting results
 outputChan := make(chan string, 1000) // Buffered channel for efficiency
 var wg sync.WaitGroup
 semaphore := make(chan struct{}, *threadPtr) // Limit number of concurrent threads

 // Read directories from the file and start checking them
 scanner := bufio.NewScanner(file)
 for scanner.Scan() {
  directory := scanner.Text()
  wg.Add(1)
  semaphore <- struct{}{} // Acquire semaphore
  go func(dir string) {
   defer wg.Done()
   CheckDirectory(finalURL, dir, outputChan, semaphore)
  }(directory)
 }
	// Start a goroutine to collect and print results immediately
 go func() {
  for result := range outputChan {
   if outputFile != nil {
    outputFile.WriteString(result + "\n")
   } else {
    fmt.Println(result) // Print results to terminal immediately
   }
  }
 }()

 // Wait for all checks to complete
 wg.Wait()
 close(outputChan)
}
