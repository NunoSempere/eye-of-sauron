MAX_LOG_SIZE := 20000

# Globalbiodefense.com

# Setup project with uv (creates virtual environment and installs dependencies)
setup:
	uv sync

# Add new dependencies
add:
	uv add $(PACKAGE)

# Add development dependencies
add-dev:
	uv add --dev $(PACKAGE)

# Run the source processor
run:
	uv run src/main.py

# Install dependencies from existing requirements.txt (if not using pyproject.toml)
install:
	uv pip install -r requirements.txt

# Watch logs in real-time
listen:
	tail -f v2.log

# Rotate logs to prevent them from getting too large
rotate:
	# TODO: rotate postgres stuff
	tail -n $(MAX_LOG_SIZE) v2.log | tee -a v2.log.tmp
	mv v2.log.tmp v2.log

# Test database connection
test-db:
	uv run python -c "from save_source import test_database_connection; print('Success' if test_database_connection() else 'Failed')"

# Clean up virtual environment and cache
clean:
	rm -rf .venv
	rm -rf __pycache__
	rm -f v2.log*
	rm -f uv.lock

# Install as systemd service
systemd:
	sudo cp globalbiodefense.service /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable {{SOURCE_NAME}}
	sudo systemctl restart {{SOURCE_NAME}}

# Development helpers
format:
	uv run black *.py

lint:
	uv run flake8 *.py

# Show service status
status:
	sudo systemctl status {{SOURCE_NAME}}

# Show service logs
logs:
	sudo journalctl -u {{SOURCE_NAME}} -f

.PHONY: setup add add-dev run install listen rotate test-db clean systemd format lint status logs
