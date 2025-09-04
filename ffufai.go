# ffufai - AI-powered ffuf wrapper in Go

This Go program recreates the Python ffufai tool with Perplexity AI API support.

```go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Perplexity API structures
type PerplexityRequest struct {
	Model       string          `json:"model"`
	Messages    []Message       `json:"messages"`
	MaxTokens   int            `json:"max_tokens"`
	Temperature float64        `json:"temperature"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type PerplexityResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ExtensionsResponse struct {
	Extensions []string `json:"extensions"`
}

// Configuration
type Config struct {
	ffufPath      string
	maxExtensions int
	url           string
	ffufArgs      []string
}

// Get API key from environment
func getAPIKey() (string, error) {
	key := os.Getenv("PERPLEXITY_API_KEY")
	if key == "" {
		return "", fmt.Errorf("PERPLEXITY_API_KEY environment variable not set")
	}
	return key, nil
}

// Get HTTP headers for a URL
func getHeaders(urlStr string) map[string]string {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Head(urlStr)
	if err != nil {
		fmt.Printf("Error fetching headers: %v\n", err)
		return map[string]string{"Header": "Error fetching headers."}
	}
	defer resp.Body.Close()
	
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	
	return headers
}

// Get AI-suggested extensions using Perplexity API
func getAIExtensions(urlStr string, headers map[string]string, apiKey string, maxExtensions int) (*ExtensionsResponse, error) {
	// Convert headers to JSON string for the prompt
	headersJSON, err := json.Marshal(headers)
	if err != nil {
		return nil, fmt.Errorf("error marshaling headers: %v", err)
	}
	
	prompt := fmt.Sprintf(`Given the following URL and HTTP headers, suggest the most likely file extensions for fuzzing this endpoint.
Respond with a JSON object containing a list of extensions. The response will be parsed with json.Unmarshal(),
so it must be valid JSON. No preamble or yapping. Use the format: {"extensions": [".ext1", ".ext2", ...]}.
Do not suggest more than %d, but only suggest extensions that make sense. For example, if the path is 
/js/ then don't suggest .css as the extension. Also, if limited, prefer the extensions which are more interesting.
The URL path is great to look at for ideas. For example, if it says presentations, then it's likely there 
are powerpoints or pdfs in there. If the path is /js/ then it's good to use js as an extension.

Examples:
1. URL: https://example.com/presentations/FUZZ
   Headers: {"Content-Type": "application/pdf", "Content-Length": "1234567"}
   JSON Response: {"extensions": [".pdf", ".ppt", ".pptx"]}
2. URL: https://example.com/FUZZ
   Headers: {"Server": "Microsoft-IIS/10.0", "X-Powered-By": "ASP.NET"}
   JSON Response: {"extensions": [".aspx", ".asp", ".exe", ".dll"]}

URL: %s
Headers: %s
JSON Response:`, maxExtensions, urlStr, string(headersJSON))

	// Prepare the Perplexity API request
	reqBody := PerplexityRequest{
		Model:       "sonar-pro",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant that suggests file extensions for fuzzing based on URL and headers.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   1000,
		Temperature: 0.0,
	}

	// Marshal the request body
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.perplexity.ai/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Make the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	// Parse the response
	var perplexityResp PerplexityResponse
	if err := json.NewDecoder(resp.Body).Decode(&perplexityResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	if len(perplexityResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in API response")
	}

	// Parse the extensions from the AI response
	var extensionsResp ExtensionsResponse
	content := perplexityResp.Choices[0].Message.Content
	
	// Try to find JSON in the response
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("no valid JSON found in AI response")
	}
	
	jsonStr := content[start : end+1]
	if err := json.Unmarshal([]byte(jsonStr), &extensionsResp); err != nil {
		return nil, fmt.Errorf("error parsing AI response JSON: %v", err)
	}

	return &extensionsResp, nil
}

// Parse command line arguments
func parseArgs() (*Config, error) {
	config := &Config{}
	
	// Define flags
	flag.StringVar(&config.ffufPath, "ffuf-path", "ffuf", "Path to ffuf executable")
	flag.IntVar(&config.maxExtensions, "max-extensions", 4, "Maximum number of extensions to suggest")
	
	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] -u URL [ffuf options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nffufai - AI-powered ffuf wrapper with Perplexity API\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -u https://example.com/FUZZ -w /path/to/wordlist.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  PERPLEXITY_API_KEY    Perplexity AI API key (required)\n")
	}
	
	// Parse known flags
	flag.Parse()
	
	// Get remaining arguments
	remainingArgs := flag.Args()
	
	// Find -u URL in the remaining arguments
	urlIndex := -1
	for i, arg := range remainingArgs {
		if arg == "-u" && i+1 < len(remainingArgs) {
			urlIndex = i + 1
			config.url = remainingArgs[urlIndex]
			break
		}
	}
	
	if urlIndex == -1 {
		return nil, fmt.Errorf("-u URL argument is required")
	}
	
	// Store all remaining arguments for ffuf (excluding the ones we processed)
	config.ffufArgs = remainingArgs
	
	return config, nil
}

func main() {
	// Parse command line arguments
	config, err := parseArgs()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}
	
	// Validate URL contains FUZZ
	parsedURL, err := url.Parse(config.url)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		os.Exit(1)
	}
	
	pathParts := strings.Split(parsedURL.Path, "/")
	if len(pathParts) == 0 || !strings.Contains(pathParts[len(pathParts)-1], "FUZZ") {
		fmt.Println("Warning: FUZZ keyword is not at the end of the URL path. Extension fuzzing may not work as expected.")
	}
	
	// Get API key
	apiKey, err := getAPIKey()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Please set the PERPLEXITY_API_KEY environment variable.")
		os.Exit(1)
	}
	
	// Get headers from base URL
	baseURL := strings.Replace(config.url, "FUZZ", "", 1)
	headers := getHeaders(baseURL)
	
	// Get AI suggestions for extensions
	fmt.Println("Getting AI suggestions for file extensions...")
	extensionsResp, err := getAIExtensions(config.url, headers, apiKey, config.maxExtensions)
	if err != nil {
		fmt.Printf("Error getting AI extensions: %v\n", err)
		os.Exit(1)
	}
	
	if len(extensionsResp.Extensions) == 0 {
		fmt.Println("No extensions suggested by AI.")
		os.Exit(1)
	}
	
	// Limit extensions to maxExtensions
	extensions := extensionsResp.Extensions
	if len(extensions) > config.maxExtensions {
		extensions = extensions[:config.maxExtensions]
	}
	
	fmt.Printf("AI suggested extensions: %v\n", extensions)
	
	// Prepare ffuf command
	ffufCmd := []string{config.ffufPath}
	ffufCmd = append(ffufCmd, config.ffufArgs...)
	ffufCmd = append(ffufCmd, "-e", strings.Join(extensions, ","))
	
	// Execute ffuf
	fmt.Printf("Running: %s\n", strings.Join(ffufCmd, " "))
	cmd := exec.Command(ffufCmd[0], ffufCmd[1:]...)
	
	// Inherit stdout and stderr so we can see ffuf output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	// Run the command
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error running ffuf: %v\n", err)
		os.Exit(1)
	}
}
```

## Usage

1. **Set up environment:**
```bash
export PERPLEXITY_API_KEY="your_perplexity_api_key_here"
```

2. **Build the program:**
```bash
go build -o ffufai ffufai.go
```

3. **Run with ffuf options:**
```bash
./ffufai -u https://example.com/FUZZ -w /path/to/wordlist.txt
./ffufai --ffuf-path /custom/path/to/ffuf --max-extensions 6 -u https://example.com/admin/FUZZ -w wordlist.txt -fc 404
```

## Key Changes from Python Version

1. **Perplexity API Integration:** Replaced OpenAI/Anthropic with Perplexity API
2. **Go Standard Library:** Used `flag`, `net/http`, `os/exec`, and `encoding/json`
3. **Error Handling:** Go-style error handling with proper error messages
4. **JSON Parsing:** Robust JSON extraction from AI responses
5. **HTTP Client:** Custom timeout and proper header handling
6. **Command Execution:** Direct process execution with inherited I/O

## Features

- ✅ Perplexity API integration with `sonar-pro` model
- ✅ HTTP HEAD request to analyze target headers
- ✅ AI-powered extension suggestions based on URL and headers  
- ✅ Flexible command-line argument parsing
- ✅ Seamless ffuf integration
- ✅ Environment variable configuration
- ✅ Proper error handling and validation
- ✅ Cross-platform compatibility

The tool maintains the same functionality as the original Python version while leveraging Go's performance and Perplexity's real-time web search capabilities for more accurate extension suggestions.