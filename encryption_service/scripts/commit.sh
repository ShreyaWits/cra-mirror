#!/bin/bash

# Check if Python virtual environment exists
if [ ! -d ".venv" ]; then
    echo "Creating Python virtual environment..."
    python3 -m venv .venv
fi

# Activate virtual environment
source .venv/bin/activate

# Check if requirements are installed
if [ ! -f ".venv/requirements_installed" ]; then
    echo "Installing requirements..."
    pip install -r scripts/requirements.txt
    touch .venv/requirements_installed
fi

# Check if there are staged changes
if ! git diff --cached --quiet; then
    # Run the AI commit message generator
    python3 scripts/ai_commit_message.py
else
    echo "No staged changes found. Please stage your changes first using 'git add'."
    exit 1
fi

# Deactivate virtual environment
deactivate 