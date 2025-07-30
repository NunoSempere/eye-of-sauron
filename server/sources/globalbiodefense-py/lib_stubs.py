#!/usr/bin/env python3

"""
Python implementations of the Go library functions.
These replace the stub implementations with actual working code.
"""

import logging
import re
import hashlib
from typing import Dict, Any, Optional
from urllib.parse import urlparse
import psycopg2
import requests
from bs4 import BeautifulSoup
import trafilatura
from newspaper import Article
from readability import Document
import openai
import json


logger = logging.getLogger(__name__)


class filters:
    """Filter library functions."""
    
    @staticmethod
    def is_dupe(source: Dict[str, str], database_url: str) -> bool:
        """
        Check if source is a duplicate by querying the database.
        """
        try:
            conn = psycopg2.connect(database_url)
            cursor = conn.cursor()
            
            # Check for duplicate title or link (case-insensitive title)
            cursor.execute(
                "SELECT EXISTS(SELECT 1 FROM sources WHERE UPPER(title) = %s OR link = %s)",
                (source['title'].upper(), source['link'])
            )
            exists = cursor.fetchone()[0]
            
            cursor.close()
            conn.close()
            
            if exists:
                logger.info(f"Skipping duplicate title/link: {source['title']} {source['link']}")
            else:
                logger.info("Article is not a duplicate")
                
            return exists
            
        except Exception as e:
            logger.error(f"Error checking for duplicates: {e}")
            return False
    
    @staticmethod
    def is_good_host(source: Dict[str, str]) -> bool:
        """
        Check if the source host is acceptable.
        """
        try:
            parsed_url = urlparse(source['link'])
            skippable_hosts = [
                "www.washingtonpost.com", 
                "www.youtube.com", 
                "www.naturalnews.com", 
                "facebook.com", 
                "m.facebook.com", 
                "www.bignewsnetwork.com"
            ]
            
            is_bad_host = parsed_url.netloc in skippable_hosts
            
            if is_bad_host:
                logger.info("Article is from a bad host")
            else:
                logger.info("Article is from a good host")
                
            return not is_bad_host
            
        except Exception as e:
            logger.error(f"Error checking host: {e}")
            return False
    
    @staticmethod
    def clean_title(title: str) -> str:
        """
        Clean up the title text by removing site markers and HTML.
        """
        def clean_title_with_marker(s: str, ending_marker: str) -> str:
            if len(s) > 25:
                # Find the last occurrence of the marker after position 25
                pos = s[25:].rfind(ending_marker)
                if pos != -1:
                    return s[:25 + pos]
            return s
        
        # Apply cleaning steps like the Go version
        result = clean_title_with_marker(title, " – ")
        result = clean_title_with_marker(result, " - ")
        result = clean_title_with_marker(result, "|")
        
        # Remove HTML tags and entities
        result = result.replace("<b>", "")
        result = result.replace("</b>", "")
        result = result.replace("&#39;", "'")
        
        return result.strip()


class readability:
    """Readability library functions."""
    
    @staticmethod
    def extract_title(url: str) -> str:
        """
        Extract title from webpage using multiple methods.
        """
        try:
            # Try newspaper3k first
            article = Article(url)
            article.download()
            article.parse()
            
            if article.title:
                logger.info(f"Found title from newspaper3k: {article.title}")
                return article.title.strip()
                
        except Exception as e:
            logger.warning(f"newspaper3k failed for title extraction: {e}")
        
        try:
            # Fallback to manual HTML parsing
            headers = {
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
            }
            response = requests.get(url, headers=headers, timeout=30)
            response.raise_for_status()
            
            soup = BeautifulSoup(response.content, 'html.parser')
            title_tag = soup.find('title')
            
            if title_tag:
                title = title_tag.get_text().strip()
                logger.info(f"Found title from HTML parsing: {title}")
                return title
                
        except Exception as e:
            logger.warning(f"HTML parsing failed for title extraction: {e}")
        
        return ""
    
    @staticmethod
    def get_article_content(url: str) -> str:
        """
        Extract article content from webpage using readability algorithm.
        """
        # Try trafilatura first (most reliable)
        try:
            content = trafilatura.fetch_url(url)
            if content:
                extracted = trafilatura.extract(content)
                if extracted and len(extracted) > 200:
                    logger.info(f"Successfully extracted content using trafilatura: {len(extracted)} chars")
                    return extracted
        except Exception as e:
            logger.warning(f"trafilatura failed: {e}")
        
        # Try newspaper3k
        try:
            article = Article(url)
            article.download()
            article.parse()
            
            if article.text and len(article.text) > 200:
                logger.info(f"Successfully extracted content using newspaper3k: {len(article.text)} chars")
                return article.text
                
        except Exception as e:
            logger.warning(f"newspaper3k failed: {e}")
        
        # Try readability-lxml as fallback
        try:
            headers = {
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
            }
            response = requests.get(url, headers=headers, timeout=30)
            response.raise_for_status()
            
            doc = Document(response.content)
            content = doc.summary()
            
            # Extract text from HTML
            soup = BeautifulSoup(content, 'html.parser')
            text = soup.get_text()
            
            if text and len(text) > 200:
                logger.info(f"Successfully extracted content using readability-lxml: {len(text)} chars")
                return text.strip()
                
        except Exception as e:
            logger.warning(f"readability-lxml failed: {e}")
        
        logger.error(f"All content extraction methods failed for: {url}")
        return ""


class llm:
    """LLM library functions."""
    
    @staticmethod
    def summarize(content: str, openai_key: str) -> str:
        """
        Summarize article content using OpenAI.
        """
        try:
            client = openai.OpenAI(api_key=openai_key)
            
            prompt = (
                "The json API endpoint returns a {summary, error} object, like {summary: \"The article is about xyz\", error: null}. "
                "The summary contains, as a string, first a general summary of the contents of the article in two paragraphs or less, "
                "and then an outline with the most salient, new and informative facts in an additional paragraph. "
                "The summary just states the contents of the article, and doesn't say \"The article says\" or similar introductions. "
                f"For example, given the following article\n\n<INPUT>{content}</INPUT>\n\n"
                "The output is as follows (as a reminder, the json API endpoint returns a {summary, error} object, "
                "like {summary: \"The article is about xyz\", error: null}. The summary contains, as a string, "
                "first a general summary of the article in two paragraphs or less, and then an outline outlines "
                "the most salient, new and informative facts in an additional paragraph):"
            )
            
            response = client.chat.completions.create(
                model="gpt-4o-mini",
                messages=[
                    {"role": "user", "content": prompt}
                ],
                response_format={"type": "json_object"}
            )
            
            result = response.choices[0].message.content
            summary_data = json.loads(result)
            
            if summary_data.get('error'):
                logger.error(f"OpenAI returned error: {summary_data['error']}")
                return ""
                
            summary = summary_data.get('summary', '')
            logger.info(f"Successfully generated summary: {len(summary)} chars")
            return summary
            
        except Exception as e:
            logger.error(f"Summarization failed: {e}")
            return ""
    
    @staticmethod
    def check_existential_importance(snippet: str, openai_key: str) -> Optional[Dict[str, Any]]:
        """
        Check existential importance using OpenAI.
        """
        try:
            client = openai.OpenAI(api_key=openai_key)
            
            prompt = (
                "The existential importance json API endpoint returns a {existential_importance_reasoning, "
                "existential_importance_bool, high_importance_bool, error} object.\n\n"
                "The existential_importance_reasoning field contains, as a string, a determination of whether "
                "the input describes an event of global importance. existential_importance_bool contains the result "
                "of that determination as a true/false boolean. high_importance_bool contains, as a true/false boolean, "
                "whether the event is highly important, even if it is not of \"existential\" importance.\n\n"
                "Items are of existential importance if:\n"
                "- They involve more than a hundred deaths.\n"
                "- They involve many cases of a sickness that might spread, or a new pathogen\n"
                "- They involve conflict between nuclear powers\n"
                "- They involve conflict that could escalate into global conflict, even if it hasn't already\n"
                "- They involve terrorist groups displaying new capabilities\n"
                "- ... and in general, if they involve events that could threaten humanity as a whole\n\n"
                f"For a longer example, given the following item\n\n<INPUT>{snippet}</INPUT>\n\n"
                "The output is as follows: (As a reminder, the existential importance json API endpoint returns a "
                "{existential_importance_reasoning, existential_importance_bool, high_importance_bool, error} object, "
                "opinion pieces, or editorials are not categorized as existentially important.)\n"
            )
            
            response = client.chat.completions.create(
                model="gpt-4o-mini",
                messages=[
                    {"role": "user", "content": prompt}
                ],
                response_format={"type": "json_object"}
            )
            
            result = response.choices[0].message.content
            importance_data = json.loads(result)
            
            if importance_data.get('error'):
                logger.error(f"OpenAI returned error: {importance_data['error']}")
                return None
                
            logger.info(f"Importance check result: {importance_data.get('existential_importance_bool')}")
            return importance_data
            
        except Exception as e:
            logger.error(f"Importance check failed: {e}")
            return None