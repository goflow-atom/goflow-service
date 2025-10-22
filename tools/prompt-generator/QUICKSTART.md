# Quick Start Guide

Get started with the GoFlow Prompt Generator in 5 minutes!

## Step 1: Install Dependencies

```bash
cd tools/prompt-generator
pip install -r requirements.txt
```

## Step 2: Get Your API Key

1. Visit [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Sign in with your Google account
3. Click "Create API Key"
4. Copy your API key

## Step 3: Set Your API Key

**Linux/Mac:**
```bash
export GEMINI_API_KEY="your-api-key-here"
```

**Windows (PowerShell):**
```powershell
$env:GEMINI_API_KEY="your-api-key-here"
```

**Windows (CMD):**
```cmd
set GEMINI_API_KEY=your-api-key-here
```

## Step 4: Generate Your First Prompt

```bash
python main.py --task-id INIT-001
```

## Step 5: Check the Output

Your generated prompt will be saved in:
- `../../docs/tasks/prompts/INIT-001_latest_prompt.md`

## What's Next?

- Generate prompts for multiple tasks: `python main.py --task-id INIT-001 INIT-002 INIT-003`
- Use verbose mode for debugging: `python main.py --task-id INIT-001 --verbose`
- Create a config file for easier usage: `cp config.example.json config.json`

## Common Issues

**"API key not found"**
- Make sure you've set the `GEMINI_API_KEY` environment variable
- Or use `--api-key YOUR_KEY` flag

**"Template file not found"**
- Make sure you're running from the project root directory
- Or use absolute paths with `--template` and `--roadmap` flags

**"Task not found"**
- Check that the task ID exists in `docs/tasks/01_IMPLEMENTATION_ROADMAP.md`
- Task IDs are case-sensitive (e.g., `INIT-001` not `init-001`)

## Need Help?

Check the full [README.md](README.md) for detailed documentation.

