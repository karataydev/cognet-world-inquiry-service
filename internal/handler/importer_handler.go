package handler

import (
	"bufio"
	"cognet-world-inquiry-service/internal/service"

	"github.com/gofiber/fiber/v2"
)

type ImportHandler struct {
	dataImporter service.DataImporter
}

func NewImportHandler(dataImporter service.DataImporter) *ImportHandler {
	return &ImportHandler{
		dataImporter: dataImporter,
	}
}

func (h *ImportHandler) ImportTSV(c *fiber.Ctx) error {
	// Get the file from form data
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded: " + err.Error(),
		})
	}

	// Open the uploaded file
	uploadedFile, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open uploaded file: " + err.Error(),
		})
	}
	defer uploadedFile.Close()

	// Use a buffered reader
	reader := bufio.NewReaderSize(uploadedFile, 1024*1024) // 1MB buffer

	// Start the import process
	if err := h.dataImporter.ImportFromReader(c.Context(), reader); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Import completed successfully",
	})
}

func (h *ImportHandler) GetStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": h.dataImporter.GetImportStatus(),
	})
}

func (h *ImportHandler) ClearDatabase(c *fiber.Ctx) error {
	if err := h.dataImporter.ClearDatabase(c.Context()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Database cleared successfully",
	})
}
