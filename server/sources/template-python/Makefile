PYTHON := python3
VENV := venv
MAX_LOG_SIZE := 20000

# {{SOURCE_NAME}}

# Setup virtual environment and install dependencies
setup:
	$(PYTHON) -m venv $(VENV)
	./$(VENV)/bin/pip install --upgrade pip
	./$(VENV)/bin/pip install -r requirements.txt

# Run the source processor
run:
	./$(VENV)/bin/python main.py

# Run without virtual environment (if system-wide Python is used)
run-system:
	$(PYTHON) main.py

# Install dependencies system-wide
install:
	pip install -r requirements.txt

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
	./$(VENV)/bin/python -c "from save_source import test_database_connection; print('Success' if test_database_connection() else 'Failed')"

# Clean up virtual environment
clean:
	rm -rf $(VENV)
	rm -f v2.log*

# Install as systemd service
systemd:
	sudo cp {{SOURCE_NAME}}.service /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable {{SOURCE_NAME}}
	sudo systemctl restart {{SOURCE_NAME}}

# Development helpers
format:
	./$(VENV)/bin/python -m black *.py

lint:
	./$(VENV)/bin/python -m flake8 *.py

# Show service status
status:
	sudo systemctl status {{SOURCE_NAME}}

# Show service logs
logs:
	sudo journalctl -u {{SOURCE_NAME}} -f

.PHONY: setup run run-system install listen rotate test-db clean systemd format lint status logs