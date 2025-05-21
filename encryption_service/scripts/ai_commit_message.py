#!/usr/bin/env python3

import subprocess
import re
from typing import List, Dict, Tuple
import sys
import os
from collections import defaultdict
import google.generativeai as genai
from dotenv import load_dotenv
import requests
import json

# Load environment variables
load_dotenv()

# Configure Gemini API
GEMINI_API_KEY = os.getenv('GEMINI_API_KEY')
if not GEMINI_API_KEY:
    print("Error: GEMINI_API_KEY environment variable is not set")
    print("Please set it using: export GEMINI_API_KEY='your-api-key'")
    sys.exit(1)

GEMINI_API_URL = f"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key={GEMINI_API_KEY}"

def get_git_diff() -> str:
    """Get the git diff of staged changes."""
    try:
        result = subprocess.run(['git', 'diff', '--cached'], 
                              capture_output=True, 
                              text=True, 
                              check=True)
        return result.stdout
    except subprocess.CalledProcessError:
        print("Error: No staged changes found")
        sys.exit(1)

def get_changed_files() -> List[str]:
    """Get list of changed files."""
    try:
        result = subprocess.run(['git', 'diff', '--cached', '--name-only'],
                              capture_output=True,
                              text=True,
                              check=True)
        return result.stdout.strip().split('\n')
    except subprocess.CalledProcessError:
        return []

def get_current_branch() -> str:
    """Get the current git branch name."""
    try:
        result = subprocess.run(['git', 'rev-parse', '--abbrev-ref', 'HEAD'],
                              capture_output=True,
                              text=True,
                              check=True)
        return result.stdout.strip()
    except subprocess.CalledProcessError:
        return "unknown"

def get_file_changes() -> Dict[str, str]:
    """Get detailed changes for each file."""
    try:
        result = subprocess.run(['git', 'diff', '--cached', '--stat'],
                              capture_output=True,
                              text=True,
                              check=True)
        changes = {}
        for line in result.stdout.split('\n'):
            if '|' in line:
                file_name = line.split('|')[0].strip()
                changes[file_name] = line.split('|')[1].strip()
        return changes
    except subprocess.CalledProcessError:
        return {}

def get_recent_commits() -> List[str]:
    """Get recent commit messages for context."""
    try:
        result = subprocess.run(['git', 'log', '-n', '5', '--pretty=format:%s'],
                              capture_output=True,
                              text=True,
                              check=True)
        return result.stdout.strip().split('\n')
    except subprocess.CalledProcessError:
        return []

def analyze_changes_with_ai(diff: str, files: List[str], branch_name: str, additional_context: str = None) -> str:
    """Use Gemini API to analyze changes and generate a commit message."""
    try:
        # Prepare the context for the AI
        context_section = f"\nAdditional Context: {additional_context}\n" if additional_context else ""
        
        prompt = f"""You are a Git commit message generator. Analyze the following changes and generate a commit message that follows the Conventional Commits specification.{context_section}

Branch: {branch_name}
Changed files: {', '.join(files)}

Git diff:
{diff}

Recent commit history:
{chr(10).join(get_recent_commits())}

Generate a commit message that:
1. Uses the format: <type>(<scope>): <description>
   Types:
   - feat: New feature
   - fix: Bug fix
   - docs: Documentation changes
   - style: Code style changes (formatting, etc)
   - refactor: Code refactoring
   - test: Adding or modifying tests
   - chore: Maintenance tasks
   - ci: CI/CD related changes
   - build: Build system changes
   - perf: Performance improvements

2. Scope should be one of:
   - ci: CI/CD related changes
   - build: Build system changes
   - core: Core functionality
   - api: API changes
   - docs: Documentation
   - deps: Dependencies
   - scripts: Scripts and tools

3. Description should:
   - Start with a verb in present tense
   - Be under 72 characters
   - Be clear and specific
   - Focus on the "why" rather than the "what"

4. After the first line, add a blank line followed by:
   - Bullet points for major changes
   - Each bullet point should start with a verb
   - Keep bullet points concise
   - Focus on impactful changes

The message should be professional, clear, and follow Git best practices."""

        # Prepare the request payload
        payload = {
            "contents": [{
                "parts": [{"text": prompt}]
            }]
        }

        # Make the API request
        headers = {
            'Content-Type': 'application/json'
        }
        
        response = requests.post(GEMINI_API_URL, headers=headers, json=payload)
        
        if response.status_code == 200:
            result = response.json()
            if 'candidates' in result and len(result['candidates']) > 0:
                message = result['candidates'][0]['content']['parts'][0]['text']
                # Clean up the response
                message = message.strip()
                # Remove any markdown code block formatting if present
                message = re.sub(r'```[^\n]*\n', '', message)
                message = re.sub(r'```', '', message)
                return message.strip()
            else:
                raise Exception("No response content in the API response")
        else:
            raise Exception(f"API request failed with status code: {response.status_code}")
            
    except Exception as e:
        print(f"Error generating AI commit message: {e}")
        return None

def main():
    try:
        # Check if there are staged changes
        try:
            subprocess.run(['git', 'diff', '--cached', '--quiet'], check=True)
            print("No staged changes found")
            sys.exit(0)
        except subprocess.CalledProcessError:
            pass
        
        # Get changes and context
        diff = get_git_diff()
        files = get_changed_files()
        branch_name = get_current_branch()
        additional_context = None
        skip_ai_generation = False
        
        while True:
            try:
                # Generate AI commit message only if not skipping
                if not skip_ai_generation:
                    print("\nAnalyzing changes with AI...")
                    commit_message = analyze_changes_with_ai(diff, files, branch_name, additional_context)
                else:
                    skip_ai_generation = False  # Reset the flag for next iteration
                
                if commit_message:
                    # Print the commit message
                    print("\nGenerated Commit Message:")
                    print("=" * 50)
                    print(commit_message)
                    print("=" * 50)
                    
                    # Ask for user action
                    print("\nOptions:")
                    print("1. Accept and commit with this message")
                    print("2. Reject and exit")
                    print("3. Generate another commit message")
                    print("4. Add additional context and generate new message")
                    print("5. Edit commit message")
                    print("q. Quit")
                    
                    response = input("\nEnter your choice (1-5 or q): ").strip().lower()
                    
                    if response == "q":
                        print("\nExiting...")
                        sys.exit(0)
                    elif response == "1":
                        # Create a temporary file with the commit message
                        with open('.git/COMMIT_EDITMSG', 'w') as f:
                            f.write(commit_message)
                        print("Commit message saved. Running git commit...")
                        # Run git commit with the message
                        try:
                            subprocess.run(['git', 'commit', '-F', '.git/COMMIT_EDITMSG'], check=True)
                            print("Successfully committed changes!")
                        except subprocess.CalledProcessError as e:
                            print(f"Error committing changes: {e}")
                        break
                    elif response == "2":
                        print("Commit message discarded")
                        break
                    elif response == "3":
                        print("Generating new commit message...")
                        additional_context = None
                        continue
                    elif response == "4":
                        try:
                            additional_context = input("\nEnter additional context for the commit message (or 'q' to quit): ").strip()
                            if additional_context.lower() == 'q':
                                print("\nExiting...")
                                sys.exit(0)
                            continue
                        except KeyboardInterrupt:
                            print("\nInput cancelled. Returning to main menu...")
                            continue
                    elif response == "5":
                        try:
                            # Create a temporary file with the current message
                            with open('.git/COMMIT_EDITMSG', 'w') as f:
                                f.write(commit_message)
                            
                            # Open the default editor for editing
                            editor = os.environ.get('EDITOR', 'vim')
                            subprocess.call([editor, '.git/COMMIT_EDITMSG'])
                            
                            # Read the edited message
                            with open('.git/COMMIT_EDITMSG', 'r') as f:
                                edited_message = f.read().strip()
                            
                            if edited_message:
                                commit_message = edited_message
                                print("\nEdited Commit Message:")
                                print("=" * 50)
                                print(commit_message)
                                print("=" * 50)
                                skip_ai_generation = True  # Skip AI generation on next iteration
                            else:
                                print("\nNo changes made to the commit message.")
                        except Exception as e:
                            print(f"\nError editing commit message: {e}")
                            continue
                    else:
                        print("Invalid choice. Please enter 1-5 or q to quit.")
                else:
                    print("Failed to generate AI commit message. Please try again.")
                    break
            except KeyboardInterrupt:
                print("\nOperation cancelled. Returning to main menu...")
                continue
            except Exception as e:
                print(f"\nAn error occurred: {str(e)}")
                print("Please try again or press Ctrl+C to exit.")
                continue
    except KeyboardInterrupt:
        print("\nExiting...")
        sys.exit(0)
    except Exception as e:
        print(f"\nFatal error: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    main() 