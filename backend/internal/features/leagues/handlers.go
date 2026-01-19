package leagues

import (
	"errors"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"

	"github.com/rj-davidson/greenrats/internal/auth"
	"github.com/rj-davidson/greenrats/internal/email"
)

type Handler struct {
	service *Service
	email   *email.Client
}

func NewHandler(service *Service, emailClient *email.Client) *Handler {
	return &Handler{service: service, email: emailClient}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	leagues := router.Group("/leagues")
	h.RegisterRoutesWithGroup(leagues)
}

func (h *Handler) RegisterRoutesWithGroup(group fiber.Router) {
	group.Post("/", h.Create)
	group.Get("/", h.ListUserLeagues)
	group.Get("/:id", h.GetByID)
	group.Get("/:id/tournaments", h.GetLeagueTournaments)
	group.Post("/join", h.JoinLeague)
	group.Post("/:id/regenerate-code", h.RegenerateJoinCode)
	group.Patch("/:id/joining", h.SetJoiningEnabled)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateLeagueRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "league name is required",
		})
	}

	userID := auth.GetDBUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	league, err := h.service.Create(c.Context(), CreateParams{
		Name:   name,
		UserID: userID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create league",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(CreateLeagueResponse{
		League: *league,
	})
}

func (h *Handler) ListUserLeagues(c *fiber.Ctx) error {
	userID := auth.GetDBUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	resp, err := h.service.ListUserLeagues(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list leagues",
		})
	}

	return c.JSON(resp)
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "league id is required",
		})
	}

	id, err := uuid.FromString(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid league id",
		})
	}

	userID := auth.GetDBUserID(c)

	var league *League
	if userID != uuid.Nil {
		league, err = h.service.GetByIDWithRole(c.Context(), id, userID)
	} else {
		league, err = h.service.GetByID(c.Context(), id)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get league",
		})
	}

	if league == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "league not found",
		})
	}

	return c.JSON(GetLeagueResponse{League: *league})
}

func (h *Handler) GetLeagueTournaments(c *fiber.Ctx) error {
	userID := auth.GetDBUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	leagueID, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid league id",
		})
	}

	resp, err := h.service.GetLeagueTournaments(c.Context(), leagueID, userID)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.JSON(resp)
}

func (h *Handler) JoinLeague(c *fiber.Ctx) error {
	userID := auth.GetDBUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	var req JoinLeagueRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	code := strings.TrimSpace(strings.ToUpper(req.Code))
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "join code is required",
		})
	}

	league, err := h.service.JoinLeague(c.Context(), userID, code)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	if h.email != nil {
		user := auth.GetDBUser(c)
		if user != nil && user.DisplayName != nil {
			go func() {
				if err := h.email.SendLeagueJoin(user.Email, email.LeagueJoinData{
					DisplayName: *user.DisplayName,
					LeagueName:  league.Name,
					IsNewMember: true,
				}); err != nil {
					log.Printf("[LEAGUES] Failed to send league join email: %v", err)
				}
			}()
		}
	}

	return c.Status(fiber.StatusCreated).JSON(JoinLeagueResponse{
		League: *league,
	})
}

func (h *Handler) RegenerateJoinCode(c *fiber.Ctx) error {
	userID := auth.GetDBUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	leagueID, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid league id",
		})
	}

	league, err := h.service.RegenerateJoinCode(c.Context(), leagueID, userID)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.JSON(RegenerateCodeResponse{League: *league})
}

func (h *Handler) SetJoiningEnabled(c *fiber.Ctx) error {
	userID := auth.GetDBUserID(c)
	if userID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication required",
		})
	}

	leagueID, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid league id",
		})
	}

	var req SetJoiningEnabledRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	league, err := h.service.SetJoiningEnabled(c.Context(), leagueID, userID, req.Enabled)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.JSON(SetJoiningEnabledResponse{League: *league})
}

func (h *Handler) handleServiceError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrLeagueNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "league not found",
		})
	case errors.Is(err, ErrInvalidJoinCode):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid join code",
		})
	case errors.Is(err, ErrAlreadyMember):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "you are already a member of this league",
		})
	case errors.Is(err, ErrJoiningDisabled):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "joining is disabled for this league",
		})
	case errors.Is(err, ErrLeagueFull):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "league has reached maximum members",
		})
	case errors.Is(err, ErrNotCommissioner):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "only the commissioner can perform this action",
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}
}
