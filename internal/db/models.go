package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Comment struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ArticleID  uuid.UUID `gorm:"type:uuid;not null;index" json:"article_id"`
	AuthorName *string   `gorm:"type:text" json:"author_name"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	CreatedAt  time.Time `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
}

type Category struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name string    `gorm:"type:text;not null;uniqueIndex" json:"name"`
	Slug *string   `gorm:"type:text" json:"slug"`
}

type Article struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Slug              string         `gorm:"type:text;not null;uniqueIndex" json:"slug"`
	Title             string         `gorm:"type:text;not null" json:"title"`
	Description       string         `gorm:"type:text;not null" json:"description"`
	Contents          JSONB          `gorm:"type:jsonb;not null" json:"contents"`
	CoverImage        *string        `gorm:"type:text" json:"cover_image"`
	PublishedAt       *time.Time     `gorm:"type:timestamptz" json:"published_at"`
	Tags              pq.StringArray `gorm:"type:text[];not null;default:'{}'" json:"tags"`
	CategoryID        *uuid.UUID     `gorm:"type:uuid;index" json:"category_id"`
	Status            string         `gorm:"type:text;not null;index" json:"status"`
	Type              string         `gorm:"type:text;not null" json:"type"`
	EstimatedReadTime int            `gorm:"not null" json:"estimated_read_time"`
	CreatedAt         time.Time      `gorm:"type:timestamptz;not null;default:now();index" json:"created_at"`
}

type Subscriber struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email            string    `gorm:"type:text;not null;uniqueIndex" json:"email"`
	SubscribedAt     time.Time `gorm:"type:timestamptz;not null;default:now()" json:"subscribed_at"`
	IsVerified       bool      `gorm:"not null;default:false" json:"is_verified"`
	UnsubscribeToken *string   `gorm:"type:text" json:"unsubscribe_token"`
}

type JSONB json.RawMessage

func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 || !json.Valid(j) {
		return nil, errors.New("JSONB must contain valid JSON")
	}
	return string(j), nil
}

func (j *JSONB) Scan(value any) error {
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[:0], v...)
	case string:
		*j = append((*j)[:0], v...)
	case nil:
		*j = nil
	default:
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}
	return nil
}

func (JSONB) GormDataType() string { return "jsonb" }

func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	if !json.Valid(j) {
		return nil, errors.New("JSONB contains invalid JSON")
	}
	return j, nil
}

func (j *JSONB) UnmarshalJSON(data []byte) error {
	if !json.Valid(data) {
		return errors.New("invalid JSON document")
	}
	*j = append((*j)[:0], data...)
	return nil
}

func (c *Category) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (a *Article) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (s *Subscriber) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (c *Comment) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
