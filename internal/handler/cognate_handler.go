package handler

import (
	"cognet-world-inquiry-service/internal/service"

	"github.com/gofiber/fiber/v2"
)

type CognateHandler struct {
    cognateSearch service.CognateSearch
}

func NewCognateHandler(cognateSearch service.CognateSearch) *CognateHandler {
    return &CognateHandler{
        cognateSearch: cognateSearch,
    }
}

// GetSuggestions handles prefix-based word suggestions
func (h *CognateHandler) GetSuggestions(c *fiber.Ctx) error {
    prefix := c.Query("prefix")
    if prefix == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "prefix is required",
        })
    }

    suggestions, err := h.cognateSearch.GetWordSuggestions(c.Context(), prefix)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "data": suggestions,
    })
}

// GetByConceptID handles getting cognates by concept ID
func (h *CognateHandler) GetByConceptID(c *fiber.Ctx) error {
    conceptID := c.Params("id")
    if conceptID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "concept ID is required",
        })
    }

    cognates, err := h.cognateSearch.FindByConceptID(c.Context(), conceptID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "data": cognates,
    })
}
