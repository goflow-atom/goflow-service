# GoFlow Prompt Generator

Automated prompt generation tool for GoFlow implementation tasks. This script reads task information from the roadmap, parses the prompt template, and uses Google Gemini AI to generate complete, ready-to-use implementation prompts.

## Features

- ✅ **Template Parsing**: Automatically extracts placeholders from the prompt template
- ✅ **Task Data Extraction**: Reads task information from the implementation roadmap
- ✅ **AI Content Generation**: Uses Google Gemini API to generate context-aware content
- ✅ **Smart Placeholder Replacement**: Fills both simple and complex placeholders
- ✅ **Multiple Task Support**: Generate prompts for multiple tasks in one run
- ✅ **Comprehensive Logging**: Detailed logs for debugging and monitoring
- ✅ **Flexible Configuration**: Support for config files and environment variables
- ✅ **Error Handling**: Robust error handling with informative messages

## Prerequisites

- Python 3.8 or higher
- Google Gemini API key ([Get one here](https://makersuite.google.com/app/apikey))

## Installation

1. **Navigate to the tool directory**:
   ```bash
   cd tools/prompt-generator
   ```

2. **Install dependencies**:
   ```bash
   pip install -r requirements.txt
   ```

3. **Set up your API key** (choose one method):

   **Option A: Environment Variable** (Recommended)
   ```bash
   # Linux/Mac
   export GEMINI_API_KEY="your-api-key-here"

   # Windows (PowerShell)
   $env:GEMINI_API_KEY="your-api-key-here"

   # Windows (CMD)
   set GEMINI_API_KEY=your-api-key-here
   ```

   **Option B: Configuration File**
   ```bash
   cp config.example.json config.json
   # Edit config.json and add your API key
   ```

## Usage

### Basic Usage

Generate a prompt for a single task:
```bash
python main.py --task-id INIT-001
```

### Multiple Tasks

Generate prompts for multiple tasks:
```bash
python main.py --task-id INIT-001 INIT-002 INIT-003
```

### With Explicit API Key

```bash
python main.py --task-id INIT-001 --api-key YOUR_API_KEY
```

### With Custom Configuration

```bash
python main.py --task-id INIT-001 --config config.json
```

### With Custom Paths

```bash
python main.py --task-id INIT-001 \
  --template docs/tasks/03_PROMPT_IMPLEMENTATION.md \
  --roadmap docs/tasks/01_IMPLEMENTATION_ROADMAP.md \
  --output-dir docs/tasks/prompts
```

### Verbose Mode

Enable detailed logging:
```bash
python main.py --task-id INIT-001 --verbose
```

## Command Line Options

| Option | Description | Required |
|--------|-------------|----------|
| `--task-id` | Task ID(s) to generate prompts for | Yes |
| `--api-key` | Google Gemini API key | No* |
| `--config` | Path to configuration JSON file | No |
| `--template` | Path to template file | No |
| `--roadmap` | Path to roadmap file | No |
| `--output-dir` | Output directory for prompts | No |
| `--verbose` | Enable verbose logging | No |

*Required if not set via environment variable or config file

## Configuration

### Configuration File Format

Create a `config.json` file with the following structure:

```json
{
  "api_key": "YOUR_GEMINI_API_KEY",
  "model_name": "gemini-1.5-pro",
  "template_path": "docs/tasks/03_PROMPT_IMPLEMENTATION.md",
  "roadmap_path": "docs/tasks/01_IMPLEMENTATION_ROADMAP.md",
  "output_dir": "docs/tasks/prompts",
  "temperature": 0.7,
  "max_tokens": 8000,
  "timeout": 60
}
```

## Output

Generated prompts are saved in two formats:

1. **Timestamped version**: `{TASK_ID}_{TIMESTAMP}_prompt.md`
2. **Latest version**: `{TASK_ID}_latest_prompt.md`

## How It Works

1. **Template Parsing**: Reads the prompt template and identifies all placeholders
2. **Task Extraction**: Parses the roadmap file to extract task information
3. **Basic Replacement**: Replaces simple placeholders with extracted task data
4. **AI Content Generation**: Generates context-aware content using Gemini API
5. **File Generation**: Saves the complete prompt to the output directory

## Troubleshooting

### Error: "API key not found"
Set the `GEMINI_API_KEY` environment variable or provide it via `--api-key`.

### Error: "Template file not found"
Ensure you're running the script from the project root or provide the full path.

### Error: "Task not found in roadmap"
Verify the task ID exists in the roadmap file. Task IDs are case-sensitive.

## Architecture

- **`PromptGeneratorConfig`**: Configuration management
- **`TemplateParser`**: Template parsing and placeholder extraction
- **`TaskDataExtractor`**: Roadmap parsing and task extraction
- **`GeminiContentGenerator`**: AI content generation
- **`PromptGenerator`**: Main orchestrator

## Version History

- **v1.0.0** (2025-01-22): Initial release
