.PHONY: setup install run clean

# Setup virtual environment
setup:
	uv venv

# Install dependencies
install: setup
	uv pip install -r requirements.txt

# Run the script
run: install
	uv run python hbdscan.py

# Clean up
clean:
	rm -rf .venv
