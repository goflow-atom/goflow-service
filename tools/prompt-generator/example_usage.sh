#!/bin/bash
# Example usage script for GoFlow Prompt Generator
# This script demonstrates various ways to use the prompt generator

set -e  # Exit on error

echo "=========================================="
echo "GoFlow Prompt Generator - Example Usage"
echo "=========================================="
echo ""

# Check if Python is installed
if ! command -v python3 &> /dev/null; then
    echo "Error: Python 3 is not installed"
    exit 1
fi

# Check if dependencies are installed
if ! python3 -c "import google.generativeai" 2>/dev/null; then
    echo "Installing dependencies..."
    pip install -r requirements.txt
fi

# Check if API key is set
if [ -z "$GEMINI_API_KEY" ]; then
    echo "Warning: GEMINI_API_KEY environment variable is not set"
    echo "Please set it with: export GEMINI_API_KEY='your-api-key'"
    echo ""
    read -p "Enter your Gemini API key (or press Enter to skip): " api_key
    if [ -n "$api_key" ]; then
        export GEMINI_API_KEY="$api_key"
    else
        echo "Skipping examples that require API key"
        exit 0
    fi
fi

echo ""
echo "Example 1: Generate prompt for a single task"
echo "--------------------------------------------"
python3 main.py --task-id INIT-001

echo ""
echo "Example 2: Generate prompts for multiple tasks"
echo "-----------------------------------------------"
python3 main.py --task-id INIT-001 INIT-002 INIT-003

echo ""
echo "Example 3: Generate prompt with verbose logging"
echo "------------------------------------------------"
python3 main.py --task-id GIN-001 --verbose

echo ""
echo "Example 4: Generate prompt with custom output directory"
echo "--------------------------------------------------------"
python3 main.py --task-id MW-001 --output-dir custom-prompts

echo ""
echo "=========================================="
echo "All examples completed successfully!"
echo "Check the output directory for generated prompts"
echo "=========================================="

