package services

import (
	"errors"
	"log"
	"sort"
	"time"

	"tournament/models"

	"github.com/google/uuid"
)

type TournamentService struct {
	db Database
}

type Database interface {
	GetTournament(id string) (*models.Tournament, error)
	UpdateTournament(id string, input *models.UpdateTournamentInput) error
	AddParticipant(participant *models.TournamentParticipant) error
	DeleteParticipant(tournamentID, userID string) error
	GetParticipant(tournamentID, userID string) (*models.TournamentParticipant, error)
	UpdateParticipant(participant *models.TournamentParticipant) error
	GetParticipants(tournamentID string) ([]*models.TournamentParticipant, error)
	AddResult(result *models.TournamentResult) error
	AddSolvedSudoku(tournamentID, userID, sudokuID string) error
	IsSudokuSolved(tournamentID, userID, sudokuID string) (bool, error)
}

func NewTournamentService(db Database) *TournamentService {
	return &TournamentService{db: db}
}

// RegisterParticipant регистрирует пользователя на турнир
func (s *TournamentService) RegisterParticipant(tournamentID, userID string) error {
	tournament, err := s.db.GetTournament(tournamentID)
	if err != nil {
		log.Println("Error getting tournament:", err)
		return err
	}

	log.Println("Tournament status:", tournament.Status)

	if tournament.Status != models.TournamentStatusPending {
		log.Println("Registration is only available for pending tournaments")
		return errors.New("registration is only available for pending tournaments")
	}

	// Проверяем, не зарегистрирован ли уже участник
	existingParticipant, err := s.db.GetParticipant(tournamentID, userID)
	if err == nil && existingParticipant != nil {
		log.Println("User is already registered for this tournament")
		return errors.New("user is already registered for this tournament")
	}

	participant := &models.TournamentParticipant{
		ID:           uuid.New().String(),
		TournamentID: tournamentID,
		UserID:       userID,
		Score:        0,
		SolvedCount:  0,
		JoinedAt:     time.Now(),
	}

	return s.db.AddParticipant(participant)
}

func (s *TournamentService) DeleteParticipant(tournamentID, userID string) error {
	return s.db.DeleteParticipant(tournamentID, userID)
}

// StartTournament начинает турнир
func (s *TournamentService) StartTournament(tournamentID string) error {
	tournament, err := s.db.GetTournament(tournamentID)
	if err != nil {
		log.Println("Error getting tournament:", err)
		return err
	}

	if tournament.Status != models.TournamentStatusPending {
		return errors.New("only pending tournaments can be started")
	}

	status := models.TournamentStatusActive
	input := &models.UpdateTournamentInput{
		Status: &status,
	}
	return s.db.UpdateTournament(tournamentID, input)
}

type ProgressUpdate struct {
	SudokuID    string
	Difficulty  string
	SolveTimeMs int64
}

// calculateScore рассчитывает очки на основе данных о решенной судоку
// TODO: ДОДУМАТЬ КАК СЧИТАТЬ ОЧКИ
func calculateScore(difficulty string, solveTimeMs int64) int {
	baseScore := 0
	switch difficulty {
	case "easy":
		baseScore = 100
	case "medium":
		baseScore = 200
	case "hard":
		baseScore = 300
	case "expert":
		baseScore = 400
	}

	// Конвертируем время в минуты
	solveTimeMinutes := float64(solveTimeMs) / (1000 * 60)

	// Бонус за скорость (максимум 50% от базового счета)
	timeBonus := 0
	if solveTimeMinutes < 5 {
		timeBonus = baseScore / 2
	} else if solveTimeMinutes < 10 {
		timeBonus = baseScore / 4
	}

	finalScore := baseScore + timeBonus
	return finalScore
}

// UpdateParticipantProgress обновляет прогресс участника
func (s *TournamentService) UpdateParticipantProgress(tournamentID, userID string, update ProgressUpdate) error {
	participant, err := s.db.GetParticipant(tournamentID, userID)
	if err != nil {
		log.Println("Error getting participant:", err)
		return err
	}

	tournament, err := s.db.GetTournament(tournamentID)
	if err != nil {
		log.Println("Error getting tournament:", err)
		return err
	}

	if tournament.Status != models.TournamentStatusActive {
		log.Println("Can only update progress for active tournaments")
		return errors.New("can only update progress for active tournaments")
	}

	// Проверяем, не решал ли участник эту судоку ранее
	isSolved, err := s.db.IsSudokuSolved(tournamentID, userID, update.SudokuID)
	if err != nil {
		log.Println("Error checking if sudoku is solved:", err)
		return err
	}
	if isSolved {
		log.Println("This sudoku has already been solved by the participant")
		return errors.New("this sudoku has already been solved")
	}

	// Рассчитываем очки
	score := calculateScore(update.Difficulty, update.SolveTimeMs)

	// Обновляем статистику
	participant.Score += score
	participant.SolvedCount++
	participant.LastSolvedAt = time.Now()
	participant.LastSolvedSudokuID = update.SudokuID

	// Добавляем запись о решенной судоку
	if err := s.db.AddSolvedSudoku(tournamentID, userID, update.SudokuID); err != nil {
		log.Println("Error adding solved sudoku:", err)
		return err
	}

	return s.db.UpdateParticipant(participant)
}

// FinishTournament завершает турнир и подводит итоги
func (s *TournamentService) FinishTournament(tournamentID string) error {
	tournament, err := s.db.GetTournament(tournamentID)
	if err != nil {
		log.Println("Error getting tournament:", err)
		return err
	}

	if tournament.Status != models.TournamentStatusActive {
		return errors.New("only active tournaments can be finished")
	}

	participants, err := s.db.GetParticipants(tournamentID)
	if err != nil {
		log.Println("Error getting participants:", err)
		return err
	}

	// Сортируем участников по очкам
	sortParticipantsByScore(participants)

	// Создаем результаты для каждого участника
	for rank, participant := range participants {
		result := &models.TournamentResult{
			ID:           uuid.New().String(),
			TournamentID: tournamentID,
			UserID:       participant.UserID,
			Score:        participant.Score,
			Rank:         rank + 1,
			SolvedCount:  participant.SolvedCount,
			FinishedAt:   time.Now(),
		}
		if err := s.db.AddResult(result); err != nil {
			log.Println("Error adding result:", err)
			return err
		}
	}

	status := models.TournamentStatusFinished
	input := &models.UpdateTournamentInput{
		Status: &status,
	}
	return s.db.UpdateTournament(tournamentID, input)
}

func sortParticipantsByScore(participants []*models.TournamentParticipant) {
	// Сортировка по убыванию очков
	sort.Slice(participants, func(i, j int) bool {
		if participants[i].Score == participants[j].Score {
			// При равных очках сортируем по времени последнего решения
			return participants[i].LastSolvedAt.Before(participants[j].LastSolvedAt)
		}
		return participants[i].Score > participants[j].Score
	})
}

// UpdateTournamentStats обновляет статистику турнира tournament_results
func (s *TournamentService) UpdateTournamentStats(tournamentID string) error {
	return nil
}

// GetDashboard возвращает дашбоард турнира
func (s *TournamentService) GetDashboard(tournamentID, userID string) (*models.TournamentDashboard, error) {
	return nil, nil
}
