#!/usr/bin/env python3

import logging
import os
import sys
import time
from datetime import datetime
from pathlib import Path

from dotenv import load_dotenv

from fetch import fetch_sources
from filter_and_expand import filter_and_expand_source
from save_source import save_source


def setup_logging():
    """Set up logging to both console and file."""
    log_format = '%(asctime)s - %(levelname)s - %(message)s'
    
    # Create handlers
    console_handler = logging.StreamHandler(sys.stdout)
    file_handler = logging.FileHandler('v2.log', mode='a')
    
    # Set format
    formatter = logging.Formatter(log_format)
    console_handler.setFormatter(formatter)
    file_handler.setFormatter(formatter)
    
    # Configure root logger
    logging.basicConfig(
        level=logging.INFO,
        handlers=[console_handler, file_handler]
    )


def main():
    """Main processing loop for {{SOURCE_NAME}} source."""
    setup_logging()
    logger = logging.getLogger(__name__)
    
    # Load environment variables
    load_dotenv()
    openai_key = os.getenv("OPENAI_KEY")
    pg_database_url = os.getenv("DATABASE_POOL_URL")
    
    if not openai_key or not pg_database_url:
        logger.error("Missing required environment variables: OPENAI_KEY, DATABASE_POOL_URL")
        sys.exit(1)
    
    while True:
        try:
            logger.info("Starting {{SOURCE_NAME}} processing")
            
            # TODO: Replace with your source-specific fetching logic
            sources = fetch_sources()
            logger.info(f"Found {len(sources)} sources")
            
            # Process each source
            for i, source in enumerate(sources, 1):
                logger.info(f"\nProcessing source {i}/{len(sources)}: {source['title']}")
                
                expanded_source, passes_filters = filter_and_expand_source(
                    source, openai_key, pg_database_url
                )
                
                if passes_filters:
                    save_source(expanded_source)
            
            logger.info("Finished processing {{SOURCE_NAME}}, sleeping for {{SLEEP_DURATION}}")
            time.sleep({{SLEEP_SECONDS}})  # TODO: Replace with appropriate duration in seconds
            
        except KeyboardInterrupt:
            logger.info("Received interrupt signal, shutting down...")
            break
        except Exception as e:
            logger.error(f"Unexpected error in main loop: {e}", exc_info=True)
            logger.info("Continuing after error...")
            time.sleep(60)  # Wait a minute before retrying


if __name__ == "__main__":
    main()