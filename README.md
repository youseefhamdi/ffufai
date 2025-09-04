# ffufai - AI-Powered Web Fuzzer

A Go implementation of an AI-powered wrapper for [ffuf](https://github.com/ffuf/ffuf) that leverages the Perplexity API to intelligently suggest file extensions based on target URL analysis and HTTP headers.

## 🚀 Features

- **AI-Powered Extension Suggestions**: Uses Perplexity's real-time web search capabilities to suggest relevant file extensions
- **HTTP Header Analysis**: Analyzes target headers to provide context-aware suggestions  
- **Seamless ffuf Integration**: Passes through all ffuf arguments while adding AI-suggested extensions
- **Multiple Models**: Support for different Perplexity models (sonar-pro, sonar-small-online, etc.)
- **Robust Error Handling**: Comprehensive validation and error reporting
- **Signal Handling**: Graceful interruption support
- **Dry Run Mode**: Preview commands before execution
- **Verbose Mode**: Detailed logging for debugging

## 📋 Prerequisites

- Go 1.21 or later
- [ffuf](https://github.com/ffuf/ffuf) installed and accessible in PATH
- Perplexity AI API key

## 🛠️ Installation

### Option 1: Build from Source

```bash
# Clone or download the source files
git clone <repository-url>
cd ffufai

# Build the executable
go build -o ffufai ffufai-improved.go

# Or use the simpler version
go build -o ffufai ffufai.go
```

### Option 2: Direct Run

```bash
# Run directly with Go
go run ffufai-improved.go [options]
```

### Option 3: Install ffuf (if not already installed)

```bash
# Install ffuf
go install github.com/ffuf/ffuf@latest

# Or download from releases
# https://github.com/ffuf/ffuf/releases
```

## 🔑 API Key Setup

1. Visit [Perplexity AI Settings](https://www.perplexity.ai/settings/api)
2. Generate your API key
3. Set the environment variable:

```bash
export PERPLEXITY_API_KEY="your_api_key_here"

# For persistent setup, add to your shell profile:
echo 'export PERPLEXITY_API_KEY="your_api_key_here"' >> ~/.bashrc
source ~/.bashrc
```

## 📖 Usage

### Basic Usage

```bash
./ffufai -u https://example.com/FUZZ -w /path/to/wordlist.txt
```

### Advanced Usage

```bash
# Specify custom ffuf path and more extensions
./ffufai --ffuf-path /custom/path/to/ffuf --max-extensions 6 -u https://example.com/admin/FUZZ -w wordlist.txt -fc 404

# Verbose mode with custom model
./ffufai --verbose --model sonar-small-online -u https://example.com/api/FUZZ -w /usr/share/wordlists/dirb/common.txt

# Dry run to preview command
./ffufai --dry-run -u https://example.com/FUZZ -w wordlist.txt

# Filter out specific status codes and save output
./ffufai -u https://target.com/FUZZ -w wordlist.txt -fc 404,403 -o results.json
```

### Command Line Options

```bash
Usage: ffufai [options] -u URL [ffuf options]

Options:
  -ffuf-path string
        Path to ffuf executable (default "ffuf")
  -max-extensions int
        Maximum number of extensions to suggest (1-10) (default 4)
  -model string
        Perplexity model to use (default "sonar-pro")
  -verbose
        Enable verbose output
  -dry-run
        Show what would be executed without running ffuf
  -version
        Show version information
```

### Common ffuf Options (Passed Through)

- `-w wordlist.txt` - Wordlist to use
- `-fc 404` - Filter out HTTP 404 responses
- `-mc 200` - Match only HTTP 200 responses
- `-o output.json` - Save output to file
- `-t 100` - Number of concurrent threads
- `-p 1-3` - Delay between requests (seconds)
- `-H "Header: Value"` - Custom HTTP header

## 🎯 Use Cases

### Directory Discovery
```bash
./ffufai -u https://example.com/FUZZ -w /usr/share/wordlists/dirb/common.txt
```

### Admin Panel Discovery
```bash
./ffufai -u https://example.com/admin/FUZZ -w admin-wordlist.txt -fc 404,403
```

### API Endpoint Discovery
```bash
./ffufai -u https://example.com/api/FUZZ -w api-endpoints.txt -mc 200
```

### Backup File Discovery
```bash
./ffufai -u https://example.com/FUZZ -w backup-files.txt --max-extensions 8
```

## 🧠 How It Works

1. **URL Analysis**: Parses the target URL and extracts path information
2. **Header Retrieval**: Performs HTTP HEAD request to analyze server headers
3. **AI Processing**: Sends URL and headers to Perplexity API for intelligent analysis
4. **Extension Suggestions**: Receives contextually relevant file extensions
5. **ffuf Execution**: Runs ffuf with AI-suggested extensions plus user arguments

## 🔧 Configuration

### Environment Variables

- `PERPLEXITY_API_KEY` - Your Perplexity API key (required)

### Supported Perplexity Models

- `sonar-pro` (default) - Advanced model with comprehensive search
- `sonar-small-online` - Faster, lighter model
- `sonar-medium-online` - Balanced performance and capability

## 📊 Example AI Suggestions

### For `/admin/FUZZ`:
- Server: Microsoft-IIS/10.0
- Suggested: `.aspx`, `.asp`, `.config`, `.xml`

### For `/api/FUZZ`:
- Content-Type: application/json
- Suggested: `.json`, `.xml`, `.php`, `.py`

### For `/js/FUZZ`:
- Path-based analysis
- Suggested: `.js`, `.min.js`, `.map`, `.ts`

## 🛡️ Security Considerations

- API keys are loaded from environment variables (not hardcoded)
- HTTP requests include proper timeouts
- Input validation prevents command injection
- Graceful error handling for network issues

## 🐛 Troubleshooting

### Common Issues

1. **"PERPLEXITY_API_KEY environment variable not set"**
   ```bash
   export PERPLEXITY_API_KEY="your_key_here"
   ```

2. **"ffuf not found"**
   ```bash
   # Install ffuf or specify custom path
   ./ffufai --ffuf-path /path/to/ffuf -u https://example.com/FUZZ -w wordlist.txt
   ```

3. **"No valid JSON found in AI response"**
   - Check your API key validity
   - Try with `--verbose` flag to see raw AI response
   - Reduce `--max-extensions` number

4. **Network timeouts**
   - Check internet connectivity
   - Verify the target URL is accessible

### Debug Mode

```bash
# Run with verbose output
./ffufai --verbose -u https://example.com/FUZZ -w wordlist.txt

# Use dry-run to test configuration
./ffufai --dry-run -u https://example.com/FUZZ -w wordlist.txt
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- [ffuf](https://github.com/ffuf/ffuf) - The amazing fuzzing tool this wraps
- [Perplexity AI](https://www.perplexity.ai/) - AI-powered search and analysis
- Original Python ffufai - Inspiration for this Go implementation

## 📚 Related Tools

- [ffuf](https://github.com/ffuf/ffuf) - Fast web fuzzer written in Go
- [gobuster](https://github.com/OJ/gobuster) - Directory/File/DNS bruteforcer  
- [dirsearch](https://github.com/maurosoria/dirsearch) - Web path scanner
- [SecLists](https://github.com/danielmiessler/SecLists) - Collection of wordlists

## 🔗 Resources

- [ffuf Wiki](https://github.com/ffuf/ffuf/wiki) - Comprehensive ffuf documentation
- [Perplexity API Docs](https://docs.perplexity.ai/) - Official API documentation
- [Web Security Testing Guide](https://owasp.org/www-project-web-security-testing-guide/) - OWASP testing methodologies