package handler

import "github.com/gin-gonic/gin"

type errorResponse struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

func writeError(c *gin.Context, status int, code, message string, details interface{}) {
    c.AbortWithStatusJSON(status, errorResponse{
        Code:    code,
        Message: message,
        Details: details,
    })
}
