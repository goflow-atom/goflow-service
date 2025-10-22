#!/usr/bin/env python3
"""
Setup Test Script for GoFlow Prompt Generator

This script verifies that all dependencies and configurations are correct
before running the main prompt generator.

Usage:
    python test_setup.py
"""

import sys
import os
from pathlib import Path


def print_header(text):
    """Print a formatted header"""
    print("\n" + "=" * 60)
    print(f"  {text}")
    print("=" * 60)


def print_status(check_name, passed, message=""):
    """Print status of a check"""
    status = "✅ PASS" if passed else "❌ FAIL"
    print(f"{status} - {check_name}")
    if message:
        print(f"       {message}")


def check_python_version():
    """Check if Python version is 3.8 or higher"""
    version = sys.version_info
    passed = version.major == 3 and version.minor >= 8
    message = f"Python {version.major}.{version.minor}.{version.micro}"
    if not passed:
        message += " (Required: Python 3.8+)"
    print_status("Python Version", passed, message)
    return passed


def check_dependencies():
    """Check if required dependencies are installed"""
    dependencies = {
        'google.generativeai': 'google-generativeai',
    }
    
    all_passed = True
    for module, package in dependencies.items():
        try:
            __import__(module)
            print_status(f"Dependency: {package}", True, "Installed")
        except ImportError:
            print_status(f"Dependency: {package}", False, 
                        f"Not installed. Run: pip install {package}")
            all_passed = False
    
    return all_passed


def check_api_key():
    """Check if API key is configured"""
    api_key = os.environ.get('GEMINI_API_KEY', '')
    
    if api_key:
        masked_key = api_key[:8] + "..." + api_key[-4:] if len(api_key) > 12 else "***"
        print_status("API Key (Environment)", True, f"Set: {masked_key}")
        return True
    
    # Check if config file exists
    config_path = Path('config.json')
    if config_path.exists():
        try:
            import json
            with open(config_path, 'r') as f:
                config = json.load(f)
                if config.get('api_key'):
                    print_status("API Key (Config File)", True, "Set in config.json")
                    return True
        except Exception as e:
            print_status("API Key (Config File)", False, f"Error reading config: {e}")
    
    print_status("API Key", False, 
                "Not found. Set GEMINI_API_KEY env var or create config.json")
    return False


def check_template_file():
    """Check if template file exists"""
    template_path = Path('../../docs/tasks/03_PROMPT_IMPLEMENTATION.md')
    
    if template_path.exists():
        size = template_path.stat().st_size
        print_status("Template File", True, f"Found ({size} bytes)")
        return True
    else:
        print_status("Template File", False, 
                    f"Not found at: {template_path}")
        return False


def check_roadmap_file():
    """Check if roadmap file exists"""
    roadmap_path = Path('../../docs/tasks/01_IMPLEMENTATION_ROADMAP.md')
    
    if roadmap_path.exists():
        size = roadmap_path.stat().st_size
        print_status("Roadmap File", True, f"Found ({size} bytes)")
        return True
    else:
        print_status("Roadmap File", False, 
                    f"Not found at: {roadmap_path}")
        return False


def check_output_directory():
    """Check if output directory exists or can be created"""
    output_dir = Path('../../docs/tasks/prompts')
    
    if output_dir.exists():
        print_status("Output Directory", True, f"Exists: {output_dir}")
        return True
    else:
        try:
            output_dir.mkdir(parents=True, exist_ok=True)
            print_status("Output Directory", True, f"Created: {output_dir}")
            return True
        except Exception as e:
            print_status("Output Directory", False, 
                        f"Cannot create: {e}")
            return False


def check_write_permissions():
    """Check if we have write permissions in output directory"""
    output_dir = Path('../../docs/tasks/prompts')
    test_file = output_dir / '.test_write_permission'
    
    try:
        with open(test_file, 'w') as f:
            f.write('test')
        test_file.unlink()
        print_status("Write Permissions", True, "Can write to output directory")
        return True
    except Exception as e:
        print_status("Write Permissions", False, 
                    f"Cannot write to output directory: {e}")
        return False


def test_api_connection():
    """Test connection to Gemini API"""
    try:
        import google.generativeai as genai
        
        api_key = os.environ.get('GEMINI_API_KEY', '')
        if not api_key:
            # Try to load from config
            config_path = Path('config.json')
            if config_path.exists():
                import json
                with open(config_path, 'r') as f:
                    config = json.load(f)
                    api_key = config.get('api_key', '')
        
        if not api_key:
            print_status("API Connection", False, "No API key available to test")
            return False
        
        genai.configure(api_key=api_key)
        model = genai.GenerativeModel('gemini-1.5-pro')
        
        # Try a simple generation
        response = model.generate_content("Say 'test successful' if you can read this.")
        
        if response and response.text:
            print_status("API Connection", True, "Successfully connected to Gemini API")
            return True
        else:
            print_status("API Connection", False, "API responded but no text returned")
            return False
            
    except Exception as e:
        print_status("API Connection", False, f"Error: {str(e)}")
        return False


def main():
    """Run all setup checks"""
    print_header("GoFlow Prompt Generator - Setup Test")
    
    print("\nRunning setup checks...\n")
    
    checks = [
        ("Python Version", check_python_version),
        ("Dependencies", check_dependencies),
        ("API Key", check_api_key),
        ("Template File", check_template_file),
        ("Roadmap File", check_roadmap_file),
        ("Output Directory", check_output_directory),
        ("Write Permissions", check_write_permissions),
    ]
    
    results = {}
    for name, check_func in checks:
        results[name] = check_func()
    
    # Optional API connection test
    print("\n" + "-" * 60)
    print("Optional: Testing API Connection (may take a few seconds)...")
    print("-" * 60)
    results["API Connection"] = test_api_connection()
    
    # Summary
    print_header("Summary")
    
    passed = sum(1 for v in results.values() if v)
    total = len(results)
    
    print(f"\nPassed: {passed}/{total} checks")
    
    if passed == total:
        print("\n✅ All checks passed! You're ready to use the prompt generator.")
        print("\nNext steps:")
        print("  1. Run: python main.py --task-id INIT-001")
        print("  2. Check output in: docs/tasks/prompts/")
        return 0
    else:
        print("\n❌ Some checks failed. Please fix the issues above.")
        print("\nCommon fixes:")
        print("  - Install dependencies: pip install -r requirements.txt")
        print("  - Set API key: export GEMINI_API_KEY='your-key'")
        print("  - Run from project root directory")
        return 1


if __name__ == '__main__':
    sys.exit(main())

