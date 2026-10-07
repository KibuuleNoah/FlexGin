package httpx

import (
	"strconv"

	"flexgin/internal/apperror"

	"github.com/gin-gonic/gin"
)

const (
	defaultPerPage = 20
	maxPerPage     = 100
)

type Page struct {
	Page    int
	PerPage int
}

func (p Page) Limit() int  { return p.PerPage }
func (p Page) Offset() int { return (p.Page - 1) * p.PerPage }

func (p Page) Meta(total int) Meta {
	return Meta{Page: p.Page, PerPage: p.PerPage, Total: total}
}

// ParsePage reads ?page= and ?per_page= with defaults and an upper bound.
func ParsePage(c *gin.Context) (Page, error) {
	p := Page{Page: 1, PerPage: defaultPerPage}
	fields := map[string]string{}

	if v := c.Query("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			fields["page"] = "must be a positive integer"
		} else {
			p.Page = n
		}
	}
	if v := c.Query("per_page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxPerPage {
			fields["per_page"] = "must be between 1 and " + strconv.Itoa(maxPerPage)
		} else {
			p.PerPage = n
		}
	}

	if len(fields) > 0 {
		return Page{}, apperror.Validation("invalid pagination parameters", fields)
	}
	return p, nil
}
