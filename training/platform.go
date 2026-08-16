package training

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrPlatformRejected = errors.New("industry platform rejected submission")

type IndustryPlatform interface {
	Submit(context.Context, Submission) error
}

type FixturePlatform struct {
	mu          sync.Mutex
	rejections  map[string]string
	submissions []Submission
}

func NewFixturePlatform(rejections map[string]string) *FixturePlatform {
	copyOfRejections := make(map[string]string, len(rejections))
	for certificateID, reason := range rejections {
		copyOfRejections[certificateID] = reason
	}
	return &FixturePlatform{rejections: copyOfRejections}
}

func (p *FixturePlatform) Submit(_ context.Context, submission Submission) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.submissions = append(p.submissions, submission)
	if reason, rejected := p.rejections[submission.CertificateID]; rejected {
		return fmt.Errorf("%w: %s", ErrPlatformRejected, reason)
	}
	return nil
}

func (p *FixturePlatform) Submissions() []Submission {
	p.mu.Lock()
	defer p.mu.Unlock()

	return append([]Submission(nil), p.submissions...)
}
