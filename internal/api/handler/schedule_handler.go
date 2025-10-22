package handler

import (
	"net/http"
	"time"

	"github.com/goflow-atom/goflow-service/internal/api/dto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ScheduleHandler handles schedule-related endpoints
type ScheduleHandler struct {
	// scheduleService service.ScheduleService // TODO: inject service when implemented
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler() *ScheduleHandler {
	return &ScheduleHandler{}
}

// ListSchedules handles GET /api/v1/schedules
// @Summary List Schedules
// @Description Retrieve all scheduled workflows for the authenticated user
// @Tags Schedules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ListSchedulesResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/schedules [get]
// @ID listSchedules
func (h *ScheduleHandler) ListSchedules(c *gin.Context) {
	// TODO: Call service layer to fetch schedules
	// Mock response
	nextRun := time.Now().Add(24 * time.Hour)
	response := dto.ListSchedulesResponse{
		Schedules: []dto.ScheduleResponse{
			{
				ID:         uuid.New().String(),
				WorkflowID: uuid.New().String(),
				Cron:       "0 9 * * 1-5",
				Input:      map[string]interface{}{"automated": true},
				Timezone:   "UTC",
				Enabled:    true,
				NextRunAt:  &nextRun,
				CreatedAt:  time.Now().Add(-7 * 24 * time.Hour),
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

// CreateSchedule handles POST /api/v1/schedules
// @Summary Create Schedule
// @Description Schedule a workflow to run on a cron pattern
// @Tags Schedules
// @Accept json
// @Produce json
// @Param schedule body dto.CreateScheduleRequest true "Schedule configuration"
// @Security BearerAuth
// @Success 201 {object} dto.ScheduleResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/schedules [post]
// @ID createSchedule
func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	var req dto.CreateScheduleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// TODO: Validate cron expression
	// TODO: Call service layer to create schedule

	nextRun := time.Now().Add(24 * time.Hour)
	response := dto.ScheduleResponse{
		ID:         uuid.New().String(),
		WorkflowID: req.WorkflowID,
		Cron:       req.Cron,
		Input:      req.Input,
		Timezone:   req.Timezone,
		Enabled:    req.Enabled,
		NextRunAt:  &nextRun,
		CreatedAt:  time.Now(),
	}

	c.JSON(http.StatusCreated, response)
}

// GetSchedule handles GET /api/v1/schedules/:id
// @Summary Get Schedule
// @Description Retrieve details of a specific schedule
// @Tags Schedules
// @Accept json
// @Produce json
// @Param id path string true "Schedule ID" format(uuid)
// @Security BearerAuth
// @Success 200 {object} dto.ScheduleResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/schedules/{id} [get]
// @ID getSchedule
func (h *ScheduleHandler) GetSchedule(c *gin.Context) {
	id := c.Param("id")

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_UUID",
				Message: "Invalid schedule ID format",
			},
		})
		return
	}

	// TODO: Call service layer to fetch schedule
	nextRun := time.Now().Add(24 * time.Hour)
	response := dto.ScheduleResponse{
		ID:         id,
		WorkflowID: uuid.New().String(),
		Cron:       "0 9 * * 1-5",
		Input:      map[string]interface{}{"automated": true},
		Timezone:   "UTC",
		Enabled:    true,
		NextRunAt:  &nextRun,
		CreatedAt:  time.Now().Add(-7 * 24 * time.Hour),
	}

	c.JSON(http.StatusOK, response)
}

// DeleteSchedule handles DELETE /api/v1/schedules/:id
// @Summary Delete Schedule
// @Description Remove a schedule
// @Tags Schedules
// @Accept json
// @Produce json
// @Param id path string true "Schedule ID" format(uuid)
// @Security BearerAuth
// @Success 204 "Schedule deleted"
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/schedules/{id} [delete]
// @ID deleteSchedule
func (h *ScheduleHandler) DeleteSchedule(c *gin.Context) {
	id := c.Param("id")

	// Validate UUID
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_UUID",
				Message: "Invalid schedule ID format",
			},
		})
		return
	}

	// TODO: Call service layer to delete schedule

	c.Status(http.StatusNoContent)
}
