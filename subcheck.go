package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)


var defaultSubdomains = []string{
	"www", "api", "mail", "blog", "dev", "test", "shop", "ftp", "support",
	"admin", "m", "dashboard", "static", "cms", "staging", "cdn", "secure", "beta",
	"docs", "help", "assets", "media", "portal", "gateway", "office", "download",
	"mobile", "search", "cloud", "services", "pay", "webmail", "smtp", "vpn", "app",
	"auth", "login", "logout", "register", "upload", "images", "backup", "status",
	"chat", "my", "devops", "partners", "analytics", "monitoring", "graph", "reporting",
	"proxy", "archive", "cache", "test1", "test2", "api-v1", "api-v2", "resources",
}

func main() {
	
	fmt.Println("  SSSSS  U   U  BBBBB   CCCC  H   H  EEEEE  CCCC  K   K")
	fmt.Println(" S        U   U  B    B C      H   H  E      C      K  K")
	fmt.Println("  SSS     U   U  BBBBB  C      HHHHH  EEEE   C      KKK")
	fmt.Println("     S    U   U  B    B C      H   H  E      C      K  K")
	fmt.Println(" SSSSS    UUUU   BBBBB   CCCC  H   H  EEEEE  CCCC  K   K")
	fmt.Println()

	
	listFile := flag.String("l", "", "Path to the file containing subdomains list (optional).")
	outputFile := flag.String("o", "", "Path to the output file (optional).")
	baseURL := flag.String("u", "", "Base URL to scan for subdomains (required).")
	flag.Parse()


	if *baseURL == "" {
		fmt.Println("Please provide a URL. Usage: -u https://example.com")
		return
	}

	
	subdomains := defaultSubdomains
	if *listFile != "" {
		var err error
		subdomains, err = readSubdomainsFromFile(*listFile)
		if err != nil {
			fmt.Printf("Error reading subdomains list: %v\n", err)
			return
		}
	}

	
	trimmedURL := strings.TrimPrefix(strings.TrimPrefix(*baseURL, "http://"), "https://")
	parts := strings.Split(trimmedURL, "/")
	domain := parts[0]

	
	var results []string
	for _, subdomain := range subdomains {
		fullURL := fmt.Sprintf("https://%s.%s", subdomain, domain)
		status := checkSubdomainStatus(fullURL)
		results = append(results, status)
	}

	
	if *outputFile != "" {
		err := writeResultsToFile(*outputFile, results)
		if err != nil {
			fmt.Printf("Error writing to output file: %v\n", err)
			return
		}
		fmt.Printf("Results saved to %s\n", *outputFile)
	} else {
		// نمایش نتایج در کنسول
		for _, result := range results {
			fmt.Println(result)
		}
	}
}


func readSubdomainsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var subdomains []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			subdomains = append(subdomains, line)
		}
	}
	return subdomains, scanner.Err()
}


func checkSubdomainStatus(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		
		return fmt.Sprintf("%s: Not Found", url)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// اگر وضعیت 404 برگشت، "Not Found" را نمایش می‌دهیم
		return fmt.Sprintf("%s: Not Found (404)", url)
	}

	
	return colorizeStatus(url, resp.StatusCode, fmt.Sprintf("%d", resp.StatusCode))
}


func colorizeStatus(url string, statusCode int, status string) string {
	var colorCode string
	switch {
	case statusCode >= 200 && statusCode < 300:
		colorCode = "\033[32m"
	case statusCode >= 300 && statusCode < 400:
		colorCode = "\033[34m" 
	case statusCode >= 400 && statusCode < 500:
		colorCode = "\033[35m" 
	case statusCode >= 500 && statusCode < 600:
		colorCode = "\033[31m" 
	default:
		colorCode = "\033[0m" 
	}

	return fmt.Sprintf("%s%s: %s\033[0m", colorCode, url, status)
}


func writeResultsToFile(path string, results []string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, result := range results {
		_, err := writer.WriteString(result + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}
