I spent a bit of time looking into Common Crawl. It does seem reasonably approachable for some use cases; the problem is not the programming, it's the sheer size; 6TB of *compressed* text every month, in 100K chunks of 60Mb each 

Some links to get people started

https://commoncrawl.org/get-started
https://commoncrawl.org/overview
https://data.commoncrawl.org/crawl-data/CC-MAIN-2025-26/index.html

And some commands:

curl -s https://data.commoncrawl.org/crawl-data/CC-MAIN-2025-30/wet.paths.gz | funzip
curl -s https://data.commoncrawl.org/crawl-data/CC-MAIN-2025-30/wet.paths.gz | funzip | wc -l
curl -s https://data.commoncrawl.org/crawl-data/CC-MAIN-2025-30/wet.paths.gz | funzip | head -n 1 | xargs -I{} wget https://data.commoncrawl.org/{}
gunzip CC-MAIN-20250707183638-20250707213638-00000.warc.wet.gz
cat CC-MAIN-20250707183638-20250707213638-00000.warc.wet | head -n 200

Actually let me estimate how expensive it would be to feed this to an LLM

cat CC-MAIN-20250707183638-20250707213638-00000.warc.wet | wc -w # 23524492 = 23.52M words in that 60Mb chunk

https://platform.openai.com/docs/pricing?latest-pricing=batch
$0.025 per 1M tokens is the cheapest from OpenAI
Deepseek is not that much cheaper atm. https://api-docs.deepseek.com/quick_start/pricing
So the cost of feeding this into an llm is 23.52M * 0.025 $/M * 100K = 58.8K per month

With batch embeddings, you have a costs of 0.0001 USD/million tokens, so 235 USD/month, which is achievable. E.g., I could just send the embeddings, compute their distance from an "existential danger" embedding, and do further processing.

Inspired by reading <https://blog.wilsonl.in/search-engine/#live-demo>, which builds a search engine. This seems very involved. But building a very search engine at 235 USD/query, distributed across a month, seems much easier. 

Because of the large size, this would probably be better for simple queries (e.g., "mentions of Epstein"), rather than open ended LLM queries.
