#!/usr/bin/env python3

import logging
import os
from datetime import datetime
from typing import Dict, Any

import psycopg2
from psycopg2.extras import RealDictCursor


logger = logging.getLogger(__name__)


def save_source(source: Dict[str, Any]) -> bool:
    """
    Connect to the database and insert the expanded source.
    
    Args:
        source: Dictionary containing expanded source data
        
    Returns:
        True if successful, False otherwise
    """
    # Now that filtering functions are implemented, we can save sources
    try:
        # Get database URL from environment
        database_url = os.getenv("DATABASE_POOL_URL")
        if not database_url:
            logger.error("DATABASE_POOL_URL environment variable not set")
            return False
        
        # Connect to database
        with psycopg2.connect(database_url) as conn:
            with conn.cursor(cursor_factory=RealDictCursor) as cursor:
                # Parse the date - handle various formats
                date_obj = parse_date_for_db(source.get('date', ''))
                
                # Insert into database with conflict handling
                cursor.execute("""
                    INSERT INTO sources (title, link, date, summary, importance_bool, importance_reasoning)
                    VALUES (%s, %s, %s, %s, %s, %s)
                    ON CONFLICT (link) DO NOTHING
                """, (
                    source.get('title', ''),
                    source.get('link', ''),
                    date_obj,
                    source.get('summary', ''),
                    source.get('importance_bool', False),
                    source.get('importance_reasoning', '')
                ))
                
                # Check if row was actually inserted
                if cursor.rowcount > 0:
                    logger.info(f"Saved source: {source.get('title', '')}")
                else:
                    logger.info(f"Source already exists (skipped): {source.get('title', '')}")
                
                conn.commit()
                return True
                
    except psycopg2.Error as e:
        logger.error(f"Database error saving source: {e}")
        return False
    except Exception as e:
        logger.error(f"Unexpected error saving source: {e}")
        return False


def parse_date_for_db(date_str: str) -> datetime:
    """
    Parse date string into datetime object for database storage.
    
    Args:
        date_str: Date string in various formats
        
    Returns:
        datetime object (current time if parsing fails)
    """
    if not date_str:
        return datetime.now()
    
    # Common date formats to try
    formats = [
        '%Y-%m-%dT%H:%M:%S%z',      # ISO 8601 with timezone
        '%Y-%m-%dT%H:%M:%S.%f%z',   # ISO 8601 with microseconds and timezone
        '%Y-%m-%dT%H:%M:%S',        # ISO 8601 without timezone
        '%Y-%m-%dT%H:%M:%S.%f',     # ISO 8601 with microseconds
        '%Y-%m-%d %H:%M:%S',        # Standard datetime
        '%Y-%m-%d',                 # Date only
        '%a, %d %b %Y %H:%M:%S GMT', # RFC 2822 GMT
        '%a, %d %b %Y %H:%M:%S %Z', # RFC 2822 with timezone name
        '%a, %d %b %Y %H:%M:%S %z', # RFC 2822 with numeric timezone
        '%d %b %Y %H:%M:%S GMT',    # Alternative RFC format
        '%d/%m/%Y %H:%M:%S',        # DD/MM/YYYY format
        '%m/%d/%Y %H:%M:%S',        # MM/DD/YYYY format
    ]
    
    for fmt in formats:
        try:
            parsed = datetime.strptime(date_str, fmt)
            # Remove timezone info if present (PostgreSQL will handle timezone)
            if parsed.tzinfo is not None:
                parsed = parsed.replace(tzinfo=None)
            return parsed
        except ValueError:
            continue
    
    # If all parsing attempts fail, log warning and return current time
    logger.warning(f"Could not parse date '{date_str}', using current time")
    return datetime.now()


def test_database_connection() -> bool:
    """
    Test database connectivity.
    
    Returns:
        True if connection successful, False otherwise
    """
    try:
        database_url = os.getenv("DATABASE_POOL_URL")
        if not database_url:
            logger.error("DATABASE_POOL_URL environment variable not set")
            return False
        
        with psycopg2.connect(database_url) as conn:
            with conn.cursor() as cursor:
                cursor.execute("SELECT 1")
                result = cursor.fetchone()
                if result and result[0] == 1:
                    logger.info("Database connection successful")
                    return True
                else:
                    logger.error("Unexpected database response")
                    return False
                    
    except psycopg2.Error as e:
        logger.error(f"Database connection failed: {e}")
        return False
    except Exception as e:
        logger.error(f"Unexpected error testing database connection: {e}")
        return False


def get_source_count() -> int:
    """
    Get the total number of sources in the database.
    
    Returns:
        Number of sources, or -1 if error
    """
    try:
        database_url = os.getenv("DATABASE_POOL_URL")
        if not database_url:
            return -1
        
        with psycopg2.connect(database_url) as conn:
            with conn.cursor() as cursor:
                cursor.execute("SELECT COUNT(*) FROM sources")
                result = cursor.fetchone()
                return result[0] if result else -1
                
    except psycopg2.Error as e:
        logger.error(f"Error getting source count: {e}")
        return -1
    except Exception as e:
        logger.error(f"Unexpected error getting source count: {e}")
        return -1


def cleanup_old_sources(days_old: int = 30) -> int:
    """
    Remove sources older than specified number of days.
    
    Args:
        days_old: Remove sources older than this many days
        
    Returns:
        Number of sources removed, or -1 if error
    """
    try:
        database_url = os.getenv("DATABASE_POOL_URL")
        if not database_url:
            return -1
        
        with psycopg2.connect(database_url) as conn:
            with conn.cursor() as cursor:
                cursor.execute("""
                    DELETE FROM sources 
                    WHERE date < NOW() - INTERVAL '%s days'
                """, (days_old,))
                
                deleted_count = cursor.rowcount
                conn.commit()
                
                logger.info(f"Deleted {deleted_count} sources older than {days_old} days")
                return deleted_count
                
    except psycopg2.Error as e:
        logger.error(f"Error cleaning up old sources: {e}")
        return -1
    except Exception as e:
        logger.error(f"Unexpected error cleaning up old sources: {e}")
        return -1
