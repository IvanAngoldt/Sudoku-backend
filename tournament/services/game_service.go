package services

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"math/rand"
// 	"net/http"
// 	"tournament/config"
// )

// type GameService struct {
// 	cfg *config.Config
// }

// type SudokuInfo struct {
// 	ID               string `json:"id"`
// 	Complexity       string `json:"complexity"`
// 	CreatedAt        string `json:"created_at"`
// 	InitialField     string `json:"initial_field"`
// 	Solution         string `json:"solution"`
// 	SolveAttempts    int    `json:"solve_attempts"`
// 	SolvesSuccessful int    `json:"solves_successful"`
// 	SuccessRate      string `json:"success_rate"`
// 	AvgSolveTimeMs   int    `json:"avg_solve_time_ms"`
// }

// func NewGameService(cfg *config.Config) *GameService {
// 	return &GameService{cfg: cfg}
// }

// func (s *GameService) GetSudokuInfo(fieldID string) (*SudokuInfo, error) {

// 	resp, err := http.Get(s.cfg.GameServiceURL + "/sudoku/" + fieldID)
// 	if err != nil {
// 		log.Println("Error getting sudoku info:", err)
// 		return nil, fmt.Errorf("failed to get sudoku info: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		log.Println("Game service returned status:", resp.StatusCode)
// 		return nil, fmt.Errorf("game service returned status: %d", resp.StatusCode)
// 	}

// 	var info SudokuInfo
// 	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
// 		log.Println("Error decoding response:", err)
// 		return nil, fmt.Errorf("failed to decode response: %w", err)
// 	}

// 	return &info, nil
// }

// func (s *GameService) GetSudokuByDifficulty(difficulty string) (*SudokuInfo, error) {
// 	resp, err := http.Get(s.cfg.GameServiceURL + "/get-random-sudoku/" + difficulty)
// 	if err != nil {
// 		log.Println("Error getting sudoku by difficulty:", err)
// 		return nil, fmt.Errorf("failed to get sudoku by difficulty: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		log.Println("Game service returned status:", resp.StatusCode)
// 		return nil, fmt.Errorf("game service returned status: %d", resp.StatusCode)
// 	}

// 	var sudoku SudokuInfo
// 	if err := json.NewDecoder(resp.Body).Decode(&sudoku); err != nil {
// 		log.Println("Error decoding response:", err)
// 		return nil, fmt.Errorf("failed to decode response: %w", err)
// 	}

// 	return &sudoku, nil
// }

// // GetUnsolvedSudokuByDifficulty получает нерешенную судоку определенной сложности для турнира
// func (s *GameService) GetUnsolvedSudokuByDifficulty(tournamentID, userID, difficulty string, db Database) (*SudokuInfo, error) {
// 	// Получаем список всех судоку данной сложности
// 	resp, err := http.Get(s.cfg.GameServiceURL + "/get-sudokus-by-difficulty/" + difficulty)
// 	if err != nil {
// 		log.Println("Error getting sudokus by difficulty:", err)
// 		return nil, fmt.Errorf("failed to get sudokus by difficulty: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		log.Println("Game service returned status:", resp.StatusCode)
// 		return nil, fmt.Errorf("game service returned status: %d", resp.StatusCode)
// 	}

// 	var sudokus []SudokuInfo
// 	if err := json.NewDecoder(resp.Body).Decode(&sudokus); err != nil {
// 		log.Println("Error decoding response:", err)
// 		return nil, fmt.Errorf("failed to decode response: %w", err)
// 	}

// 	// Перемешиваем список судоку для случайного выбора
// 	rand.Shuffle(len(sudokus), func(i, j int) {
// 		sudokus[i], sudokus[j] = sudokus[j], sudokus[i]
// 	})

// 	// Ищем первую нерешенную судоку
// 	for _, sudoku := range sudokus {
// 		isSolved, err := db.IsSudokuSolved(tournamentID, userID, sudoku.ID)
// 		if err != nil {
// 			log.Println("Error checking if sudoku is solved:", err)
// 			continue
// 		}
// 		if !isSolved {
// 			return &sudoku, nil
// 		} else {
// 			log.Println("Sudoku is solved:", sudoku.ID, "for tournament:", tournamentID, "and user:", userID)
// 		}
// 	}

// 	return nil, fmt.Errorf("no unsolved sudokus of difficulty %s found", difficulty)
// }
