package v1

import (
	"net/http"
	"strconv"

	"github.com/anton1ks96/college-core-api/internal/domain"
	"github.com/gin-gonic/gin"
)

func (h *Handler) createTopic(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userName, _ := c.Get("username")

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
		userName.(string),
		req.Title,
		req.Description,
		req.Students,
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

func (h *Handler) getAllTopics(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	topics, total, err := h.services.Topic.GetAllTopics(
		c.Request.Context(),
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
	userName, _ := c.Get("username")
	role, _ := c.Get("role")

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
		userName.(string),
		role.(string),
		req.Students,
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
	role, _ := c.Get("role")

	students, err := h.services.Topic.GetTopicStudents(
		c.Request.Context(),
		topicID,
		userID.(string),
		role.(string),
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

	students, total, err := h.services.Topic.SearchStudents(
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
		"total":    total,
	})
}

func (h *Handler) searchTeachers(c *gin.Context) {
	var req domain.SearchStudentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	teachers, total, err := h.services.Topic.SearchTeachers(
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
		"teachers": teachers,
		"total":    total,
	})
}

func (h *Handler) removeStudentFromTopic(c *gin.Context) {
	topicID := c.Param("id")
	studentID := c.Param("student_id")

	if topicID == "" || studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "topic id and student id are required",
		})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	err := h.services.Topic.RemoveStudent(
		c.Request.Context(),
		topicID,
		studentID,
		userID.(string),
		role.(string),
	)

	if err != nil {
		if err.Error() == "access denied: only topic creator or admin can remove students" {
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
		if err.Error() == "failed to remove student: assignment not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "student assignment not found",
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
		"message": "Student removed from topic successfully",
	})
}
