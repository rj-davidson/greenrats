package golfers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	golfers := router.Group("/golfers")
	h.RegisterRoutesWithGroup(golfers)
}

func (h *Handler) RegisterRoutesWithGroup(group fiber.Router) {
	group.Get("/:id", h.GetByID)
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "golfer id is required",
		})
	}

	golfer, err := h.service.GetByID(c.UserContext(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidGolferID):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid golfer id",
			})
		case errors.Is(err, ErrGolferNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "golfer not found",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get golfer",
			})
		}
	}

	return c.JSON(GetGolferResponse{Golfer: *golfer})
}
