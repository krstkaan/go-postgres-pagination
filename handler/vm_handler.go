package handler

import (
	"net/http"
	"strconv"
	"strings"

	"go-postgres-pagination/service"

	"github.com/gin-gonic/gin"
)

type VMHandler struct {
	service service.VMService
}

func NewVMHandler(s service.VMService) *VMHandler {
	return &VMHandler{s}
}

func (h *VMHandler) GetVMs(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	// Multi-filter support
	filters := map[string][]string{}
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 && len(key) > 7 && key[:7] == "search[" && key[len(key)-1:] == "]" {
			column := key[7 : len(key)-1]
			// Split values by comma
			filterValues := []string{}
			for _, raw := range values {
				for _, val := range splitCSV(raw) {
					filterValues = append(filterValues, val)
				}
			}
			filters[column] = filterValues
		}
	}

	result, err := h.service.GetPaginatedVMs(c.Request.Context(), page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *VMHandler) SearchVMs(c *gin.Context) {
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

	// Security: only allow searching in specific columns
	allowedColumns := map[string]bool{
		"vmid":             true,
		"hypervisor":       true,
		"name":             true,
		"status":           true,
		"node":             true,
		"cluster":          true,
		"datacenter":       true,
		"guestos":          true,
		"hypervisor_agent": true,
		"tags":             true,
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

func (h *VMHandler) GetVMByID(c *gin.Context) {
	vmid := c.Param("id")
	if vmid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid VM ID"})
		return
	}

	vm, err := h.service.GetVMByID(c.Request.Context(), vmid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if vm == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "VM not found"})
		return
	}

	c.JSON(http.StatusOK, vm)
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
