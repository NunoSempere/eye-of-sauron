package main

import (

	"context"
	"time"
	"os"
	"github.com/jackc/pgx/v4"
	"log"
	"fmt"
)

func markRelevantPerHumanCheckInServer(state string, id int) error {
	flag := true
	if flag {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_POOL_URL"))
		if err != nil {
			log.Printf("failed to connect to database: %v", err)
			return fmt.Errorf("database connection error: %v", err)
		}
		defer conn.Close(ctx)

		_, err = conn.Exec(ctx, "UPDATE sources SET relevant_per_human_check = $1 WHERE id = $2", state, id)
		if err != nil {
			log.Printf("failed to mark source as relevant: %v", err)
			return fmt.Errorf("database update error: %v", err)
		}
	}
	return nil
}

func (a *App) markRelevantPerHumanCheck(state string, i int) error {
	if len(a.sources) == 0 {
		return nil
	}

	// Toggle processed state in UI immediately
	a.sources[i].RelevantPerHumanCheck = state

	// Update database asynchronously
	a.waitgroup.Add(1)
	go func() {
		defer a.waitgroup.Done()
		err := markRelevantPerHumanCheckInServer(state, a.sources[i].ID)
		if err != nil {
			fmt.Printf("%v", err)
			go func() {
				a.failureMark = true
				time.Sleep(2)
				a.failureMark = false
			}()
			a.sources[i].RelevantPerHumanCheck = state
		}
	}()

	return nil
}

func markProcessedInServer(state bool, id int, source Source) error {
	flag := true
	if flag {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_POOL_URL"))
		if err != nil {
			log.Printf("failed to connect to database: %v", err)
			return fmt.Errorf("database connection error: %v", err)
		}
		defer conn.Close(ctx)

		_, err = conn.Exec(ctx, "UPDATE sources SET processed = $1 WHERE id = $2", state, id)
		if err != nil {
			log.Printf("failed to mark source as processed: %v", err)
			log.Printf("source: %v", source)
			return fmt.Errorf("database update error: %v", err)
		}
	}
	return nil
}

func (a *App) markProcessed(i int, source Source) error {
	if len(a.sources) == 0 {
		return nil
	}

	// Toggle processed state in UI immediately
	newState := !a.sources[i].Processed
	a.sources[i].Processed = newState

	// Update database asynchronously
	a.waitgroup.Add(1)
	go func() {
		defer a.waitgroup.Done()
		err := markProcessedInServer(newState, a.sources[i].ID, source)
		if err != nil {
			log.Printf("%v", err)
			go func() {
				a.failureMark = true
				time.Sleep(2)
				a.failureMark = false
			}()
			a.sources[i].Processed = !newState
		}
	}()

	//
	if a.sources[i].RelevantPerHumanCheck != RELEVANT_PER_HUMAN_CHECK_YES {
		a.markRelevantPerHumanCheck(RELEVANT_PER_HUMAN_CHECK_NO, i)
	}

	return nil
}
