package data

import (
	ctx "context"
	"errors"
	"regexp"
	"strings"
)

var (
	emailRegex   = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	subjectRegex = regexp.MustCompile(`^[A-Za-z0-9\s]{1,100}$`)
	messageRegex = regexp.MustCompile(`^[\w\s.,!?()-]{1,1000}$`)
)

// Data defines the Mailbox database interface.
type Data interface {

	// Count returns the number of entries in the mailbo.
	Count(ctx.Context) int64

	// ReadAll fetches batches of entries of arbitrary size.
	// See the implementation in mongo.go.
	ReadAll(ctx.Context, int64, int64) ([]Form, error)

	// Read fetches a single entry by referencing it's ID.
	Read(ctx.Context, string) (Form, error)

	// Create a new mailbox entry.
	Create(ctx.Context, Form) (string, error)

	// Delete a mailbox entry by referencing its ID.
	Delete(ctx.Context, string) error
}

// Form defines a single mailbox entry.
type Form struct {
	ID      string `json:"id" bson:"_id,omitempty"`
	From    string `json:"from" bson:"from" binding:"required"`
	Subject string `json:"subject" bson:"subject" binding:"required"`
	Message string `json:"message" bson:"message" binding:"required"`
}

// FormWithCaptcha encapsulates a Form with a captcha token and
// RemoteIP attribute. It is used to validate Turnstile captchas.
type FormWithCaptcha struct {
	Form
	Captcha  string `json:"captcha" binding:"required"`
	RemoteIP string `json:"remote_ip" binding:"required"`
}

// Validate a form's 'From', 'Subject', and 'Message' fields.
// returns a human-friendly error message.
func (f Form) Validate() error {
	if strings.TrimSpace(f.From) == "" {
		return errors.New("'from' field is required")
	}

	if !emailRegex.MatchString(f.From) {
		return errors.New("'from' field must be a valid email address")
	}

	if strings.TrimSpace(f.Subject) == "" {
		return errors.New("'subject' field is required")
	}

	if !subjectRegex.MatchString(f.Subject) {
		return errors.New("'subject' field must be 1-100 characters long and contain only letters, numbers, and spaces")
	}

	if strings.TrimSpace(f.Message) == "" {
		return errors.New("'message' field is required")
	}

	if !messageRegex.MatchString(f.Message) {
		return errors.New("'message' field must be 1-1000 characters long and contain only letters, numbers, spaces, and basic punctuation")
	}

	return nil
}
