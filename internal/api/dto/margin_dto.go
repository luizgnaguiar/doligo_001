package dto

type ProductMarginReportRequest struct {
	ProductID string `param:"productID" validate:"required,uuid"`
	StartDate string `query:"startDate" validate:"required,datetime=2006-01-02"`
	EndDate   string `query:"endDate" validate:"required,datetime=2006-01-02"`
}

type OverallMarginReportRequest struct {
	StartDate string `query:"startDate" validate:"required,datetime=2006-01-02"`
	EndDate   string `query:"endDate" validate:"required,datetime=2006-01-02"`
}
