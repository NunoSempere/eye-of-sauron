DELETE FROM "sources-ai" 
WHERE id NOT IN (
    SELECT MIN(id) 
    FROM "sources-ai" 
    GROUP BY link
);
