package utils

import (
	"net/http"
	"strconv"
	"strings"
)

type FeedQuery struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=20"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=100"`
	Since  string   `json:"since" validate:"datetime"`
	Until  string   `json:"until" validate:"datetime"`
}

func (fq FeedQuery) Parse(r *http.Request) (FeedQuery, error) {
	query := r.URL.Query()

	limit := query.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return fq, nil
		}
		fq.Limit = l
	}

	offset := query.Get("offset")
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return fq, nil
		}
		fq.Offset = o
	}

	sort := query.Get("sort")
	if sort != "" {
		fq.Sort = sort
	}

	tags := query.Get("tags")
	if tags != "" {
		fq.Tags = strings.Split(tags, ",")
	}

	search := query.Get("search")
	if search != "" {
		fq.Search = search
	}

	return fq, nil
}
