package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}

	fmt.Println("insertion finished")
}

type User struct {
	DateOfBirth time.Time `db:"date_of_birth"`
}

func run() error {
	db, err := sqlx.Open("mysql", "root:password@tcp(localhost:3306)/users?parseTime=true")
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	dateOfBirth, err := time.Parse(time.DateTime, "2006-01-02 15:04:05")
	if err != nil {
		return err
	}

	group, _ := errgroup.WithContext(context.Background())
	group.SetLimit(20)

	insertLatencyDurationsCh := make(chan int64, 2000)

	var waitGroup sync.WaitGroup

	waitGroup.Add(1)
	go func() {
		var counter int64
		var sum int64
		for duration := range insertLatencyDurationsCh {
			sum += duration
			counter++
		}

		avgLatency := sum / counter

		fmt.Println("AVERAGE INSERTION LATENCY: ", time.Duration(avgLatency))
		waitGroup.Done()
	}()

	batchSize := 1000
	var i uint32
	users := make([]User, 0, batchSize)
	for ; i < uint32(40_000_000); i++ {
		if len(users) == batchSize {
			group.Go(func() error {
				return UpsertBatch(db, users, insertLatencyDurationsCh)
			})
			fmt.Println("passed batch for insert: ", i)
			users = make([]User, 0, batchSize)
		}
		dateOfBirth = dateOfBirth.Add(10 * time.Second)
		users = append(users, User{dateOfBirth})
	}

	if len(users) > 0 {
		group.Go(func() error {
			return UpsertBatch(db, users, insertLatencyDurationsCh)
		})
		fmt.Println("passed last batch for insert...")
	}

	if err := group.Wait(); err != nil {
		return err
	}

	close(insertLatencyDurationsCh)
	waitGroup.Wait()

	return nil
}

func UpsertBatch(db *sqlx.DB, users []User, latencyCh chan<- int64) error {
	if len(users) == 0 {
		return nil
	}

	columns := []string{
		"date_of_birth",
	}

	insertQuery := `INSERT INTO users (` + strings.Join(columns, ",") + `) 
		VALUES (:` + strings.Join(columns, ",:") + `)`

	t1 := time.Now()
	_, err := db.NamedExec(insertQuery, users)
	if err != nil {
		return err
	}
	latencyCh <- time.Since(t1).Nanoseconds()

	return nil
}
