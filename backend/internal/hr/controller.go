package hr

import (
	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httprequest"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
)

const idParam = "id"

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) ListEmployees(ctx *gin.Context) {
	var query EmployeeQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListEmployees(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) SummarizeEmployees(ctx *gin.Context) {
	summary, err := c.service.SummarizeEmployees(ctx.Request.Context())
	httpresponse.Respond(ctx, summary, err)
}

func (c *Controller) GetEmployee(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	employee, err := c.service.GetEmployee(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, employee, err)
}

func (c *Controller) CreateEmployee(ctx *gin.Context) {
	var request EmployeeRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	employee, err := c.service.CreateEmployee(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, employee, err)
}

func (c *Controller) UpdateEmployee(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request EmployeeRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	employee, err := c.service.UpdateEmployee(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, employee, err)
}

func (c *Controller) DeleteEmployee(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteEmployee(ctx.Request.Context(), id))
}

func (c *Controller) ListAttendances(ctx *gin.Context) {
	var query AttendanceQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListAttendances(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetAttendance(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	attendance, err := c.service.GetAttendance(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, attendance, err)
}

func (c *Controller) CreateAttendance(ctx *gin.Context) {
	var request AttendanceRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	attendance, err := c.service.CreateAttendance(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, attendance, err)
}

func (c *Controller) UpdateAttendance(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request AttendanceRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	attendance, err := c.service.UpdateAttendance(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, attendance, err)
}

func (c *Controller) DeleteAttendance(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeleteAttendance(ctx.Request.Context(), id))
}

func (c *Controller) ListPayrolls(ctx *gin.Context) {
	var query PayrollQuery
	if err := httprequest.BindQuery(ctx, &query); err != nil {
		_ = ctx.Error(err)
		return
	}
	page, err := c.service.ListPayrolls(ctx.Request.Context(), query)
	httpresponse.Respond(ctx, page, err)
}

func (c *Controller) GetPayroll(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	payroll, err := c.service.GetPayroll(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, payroll, err)
}

func (c *Controller) CreatePayroll(ctx *gin.Context) {
	var request PayrollRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	payroll, err := c.service.CreatePayroll(ctx.Request.Context(), request)
	httpresponse.RespondCreated(ctx, payroll, err)
}

func (c *Controller) UpdatePayroll(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	var request PayrollRequest
	if err := httprequest.BindJSON(ctx, &request); err != nil {
		_ = ctx.Error(err)
		return
	}
	payroll, err := c.service.UpdatePayroll(ctx.Request.Context(), id, request)
	httpresponse.Respond(ctx, payroll, err)
}

func (c *Controller) PayPayroll(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	payroll, err := c.service.PayPayroll(ctx.Request.Context(), id)
	httpresponse.Respond(ctx, payroll, err)
}

func (c *Controller) DeletePayroll(ctx *gin.Context) {
	id, err := httprequest.PathID(ctx, idParam)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	httpresponse.RespondNoContent(ctx, c.service.DeletePayroll(ctx.Request.Context(), id))
}
