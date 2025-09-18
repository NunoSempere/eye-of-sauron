DELETE FROM sources 
WHERE id NOT IN (
    SELECT MIN(id) 
    FROM sources 
    GROUP BY link
);
