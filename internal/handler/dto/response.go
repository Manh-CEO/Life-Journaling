package dto

// Response is the standard API response envelope.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta holds pagination metadata.
type Meta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// NewSuccessResponse creates a success response with data.
func NewSuccessResponse(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

// NewPaginatedResponse creates a success response with pagination metadata.
func NewPaginatedResponse(data interface{}, total, limit, offset int) Response {
	return Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}
}

// NewErrorResponse creates an error response.
func NewErrorResponse(message string) Response {
	return Response{
		Success: false,
		Error:   message,
	}
}
