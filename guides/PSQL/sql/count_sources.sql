SELECT 'main' as table_name, COUNT(*) as total 
FROM sources 
UNION ALL 
SELECT 'ai' as table_name, COUNT(*) as total 
FROM "sources-ai";
