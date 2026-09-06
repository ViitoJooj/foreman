package http

import (
	"net/http"
	"uuid"

	"github.com/gin-gonic/gin"

	"github.com/ViitoJooj/foreman/internal/core/domain"
	"github.com/ViitoJooj/foreman/internal/core/service"
)

// TaskHandler serves the task endpoints.
type TaskHandler struct {
	create *service.CreateTask
	list   *service.ListTasks
}

// NewTaskHandler wires the handler to its use cases.
func NewTaskHandler(create *service.CreateTask, list *service.ListTasks) *TaskHandler {
	return &TaskHandler{create: create, list: list}
}

// Create handles POST /api/v1/tasks.
func (h *TaskHandler) Create(c *gin.Context) {
	var req createTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company_id"})
		return
	}

	task, err := h.create.Execute(c.Request.Context(), service.CreateTaskInput{
		CompanyID:   companyID,
		Title:       req.Title,
		Description: req.Description,
		Risk:        domain.RiskLevel(req.Risk),
		MaxRetries:  req.MaxRetries,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, newTaskResponse(task))
}

// List handles GET /api/v1/tasks with optional company and state query filters.
func (h *TaskHandler) List(c *gin.Context) {
	var filter service.ListTasksFilter

	if raw := c.Query("company"); raw != "" {
		companyID, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company"})
			return
		}
		filter.CompanyID = companyID
	}
	filter.State = domain.TaskState(c.Query("state"))

	tasks, err := h.list.Execute(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}

	resp := make([]taskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, newTaskResponse(t))
	}
	c.JSON(http.StatusOK, resp)
}
