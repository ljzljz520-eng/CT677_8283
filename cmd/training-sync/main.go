package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"example.com/vocational-training/training"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	file, err := os.Open("fixtures/enrollments.csv")
	if err != nil {
		return fmt.Errorf("open fixture: %w", err)
	}
	defer file.Close()

	inputs, err := training.ParseEnrollmentCSV(file)
	if err != nil {
		return err
	}
	store := training.NewMemoryStore()
	platform := training.NewFixturePlatform(nil)
	clock := training.NewStepClock(time.Date(2026, time.January, 2, 9, 0, 0, 0, time.UTC), time.Second)
	service := training.NewService(store, platform, training.DefaultCourseCatalog(), clock)
	report, err := service.Import(context.Background(), inputs)
	if err != nil {
		return fmt.Errorf("import records: %w", err)
	}
	fmt.Printf("import total=%d valid=%d invalid=%d\n", report.Total, report.Valid, report.Invalid)

	template, err := service.CorrectionTemplate()
	if err != nil {
		return fmt.Errorf("create correction template: %w", err)
	}
	fmt.Printf("correction template:\n%s", template)

	task, err := service.StartUpload(context.Background())
	if err != nil {
		return fmt.Errorf("start upload: %w", err)
	}
	progress, err := service.Wait(context.Background(), task.TaskID)
	if err != nil {
		return fmt.Errorf("wait for upload: %w", err)
	}
	encoder := json.NewEncoder(os.Stdout)
	if err := encoder.Encode(progress); err != nil {
		return fmt.Errorf("print progress: %w", err)
	}
	logs, err := service.Logs(task.TaskID)
	if err != nil {
		return fmt.Errorf("read task logs: %w", err)
	}
	for _, entry := range logs {
		fmt.Printf("log sequence=%d level=%s record=%s message=%q\n", entry.Sequence, entry.Level, entry.RecordID, entry.Message)
	}
	return nil
}
