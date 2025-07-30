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
    Fetch sources from {{SOURCE_NAME}}.
    
    Returns:
        List of source dictionaries with keys: title, link, date
    """
    # TODO: Choose and implement the appropriate method for your source
    
    # Example for RSS feed:
    # return fetch_from_rss("{{RSS_URL}}")
    
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


def fetch_from_api(url: str, headers: Optional[Dict[str, str]] = None) -> List[Dict[str, str]]:
    """
    Fetch sources from a JSON API.
    
    Args:
        url: API endpoint URL
        headers: Optional HTTP headers for authentication
        
    Returns:
        List of source dictionaries
    """
    try:
        logger.info(f"Fetching from API: {url}")
        
        if headers is None:
            headers = {}
        
        # Add user agent
        headers.setdefault('User-Agent', 'eye-of-sauron/1.0')
        
        response = requests.get(url, headers=headers, timeout=30)
        response.raise_for_status()
        
        data = response.json()
        
        # TODO: Customize this based on your API's response structure
        sources = []
        
        # Example: if API returns {"articles": [...]}
        if isinstance(data, dict) and 'articles' in data:
            for item in data['articles']:
                source = {
                    'title': item.get('title', ''),
                    'link': item.get('url', ''),
                    'date': item.get('publishedAt', '')  # Common in news APIs
                }
                sources.append(source)
        
        # Example: if API returns a list directly
        elif isinstance(data, list):
            for item in data:
                source = {
                    'title': item.get('title', ''),
                    'link': item.get('link', item.get('url', '')),
                    'date': item.get('date', item.get('published', ''))
                }
                sources.append(source)
        
        logger.info(f"Fetched {len(sources)} sources from API")
        return sources
        
    except requests.RequestException as e:
        logger.error(f"Error fetching from API: {e}")
        return []
    except json.JSONDecodeError as e:
        logger.error(f"Error parsing JSON response: {e}")
        return []


def fetch_from_webpage(url: str) -> List[Dict[str, str]]:
    """
    Scrape sources from a webpage.
    
    Args:
        url: Webpage URL to scrape
        
    Returns:
        List of source dictionaries
    """
    try:
        logger.info(f"Scraping webpage: {url}")
        
        headers = {
            'User-Agent': 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
        }
        
        response = requests.get(url, headers=headers, timeout=30)
        response.raise_for_status()
        
        soup = BeautifulSoup(response.content, 'html.parser')
        
        sources = []
        
        # TODO: Customize these selectors based on the target webpage structure
        
        # Example: Extract links from a news site
        article_links = soup.find_all('a', class_='article-link')  # Adjust selector
        for link in article_links:
            href = link.get('href', '')
            if href:
                # Convert relative URLs to absolute
                full_url = urljoin(url, href)
                
                source = {
                    'title': link.get_text(strip=True),
                    'link': full_url,
                    'date': ''  # Try to extract date from nearby elements if available
                }
                sources.append(source)
        
        # Alternative example: Extract from structured data
        # articles = soup.find_all('article')
        # for article in articles:
        #     title_elem = article.find('h2') or article.find('h3')  
        #     link_elem = article.find('a')
        #     date_elem = article.find(class_='date') or article.find('time')
        #     
        #     if title_elem and link_elem:
        #         source = {
        #             'title': title_elem.get_text(strip=True),
        #             'link': urljoin(url, link_elem.get('href', '')),
        #             'date': date_elem.get_text(strip=True) if date_elem else ''
        #         }
        #         sources.append(source)
        
        # Filter out empty links
        sources = [s for s in sources if s['link'] and s['title']]
        
        logger.info(f"Scraped {len(sources)} sources from webpage")
        return sources
        
    except requests.RequestException as e:
        logger.error(f"Error scraping webpage: {e}")
        return []
    except Exception as e:
        logger.error(f"Error parsing webpage: {e}")
        return []


# Utility function for handling pagination in APIs
def fetch_paginated_api(base_url: str, max_pages: int = 5, 
                       headers: Optional[Dict[str, str]] = None) -> List[Dict[str, str]]:
    """
    Fetch from a paginated API.
    
    Args:
        base_url: Base API URL (should include page parameter placeholder)
        max_pages: Maximum number of pages to fetch
        headers: Optional HTTP headers
        
    Returns:
        List of all sources from all pages
    """
    all_sources = []
    
    for page in range(1, max_pages + 1):
        # TODO: Customize URL pattern for your API
        page_url = f"{base_url}&page={page}"  # or however pagination works
        
        sources = fetch_from_api(page_url, headers)
        if not sources:
            # No more data, stop pagination
            break
            
        all_sources.extend(sources)
        logger.info(f"Fetched page {page}, total sources so far: {len(all_sources)}")
    
    return all_sources