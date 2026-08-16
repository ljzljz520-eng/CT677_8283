package training

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

var enrollmentHeader = []string{"name", "certificate_id", "course", "hours"}

func ParseEnrollmentCSV(reader io.Reader) ([]EnrollmentInput, error) {
	rows, err := csv.NewReader(reader).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read enrollment CSV: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("read enrollment CSV: missing header")
	}
	if len(rows[0]) != len(enrollmentHeader) {
		return nil, fmt.Errorf("read enrollment CSV: invalid header")
	}
	for index, field := range enrollmentHeader {
		if rows[0][index] != field {
			return nil, fmt.Errorf("read enrollment CSV: column %d must be %q", index+1, field)
		}
	}

	inputs := make([]EnrollmentInput, 0, len(rows)-1)
	for index, row := range rows[1:] {
		if len(row) != len(enrollmentHeader) {
			return nil, fmt.Errorf("read enrollment CSV: row %d has %d columns", index+2, len(row))
		}
		hours, err := strconv.Atoi(row[3])
		if err != nil {
			return nil, fmt.Errorf("read enrollment CSV: row %d has invalid hours: %w", index+2, err)
		}
		inputs = append(inputs, EnrollmentInput{Name: row[0], CertificateID: row[1], Course: row[2], Hours: hours})
	}
	return inputs, nil
}
