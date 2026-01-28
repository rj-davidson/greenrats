package leaderboards

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterLeagueRoutes(group fiber.Router) {
	group.Get("/:id/leaderboard", h.GetLeagueLeaderboard)
	group.Get("/:id/standings", h.GetLeagueStandings)
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

func (h *Handler) GetLeagueStandings(c *fiber.Ctx) error {
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

	includePicks := c.Query("include") == "picks"
	requestingUserID := auth.GetDBUserID(c)

	resp, err := h.service.GetLeagueStandings(c.UserContext(), leagueID, seasonYear, includePicks, requestingUserID)
	if err != nil {
		if err.Error() == "league not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "league not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get standings",
		})
	}

	return c.JSON(resp)
}
