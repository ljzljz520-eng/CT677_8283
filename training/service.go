package training

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrNoPendingRecords = errors.New("no pending records")

type Service struct {
	store    *MemoryStore
	platform IndustryPlatform
	catalog  map[string]CourseRule
	clock    Clock
}

func NewService(store *MemoryStore, platform IndustryPlatform, catalog map[string]CourseRule, clock Clock) *Service {
	catalogCopy := make(map[string]CourseRule, len(catalog))
	for course, rule := range catalog {
		catalogCopy[course] = rule
	}
	return &Service{store: store, platform: platform, catalog: catalogCopy, clock: clock}
}

func (s *Service) Import(ctx context.Context, inputs []EnrollmentInput) (ImportReport, error) {
	report := ImportReport{Total: len(inputs)}
	for _, input := range inputs {
		if err := ctx.Err(); err != nil {
			return report, err
		}
		normalized, issues := validateEnrollment(input, s.catalog)
		s.store.addRecord(normalized, issues)
		if len(issues) == 0 {
			report.Valid++
		} else {
			report.Invalid++
		}
	}
	return report, nil
}

func (s *Service) CorrectionTemplate() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	writer := csv.NewWriter(buffer)
	if err := writer.Write([]string{"record_id", "name", "certificate_id", "course", "hours", "issues"}); err != nil {
		return nil, err
	}
	for _, record := range s.store.invalidRecords() {
		issueCodes := make([]string, len(record.Issues))
		for index, issue := range record.Issues {
			issueCodes[index] = issue.Field + ":" + issue.Code
		}
		row := []string{
			record.ID,
			record.Name,
			record.CertificateID,
			record.Course,
			strconv.Itoa(record.Hours),
			strings.Join(issueCodes, ";"),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (s *Service) StartUpload(ctx context.Context) (TaskProgress, error) {
	if err := ctx.Err(); err != nil {
		return TaskProgress{}, err
	}
	records := s.store.pendingRecords()
	if len(records) == 0 {
		return TaskProgress{}, ErrNoPendingRecords
	}
	task := s.store.createTask(len(records))
	workerContext, cancel := context.WithCancel(ctx)
	go s.runUpload(workerContext, cancel, task.TaskID, records)
	return task, nil
}

func (s *Service) runUpload(ctx context.Context, cancel context.CancelFunc, taskID string, records []EnrollmentRecord) {
	defer cancel()
	if err := ctx.Err(); err != nil {
		s.store.appendLog(taskID, s.clock.Now(), LogError, "", "task canceled before upload")
		s.store.finishTask(taskID, TaskCanceled, s.clock.Now())
		return
	}

	s.store.startTask(taskID, s.clock.Now())
	s.store.appendLog(taskID, s.clock.Now(), LogInfo, "", "upload task started")
	for _, record := range records {
		submission := Submission{
			RecordID:      record.ID,
			Name:          record.Name,
			CertificateID: record.CertificateID,
			Course:        record.Course,
			Hours:         record.Hours,
		}
		if err := s.platform.Submit(ctx, submission); err != nil {
			s.store.rejectRecord(taskID, record.ID)
			s.store.appendLog(taskID, s.clock.Now(), LogError, record.ID, fmt.Sprintf("submission rejected: %v", err))
			cancel()
			break
		}
		s.store.submitRecord(taskID, record.ID)
		s.store.appendLog(taskID, s.clock.Now(), LogInfo, record.ID, "submission accepted")
	}

	status := TaskSucceeded
	message := "upload task completed"
	level := LogInfo
	if ctx.Err() != nil {
		status = TaskCanceled
		message = "upload task canceled"
		level = LogError
	}
	s.store.appendLog(taskID, s.clock.Now(), level, "", message)
	s.store.finishTask(taskID, status, s.clock.Now())
}

func (s *Service) Wait(ctx context.Context, taskID string) (TaskProgress, error) {
	done, err := s.store.taskDone(taskID)
	if err != nil {
		return TaskProgress{}, err
	}
	select {
	case <-done:
		return s.store.TaskProgress(taskID)
	case <-ctx.Done():
		return TaskProgress{}, ctx.Err()
	}
}

func (s *Service) Progress(taskID string) (TaskProgress, error) {
	return s.store.TaskProgress(taskID)
}

func (s *Service) Logs(taskID string) ([]TaskLog, error) {
	return s.store.TaskLogs(taskID)
}
