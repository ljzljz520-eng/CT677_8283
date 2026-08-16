package training

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var certificatePattern = regexp.MustCompile(`^[A-Z]{2}[0-9]{8}$`)

func validateEnrollment(input EnrollmentInput, catalog map[string]CourseRule) (EnrollmentInput, []ValidationIssue) {
	input.Name = strings.TrimSpace(input.Name)
	input.CertificateID = strings.ToUpper(strings.TrimSpace(input.CertificateID))
	input.Course = strings.TrimSpace(input.Course)

	issues := make([]ValidationIssue, 0, 4)
	nameLength := utf8.RuneCountInString(input.Name)
	if nameLength < 2 || nameLength > 50 {
		issues = append(issues, ValidationIssue{Field: "name", Code: "invalid_length", Message: "name must contain 2 to 50 characters"})
	}
	if !certificatePattern.MatchString(input.CertificateID) {
		issues = append(issues, ValidationIssue{Field: "certificate_id", Code: "invalid_format", Message: "certificate ID must contain 2 letters and 8 digits"})
	}
	rule, ok := catalog[input.Course]
	if !ok {
		issues = append(issues, ValidationIssue{Field: "course", Code: "unknown_course", Message: "course is not in the training catalog"})
	}
	if input.Hours <= 0 {
		issues = append(issues, ValidationIssue{Field: "hours", Code: "not_positive", Message: "hours must be positive"})
	} else if ok && (input.Hours < rule.MinHours || input.Hours > rule.MaxHours) {
		issues = append(issues, ValidationIssue{Field: "hours", Code: "out_of_range", Message: "hours are outside the course range"})
	}

	return input, issues
}
