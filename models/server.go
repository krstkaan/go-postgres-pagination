package models

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
