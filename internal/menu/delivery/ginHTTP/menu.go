package ginHTTP

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/coquizen/servercarte/domain/authentication"
	"github.com/coquizen/servercarte/domain/menu"
)

type menuHandler struct {
	menuSvc menu.Service
	authSvc authentication.Service
}

// RegisterRoutes sets up menu API endpoint using Gin has the delivery.
func RegisterRoutes(svc menu.Service, authSvc authentication.Service, r *gin.Engine, authMiddleWare gin.HandlerFunc, authorizationMiddleWare gin.HandlerFunc) {
	h := menuHandler{svc, authSvc}
	publicRoutes(r, &h)
	privateRoutes(r, &h, authMiddleWare, authorizationMiddleWare)
}

func publicRoutes(r *gin.Engine, h *menuHandler) {
	menuGroup := r.Group("/api/v1")
	menuViewGroup := menuGroup.Group("")
	menuViewGroup.GET("/menus", h.listMenus)
	menuViewGroup.GET("/sections", h.listSections)
	menuViewGroup.GET("/sections/:id", h.findSectionByID)
	menuViewGroup.GET("/items", h.listItems)
	menuViewGroup.GET("/items/:id", h.findItemByID)
}

func privateRoutes(r *gin.Engine, h *menuHandler, authMiddleWare, authorizationMiddleware gin.HandlerFunc) {
	menuEditGroup := r.Group("/api/v1", authMiddleWare, authorizationMiddleware)
	menuEditGroup.POST("/sections", h.createSection)
	menuEditGroup.PATCH("/sections/:id", h.updateSection)
	menuEditGroup.DELETE("/sections/:id", h.deleteSection)
	menuEditGroup.POST("/items", h.createItem)
	menuEditGroup.PATCH("/items/:id", h.updateItem)
	menuEditGroup.DELETE("/items/:id", h.deleteItem)
}

// ---   Menus  --- //
func (h *menuHandler) listMenus(ctx *gin.Context) {
	menus, err := h.menuSvc.Menus(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": menus})

}

// --- Sections --- //
func (h *menuHandler) listSections(ctx *gin.Context) {
	sections, err := h.menuSvc.Sections(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": sections})

}

func (h *menuHandler) findSectionByID(ctx *gin.Context) {
	rawID := ctx.Param("id")
	log.Printf("ID: %s", rawID)

	section, err := h.menuSvc.SectionByID(ctx, rawID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": section})

}

type newSectionRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Active      bool    `json:"active"`
	Visible     bool    `json:"visible"`
	Type        menu.SectionType     `json:"type"`
	ListOrder   uint    `json:"list_order"`
	SectionID   *string `json:"section_id,omitempty"`
}

type updateSectionRequest struct {
	ID          *string `json:"id,omitempty"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Active      *bool   `json:"active,omitempty"`
	Visible     *bool   `json:"visible,omitempty"`
	Type        *menu.SectionType    `json:"type,omitempty"`
	ListOrder   *uint   `json:"list_order,omitempty"`
}

// createSection creates a new section.
func (h *menuHandler) createSection(ctx *gin.Context) {
	var reqSection newSectionRequest

	if err := ctx.ShouldBindJSON(&reqSection); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	section := menu.Section{
		Title:       reqSection.Title,
		Description: reqSection.Description,
		Active:      reqSection.Active,
		Type:        reqSection.Type,
		Visible:     reqSection.Visible,
		ListOrder:   reqSection.ListOrder,
	}

	if reqSection.SectionID != nil {
		requestSectionUUID, err := uuid.Parse(*reqSection.SectionID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, err)
			return
		}
		section.SectionID = &requestSectionUUID
	}
	if err := h.menuSvc.NewSection(ctx, &section); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, &section)
}

// updateSection update section's data.
func (h *menuHandler) updateSection(ctx *gin.Context) {
	var updatedSection updateSectionRequest

	if err := ctx.ShouldBindJSON(&updatedSection); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rawID := ctx.Param("id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var section = menu.Section{
		Title:       *updatedSection.Title,
		Description: updatedSection.Description,
		Active:      *updatedSection.Active,
		Visible:     *updatedSection.Visible,
		Type:		 *updatedSection.Type,
		ListOrder:   *updatedSection.ListOrder,
	}
	section.ID = id
	if err := h.menuSvc.UpdateSectionContent(ctx, &section); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": section})
}

func (h *menuHandler) deleteSection(ctx *gin.Context) {
	rawID := ctx.Param("id")
	if err := h.menuSvc.DeleteSection(ctx, rawID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "section deleted"})

}

// ---  Item  --- //
func (h *menuHandler) listItems(ctx *gin.Context) {
	items, err := h.menuSvc.Items(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": items})

}

type newItemRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Active      bool    `json:"active"`
	Type        menu.ItemType     `json:"type"`
	ListOrder   uint    `json:"list_order"`
	Price       uint64  `json:"visible"`
	SectionID   string  `json:"section_id"`
}

type updateItemRequest struct {
	ID          *string `json:"id,omitempty"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Active      *bool   `json:"active,omitempty"`
	Type        *menu.ItemType   `json:"type,omitempty"`
	ListOrder   *uint   `json:"list_order,omitempty"`
	Price       *uint64 `json:"visible,omitempty"`
}

// createSection creates a new section.
func (h *menuHandler) createItem(ctx *gin.Context) {
	var req newItemRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sectionUUID, err := uuid.Parse(req.SectionID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	item := menu.Item{

		Title:       req.Title,
		Description: req.Description,
		Active:      req.Active,
		Price:       req.Price,
		Type:        req.Type,
		ListOrder:   req.ListOrder,
		SectionID:   &sectionUUID,
	}

	if err := h.menuSvc.NewItem(ctx, &item); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, &item)
}

// updateSection creates a new section.
func (h *menuHandler) updateItem(ctx *gin.Context) {
	rawID := ctx.Param("id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var updatedItem updateItemRequest
	if err := ctx.ShouldBindJSON(&updatedItem); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var item = menu.Item{
		Title:       *updatedItem.Title,
		Description: updatedItem.Description,
		Active:      *updatedItem.Active,
		Type:        *updatedItem.Type,
		ListOrder:   *updatedItem.ListOrder,
		Price:       *updatedItem.Price,
	}
	item.ID = id

	if err := h.menuSvc.UpdateItemContent(ctx, &item); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": item})
}
func (h *menuHandler) findItemByID(ctx *gin.Context) {
	rawID := ctx.Param("id")
	item, err := h.menuSvc.ItemByID(ctx, rawID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": item})

}

func (h *menuHandler) deleteItem(ctx *gin.Context) {
	rawID := ctx.Param("id")

	if err := h.menuSvc.DeleteItem(ctx, rawID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "item deleted"})

}
