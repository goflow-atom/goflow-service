@echo off
REM Example usage script for GoFlow Prompt Generator (Windows)
REM This script demonstrates various ways to use the prompt generator

echo ==========================================
echo GoFlow Prompt Generator - Example Usage
echo ==========================================
echo.

REM Check if Python is installed
python --version >nul 2>&1
if errorlevel 1 (
    echo Error: Python is not installed
    exit /b 1
)

REM Check if dependencies are installed
python -c "import google.generativeai" >nul 2>&1
if errorlevel 1 (
    echo Installing dependencies...
    pip install -r requirements.txt
)

REM Check if API key is set
if "%GEMINI_API_KEY%"=="" (
    echo Warning: GEMINI_API_KEY environment variable is not set
    echo Please set it with: set GEMINI_API_KEY=your-api-key
    echo.
    set /p api_key="Enter your Gemini API key (or press Enter to skip): "
    if not "!api_key!"=="" (
        set GEMINI_API_KEY=!api_key!
    ) else (
        echo Skipping examples that require API key
        exit /b 0
    )
)

echo.
echo Example 1: Generate prompt for a single task
echo --------------------------------------------
python main.py --task-id INIT-001

echo.
echo Example 2: Generate prompts for multiple tasks
echo -----------------------------------------------
python main.py --task-id INIT-001 INIT-002 INIT-003

echo.
echo Example 3: Generate prompt with verbose logging
echo ------------------------------------------------
python main.py --task-id GIN-001 --verbose

echo.
echo Example 4: Generate prompt with custom output directory
echo --------------------------------------------------------
python main.py --task-id MW-001 --output-dir custom-prompts

echo.
echo ==========================================
echo All examples completed successfully!
echo Check the output directory for generated prompts
echo ==========================================
pause

