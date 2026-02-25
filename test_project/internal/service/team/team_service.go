package team

import (
	"context"
	"fmt"
	"math/rand/v2"

	"avito-intership-2025/internal/http/api"
	"avito-intership-2025/internal/models"
	repo "avito-intership-2025/internal/repository"
	"avito-intership-2025/internal/service"
)

type TeamProvider interface {
	Create(ctx context.Context, teamName string) (int, error)
	GetByTeamName(ctx context.Context, teamName string) (*models.Team, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=UserProvider
type UserProvider interface {
	Save(ctx context.Context, user *models.User) (string, error)
	GetUsersInTeam(ctx context.Context, teamName string) ([]*models.User, error)
	GetById(ctx context.Context, userID string) (*models.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) error
	GetActiveUsersIDInTeam(ctx context.Context, teamID int) ([]string, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=PrProvider
type PrProvider interface {
	GetUserReviews(ctx context.Context, userID string) ([]*models.PullRequest, error)
	GetPrReviewers(ctx context.Context, prID string) ([]string, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) error
	DeleteReviewer(ctx context.Context, prID, userID string) error
}

type TeamService struct {
	teamProvider TeamProvider
	userProvider UserProvider
	prProvider   PrProvider
	trm          service.TransactionManager
}

func NewTeamService(
	trm service.TransactionManager,
	teamProvider TeamProvider,
	userProvider UserProvider,
	prProvider PrProvider,
) *TeamService {
	return &TeamService{
		teamProvider: teamProvider,
		userProvider: userProvider,
		prProvider:   prProvider,
		trm:          trm,
	}
}

func (s *TeamService) Add(ctx context.Context, teamName string, users []api.TeamMember) (*api.TeamSchema, error) {
	resp := &api.TeamSchema{}
	members := make([]api.TeamMember, 0, len(users))

	err := s.trm.Do(ctx, func(ctx context.Context) error {
		teamID, err := s.teamProvider.Create(ctx, teamName)
		if err != nil {
			return err
		}

		for _, u := range users {
			user := &models.User{
				ID:       u.UserID,
				Name:     u.Username,
				TeamID:   teamID,
				IsActive: u.IsActive,
			}

			_, err := s.userProvider.Save(ctx, user)
			if err != nil {
				return err
			}

			member := api.TeamMember{
				UserID:   user.ID,
				Username: user.Name,
				IsActive: user.IsActive,
			}

			members = append(members, member)
		}

		resp.TeamName = teamName
		resp.Members = members

		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *TeamService) Get(ctx context.Context, teamName string) (*api.TeamSchema, error) {
	resp := &api.TeamSchema{}

	_, err := s.teamProvider.GetByTeamName(ctx, teamName)
	if err != nil {
		return nil, err
	}

	users, err := s.userProvider.GetUsersInTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}

	members := make([]api.TeamMember, 0, len(users))
	for _, u := range users {
		member := api.TeamMember{
			UserID:   u.ID,
			Username: u.Name,
			IsActive: u.IsActive,
		}

		members = append(members, member)
	}

	resp.TeamName = teamName
	resp.Members = members

	return resp, nil
}

// DeactivateUsers массово деактивирует пользователей команды и переназначает их PR
func (s *TeamService) DeactivateUsers(ctx context.Context, teamName string, userIDs []string) (*api.DeactivateUsersResponse, error) {
	resp := &api.DeactivateUsersResponse{
		DeactivatedUsers: []string{},
		ReassignedPRs:    0,
	}

	err := s.trm.Do(ctx, func(ctx context.Context) error {
		// 1. Проверяем существование команды
		team, err := s.teamProvider.GetByTeamName(ctx, teamName)
		if err != nil {
			return err
		}

		// 2. Проверяем что все пользователи принадлежат указанной команде
		for _, userID := range userIDs {
			user, err := s.userProvider.GetById(ctx, userID)
			if err != nil {
				return err
			}
			if user.TeamID != team.ID {
				return fmt.Errorf("user %s does not belong to team %s", userID, teamName)
			}
		}

		// 3. Собираем все открытые PR где эти пользователи - ревьюверы
		prToReassign := make(map[string][]string) // prID -> []userIDs to replace

		for _, userID := range userIDs {
			prs, err := s.prProvider.GetUserReviews(ctx, userID)
			if err != nil {
				return err
			}

			for _, pr := range prs {
				// Только открытые PR
				if pr.Status == "OPEN" {
					if _, exists := prToReassign[pr.ID]; !exists {
						prToReassign[pr.ID] = []string{}
					}
					prToReassign[pr.ID] = append(prToReassign[pr.ID], userID)
				}
			}
		}

		// 4. Переназначаем каждый PR
		for prID, reviewersToReplace := range prToReassign {
			// Получаем информацию о PR
			allReviewers, err := s.prProvider.GetPrReviewers(ctx, prID)
			if err != nil {
				return err
			}

			// Нужно получить автора PR чтобы исключить его из кандидатов
			var authorID string
			for _, userID := range reviewersToReplace {
				prs, err := s.prProvider.GetUserReviews(ctx, userID)
				if err != nil {
					continue
				}
				for _, pr := range prs {
					if pr.ID == prID {
						authorID = pr.AuthorId
						break
					}
				}
				if authorID != "" {
					break
				}
			}

			// Получаем список активных пользователей команды
			activeCandidates, err := s.userProvider.GetActiveUsersIDInTeam(ctx, team.ID)
			if err != nil {
				return err
			}

			// Переназначаем каждого деактивируемого ревьювера
			for _, oldReviewerID := range reviewersToReplace {
				// Фильтруем кандидатов: исключаем автора, текущих ревьюверов и деактивируемых
				excludeList := append([]string{authorID}, allReviewers...)
				excludeList = append(excludeList, userIDs...)

				available := filterCandidates(activeCandidates, excludeList...)

				if len(available) > 0 {
					// Выбираем случайного кандидата
					newReviewerID := selectRandom(available)

					err = s.prProvider.ReassignReviewer(ctx, prID, oldReviewerID, newReviewerID)
					if err != nil {
						return err
					}
					resp.ReassignedPRs++
					// Обновляем список текущих ревьюверов чтобы не назначить дважды
					allReviewers = append(allReviewers, newReviewerID)
				} else {
					// Нет доступных кандидатов - возвращаем ошибку
					return repo.ErrNoCandidate
				}
			}
		}

		// 5. Деактивируем пользователей
		for _, userID := range userIDs {
			err := s.userProvider.SetIsActive(ctx, userID, false)
			if err != nil {
				// Если не удалось деактивировать - это критичная ошибка, откатываем транзакцию
				return fmt.Errorf("failed to deactivate user %s: %w", userID, err)
			}
			resp.DeactivatedUsers = append(resp.DeactivatedUsers, userID)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func filterCandidates(candidates []string, excludeIDs ...string) []string {
	excludeMap := make(map[string]struct{}, len(excludeIDs))
	for _, id := range excludeIDs {
		excludeMap[id] = struct{}{}
	}

	var filtered []string
	for _, candidate := range candidates {
		if _, excluded := excludeMap[candidate]; !excluded {
			filtered = append(filtered, candidate)
		}
	}
	return filtered
}

func selectRandom(items []string) string {
	if len(items) == 0 {
		return ""
	}
	if len(items) == 1 {
		return items[0]
	}
	return items[rand.IntN(len(items))] //nolint:gosec
}
