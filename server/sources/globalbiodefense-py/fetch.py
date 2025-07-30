#!/usr/bin/env python3

import json
import logging
from datetime import datetime
from typing import List, Dict, Any, Optional
from urllib.parse import urljoin

import requests
from bs4 import BeautifulSoup
import feedparser


logger = logging.getLogger(__name__)


def fetch_sources() -> List[Dict[str, str]]:
    """
    Fetch sources from globalbiodefense.com
    
    Returns:
        List of source dictionaries with keys: title, link, date
    """
    # TODO: Choose and implement the appropriate method for your source
    
    # Example for RSS feed:
    return fetch_from_rss("https://globalbiodefense.com/feed/")
    
    # Example for JSON API:
    # return fetch_from_api("{{API_URL}}")
    
    # Example for web scraping:
    # return fetch_from_webpage("{{WEBPAGE_URL}}")
    
    # TODO: Replace with actual implementation
    logger.warning("fetch_sources() not implemented - returning empty list")
    return []


def fetch_from_rss(url: str) -> List[Dict[str, str]]:
    """
    Fetch sources from an RSS feed.
    
    Args:
        url: RSS feed URL
        
    Returns:
        List of source dictionaries
    """
    try:
        logger.info(f"Fetching RSS feed from: {url}")
        feed = feedparser.parse(url)
        
        if feed.bozo:
            logger.warning(f"RSS feed may be malformed: {feed.bozo_exception}")
        
        sources = []
        for entry in feed.entries:
            # Extract publication date
            date_str = ""
            if hasattr(entry, 'published'):
                date_str = entry.published
            elif hasattr(entry, 'updated'):
                date_str = entry.updated
            
            # Convert to ISO format if possible
            if date_str:
                try:
                    # feedparser usually parses dates into time.struct_time
                    if hasattr(entry, 'published_parsed') and entry.published_parsed:
                        dt = datetime(*entry.published_parsed[:6])
                        date_str = dt.isoformat()
                    elif hasattr(entry, 'updated_parsed') and entry.updated_parsed:
                        dt = datetime(*entry.updated_parsed[:6])
                        date_str = dt.isoformat()
                except Exception as e:
                    logger.warning(f"Failed to parse date '{date_str}': {e}")
            
            source = {
                'title': entry.get('title', ''),
                'link': entry.get('link', ''),
                'date': date_str
            }
            sources.append(source)
        
        logger.info(f"Fetched {len(sources)} sources from RSS feed")
        return sources
        
    except Exception as e:
        logger.error(f"Error fetching RSS feed: {e}")
        return []

