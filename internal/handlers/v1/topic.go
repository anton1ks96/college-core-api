package v1

import (
	"net/http"
	"strconv"

	"github.com/anton1ks96/college-core-api/internal/domain"
	"github.com/gin-gonic/gin"
)

func (h *Handler) createTopic(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req domain.CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	topic, err := h.services.Topic.CreateTopic(
		c.Request.Context(),
		userID.(string),
		req.Title,
		req.Description,
		req.StudentIDs,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          topic.ID,
		"title":       topic.Title,
		"description": topic.Description,
		"created_at":  topic.CreatedAt,
		"message":     "Topic created successfully",
	})
}

func (h *Handler) getMyTopics(c *gin.Context) {
	userID, _ := c.Get("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	topics, total, err := h.services.Topic.GetMyTopics(
		c.Request.Context(),
		userID.(string),
		page,
		limit,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"topics": topics,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

func (h *Handler) getAssignedTopics(c *gin.Context) {
	userID, _ := c.Get("user_id")

	topics, err := h.services.Topic.GetAssignedTopics(
		c.Request.Context(),
		userID.(string),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"assignments": topics,
	})
}

func (h *Handler) addStudentsToTopic(c *gin.Context) {
	topicID := c.Param("id")
	if topicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "topic id is required",
		})
		return
	}

	userID, _ := c.Get("user_id")

	var req domain.AddStudentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	err := h.services.Topic.AddStudents(
		c.Request.Context(),
		topicID,
		userID.(string),
		req.StudentIDs,
	)

	if err != nil {
		if err.Error() == "access denied: only topic creator can add students" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
			return
		}
		if err.Error() == "topic not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "topic not found",
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
		"message": "Students added successfully",
	})
}

func (h *Handler) getTopicStudents(c *gin.Context) {
	topicID := c.Param("id")
	if topicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "topic id is required",
		})
		return
	}

	userID, _ := c.Get("user_id")

	students, err := h.services.Topic.GetTopicStudents(
		c.Request.Context(),
		topicID,
		userID.(string),
	)

	if err != nil {
		if err.Error() == "access denied: only topic creator can view students" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
			return
		}
		if err.Error() == "topic not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "topic not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"students": students,
	})
}

func (h *Handler) searchStudents(c *gin.Context) {
	var req domain.SearchStudentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	students, err := h.services.Topic.SearchStudents(
		c.Request.Context(),
		req.Query,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"students": students,
	})
}