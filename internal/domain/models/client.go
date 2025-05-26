package models

import (
	"fmt"
	"time"
)

type Client struct {
	ID        int               `json:"id"`
	Clave     string            `json:"clave"`
	Nombre    string            `json:"nombre"`
	Correo    string            `json:"correo"`
	Telefono  string            `json:"telefono"`
	Errors    map[string]string `json:"errors,omitempty"`
	IsValid   bool              `json:"is_valid"`
	RowNumber int               `json:"row_number"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type ClientFilter struct {
	Clave     string `json:"clave,omitempty"`
	Nombre    string `json:"nombre,omitempty"`
	Correo    string `json:"correo,omitempty"`
	Telefono  string `json:"telefono,omitempty"`
	HasErrors *bool  `json:"has_errors,omitempty"`
	Page      int    `json:"page,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ClientStats struct {
	Total         int            `json:"total"`
	Valid         int            `json:"valid"`
	Invalid       int            `json:"invalid"`
	ErrorsByField map[string]int `json:"errors_by_field"`
}

// Métodos del modelo Client
func (c *Client) AddError(field, message string) {
	if c.Errors == nil {
		c.Errors = make(map[string]string)
	}
	c.Errors[field] = message
	c.IsValid = false
}

func (c *Client) ClearErrors() {
	c.Errors = make(map[string]string)
	c.IsValid = true
}

func (c *Client) HasError(field string) bool {
	_, exists := c.Errors[field]
	return exists
}

func (c *Client) GetError(field string) string {
	if err, exists := c.Errors[field]; exists {
		return err
	}
	return ""
}

func (c *Client) String() string {
	return fmt.Sprintf("Client{ID: %d, Clave: %s, Nombre: %s, Valid: %t}",
		c.ID, c.Clave, c.Nombre, c.IsValid)
}
