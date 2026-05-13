package handlers

import (
	"errors"
	"net/http"

	"github.com/czanix/boilerplate-api-go/internal/application"
	"github.com/czanix/boilerplate-api-go/internal/domain"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	createOrder *application.CreateOrderUseCase
}

func NewOrderHandler(createOrder *application.CreateOrderUseCase) *OrderHandler {
	return &OrderHandler{createOrder: createOrder}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var input application.CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	output, err := h.createOrder.Execute(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyOrder) || errors.Is(err, domain.ErrInvalidQuantity) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, output)
}
