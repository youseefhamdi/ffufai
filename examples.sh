#!/bin/bash

# ffufai examples and demonstration script
# This script shows various usage patterns for ffufai

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

echo "🎯 ffufai Usage Examples"
echo "========================"
echo

# Check if ffufai is available
if ! command -v ffufai &> /dev/null; then
    if [ -f "./ffufai" ]; then
        FFUFAI="./ffufai"
        print_info "Using local ffufai binary"
    elif [ -f "./ffufai-improved.go" ]; then
        FFUFAI="go run ./ffufai-improved.go"
        print_info "Using Go source directly"
    else
        print_error "ffufai not found. Please install it first."
        exit 1
    fi
else
    FFUFAI="ffufai"
    print_info "Using installed ffufai"
fi

# Check for API key
if [ -z "${PERPLEXITY_API_KEY}" ]; then
    print_error "PERPLEXITY_API_KEY environment variable is not set"
    echo "Please set it with: export PERPLEXITY_API_KEY=\"your_api_key_here\""
    exit 1
fi

print_success "Perplexity API key found"
echo

# Example URLs for demonstration (using httpbin.org as safe target)
TARGET_BASE="https://httpbin.org"
SAFE_TARGET="${TARGET_BASE}/status/200"

echo "🔍 Example 1: Basic Directory Fuzzing (Dry Run)"
echo "================================================"
echo "Command: $FFUFAI --dry-run -u ${TARGET_BASE}/FUZZ -w /dev/null"
$FFUFAI --dry-run -u "${TARGET_BASE}/FUZZ" -w /dev/null 2>/dev/null || true
echo

echo "🔍 Example 2: Verbose Mode with Custom Extensions"
echo "=================================================="
echo "Command: $FFUFAI --verbose --max-extensions 6 --dry-run -u ${TARGET_BASE}/admin/FUZZ -w /dev/null"
$FFUFAI --verbose --max-extensions 6 --dry-run -u "${TARGET_BASE}/admin/FUZZ" -w /dev/null 2>/dev/null || true
echo

echo "🔍 Example 3: API Endpoint Discovery"
echo "===================================="
echo "Command: $FFUFAI --dry-run -u ${TARGET_BASE}/api/FUZZ -w /dev/null"
$FFUFAI --dry-run -u "${TARGET_BASE}/api/FUZZ" -w /dev/null 2>/dev/null || true
echo

echo "🔍 Example 4: Custom Model Usage"
echo "================================="
echo "Command: $FFUFAI --model sonar-small-online --dry-run -u ${TARGET_BASE}/js/FUZZ -w /dev/null"
$FFUFAI --model sonar-small-online --dry-run -u "${TARGET_BASE}/js/FUZZ" -w /dev/null 2>/dev/null || true
echo

# Create a temporary wordlist for demonstration
TEMP_WORDLIST=$(mktemp)
echo -e "index\nadmin\ntest\napi\nconfig" > "$TEMP_WORDLIST"

echo "🔍 Example 5: Real Fuzzing with Small Wordlist"
echo "==============================================="
echo "Command: $FFUFAI -u ${TARGET_BASE}/FUZZ -w $TEMP_WORDLIST -fc 404"
print_warning "This will make real HTTP requests to httpbin.org (safe testing site)"
read -p "Press Enter to continue or Ctrl+C to cancel..."

$FFUFAI -u "${TARGET_BASE}/FUZZ" -w "$TEMP_WORDLIST" -fc 404 2>/dev/null || true
echo

# Cleanup
rm -f "$TEMP_WORDLIST"

echo "📚 Additional Examples for Real-World Usage"
echo "==========================================="
echo
echo "1. Directory Bruteforcing:"
echo "   ffufai -u https://target.com/FUZZ -w /usr/share/wordlists/dirb/common.txt"
echo
echo "2. Admin Panel Discovery:"
echo "   ffufai -u https://target.com/admin/FUZZ -w admin-wordlist.txt -fc 404,403"
echo
echo "3. Backup File Discovery:"
echo "   ffufai -u https://target.com/FUZZ -w backup-files.txt --max-extensions 8"
echo
echo "4. API Endpoint Discovery:"
echo "   ffufai -u https://target.com/api/v1/FUZZ -w api-endpoints.txt -mc 200 -o results.json"
echo
echo "5. Subdomain Discovery (requires DNS setup):"
echo "   ffufai -u https://FUZZ.target.com -w subdomains.txt -H 'Host: FUZZ.target.com'"
echo
echo "6. Parameter Discovery:"
echo "   ffufai -u 'https://target.com/search?q=test&FUZZ=1' -w params.txt -fs 1234"
echo
echo "7. Extension-specific Discovery:"
echo "   ffufai -u https://target.com/uploads/FUZZ -w filenames.txt --max-extensions 10"
echo

print_info "For more examples and detailed documentation, see README.md"
print_info "Always ensure you have permission to test target websites"

echo
echo "🎉 Demo completed successfully!"