package reports

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/download", h.DownloadReport)
}

func (h *Handler) DownloadReport(c *gin.Context) {
	var req ReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.DateFrom.IsZero() {
		req.DateFrom = time.Now().AddDate(0, 0, -30)
	}
	if req.DateTo.IsZero() {
		req.DateTo = time.Now()
	}

	fileBytes, ext, err := h.service.GenerateReport(req)

	if err != nil {
		fmt.Printf("❌ ОШИБКА: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed: %v", err)})
		return
	}

	contentType := "application/pdf"
	if ext == "xlsx" {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
	if ext == "csv" {
		contentType = "text/csv"
	}

	filename := fmt.Sprintf("Report_%s_%s.%s", req.Type, time.Now().Format("20060102_1504"), ext)

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, fileBytes)
}
