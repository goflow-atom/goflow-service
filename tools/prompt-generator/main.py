#!/usr/bin/env python3
"""
GoFlow Prompt Generator

This script automates the generation of AI implementation prompts by:
1. Reading and parsing the template file
2. Extracting task information from the roadmap
3. Using Google Gemini API to generate content for placeholders
4. Creating fully populated prompt files ready for code generation

Usage:
    python main.py --task-id INIT-001 --api-key YOUR_API_KEY
    python main.py --task-id INIT-001 --config config.json
"""

import os
import sys
import re
import json
import logging
import argparse
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Tuple, Any
from dataclasses import dataclass, asdict

try:
    import google.generativeai as genai  # type: ignore
except ImportError:
    print("Error: google-generativeai package not installed.")
    print("Install it with: pip install google-generativeai")
    sys.exit(1)


# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('prompt_generator.log'),
        logging.StreamHandler(sys.stdout)
    ]
)
logger = logging.getLogger(__name__)


@dataclass
class TaskInfo:
    """Data class to hold task information extracted from roadmap"""
    task_id: str
    component_name: str
    description: str
    priority: str
    status: str
    dependencies: str
    phase_number: str
    phase_name: str
    acceptance_criteria: List[str]

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization"""
        return asdict(self)


class PromptGeneratorConfig:
    """Configuration manager for the prompt generator"""

    def __init__(self, config_path: Optional[str] = None):
        """Initialize configuration from file or environment variables"""
        self.config = self._load_config(config_path)
        self._validate_config()

    def _load_config(self, config_path: Optional[str]) -> Dict[str, Any]:
        """Load configuration from file or use defaults"""
        # Determine the repository root (2 levels up from this script)
        script_dir = Path(__file__).parent.resolve()
        repo_root = script_dir.parent.parent

        default_config: Dict[str, Any] = {
            'api_key': os.environ.get('GEMINI_API_KEY', ''),
            'model_name': 'gemini-2.5-flash',
            'template_path': str(repo_root / 'docs' / 'tasks' / '03_PROMPT_IMPLEMENTATION.md'),
            'roadmap_path': str(repo_root / 'docs' / 'tasks' / '01_IMPLEMENTATION_ROADMAP.md'),
            'output_dir': str(repo_root / 'docs' / 'tasks' / 'prompts'),
            'temperature': 0.7,
            'max_tokens': 8000,
            'timeout': 60
        }

        if config_path and os.path.exists(config_path):
            try:
                with open(config_path, 'r', encoding='utf-8') as f:
                    file_config = json.load(f)
                    default_config.update(file_config)
                    logger.info(f"Loaded configuration from {config_path}")
            except Exception as e:
                logger.warning(f"Failed to load config file: {e}. Using defaults.")

        return default_config

    def _validate_config(self) -> None:
        """Validate required configuration values"""
        if not self.config.get('api_key'):
            raise ValueError(
                "API key not found. Set GEMINI_API_KEY environment variable "
                "or provide it in config file."
            )

        # Validate paths exist
        template_path = str(self.config['template_path'])
        roadmap_path = str(self.config['roadmap_path'])

        if not os.path.exists(template_path):
            raise FileNotFoundError(f"Template file not found: {template_path}")

        if not os.path.exists(roadmap_path):
            raise FileNotFoundError(f"Roadmap file not found: {roadmap_path}")

    def validate_config(self) -> None:
        """Public method to validate configuration"""
        self._validate_config()

    def get(self, key: str, default: Any = None) -> Any:
        """Get configuration value"""
        return self.config.get(key, default)


class TemplateParser:
    """Parser for extracting placeholders from template file"""

    # Regex pattern to match {{PLACEHOLDER}} format
    PLACEHOLDER_PATTERN = re.compile(r'\{\{([A-Z_]+)\}\}')

    def __init__(self, template_path: str):
        """Initialize parser with template file path"""
        self.template_path = template_path
        self.template_content = self._load_template()
        self.placeholders = self._extract_placeholders()
        logger.info(f"Found {len(self.placeholders)} unique placeholders in template")

    def _load_template(self) -> str:
        """Load template file content"""
        try:
            with open(self.template_path, 'r', encoding='utf-8') as f:
                content = f.read()
            logger.info(f"Loaded template from {self.template_path}")
            return content
        except Exception as e:
            logger.error(f"Failed to load template: {e}")
            raise

    def _extract_placeholders(self) -> List[str]:
        """Extract all unique placeholders from template"""
        matches = self.PLACEHOLDER_PATTERN.findall(self.template_content)
        unique_placeholders = sorted(set(matches))
        return unique_placeholders


    def get_template_section(self, start_marker: str, end_marker: str) -> str:
        """Extract a specific section from the template"""
        try:
            start_idx = self.template_content.find(start_marker)
            end_idx = self.template_content.find(end_marker, start_idx)

            if start_idx == -1 or end_idx == -1:
                return ""

            return self.template_content[start_idx:end_idx]
        except Exception as e:
            logger.warning(f"Failed to extract section: {e}")
            return ""


class TaskDataExtractor:
    """Extractor for task information from roadmap file"""

    def __init__(self, roadmap_path: str):
        """Initialize extractor with roadmap file path"""
        self.roadmap_path = roadmap_path
        self.roadmap_content = self._load_roadmap()

    def _load_roadmap(self) -> str:
        """Load roadmap file content"""
        try:
            with open(self.roadmap_path, 'r', encoding='utf-8') as f:
                content = f.read()
            logger.info(f"Loaded roadmap from {self.roadmap_path}")
            return content
        except Exception as e:
            logger.error(f"Failed to load roadmap: {e}")
            raise

    def extract_task_info(self, task_id: str) -> Optional[TaskInfo]:
        """Extract task information for given task ID"""
        try:
            # Find the task line in the roadmap
            task_pattern = re.compile(
                rf'\|\s*{re.escape(task_id)}\s*\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|'
                rf'\s*([^|]+)\s*\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|\s*([^|]+)\s*\|'
            )

            match = task_pattern.search(self.roadmap_content)
            if not match:
                logger.error(f"Task {task_id} not found in roadmap")
                return None

            # Extract basic task information
            component_name = match.group(1).strip()
            description = match.group(2).strip()
            priority = match.group(3).strip()
            status = match.group(4).strip()
            dependencies = match.group(5).strip()

            # Extract phase information
            phase_info = self._extract_phase_info(task_id)

            # Extract acceptance criteria
            acceptance_criteria = self._extract_acceptance_criteria(task_id)

            task_info = TaskInfo(
                task_id=task_id,
                component_name=component_name,
                description=description,
                priority=priority,
                status=status,
                dependencies=dependencies,
                phase_number=phase_info[0],
                phase_name=phase_info[1],
                acceptance_criteria=acceptance_criteria
            )

            logger.info(f"Extracted task info for {task_id}: {component_name}")
            return task_info

        except Exception as e:
            logger.error(f"Failed to extract task info: {e}")
            return None

    def _extract_phase_info(self, task_id: str) -> Tuple[str, str]:
        """Extract phase number and name for the task"""
        # Find the phase section containing this task
        phase_pattern = re.compile(
            r'##\s+Phase\s+(\d+(?:\.\d+)?):?\s+([^\n]+)',
            re.IGNORECASE
        )

        task_pos = self.roadmap_content.find(task_id)
        if task_pos == -1:
            return ("Unknown", "Unknown Phase")

        # Search backwards for the phase header
        content_before = self.roadmap_content[:task_pos]
        matches = list(phase_pattern.finditer(content_before))

        if matches:
            last_match = matches[-1]
            return (last_match.group(1), last_match.group(2).strip())

        return ("Unknown", "Unknown Phase")

    def _extract_acceptance_criteria(self, task_id: str) -> List[str]:
        """Extract acceptance criteria for the task"""
        # Find acceptance criteria section near the task
        task_pos = self.roadmap_content.find(task_id)
        if task_pos == -1:
            return []

        # Look for acceptance criteria in the next 2000 characters
        search_section = self.roadmap_content[task_pos:task_pos + 2000]

        criteria_pattern = re.compile(r'-\s+✅\s+([^\n]+)')
        matches = criteria_pattern.findall(search_section)

        return matches[:10]  # Limit to first 10 criteria


class GeminiContentGenerator:
    """Generator for AI-powered content using Google Gemini API"""

    def __init__(self, config: PromptGeneratorConfig):
        """Initialize Gemini API client"""
        self.config = config
        self._setup_api()

    def _setup_api(self) -> None:
        """Configure Gemini API"""
        try:
            api_key = str(self.config.get('api_key', ''))
            model_name = str(self.config.get('model_name', 'gemini-2.5-flash'))

            genai.configure(api_key=api_key)  # type: ignore
            self.model = genai.GenerativeModel(model_name=model_name)  # type: ignore
            logger.info(f"Initialized Gemini API with model: {model_name}")
        except Exception as e:
            logger.error(f"Failed to initialize Gemini API: {e}")
            raise

    def generate_content(self, placeholder: str, task_info: TaskInfo, context: str = "") -> str:
        """Generate content for a specific placeholder using Gemini API"""
        try:
            # Create a prompt for the AI to generate appropriate content
            prompt = self._create_generation_prompt(placeholder, task_info, context)

            logger.info(f"Generating content for placeholder: {placeholder}")

            # Get configuration values
            temperature = float(self.config.get('temperature', 0.7))
            max_tokens = int(self.config.get('max_tokens', 8000))

            # Generate content using Gemini
            response = self.model.generate_content(  # type: ignore
                prompt,
                generation_config=genai.types.GenerationConfig(  # type: ignore
                    temperature=temperature,
                    max_output_tokens=max_tokens
                )
            )

            if response and hasattr(response, 'text') and response.text:
                content = response.text.strip()
                logger.info(f"Generated {len(content)} characters for {placeholder}")
                return content
            else:
                logger.warning(f"Empty response for {placeholder}")
                return f"[Content for {placeholder} - to be filled]"

        except Exception as e:
            logger.error(f"Failed to generate content for {placeholder}: {e}")
            return f"[Error generating {placeholder}: {str(e)}]"

    def _create_generation_prompt(self, placeholder: str, task_info: TaskInfo, context: str) -> str:
        """Create a prompt for the AI to generate content for a placeholder"""

        # Base context about the project
        base_context = f"""
You are helping to generate implementation prompts for the GoFlow Workflow Engine project.
This is a production-grade workflow orchestration system built in Go.

Task Information:
- Task ID: {task_info.task_id}
- Component: {task_info.component_name}
- Description: {task_info.description}
- Phase: {task_info.phase_number} - {task_info.phase_name}
- Priority: {task_info.priority}
- Dependencies: {task_info.dependencies}

Acceptance Criteria:
{chr(10).join(f"- {criteria}" for criteria in task_info.acceptance_criteria)}
"""

        # Placeholder-specific instructions with STRICT length limits
        placeholder_instructions = {
            'TARGET_LAYER': 'Identify which architectural layer (max 100 chars). Format: "**Layer Name** - one sentence."',
            'PRIMARY_FILES': 'List 2-3 main files ONLY (max 200 chars). Format: "- internal/[layer]/[file].go - brief desc"',
            'SUPPORTING_FILES': 'List 1-2 files if needed (max 150 chars). If none, say "None required".',
            'INTERFACE_DEFINITIONS': 'Provide ONE Go interface with 3-4 methods ONLY (max 400 chars). Use proper Go syntax.',
            'REQUIRED_DEPENDENCIES': 'List 2-3 packages ONLY (max 150 chars). Format: "- package/path - purpose"',
            'STANDARD_IMPORTS': 'Provide import block with 5-6 imports ONLY (max 250 chars).',
            'ENV_VARIABLES': 'List 2-3 variables ONLY (max 150 chars). Format: "KEY=value". If none, say "None required".',
            'CONFIG_FIELDS': 'Provide ONE struct with 3-4 fields ONLY (max 300 chars). Use proper Go syntax.',
            'DOC_REFERENCES': 'List 2 references ONLY (max 150 chars).',
            'TEST_FILE_PATH': 'Provide ONLY the test file path (max 80 chars). Example: "internal/repository/workflow_test.go"',
            'PACKAGE_NAME': 'Provide ONLY the package name (max 20 chars). Example: "repository"',
            'TEST_FUNCTIONS': 'List 3-4 test names ONLY (max 200 chars). Format: "- TestFunc_Scenario"',
            'BLOCKING_DEPENDENCIES': 'List task IDs or "None" (max 50 chars).',
            'RELATED_TASKS': 'List 1-2 task IDs ONLY (max 100 chars). Format: "- TASK-ID: brief desc"',
            'TASK_DESCRIPTION': f'Expand briefly on: {task_info.description}. MAXIMUM 600 characters. Be concise and focus on WHAT, not HOW.',
            'ACCEPTANCE_CRITERIA': 'Provide 4-5 BRIEF criteria (max 400 chars). Use numbered list, one line each.',
        }

        instruction = placeholder_instructions.get(
            placeholder,
            f'Generate appropriate content for the {placeholder} placeholder based on the task information.'
        )

        full_prompt = f"""{base_context}

Additional Context:
{context}

Task: Generate content for the {placeholder} placeholder.

Instructions:
{instruction}

CRITICAL REQUIREMENTS:
- STRICT character limit - DO NOT EXCEED the specified maximum
- Be CONCISE and BRIEF - every word counts
- Use proper Go syntax where applicable
- Follow GoFlow project conventions
- If information is not applicable, clearly state "Not applicable" or "None required"
- Generate ONLY the requested content, no explanations or meta-commentary

Generate the content now (remember the character limit):
"""

        return full_prompt


class PromptGenerator:
    """Main prompt generator that orchestrates the entire process"""

    def __init__(self, config: PromptGeneratorConfig):
        """Initialize the prompt generator"""
        self.config = config
        self.template_parser = TemplateParser(config.get('template_path'))
        self.task_extractor = TaskDataExtractor(config.get('roadmap_path'))
        self.content_generator = GeminiContentGenerator(config)
        logger.info("Prompt generator initialized successfully")

    def generate_prompt(self, task_id: str) -> Optional[str]:
        """Generate a complete prompt for the given task ID"""
        try:
            logger.info(f"Starting prompt generation for task: {task_id}")

            # Step 1: Extract task information
            task_info = self.task_extractor.extract_task_info(task_id)
            if not task_info:
                logger.error(f"Failed to extract task info for {task_id}")
                return None

            # Step 2: Split template into sections
            # Find the AI Implementation Prompt section (between ```markdown and closing ```)
            template_content = self.template_parser.template_content

            # Find the start of AI Implementation Prompt section
            prompt_start_marker = "## AI Implementation Prompt\n\n```markdown"
            prompt_start = template_content.find(prompt_start_marker)

            if prompt_start == -1:
                logger.error("Could not find AI Implementation Prompt section marker")
                return None

            # Find the end of the prompt section (first ``` after the start, excluding the opening one)
            prompt_section_start = prompt_start + len(prompt_start_marker)
            # Look for the closing ``` followed by two newlines and ---
            prompt_end_marker = "```\n\n---\n\n## Validation Checklist"
            prompt_end = template_content.find(prompt_end_marker, prompt_section_start)

            if prompt_end == -1:
                logger.error("Could not find end of AI Implementation Prompt section")
                return None

            # Split into parts - we only need header and prompt section
            header_section = template_content[:prompt_section_start]
            prompt_section = template_content[prompt_section_start:prompt_end]
            # footer_section is intentionally not used to reduce output size

            # Step 3: Replace placeholders ONLY in the prompt section
            prompt_section = self._replace_basic_placeholders(prompt_section, task_info)
            prompt_section = self._generate_ai_content(prompt_section, task_info)

            # Step 4: Reassemble with closing markdown block
            final_content = header_section + prompt_section + "\n```\n"

            # Step 5: Check 20,000 character limit
            MAX_LENGTH = 20000
            content_length = len(final_content)

            if content_length > MAX_LENGTH:
                logger.error(f"Prompt length ({content_length} chars) exceeds {MAX_LENGTH} char limit!")
                logger.error("This should not happen with the strict length limits. Please review AI-generated content.")
                # Don't truncate - this indicates a problem with the generation
                return None

            logger.info(f"Successfully generated prompt for {task_id} ({content_length} characters, {MAX_LENGTH - content_length} chars remaining)")
            return final_content

        except Exception as e:
            logger.error(f"Failed to generate prompt: {e}")
            return None

    def _replace_basic_placeholders(self, content: str, task_info: TaskInfo) -> str:
        """Replace basic placeholders with task information"""
        replacements = {
            'TASK_ID': task_info.task_id,
            'COMPONENT_NAME': task_info.component_name,
            'PHASE_NUMBER': task_info.phase_number,
            'PHASE_NAME': task_info.phase_name,
            'PRIORITY': task_info.priority,
            'DEPENDENCIES': task_info.dependencies,
            'DEVELOPER_NAME': 'AI Assistant',
            'DATE': datetime.now().strftime('%Y-%m-%d'),
        }

        for placeholder, value in replacements.items():
            content = content.replace(f'{{{{{placeholder}}}}}', value)

        logger.info(f"Replaced {len(replacements)} basic placeholders")
        return content

    def _generate_ai_content(self, content: str, task_info: TaskInfo) -> str:
        """Generate AI content for complex placeholders"""
        # List of placeholders that need AI generation
        ai_placeholders = [
            'TARGET_LAYER',
            'PRIMARY_FILES',
            'SUPPORTING_FILES',
            'INTERFACE_DEFINITIONS',
            'REQUIRED_DEPENDENCIES',
            'STANDARD_IMPORTS',
            'ENV_VARIABLES',
            'CONFIG_FIELDS',
            'DOC_REFERENCES',
            'TEST_FILE_PATH',
            'PACKAGE_NAME',
            'TEST_FUNCTIONS',
            'BLOCKING_DEPENDENCIES',
            'RELATED_TASKS',
            'TASK_DESCRIPTION',
            'ACCEPTANCE_CRITERIA',
        ]

        for placeholder in ai_placeholders:
            pattern = f'{{{{{placeholder}}}}}'
            if pattern in content:
                logger.info(f"Generating AI content for: {placeholder}")
                generated_content = self.content_generator.generate_content(
                    placeholder, task_info
                )
                content = content.replace(pattern, generated_content)

        return content

    def save_prompt(self, task_id: str, prompt_content: str) -> bool:
        """Save the generated prompt to a file"""
        try:
            # Create output directory if it doesn't exist
            output_dir = Path(str(self.config.get('output_dir', 'docs/tasks/prompts')))
            output_dir.mkdir(parents=True, exist_ok=True)

            # Generate filename
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            filename = f"{task_id}_{timestamp}_prompt.md"
            filepath = output_dir / filename

            # Save the prompt
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(prompt_content)

            logger.info(f"Saved prompt to: {filepath}")

            # Also save a "latest" version without timestamp
            latest_filename = f"{task_id}_latest_prompt.md"
            latest_filepath = output_dir / latest_filename
            with open(latest_filepath, 'w', encoding='utf-8') as f:
                f.write(prompt_content)

            logger.info(f"Saved latest prompt to: {latest_filepath}")
            return True

        except Exception as e:
            logger.error(f"Failed to save prompt: {e}")
            return False

    def generate_and_save(self, task_id: str) -> bool:
        """Generate and save a prompt for the given task ID"""
        prompt_content = self.generate_prompt(task_id)
        if prompt_content:
            return self.save_prompt(task_id, prompt_content)
        return False


def parse_arguments() -> argparse.Namespace:
    """Parse command line arguments"""
    parser = argparse.ArgumentParser(
        description='Generate AI implementation prompts for GoFlow tasks',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Generate prompt with API key from environment
  python main.py --task-id INIT-001

  # Generate prompt with explicit API key
  python main.py --task-id INIT-001 --api-key YOUR_API_KEY

  # Generate prompt with custom config file
  python main.py --task-id INIT-001 --config config.json

  # Generate prompts for multiple tasks
  python main.py --task-id INIT-001 INIT-002 INIT-003
        """
    )

    parser.add_argument(
        '--task-id',
        nargs='+',
        required=True,
        help='Task ID(s) to generate prompts for (e.g., INIT-001)'
    )

    parser.add_argument(
        '--api-key',
        help='Google Gemini API key (or set GEMINI_API_KEY env variable)'
    )

    parser.add_argument(
        '--config',
        help='Path to configuration JSON file'
    )

    parser.add_argument(
        '--template',
        help='Path to template file (overrides config)'
    )

    parser.add_argument(
        '--roadmap',
        help='Path to roadmap file (overrides config)'
    )

    parser.add_argument(
        '--output-dir',
        help='Output directory for generated prompts (overrides config)'
    )

    parser.add_argument(
        '--verbose',
        action='store_true',
        help='Enable verbose logging'
    )

    return parser.parse_args()


def main() -> None:
    """Main entry point for the script"""
    try:
        # Parse arguments
        args = parse_arguments()

        # Set logging level
        if args.verbose:
            logging.getLogger().setLevel(logging.DEBUG)
            logger.debug("Verbose logging enabled")

        # Load configuration
        config = PromptGeneratorConfig(args.config)

        # Override config with command line arguments
        if args.api_key:
            config.config['api_key'] = args.api_key
        if args.template:
            config.config['template_path'] = args.template
        if args.roadmap:
            config.config['roadmap_path'] = args.roadmap
        if args.output_dir:
            config.config['output_dir'] = args.output_dir

        # Re-validate after overrides
        config.validate_config()

        # Initialize prompt generator
        logger.info("Initializing prompt generator...")
        generator = PromptGenerator(config)

        # Generate prompts for each task ID
        task_ids = args.task_id
        success_count = 0
        failure_count = 0

        for task_id in task_ids:
            logger.info(f"\n{'='*60}")
            logger.info(f"Processing task: {task_id}")
            logger.info(f"{'='*60}\n")

            if generator.generate_and_save(task_id):
                success_count += 1
                logger.info(f"[SUCCESS] Successfully generated prompt for {task_id}")
            else:
                failure_count += 1
                logger.error(f"[FAILED] Failed to generate prompt for {task_id}")

        # Print summary
        logger.info(f"\n{'='*60}")
        logger.info("SUMMARY")
        logger.info(f"{'='*60}")
        logger.info(f"Total tasks processed: {len(task_ids)}")
        logger.info(f"Successful: {success_count}")
        logger.info(f"Failed: {failure_count}")
        logger.info(f"{'='*60}\n")

        # Exit with appropriate code
        sys.exit(0 if failure_count == 0 else 1)

    except KeyboardInterrupt:
        logger.info("\nOperation cancelled by user")
        sys.exit(130)
    except Exception as e:
        logger.error(f"Fatal error: {e}", exc_info=True)
        sys.exit(1)


if __name__ == '__main__':
    main()

