package training

import "time"

type EnrollmentInput struct {
	Name          string
	CertificateID string
	Course        string
	Hours         int
}

type ValidationIssue struct {
	Field   string
	Code    string
	Message string
}

type RecordStatus string

const (
	RecordPending   RecordStatus = "pending"
	RecordInvalid   RecordStatus = "invalid"
	RecordSubmitted RecordStatus = "submitted"
	RecordRejected  RecordStatus = "rejected"
)

type EnrollmentRecord struct {
	ID string
	EnrollmentInput
	Status RecordStatus
	Issues []ValidationIssue
}

type CourseRule struct {
	MinHours int
	MaxHours int
}

type ImportReport struct {
	Total   int
	Valid   int
	Invalid int
}

type TaskStatus string

const (
	TaskQueued    TaskStatus = "queued"
	TaskRunning   TaskStatus = "running"
	TaskSucceeded TaskStatus = "succeeded"
	TaskCanceled  TaskStatus = "canceled"
)

type TaskProgress struct {
	TaskID     string
	Status     TaskStatus
	Total      int
	Pending    int
	Submitted  int
	Rejected   int
	StartedAt  time.Time
	FinishedAt time.Time
}

type LogLevel string

const (
	LogInfo  LogLevel = "info"
	LogError LogLevel = "error"
)

type TaskLog struct {
	Sequence int
	Time     time.Time
	Level    LogLevel
	RecordID string
	Message  string
}

type Submission struct {
	RecordID      string
	Name          string
	CertificateID string
	Course        string
	Hours         int
}

func DefaultCourseCatalog() map[string]CourseRule {
	return map[string]CourseRule{
		"Equipment Inspection": {MinHours: 8, MaxHours: 48},
		"Industrial Welding":   {MinHours: 24, MaxHours: 80},
		"Safety Operations":    {MinHours: 8, MaxHours: 32},
	}
}
