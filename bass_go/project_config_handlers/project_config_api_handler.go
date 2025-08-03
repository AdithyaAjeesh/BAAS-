package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"main.go/project/models"
)

func (ph *ProjectHandler) CreateAPIForProject(c *gin.Context) {
	projectID := c.Param("id")

	var project models.Project
	if err := ph.db.First(&project, "id = ?", projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Project not found",
				"message": "No project with the given ID",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to find project",
			"message": err.Error(),
		})
		return
	}

	var req models.CreateAPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid API request",
			"message": err.Error(),
		})
		return
	}

	api := models.API{
		ID:        uuid.New().String()[:8],
		ProjectID: projectID,
		Name:      req.Name,
		Path:      req.Path,
		Method:    req.Method,
		TableName: req.TableName,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := ph.db.Create(&api).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create API config",
			"message": err.Error(),
		})
		return
	}

	response := models.APIResponse{
		ID:        api.ID,
		Name:      api.Name,
		Path:      api.Path,
		Method:    api.Method,
		TableName: api.TableName,
		Status:    api.Status,
		CreatedAt: api.CreatedAt,
		UpdatedAt: api.UpdatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "API created successfully",
		"api":     response,
	})
}
