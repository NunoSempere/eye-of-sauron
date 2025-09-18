SELECT 'sources' as table_name, 
       pg_size_pretty(pg_total_relation_size('sources')) as size 
UNION ALL 
SELECT 'sources-ai' as table_name, 
       pg_size_pretty(pg_total_relation_size('"sources-ai"')) as size;
