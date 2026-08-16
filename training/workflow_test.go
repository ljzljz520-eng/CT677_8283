package training_test

import (
	"bytes"
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"example.com/vocational-training/training"
)

func TestFixedFixtureBusinessFlow(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "fixtures", "enrollments.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	inputs, err := training.ParseEnrollmentCSV(file)
	if err != nil {
		t.Fatal(err)
	}
	store := training.NewMemoryStore()
	platform := training.NewFixturePlatform(nil)
	service := training.NewService(store, platform, training.DefaultCourseCatalog(), testClock())
	report, err := service.Import(context.Background(), inputs)
	if err != nil {
		t.Fatal(err)
	}
	if report != (training.ImportReport{Total: 4, Valid: 3, Invalid: 1}) {
		t.Fatalf("unexpected import report: %+v", report)
	}

	template, err := service.CorrectionTemplate()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(bytes.NewReader(template)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("correction template row count = %d, want 2", len(rows))
	}
	wantCorrection := []string{"REC-0003", "Eve Stone", "INVALID", "Safety Operations", "8", "certificate_id:invalid_format"}
	if !reflect.DeepEqual(rows[1], wantCorrection) {
		t.Fatalf("correction row = %v, want %v", rows[1], wantCorrection)
	}

	task, err := service.StartUpload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	progress, err := service.Wait(context.Background(), task.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	wantProgress := training.TaskProgress{
		TaskID:     "TASK-0001",
		Status:     training.TaskSucceeded,
		Total:      3,
		Pending:    0,
		Submitted:  3,
		Rejected:   0,
		StartedAt:  time.Date(2026, time.January, 2, 9, 0, 0, 0, time.UTC),
		FinishedAt: time.Date(2026, time.January, 2, 9, 0, 6, 0, time.UTC),
	}
	if progress != wantProgress {
		t.Fatalf("progress = %+v, want %+v", progress, wantProgress)
	}
	if submissions := platform.Submissions(); len(submissions) != 3 {
		t.Fatalf("submission count = %d, want 3", len(submissions))
	}
	logs, err := service.Logs(task.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 5 {
		t.Fatalf("log count = %d, want 5", len(logs))
	}
	for index, entry := range logs {
		if entry.Sequence != index+1 {
			t.Fatalf("log sequence at index %d = %d", index, entry.Sequence)
		}
	}
}

func TestRejectedTaskPreservesUnattemptedRecords(t *testing.T) {
	inputs := []training.EnrollmentInput{
		{Name: "Alice Chen", CertificateID: "AA00000001", Course: "Safety Operations", Hours: 16},
		{Name: "Bob Stone", CertificateID: "BB00000002", Course: "Safety Operations", Hours: 16},
		{Name: "Cara Miles", CertificateID: "CC00000003", Course: "Safety Operations", Hours: 16},
	}
	store := training.NewMemoryStore()
	platform := training.NewFixturePlatform(map[string]string{"AA00000001": "record is not eligible"})
	service := training.NewService(store, platform, training.DefaultCourseCatalog(), testClock())
	if _, err := service.Import(context.Background(), inputs); err != nil {
		t.Fatal(err)
	}
	task, err := service.StartUpload(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	progress, err := service.Wait(context.Background(), task.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if progress.Status != training.TaskCanceled {
		t.Errorf("task status = %s, want %s", progress.Status, training.TaskCanceled)
	}
	if progress.Pending != 2 || progress.Submitted != 0 || progress.Rejected != 1 {
		t.Errorf("task counts = pending:%d submitted:%d rejected:%d, want pending:2 submitted:0 rejected:1", progress.Pending, progress.Submitted, progress.Rejected)
	}
	if submissions := platform.Submissions(); len(submissions) != 1 {
		t.Errorf("submission count = %d, want 1", len(submissions))
	}
	records := store.Records()
	wantStatuses := []training.RecordStatus{training.RecordRejected, training.RecordPending, training.RecordPending}
	for index, want := range wantStatuses {
		if records[index].Status != want {
			t.Errorf("record %s status = %s, want %s", records[index].ID, records[index].Status, want)
		}
	}
}

func TestValidationReportsEveryField(t *testing.T) {
	store := training.NewMemoryStore()
	service := training.NewService(store, training.NewFixturePlatform(nil), training.DefaultCourseCatalog(), testClock())
	report, err := service.Import(context.Background(), []training.EnrollmentInput{{Name: "X", CertificateID: "bad", Course: "Unknown", Hours: -1}})
	if err != nil {
		t.Fatal(err)
	}
	if report != (training.ImportReport{Total: 1, Invalid: 1}) {
		t.Fatalf("unexpected import report: %+v", report)
	}
	records := store.Records()
	if len(records) != 1 {
		t.Fatalf("record count = %d, want 1", len(records))
	}
	wantFields := []string{"name", "certificate_id", "course", "hours"}
	fields := make([]string, len(records[0].Issues))
	for index, issue := range records[0].Issues {
		fields[index] = issue.Field
	}
	if !reflect.DeepEqual(fields, wantFields) {
		t.Fatalf("issue fields = %v, want %v", fields, wantFields)
	}
}

func testClock() *training.StepClock {
	return training.NewStepClock(time.Date(2026, time.January, 2, 9, 0, 0, 0, time.UTC), time.Second)
}
