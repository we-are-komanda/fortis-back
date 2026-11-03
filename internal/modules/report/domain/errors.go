package domain

import "errors"

var ErrInvalidReport = errors.New("invalid report request")
var ErrReportNotFound = errors.New("report not found")
var ErrInvalidSnapshot = errors.New("invalid report snapshot")
