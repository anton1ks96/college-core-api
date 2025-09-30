package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/anton1ks96/college-core-api/internal/domain"
	"github.com/anton1ks96/college-core-api/pkg/logger"
	"github.com/gin-gonic/gin"
)

func (h *Handler) createDataset(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid form data",
		})
		return
	}

	titles := form.Value["title"]
	if len(titles) == 0 || titles[0] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "title is required",
		})
		return
	}
	title := titles[0]

	assignmentIDs := form.Value["assignment_id"]
	if len(assignmentIDs) == 0 || assignmentIDs[0] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "assignment_id is required",
		})
		return
	}
	assignmentID := assignmentIDs[0]

	files := form.File["file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file is required",
		})
		return
	}

	file := files[0]

	if !isMarkdownFile(file.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file must be markdown (.md or .markdown)",
		})
		return
	}

	if file.Size > h.cfg.Limits.MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("file size exceeds limit of %d bytes", h.cfg.Limits.MaxFileSize),
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		logger.Error(fmt.Errorf("failed to open uploaded file: %w", err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to process file",
		})
		return
	}
	defer src.Close()

	dataset, err := h.services.Dataset.Create(c.Request.Context(), userID.(string), username.(string), title, assignmentID, src)
	if err != nil {
		if err.Error() == "assignment not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "assignment not found",
			})
			return
		}
		if err.Error() == "access denied: assignment belongs to another student" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "access denied: assignment belongs to another student",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"dataset_id": dataset.ID,
		"title":      dataset.Title,
		"created_at": dataset.CreatedAt,
		"message":    "Dataset created and queued for indexing",
	})
}

func (h *Handler) getDatasets(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	response, err := h.services.Dataset.GetList(
		c.Request.Context(),
		userID.(string),
		role.(string),
		page,
		limit,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) getDataset(c *gin.Context) {
	datasetID := c.Param("id")
	if datasetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "dataset id is required",
		})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	dataset, err := h.services.Dataset.GetByID(
		c.Request.Context(),
		datasetID,
		userID.(string),
		role.(string),
	)

	if err != nil {
		if err.Error() == "dataset not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "dataset not found",
			})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "access denied",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dataset)
}

func (h *Handler) updateDataset(c *gin.Context) {
	datasetID := c.Param("id")
	if datasetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "dataset id is required",
		})
		return
	}

	userID, _ := c.Get("user_id")

	var req domain.UpdateDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	dataset, err := h.services.Dataset.Update(
		c.Request.Context(),
		datasetID,
		userID.(string),
		req.Title,
		&req.Content,
	)

	if err != nil {
		if err.Error() == "dataset not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "dataset not found",
			})
			return
		}
		if err.Error() == "access denied: only owner can edit dataset" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "only owner can edit dataset",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         dataset.ID,
		"title":      dataset.Title,
		"updated_at": dataset.UpdatedAt,
		"message":    "Dataset updated successfully",
	})
}

func (h *Handler) deleteDataset(c *gin.Context) {
	datasetID := c.Param("id")
	if datasetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "dataset id is required",
		})
		return
	}

	userID, _ := c.Get("user_id")

	err := h.services.Dataset.Delete(
		c.Request.Context(),
		datasetID,
		userID.(string),
	)

	if err != nil {
		if err.Error() == "dataset not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "dataset not found",
			})
			return
		}
		if err.Error() == "access denied: only owner can delete dataset" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "only owner can delete dataset",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Dataset deleted successfully",
	})
}

func (h *Handler) askQuestion(c *gin.Context) {
	datasetID := c.Param("id")
	if datasetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "dataset id is required",
		})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var req domain.AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	response, err := h.services.Dataset.AskQuestion(
		c.Request.Context(),
		datasetID,
		userID.(string),
		role.(string),
		req.Question,
	)

	if err != nil {
		if err.Error() == "dataset not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "dataset not found",
			})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "access denied",
			})
			return
		}
		if err.Error() == "dataset is not indexed yet, please wait" {
			c.JSON(http.StatusPreconditionFailed, gin.H{
				"error": "dataset is not indexed yet, please wait",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) reindexDataset(c *gin.Context) {
	datasetID := c.Param("id")
	if datasetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "dataset id is required",
		})
		return
	}

	userID, _ := c.Get("user_id")

	response, err := h.services.Dataset.Reindex(
		c.Request.Context(),
		datasetID,
		userID.(string),
	)

	if err != nil {
		if err.Error() == "dataset not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "dataset not found",
			})
			return
		}
		if err.Error() == "access denied: only owner can reindex dataset" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "only owner can reindex dataset",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func isMarkdownFile(filename string) bool {
	return len(filename) > 3 && (filename[len(filename)-3:] == ".md" ||
		(len(filename) > 9 && filename[len(filename)-9:] == ".markdown"))
}
