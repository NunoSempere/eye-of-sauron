now=$(date +%s)
then=$(echo "$now - 60*60" | bc)
wget "http://hn.algolia.com/api/v1/search_by_date?tags=story&numericFilters=created_at_i>$then,created_at_i<$now" -O example.json
npx prettier -w example.json



