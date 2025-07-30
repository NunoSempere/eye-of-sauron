#!/usr/bin/env python3

import logging
from datetime import datetime
from typing import Dict, Tuple, Any, Optional

# TODO: These would need to be Python equivalents of the Go libraries
# For now, we'll define placeholder functions that match the expected interface
from lib_stubs import filters, llm, readability


logger = logging.getLogger(__name__)


def filter_and_expand_source(source: Dict[str, str], openai_key: str, 
                           database_url: str) -> Tuple[Dict[str, Any], bool]:
    """
    Process a {{SOURCE_NAME}} source through various filters,
    expand its content (via summarization and importance check),
    and return an ExpandedSource dict and a boolean indicating if it passes thresholds.
    
    Args:
        source: Dictionary with keys: title, link, date
        openai_key: OpenAI API key for LLM operations
        database_url: Database connection URL
        
    Returns:
        Tuple of (expanded_source_dict, passes_filters_bool)
    """
    # Initialize expanded source with basic info
    expanded_source = {
        'title': source['title'],
        'link': source['link'],
        'date': source['date'],
        'summary': '',
        'importance_bool': False,
        'importance_reasoning': ''
    }

    # If no date provided, use current time
    if not expanded_source['date']:
        expanded_source['date'] = datetime.now().isoformat()

    # Check for duplicates
    try:
        is_dupe = filters.is_dupe(source, database_url)
        if is_dupe:
            logger.info(f"Duplicate source found: {source['link']}")
            return expanded_source, False
    except Exception as e:
        logger.error(f"Error checking for duplicates: {e}")
        return expanded_source, False

    # TODO: Add source-specific freshness check if needed
    # For sources without timestamps, you might want to assume freshness
    # or try to extract date from the article content
    
    # Check if host is acceptable
    try:
        is_good_host = filters.is_good_host(source)
        if not is_good_host:
            logger.info(f"Host not acceptable: {source['link']}")
            return expanded_source, False
    except Exception as e:
        logger.error(f"Error checking host: {e}")
        return expanded_source, False

    # Try to get a better title from the source HTML
    try:
        title = readability.extract_title(source['link'])
        if title:
            expanded_source['title'] = title
            logger.info(f"Found title from HTML: {title}")
    except Exception as e:
        logger.warning(f"Failed to extract title from HTML: {e}")

    # Clean up the title
    try:
        expanded_source['title'] = filters.clean_title(expanded_source['title'])
    except Exception as e:
        logger.warning(f"Failed to clean title: {e}")

    # Get article content using a readability extractor
    try:
        content = readability.get_article_content(source['link'])
        if not content:
            logger.error(f"No content extracted for: {source['link']}")
            return expanded_source, False
    except Exception as e:
        logger.error(f"Readability extraction failed for {source['link']}: {e}")
        return expanded_source, False
    
    # Summarize the article using an LLM
    try:
        summary = llm.summarize(content, openai_key)
        if not summary:
            logger.error(f"Empty summary for: {source['link']}")
            return expanded_source, False
        expanded_source['summary'] = summary
        logger.info(f"Summary: {summary}")
    except Exception as e:
        logger.error(f"Summarization failed for {source['link']}: {e}")
        return expanded_source, False

    # Check existential or importance threshold
    try:
        existential_importance_snippet = f"# {expanded_source['title']}\n\n{summary}"
        importance_result = llm.check_existential_importance(
            existential_importance_snippet, openai_key
        )
        
        if not importance_result:
            logger.error(f"Importance check failed for {source['link']}")
            return expanded_source, False
            
        expanded_source['importance_bool'] = importance_result.get('existential_importance_bool', False)
        expanded_source['importance_reasoning'] = importance_result.get('existential_importance_reasoning', '')
        
        logger.info(f"Importance bool: {expanded_source['importance_bool']}")
        logger.info(f"Reasoning: {expanded_source['importance_reasoning']}")
        
    except Exception as e:
        logger.error(f"Importance check failed for {source['link']}: {e}")
        return expanded_source, False

    return expanded_source, expanded_source['importance_bool']


def parse_date_string(date_str: str) -> Optional[datetime]:
    """
    Parse various date string formats into datetime object.
    
    Args:
        date_str: Date string in various formats
        
    Returns:
        datetime object or None if parsing fails
    """
    if not date_str:
        return None
    
    # Common date formats to try
    formats = [
        '%Y-%m-%dT%H:%M:%S%z',      # ISO 8601 with timezone
        '%Y-%m-%dT%H:%M:%S',        # ISO 8601 without timezone
        '%Y-%m-%d %H:%M:%S',        # Standard datetime
        '%Y-%m-%d',                 # Date only
        '%a, %d %b %Y %H:%M:%S %Z', # RFC 2822 (RSS format)
        '%a, %d %b %Y %H:%M:%S %z', # RFC 2822 with numeric timezone
    ]
    
    for fmt in formats:
        try:
            return datetime.strptime(date_str, fmt)
        except ValueError:
            continue
    
    logger.warning(f"Could not parse date string: {date_str}")
    return None


def is_fresh_enough(date_str: str, max_age_hours: int = 48) -> bool:
    """
    Check if a source is fresh enough based on its date.
    
    Args:
        date_str: Date string from the source
        max_age_hours: Maximum age in hours to consider fresh
        
    Returns:
        True if fresh enough, False otherwise
    """
    if not date_str:
        # If no date available, assume it's fresh
        return True
    
    parsed_date = parse_date_string(date_str)
    if not parsed_date:
        # If we can't parse the date, assume it's fresh
        return True
    
    # Remove timezone info for comparison if present
    if parsed_date.tzinfo is not None:
        parsed_date = parsed_date.replace(tzinfo=None)
    
    now = datetime.now()
    age_hours = (now - parsed_date).total_seconds() / 3600
    
    is_fresh = age_hours <= max_age_hours
    logger.debug(f"Source age: {age_hours:.1f} hours, fresh: {is_fresh}")
    
    return is_fresh