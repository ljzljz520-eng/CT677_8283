package training

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrTaskNotFound = errors.New("upload task not found")

type storedTask struct {
	progress TaskProgress
	logs     []TaskLog
	done     chan struct{}
}

type MemoryStore struct {
	mu             sync.RWMutex
	recordSequence int
	taskSequence   int
	records        map[string]EnrollmentRecord
	recordOrder    []string
	tasks          map[string]*storedTask
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		records: make(map[string]EnrollmentRecord),
		tasks:   make(map[string]*storedTask),
	}
}

func (s *MemoryStore) addRecord(input EnrollmentInput, issues []ValidationIssue) EnrollmentRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.recordSequence++
	status := RecordPending
	if len(issues) > 0 {
		status = RecordInvalid
	}
	record := EnrollmentRecord{
		ID:              fmt.Sprintf("REC-%04d", s.recordSequence),
		EnrollmentInput: input,
		Status:          status,
		Issues:          append([]ValidationIssue(nil), issues...),
	}
	s.records[record.ID] = record
	s.recordOrder = append(s.recordOrder, record.ID)
	return cloneRecord(record)
}

func (s *MemoryStore) Records() []EnrollmentRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := make([]EnrollmentRecord, 0, len(s.recordOrder))
	for _, id := range s.recordOrder {
		records = append(records, cloneRecord(s.records[id]))
	}
	return records
}

func (s *MemoryStore) pendingRecords() []EnrollmentRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := make([]EnrollmentRecord, 0, len(s.recordOrder))
	for _, id := range s.recordOrder {
		record := s.records[id]
		if record.Status == RecordPending {
			records = append(records, cloneRecord(record))
		}
	}
	return records
}

func (s *MemoryStore) invalidRecords() []EnrollmentRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := make([]EnrollmentRecord, 0)
	for _, id := range s.recordOrder {
		record := s.records[id]
		if record.Status == RecordInvalid {
			records = append(records, cloneRecord(record))
		}
	}
	return records
}

func (s *MemoryStore) createTask(total int) TaskProgress {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.taskSequence++
	progress := TaskProgress{
		TaskID:  fmt.Sprintf("TASK-%04d", s.taskSequence),
		Status:  TaskQueued,
		Total:   total,
		Pending: total,
	}
	s.tasks[progress.TaskID] = &storedTask{progress: progress, done: make(chan struct{})}
	return progress
}

func (s *MemoryStore) startTask(taskID string, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := s.tasks[taskID]
	task.progress.Status = TaskRunning
	task.progress.StartedAt = at
}

func (s *MemoryStore) submitRecord(taskID, recordID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := s.records[recordID]
	record.Status = RecordSubmitted
	s.records[recordID] = record
	task := s.tasks[taskID]
	task.progress.Pending--
	task.progress.Submitted++
}

func (s *MemoryStore) rejectRecord(taskID, recordID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := s.records[recordID]
	record.Status = RecordRejected
	s.records[recordID] = record
	task := s.tasks[taskID]
	task.progress.Pending--
	task.progress.Rejected++
}

func (s *MemoryStore) appendLog(taskID string, at time.Time, level LogLevel, recordID, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := s.tasks[taskID]
	task.logs = append(task.logs, TaskLog{
		Sequence: len(task.logs) + 1,
		Time:     at,
		Level:    level,
		RecordID: recordID,
		Message:  message,
	})
}

func (s *MemoryStore) finishTask(taskID string, status TaskStatus, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := s.tasks[taskID]
	task.progress.Status = status
	task.progress.FinishedAt = at
	close(task.done)
}

func (s *MemoryStore) taskDone(taskID string) (<-chan struct{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return nil, ErrTaskNotFound
	}
	return task.done, nil
}

func (s *MemoryStore) TaskProgress(taskID string) (TaskProgress, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return TaskProgress{}, ErrTaskNotFound
	}
	return task.progress, nil
}

func (s *MemoryStore) TaskLogs(taskID string) ([]TaskLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[taskID]
	if !ok {
		return nil, ErrTaskNotFound
	}
	return append([]TaskLog(nil), task.logs...), nil
}

func cloneRecord(record EnrollmentRecord) EnrollmentRecord {
	record.Issues = append([]ValidationIssue(nil), record.Issues...)
	return record
}
