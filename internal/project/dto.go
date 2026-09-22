package project

import "time"

type CreateProjectRequest struct {
	Name string `json:"name" binding:"required,min=2,max=150"`
	Description string `json:"description"`
}

type UpdateProjectRequest struct {
	Name string `json:"name" binding:"required,min=2,max=150"`
	Description string `json:"description"`
}

type ProjectResponse struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
	OwnerID string `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}