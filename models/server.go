package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type VM struct {
	VMId            string      `json:"vmid"`
	Hypervisor      string      `json:"hypervisor"`
	Name            string      `json:"name"`
	Status          string      `json:"status"`
	CPU             float64     `json:"cpu"`
	Memory          int64       `json:"mem"`
	Disk            int64       `json:"disk"`
	IPAddresses     StringArray `json:"ip" gorm:"type:text[]"`
	Tags            StringArray `json:"tags" gorm:"type:text[]"`
	Node            string      `json:"node"`
	Cluster         string      `json:"cluster"`
	Datacenter      string      `json:"datacenter"`
	GuestOS         string      `json:"guestos"`
	HypervisorAgent string      `json:"hypervisor_agent"`
	Raw             JSONB       `json:"raw"`
	UpdatedAt       time.Time   `json:"updated_at"`
	DeletedAt       *time.Time  `json:"deleted_at"`
}

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}

	return json.Unmarshal(bytes, j)
}

// Custom StringArray type that can handle PostgreSQL array scanning
type StringArray []string

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		// Handle PostgreSQL array format like {item1,item2}
		str := string(v)
		if str == "{}" || str == "" {
			*s = []string{}
			return nil
		}

		// Remove braces and split by comma
		str = strings.Trim(str, "{}")
		if str == "" {
			*s = []string{}
			return nil
		}

		items := strings.Split(str, ",")
		result := make([]string, len(items))
		for i, item := range items {
			result[i] = strings.TrimSpace(item)
		}
		*s = result
		return nil
	case string:
		// Handle string format
		str := v
		if str == "{}" || str == "" {
			*s = []string{}
			return nil
		}

		str = strings.Trim(str, "{}")
		if str == "" {
			*s = []string{}
			return nil
		}

		items := strings.Split(str, ",")
		result := make([]string, len(items))
		for i, item := range items {
			result[i] = strings.TrimSpace(item)
		}
		*s = result
		return nil
	default:
		// Try to use pq.StringArray as fallback
		pqArray := pq.StringArray{}
		if err := pqArray.Scan(value); err == nil {
			*s = []string(pqArray)
			return nil
		}
		return fmt.Errorf("cannot scan %T into StringArray", value)
	}
}

func (s StringArray) Value() (driver.Value, error) {
	return pq.StringArray(s).Value()
}

type PaginatedVMs struct {
	Data       []VM `json:"data"`
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	TotalItems int  `json:"total_items"`
	TotalPages int  `json:"total_pages"`
}

// Legacy Server struct for backward compatibility (can be removed later)
type Server struct {
	ID           int    `json:"id"`
	CustomerName string `json:"customer_name"`
	Name         string `json:"name"`
	IPAddress    string `json:"ip_address"`
	CPU          string `json:"cpu"`
	Memory       string `json:"memory"`
	Status       string `json:"status"`
}

type PaginatedServers struct {
	Data       []Server `json:"data"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
	TotalItems int      `json:"total_items"`
	TotalPages int      `json:"total_pages"`
}
