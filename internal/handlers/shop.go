package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/middleware"
	"github.com/north-fy/levelup/internal/services"
)

type createShopItemRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	PriceGold   int    `json:"price_gold"`
}

type updateShopItemRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	PriceGold   *int    `json:"price_gold"`
}

// ShopHandler exposes the shop endpoints.
type ShopHandler struct {
	shop *services.ShopService
}

// NewShopHandler creates the shop handler.
func NewShopHandler(shop *services.ShopService) *ShopHandler {
	return &ShopHandler{shop: shop}
}

// Create godoc
//
//	@Summary	List an item for sale
//	@Tags		shop
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		body	body	createShopItemRequest	true	"Item payload"
//	@Success	201	{object}	domain.ShopItem
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Router		/shop/items [post]
func (h *ShopHandler) Create(c *gin.Context) {
	var req createShopItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
		return
	}

	item, err := h.shop.Create(c.Request.Context(), middleware.UserID(c), services.CreateShopItemInput{
		Title:       req.Title,
		Description: req.Description,
		PriceGold:   req.PriceGold,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

// List godoc
//
//	@Summary	List shop items
//	@Tags		shop
//	@Security	BearerAuth
//	@Produce	json
//	@Param		mine	query	bool	false	"List only the current user's items"
//	@Success	200	{array}	domain.ShopItem
//	@Failure	401	{object}	ErrorResponse
//	@Router		/shop/items [get]
func (h *ShopHandler) List(c *gin.Context) {
	mine := c.Query("mine") == "true"
	items, err := h.shop.List(c.Request.Context(), middleware.UserID(c), mine)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

// Update godoc
//
//	@Summary	Update one of the user's items
//	@Tags		shop
//	@Security	BearerAuth
//	@Accept		json
//	@Produce	json
//	@Param		id		path	uint					true	"Item id"
//	@Param		body	body	updateShopItemRequest	true	"Item fields"
//	@Success	200	{object}	domain.ShopItem
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/shop/items/{id} [patch]
func (h *ShopHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req updateShopItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request"})
		return
	}

	item, err := h.shop.Update(c.Request.Context(), middleware.UserID(c), id, services.UpdateShopItemInput{
		Title:       req.Title,
		Description: req.Description,
		PriceGold:   req.PriceGold,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

// Delete godoc
//
//	@Summary	Deactivate one of the user's items
//	@Tags		shop
//	@Security	BearerAuth
//	@Param		id	path	uint	true	"Item id"
//	@Success	204
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Router		/shop/items/{id} [delete]
func (h *ShopHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.shop.Delete(c.Request.Context(), middleware.UserID(c), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Buy godoc
//
//	@Summary	Buy an item
//	@Tags		shop
//	@Security	BearerAuth
//	@Produce	json
//	@Param		id	path	uint	true	"Item id"
//	@Success	200	{object}	domain.Purchase
//	@Failure	400	{object}	ErrorResponse
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Failure	409	{object}	ErrorResponse
//	@Router		/shop/items/{id}/buy [post]
func (h *ShopHandler) Buy(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	purchase, err := h.shop.Buy(c.Request.Context(), middleware.UserID(c), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, purchase)
}

// Purchases godoc
//
//	@Summary	List the current user's purchase history
//	@Tags		shop
//	@Security	BearerAuth
//	@Produce	json
//	@Success	200	{array}	domain.Purchase
//	@Failure	401	{object}	ErrorResponse
//	@Router		/shop/purchases [get]
func (h *ShopHandler) Purchases(c *gin.Context) {
	purchases, err := h.shop.PurchaseHistory(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, purchases)
}
