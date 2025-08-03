package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"main.go/project/models"
)

type ProjectHandler struct {
	db *gorm.DB
}

func NewProjectHandler(db *gorm.DB) *ProjectHandler {
	return &ProjectHandler{db: db}
}

func (ph *ProjectHandler) CreateProject(c *gin.Context) {
	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}
	projectID := generateProjectID(req.Name)

	project := &models.Project{
		ID:           projectID,
		Name:         req.Name,
		Description:  req.Description,
		DatabaseURL:  req.DatabaseURL,
		DatabaseType: req.DatabaseType,
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := ph.db.Create(project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create project",
			"message": err.Error(),
		})
		return
	}
	response := models.ProjectResponse{
		ID:           project.ID,
		Name:         project.Name,
		Description:  project.Description,
		DatabaseType: project.DatabaseType,
		Status:       project.Status,
		APIs:         []models.APIResponse{},
		CreatedAt:    project.CreatedAt,
		UpdatedAt:    project.UpdatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Project created successfully",
		"project": response,
	})
}

func (ph *ProjectHandler) GetProject(c *gin.Context) {
	projectID := c.Param("id")

	var project models.Project
	if err := ph.db.Preload("APIs").Where("id = ?", projectID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Project not found",
				"message": "Project with this ID does not exist",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get project",
			"message": err.Error(),
		})
		return
	}
	var apiResponses []models.APIResponse
	for _, api := range project.APIs {
		apiResponses = append(apiResponses, models.APIResponse{
			ID:        api.ID,
			Name:      api.Name,
			Path:      api.Path,
			Method:    api.Method,
			TableName: api.TableName,
			Status:    api.Status,
			CreatedAt: api.CreatedAt,
			UpdatedAt: api.UpdatedAt,
		})
	}
	response := models.ProjectResponse{
		ID:           project.ID,
		Name:         project.Name,
		Description:  project.Description,
		DatabaseType: project.DatabaseType,
		DatabaseURL:  project.DatabaseURL,
		Status:       project.Status,
		APIs:         apiResponses,
		CreatedAt:    project.CreatedAt,
		UpdatedAt:    project.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"project": response,
	})
}

func (ph *ProjectHandler) ListProjects(c *gin.Context) {
	var projects []models.Project
	if err := ph.db.Preload("APIs").Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list projects",
			"message": err.Error(),
		})
		return
	}

	var responses []models.ProjectResponse
	for _, project := range projects {
		var apiResponses []models.APIResponse
		for _, api := range project.APIs {
			apiResponses = append(apiResponses, models.APIResponse{
				ID:        api.ID,
				Name:      api.Name,
				Path:      api.Path,
				Method:    api.Method,
				TableName: api.TableName,
				Status:    api.Status,
				CreatedAt: api.CreatedAt,
				UpdatedAt: api.UpdatedAt,
			})
		}

		responses = append(responses, models.ProjectResponse{
			ID:           project.ID,
			Name:         project.Name,
			Description:  project.Description,
			DatabaseType: project.DatabaseType,
			Status:       project.Status,
			APIs:         apiResponses,
			CreatedAt:    project.CreatedAt,
			UpdatedAt:    project.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": responses,
		"total":    len(responses),
	})
}

func (ph *ProjectHandler) UpdateProject(c *gin.Context) {
	projectID := c.Param("id")

	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	var project models.Project
	if err := ph.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Project not found",
				"message": "Project with this ID does not exist",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to find project",
			"message": err.Error(),
		})
		return
	}
	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.DatabaseURL != "" {
		project.DatabaseURL = req.DatabaseURL
	}
	if req.DatabaseType != "" {
		project.DatabaseType = req.DatabaseType
	}
	project.UpdatedAt = time.Now()
	if err := ph.db.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update project",
			"message": err.Error(),
		})
		return
	}

	response := models.ProjectResponse{
		ID:           project.ID,
		Name:         project.Name,
		Description:  project.Description,
		DatabaseType: project.DatabaseType,
		Status:       project.Status,
		APIs:         []models.APIResponse{},
		CreatedAt:    project.CreatedAt,
		UpdatedAt:    project.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project updated successfully",
		"project": response,
	})
}

func (ph *ProjectHandler) DeleteProject(c *gin.Context) {
	projectID := c.Param("id")
	if err := ph.db.Where("project_id = ?", projectID).Delete(&models.API{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete project APIs",
			"message": err.Error(),
		})
		return
	}
	result := ph.db.Where("id = ?", projectID).Delete(&models.Project{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete project",
			"message": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Project not found",
			"message": "Project with this ID does not exist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project deleted successfully",
	})
}

func generateProjectID(name string) string {
	return uuid.New().String()[:8]
}
