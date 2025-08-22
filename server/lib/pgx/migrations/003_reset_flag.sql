UPDATE flags
SET code = 0,
    msg = 'Flag cleared',
    updated_at = CURRENT_TIMESTAMP
WHERE name = 'openai_refill';
