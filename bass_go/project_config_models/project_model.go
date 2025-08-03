package models

import "time"

type Project struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	Description  string    `json:"description"`
	DatabaseURL  string    `json:"database_url" gorm:"not null"`
	DatabaseType string    `json:"database_type"`
	Status       string    `json:"status" gorm:"default:'active'"`
	APIs         []API     `json:"apis" gorm:"foreignKey:ProjectID"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type API struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	ProjectID string    `json:"project_id" gorm:"not null;index"`
	Name      string    `json:"name" gorm:"not null"`
	Path      string    `json:"path" gorm:"not null"`
	Method    string    `json:"method" gorm:"not null"`
	TableName string    `json:"table_name"`
	Status    string    `json:"status" gorm:"default:'active'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateProjectRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	DatabaseURL  string `json:"database_url" binding:"required"`
	DatabaseType string `json:"database_type" binding:"required"`
}

type UpdateProjectRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	DatabaseURL  string `json:"database_url"`
	DatabaseType string `json:"database_type"`
}

type ProjectResponse struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	DatabaseType string        `json:"database_type"`
	DatabaseURL  string        `json:"database_url"`
	Status       string        `json:"status"`
	APIs         []APIResponse `json:"apis"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type CreateAPIRequest struct {
	Name      string `json:"name" binding:"required"`
	Path      string `json:"path" binding:"required"`
	Method    string `json:"method" binding:"required"`
	TableName string `json:"table_name" binding:"required"`
}

type UpdateAPIRequest struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Method    string `json:"method"`
	TableName string `json:"table_name"`
}

type APIResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Method    string    `json:"method"`
	TableName string    `json:"table_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
