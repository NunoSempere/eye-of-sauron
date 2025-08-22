package outbound

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    "github.com/jackc/pgx/v5"
)

const (
    FlagOpenAIRefill = "openai_refill"
    FlagCodeSet      = 1
    FlagCodeUnset    = 0
    defaultTimeout   = 5 * time.Second
)

// SetFlag connects to the database and upserts the flag status
func SetFlag() error {
    conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_POOL_URL"))
    if err != nil {
        return fmt.Errorf("unable to connect to database: %w", err)
    }
    defer conn.Close(context.Background())

    ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
    defer cancel()

    // Using UPSERT (INSERT ... ON CONFLICT DO UPDATE)
    _, err = conn.Exec(ctx, `
        INSERT INTO flags (name, code, msg)
        VALUES ($1, $2, $3)
        ON CONFLICT (name)
        DO UPDATE SET
            code = $2,
            msg = $3,
            updated_at = CURRENT_TIMESTAMP
    `, FlagOpenAIRefill, FlagCodeSet, "OpenAI balance zero or negative, please refill")

    if err != nil {
        return fmt.Errorf("failed to set flag: %w", err)
    }

    log.Printf("Flag set/updated successfully")
    return nil
}

// CheckFlag checks if the flag exists and is set
func CheckFlag() bool {
    conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_POOL_URL"))
    if err != nil {
        log.Printf("Unable to connect to database: %v\n", err)
        return false
    }
    defer conn.Close(context.Background())

    ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
    defer cancel()

    var code int
    err = conn.QueryRow(ctx, `
        SELECT code
        FROM flags
        WHERE name = $1
    `, FlagOpenAIRefill).Scan(&code)

    if err != nil {
        if err == pgx.ErrNoRows {
            // Flag doesn't exist
            return false
        }
        log.Printf("Error checking flag in database: %v\n", err)
        return false
    }

    // Return true if code is 1 (flag is set)
    return code == FlagCodeSet
}

// ClearFlag sets the flag code to 0
func ClearFlag() error {
    conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_POOL_URL"))
    if err != nil {
        return fmt.Errorf("unable to connect to database: %w", err)
    }
    defer conn.Close(context.Background())

    ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
    defer cancel()

    _, err = conn.Exec(ctx, `
        UPDATE flags
        SET code = $1,
            msg = $2,
            updated_at = CURRENT_TIMESTAMP
        WHERE name = $3
    `, FlagCodeUnset, "Flag cleared", FlagOpenAIRefill)

    if err != nil {
        return fmt.Errorf("failed to clear flag: %w", err)
    }

    log.Printf("Flag cleared successfully")
    return nil
}
