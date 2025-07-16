package handler

import (
	"net/http"
	"strconv"
	"strings"

	"go-postgres-pagination/service"

	"github.com/gin-gonic/gin"
)

type ServerHandler struct {
	service service.ServerService
}

func NewServerHandler(s service.ServerService) *ServerHandler {
	return &ServerHandler{s}
}

func (h *ServerHandler) GetServers(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	// Çoklu filtre desteği
	filters := map[string][]string{}
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 && len(key) > 7 && key[:7] == "search[" && key[len(key)-1:] == "]" {
			column := key[7 : len(key)-1]
			// Değerleri virgülle ayır
			filterValues := []string{}
			for _, raw := range values {
				for _, val := range splitCSV(raw) {
					filterValues = append(filterValues, val)
				}
			}
			filters[column] = filterValues
		}
	}

	result, err := h.service.GetPaginatedServers(c.Request.Context(), page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func (h *ServerHandler) SearchServers(c *gin.Context) {
	column := c.Query("column")
	query := c.Query("query")
	limitStr := c.DefaultQuery("limit", "10")

	if column == "" || query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "column and query parameters are required"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Güvenlik için sadece belirli kolonlarda arama yapalım
	allowedColumns := map[string]bool{
		"customer_name": true,
		"name":          true,
		"ip_address":    true,
		"cpu":           true,
		"memory":        true,
		"status":        true,
	}

	if !allowedColumns[column] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid column"})
		return
	}

	suggestions, err := h.service.SearchColumnValues(c.Request.Context(), column, query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"suggestions": suggestions,
		"column":      column,
		"query":       query,
	})
}

func (h *ServerHandler) GetServerByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	server, err := h.service.GetServerByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if server == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	c.JSON(http.StatusOK, server)
}