package project

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khoirirrosikin/task-management/internal/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	projects := rg.Group("/projects")
	projects.Use(authMiddleware)
	{
		projects.POST("", h.CreateProject)
		projects.GET("", h.ListProjects)
		projects.GET("/:id", h.GetProjectByID)
		projects.PUT("/:id", h.UpdateProject)
		projects.DELETE("/:id", h.DeleteProject)
	}
}

func (h* Handler) CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation error", response.FormatValidationError(err))
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	res, err := h.service.CreateProject(c.Request.Context(), userID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Project created successfully", res)
}

func (h *Handler) ListProjects(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	res, err := h.service.ListProjects(c.Request.Context(), userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Projects retrieved successfully", res)
}

func (h *Handler) GetProjectByID(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID", nil)
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	res, err := h.service.GetProjectByID(c.Request.Context(), projectID, userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Project retrieved successfully", res)
}

func (h *Handler) UpdateProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID", nil)
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Validation error", response.FormatValidationError(err))
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	res, err := h.service.UpdateProject(c.Request.Context(), projectID, userID, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Project updated successfully", res)
}

func (h *Handler) DeleteProject(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid project ID", nil)
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)

	if err := h.service.DeleteProject(c.Request.Context(), projectID, userID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Project deleted successfully", nil)
}