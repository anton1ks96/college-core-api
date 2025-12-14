package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/anton1ks96/college-core-api/internal/domain"
	"github.com/gin-gonic/gin"
)

func (h *Handler) createSavedChat(c *gin.Context) {
	datasetID := c.Param("id")
	if datasetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "dataset id is required",
		})
		return
	}

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("role")

	var req domain.CreateSavedChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	chat, err := h.services.SavedChat.CreateChat(
		c.Request.Context(),
		datasetID,
		userID.(string),
		username.(string),
		role.(string),
		req.Title,
		req.Messages,
	)

	if err != nil {
		if err.Error() == "dataset not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "dataset not found",
			})
			return
		}
		if strings.Contains(err.Error(), "access denied") {
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

	c.JSON(http.StatusCreated, chat)
}

func (h *Handler) getSavedChats(c *gin.Context) {
	datasetID := c.Param("id")
	if datasetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "dataset id is required",
		})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	response, err := h.services.SavedChat.GetChatsByDataset(
		c.Request.Context(),
		datasetID,
		userID.(string),
		role.(string),
		page,
		limit,
	)

	if err != nil {
		if err.Error() == "dataset not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "dataset not found",
			})
			return
		}
		if strings.Contains(err.Error(), "access denied") {
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

	c.JSON(http.StatusOK, response)
}

func (h *Handler) getSavedChat(c *gin.Context) {
	chatID := c.Param("chat_id")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "chat id is required",
		})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	chat, err := h.services.SavedChat.GetChat(
		c.Request.Context(),
		chatID,
		userID.(string),
		role.(string),
	)

	if err != nil {
		if err.Error() == "chat not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "chat not found",
			})
			return
		}
		if err.Error() == "dataset not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "dataset not found",
			})
			return
		}
		if strings.Contains(err.Error(), "access denied") {
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

	c.JSON(http.StatusOK, chat)
}

func (h *Handler) updateSavedChat(c *gin.Context) {
	chatID := c.Param("chat_id")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "chat id is required",
		})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var req domain.UpdateSavedChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	chat, err := h.services.SavedChat.UpdateChat(
		c.Request.Context(),
		chatID,
		userID.(string),
		role.(string),
		req.Title,
		req.Messages,
	)

	if err != nil {
		if err.Error() == "chat not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "chat not found",
			})
			return
		}
		if err.Error() == "dataset not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "dataset not found",
			})
			return
		}
		if strings.Contains(err.Error(), "access denied") {
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

	c.JSON(http.StatusOK, chat)
}

func (h *Handler) deleteSavedChat(c *gin.Context) {
	chatID := c.Param("chat_id")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "chat id is required",
		})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	err := h.services.SavedChat.DeleteChat(
		c.Request.Context(),
		chatID,
		userID.(string),
		role.(string),
	)

	if err != nil {
		if err.Error() == "chat not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "chat not found",
			})
			return
		}
		if strings.Contains(err.Error(), "access denied") {
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Chat deleted successfully",
	})
}

func (h *Handler) downloadSavedChat(c *gin.Context) {
	chatID := c.Param("chat_id")
	if chatID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "chat id is required",
		})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	content, filename, err := h.services.SavedChat.DownloadChatMarkdown(
		c.Request.Context(),
		chatID,
		userID.(string),
		role.(string),
	)

	if err != nil {
		if err.Error() == "chat not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "chat not found",
			})
			return
		}
		if err.Error() == "dataset not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "dataset not found",
			})
			return
		}
		if strings.Contains(err.Error(), "access denied") {
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

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.Data(http.StatusOK, "text/markdown; charset=utf-8", content)
}
