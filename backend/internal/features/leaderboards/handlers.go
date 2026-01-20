package leaderboards

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterLeagueRoutes(group fiber.Router) {
	group.Get("/:id/leaderboard", h.GetLeagueLeaderboard)
}

func (h *Handler) GetLeagueLeaderboard(c *fiber.Ctx) error {
	leagueID, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid league id",
		})
	}

	var seasonYear int
	if seasonYearStr := c.Query("season_year"); seasonYearStr != "" {
		year, err := strconv.Atoi(seasonYearStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid season_year",
			})
		}
		seasonYear = year
	}

	resp, err := h.service.GetLeagueLeaderboard(c.UserContext(), leagueID, seasonYear)
	if err != nil {
		if err.Error() == "league not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "league not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get leaderboard",
		})
	}

	return c.JSON(resp)
}
