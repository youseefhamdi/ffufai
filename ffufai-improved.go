#!/usr/bin/env go run
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
)

const (
	Version        = "1.0.0"
	PerplexityURL  = "https://api.perplexity.ai/chat/completions"
	DefaultModel   = "sonar-pro"
	RequestTimeout = 30 * time.Second
	HeaderTimeout  = 10 * time.Second
)

// Perplexity API structures
type PerplexityRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type PerplexityResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ExtensionsResponse struct {
	Extensions []string `json:"extensions"`
}

// Configuration
type Config struct {
	FfufPath      string
	MaxExtensions int
	URL           string
	FfufArgs      []string
	Model         string
	Verbose       bool
	DryRun        bool
}

// Get API key from environment
func getAPIKey() (string, error) {
	key := os.Getenv("PERPLEXITY_API_KEY")
	if key == "" {
		return "", fmt.Errorf("PERPLEXITY_API_KEY environment variable not set")
	}
	return key, nil
}

// Get HTTP headers for a URL with proper timeout and context
func getHeaders(ctx context.Context, urlStr string) (map[string]string, error) {
	client := &http.Client{
		Timeout: HeaderTimeout,
	}
	
	req, err := http.NewRequestWithContext(ctx, "HEAD", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("creating HEAD request: %w", err)
	}
	
	// Set a common User-Agent to avoid blocking
	req.Header.Set("User-Agent", "ffufai/"+Version)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing HEAD request: %w", err)
	}
	defer resp.Body.Close()
	
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	
	// Add response status for context
	headers["Status-Code"] = resp.Status
	
	return headers, nil
}

// Get AI-suggested extensions using Perplexity API
func getAIExtensions(ctx context.Context, urlStr string, headers map[string]string, apiKey string, config *Config) (*ExtensionsResponse, error) {
	// Convert headers to JSON string for the prompt
	headersJSON, err := json.MarshalIndent(headers, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling headers: %w", err)
	}
	
	prompt := fmt.Sprintf(`Given the following URL and HTTP headers, suggest the most likely file extensions for fuzzing this endpoint.
Respond with a JSON object containing a list of extensions. The response will be parsed with json.Unmarshal(),
so it must be valid JSON. No preamble or explanation needed. Use the format: {"extensions": [".ext1", ".ext2", ...]}.

Guidelines:
- Suggest up to %d extensions maximum
- Only suggest extensions that make logical sense for this URL path and headers  
- If the path contains specific technology indicators (like /js/, /css/, /api/, /admin/), prioritize related extensions
- Consider the Server header and other technology indicators in headers
- Prefer commonly exploited file types if the path suggests admin/config areas
- For generic paths, suggest a mix of web technologies (.php, .html, .js, .css, .txt, .xml, .json)

Examples:
1. URL: https://example.com/presentations/FUZZ
   Headers: {"Content-Type": "application/pdf", "Server": "Apache"}
   Response: {"extensions": [".pdf", ".ppt", ".pptx", ".doc"]}

2. URL: https://example.com/admin/FUZZ  
   Headers: {"Server": "Microsoft-IIS/10.0", "X-Powered-By": "ASP.NET"}
   Response: {"extensions": [".aspx", ".asp", ".config", ".xml"]}

3. URL: https://example.com/api/FUZZ
   Headers: {"Content-Type": "application/json", "Server": "nginx"}
   Response: {"extensions": [".json", ".xml", ".php", ".py"]}

URL: %s
Headers: %s

Response:`, config.MaxExtensions, urlStr, string(headersJSON))

	// Prepare the Perplexity API request
	reqBody := PerplexityRequest{
		Model: config.Model,
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a cybersecurity expert that suggests file extensions for web application fuzzing. You respond only with valid JSON containing an extensions array.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   500,
		Temperature: 0.1, // Low temperature for consistent results
	}

	// Marshal the request body
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling API request: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", PerplexityURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating API request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("User-Agent", "ffufai/"+Version)

	// Make the request with timeout
	client := &http.Client{
		Timeout: RequestTimeout,
	}
	
	if config.Verbose {
		fmt.Printf("Making Perplexity API request...\n")
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing API request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Parse the response
	var perplexityResp PerplexityResponse
	if err := json.NewDecoder(resp.Body).Decode(&perplexityResp); err != nil {
		return nil, fmt.Errorf("parsing API response: %w", err)
	}

	if len(perplexityResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in API response")
	}

	content := perplexityResp.Choices[0].Message.Content
	
	if config.Verbose {
		fmt.Printf("AI Response: %s\n", content)
	}

	// Extract JSON from the response using regex
	jsonRegex := regexp.MustCompile(`\{[^{}]*"extensions"\s*:\s*\[[^\]]*\][^{}]*\}`)
	matches := jsonRegex.FindAllString(content, -1)
	
	if len(matches) == 0 {
		return nil, fmt.Errorf("no valid JSON found in AI response")
	}
	
	// Try to parse the first match
	var extensionsResp ExtensionsResponse
	if err := json.Unmarshal([]byte(matches[0]), &extensionsResp); err != nil {
		return nil, fmt.Errorf("parsing AI response JSON: %w", err)
	}

	// Validate and clean extensions
	var validExtensions []string
	for _, ext := range extensionsResp.Extensions {
		// Ensure extension starts with dot
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		// Basic validation: only alphanumeric and common symbols
		if matched, _ := regexp.MatchString(`^\.[a-zA-Z0-9]+$`, ext); matched {
			validExtensions = append(validExtensions, ext)
		}
	}
	
	extensionsResp.Extensions = validExtensions
	return &extensionsResp, nil
}

// Parse command line arguments with better error handling
func parseArgs() (*Config, error) {
	config := &Config{
		Model: DefaultModel,
	}
	
	// Define flags
	flag.StringVar(&config.FfufPath, "ffuf-path", "ffuf", "Path to ffuf executable")
	flag.IntVar(&config.MaxExtensions, "max-extensions", 4, "Maximum number of extensions to suggest (1-10)")
	flag.StringVar(&config.Model, "model", DefaultModel, "Perplexity model to use")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Show what would be executed without running ffuf")
	
	// Version flag
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	
	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "ffufai v%s - AI-powered ffuf wrapper with Perplexity API\n\n", Version)
		fmt.Fprintf(os.Stderr, "Usage: %s [options] -u URL [ffuf options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -u https://example.com/FUZZ -w /path/to/wordlist.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --verbose --max-extensions 6 -u https://example.com/admin/FUZZ -w wordlist.txt -fc 404\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --dry-run -u https://example.com/api/FUZZ -w wordlist.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  PERPLEXITY_API_KEY    Perplexity AI API key (required)\n\n")
		fmt.Fprintf(os.Stderr, "Note: All ffuf options can be passed after the URL.\n")
	}
	
	// Parse known flags
	flag.Parse()
	
	if showVersion {
		fmt.Printf("ffufai version %s\n", Version)
		os.Exit(0)
	}
	
	// Validate max extensions
	if config.MaxExtensions < 1 || config.MaxExtensions > 10 {
		return nil, fmt.Errorf("max-extensions must be between 1 and 10")
	}
	
	// Get remaining arguments
	remainingArgs := flag.Args()
	
	// Find -u URL in the remaining arguments
	urlIndex := -1
	for i, arg := range remainingArgs {
		if arg == "-u" && i+1 < len(remainingArgs) {
			urlIndex = i + 1
			config.URL = remainingArgs[urlIndex]
			break
		}
	}
	
	if urlIndex == -1 {
		return nil, fmt.Errorf("-u URL argument is required")
	}
	
	// Store all remaining arguments for ffuf
	config.FfufArgs = remainingArgs
	
	return config, nil
}

// Validate URL and provide helpful warnings
func validateURL(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}
	
	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must include scheme (http:// or https://)")
	}
	
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include hostname")
	}
	
	if !strings.Contains(urlStr, "FUZZ") {
		return fmt.Errorf("URL must contain the FUZZ keyword")
	}
	
	// Check if FUZZ is at the end of path for extension fuzzing
	pathParts := strings.Split(parsedURL.Path, "/")
	if len(pathParts) == 0 || !strings.Contains(pathParts[len(pathParts)-1], "FUZZ") {
		fmt.Fprintf(os.Stderr, "Warning: FUZZ keyword is not at the end of the URL path. Extension fuzzing may not work as expected.\n")
	}
	
	return nil
}

// Execute ffuf with proper signal handling
func executeFfuf(config *Config, extensions []string) error {
	// Prepare ffuf command
	ffufCmd := []string{config.FfufPath}
	ffufCmd = append(ffufCmd, config.FfufArgs...)
	ffufCmd = append(ffufCmd, "-e", strings.Join(extensions, ","))
	
	if config.DryRun {
		fmt.Printf("Would execute: %s\n", strings.Join(ffufCmd, " "))
		return nil
	}
	
	fmt.Printf("Executing: %s\n", strings.Join(ffufCmd, " "))
	
	// Create command with context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	cmd := exec.CommandContext(ctx, ffufCmd[0], ffufCmd[1:]...)
	
	// Inherit stdout and stderr so we can see ffuf output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	// Handle interruption signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		fmt.Fprintf(os.Stderr, "\nReceived interrupt signal, stopping ffuf...\n")
		cancel()
	}()
	
	// Run the command
	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.Canceled {
			return fmt.Errorf("ffuf was interrupted")
		}
		return fmt.Errorf("ffuf execution failed: %w", err)
	}
	
	return nil
}

func main() {
	// Parse command line arguments
	config, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		flag.Usage()
		os.Exit(1)
	}
	
	// Validate URL
	if err := validateURL(config.URL); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	// Get API key
	apiKey, err := getAPIKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please set the PERPLEXITY_API_KEY environment variable.\n")
		fmt.Fprintf(os.Stderr, "Get your API key from: https://www.perplexity.ai/settings/api\n")
		os.Exit(1)
	}
	
	// Create context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	// Get headers from base URL
	baseURL := strings.Replace(config.URL, "FUZZ", "", 1)
	
	if config.Verbose {
		fmt.Printf("Analyzing target: %s\n", baseURL)
	}
	
	headers, err := getHeaders(ctx, baseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not fetch headers from %s: %v\n", baseURL, err)
		headers = map[string]string{"Header": "Error fetching headers"}
	} else if config.Verbose {
		fmt.Printf("Retrieved %d headers\n", len(headers))
	}
	
	// Get AI suggestions for extensions
	fmt.Printf("Getting AI suggestions for file extensions...\n")
	extensionsResp, err := getAIExtensions(ctx, config.URL, headers, apiKey, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting AI extensions: %v\n", err)
		os.Exit(1)
	}
	
	if len(extensionsResp.Extensions) == 0 {
		fmt.Printf("No extensions suggested by AI.\n")
		os.Exit(1)
	}
	
	// Limit extensions to maxExtensions
	extensions := extensionsResp.Extensions
	if len(extensions) > config.MaxExtensions {
		extensions = extensions[:config.MaxExtensions]
	}
	
	fmt.Printf("AI suggested extensions: %v\n", extensions)
	
	// Execute ffuf
	if err := executeFfuf(config, extensions); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	if config.Verbose {
		fmt.Printf("ffufai completed successfully\n")
	}
}