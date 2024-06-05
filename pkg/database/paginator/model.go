package paginator

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/radyatamaa/dating-apps-api/pkg/helper"
	"github.com/radyatamaa/dating-apps-api/pkg/response"
)

var (
	PageSizes        = []int{5, 10, 15, 20, 25}
	DEFAULT_PAGESIZE = 10
	MAX_PAGESIZE     = 100
	DEFAULT_PAGE     = 0
)

type MetaPaginatorResponse struct {
	CurrentPage     int    `json:"current_page"`
	PerPage         int    `json:"limit_per_page"`
	PreviousPage    int    `json:"back_page"`
	NextPage        int    `json:"next_page"`
	TotalRecords    int    `json:"total_records"`
	TotalPages      int    `json:"total_pages"`
	LabelPages      string `json:"label_pages"`
	PageSizes       []int  `json:"page_sizes"`
	DefaultPageSize int    `json:"default_page_size"`
}

func Pagination(pageRequest, pageSizeRequest int) (limit, page, offset int) {
	limit = 10
	page = 1
	offset = 0

	page = pageRequest
	limit = pageSizeRequest
	if page == 0 && (limit == 0 || limit == 10) {
		page = 1
		limit = 10
	}

	offset = (page - 1) * limit

	return
}

func (p MetaPaginatorResponse) MappingPaginator(page, limit, offset, totalAllRecords, countData int) MetaPaginatorResponse {
	totalPage := int(math.Ceil(float64(totalAllRecords) / float64(limit)))
	prev := page
	next := page

	if page != 1 {
		prev = page - 1
	}

	if page != totalPage {
		next = page + 1
	}

	p = MetaPaginatorResponse{
		CurrentPage:  page,
		PerPage:      countData,
		PreviousPage: prev,
		NextPage:     next,
		TotalRecords: totalAllRecords,
		TotalPages:   totalPage,
	}

	if totalPage == 0 {
		totalPage = 1
	}

	startPage := 0
	endPage := 0
	if totalAllRecords != 0 && countData != 0 {
		startPage = offset + 1
		endPage = offset + countData
	}
	label := fmt.Sprint(startPage, "-", endPage, " of ", totalAllRecords)
	maxLimitPerPage := totalAllRecords - (startPage - 1)

	p.LabelPages = label
	p.DefaultPageSize = p.PerPage

	p.PageSizes = append(p.PageSizes, p.PerPage)

	for _, pageSz := range PageSizes {
		check := helper.ItemExists(p.PageSizes, pageSz)
		if pageSz <= maxLimitPerPage && !check && (page*limit) <= maxLimitPerPage {
			p.PageSizes = append(p.PageSizes, pageSz)
			continue
		}
	}

	sort.Slice(p.PageSizes, func(i, j int) bool {
		return p.PageSizes[i] < p.PageSizes[j]
	})

	return p
}

func PaginationQueryParamValidation(pageSizeStr, pageStr string) (int, int, error) {
	// default
	var pageSizeDefault = DEFAULT_PAGESIZE
	var pageDefault = DEFAULT_PAGE

	pageSizeInt, err := strconv.Atoi(pageSizeStr)
	if err != nil && pageSizeStr != "" {
		return 0, 0, response.ErrQueryParamInvalid
	}
	if err == nil {
		pageSizeDefault = pageSizeInt
		if pageSizeInt == 0 {
			pageSizeDefault = DEFAULT_PAGESIZE
		}
		// dafault value maximum pageSize = 100
		if pageSizeDefault > MAX_PAGESIZE {
			pageSizeDefault = MAX_PAGESIZE
		}
		if pageSizeInt < 0 {
			return 0, 0, response.ErrQueryParamInvalid
		}
	}

	pageInt, err := strconv.Atoi(pageStr)

	if err != nil && pageStr != "" {
		return 0, 0, response.ErrQueryParamInvalid
	}
	if err == nil {
		pageDefault = pageInt
		if pageInt == 0 {
			pageDefault = DEFAULT_PAGE
		}
		if pageInt < 0 {
			return 0, 0, response.ErrQueryParamInvalid
		}
	}

	return pageSizeDefault, pageDefault, nil
}
