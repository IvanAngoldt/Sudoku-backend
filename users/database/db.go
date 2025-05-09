package database

import (
	"database/sql"
	"fmt"
	"users/config"
	"users/models"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func (d *Database) Config() {
	panic("unimplemented")
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Создаем таблицу пользователей, если она не существует
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, err
	}

	// Создаем таблицу с доп. информацией о пользователях, если она не существует
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_info (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			full_name VARCHAR(100) NOT NULL,
			age INT NOT NULL,
			city VARCHAR(50) NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_avatar (
			user_id VARCHAR(36) PRIMARY KEY,
			avatar_url TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return nil, err
	}

	// user_statistics
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_statistics (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL UNIQUE
		)
	`)
	if err != nil {
		return nil, err
	}

	// user_difficulty_stats
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_difficulty_stats (
			id SERIAL PRIMARY KEY,
			user_statistics_id INTEGER NOT NULL REFERENCES user_statistics(id) ON DELETE CASCADE,
			difficulty TEXT NOT NULL, -- значения: easy, medium, hard, ...
			total_solved INTEGER NOT NULL DEFAULT 0,
			total_time_seconds INTEGER NOT NULL DEFAULT 0,
			best_time_seconds INTEGER,
			UNIQUE (user_statistics_id, difficulty)
		)
	`)
	if err != nil {
		return nil, err
	}

	// achievements
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS achievements (
			id SERIAL PRIMARY KEY,
			code TEXT UNIQUE NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			icon_url TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		return nil, err
	}

	// user_achievements
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_achievements (
			user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			achievement_id INTEGER NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
			earned_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (user_id, achievement_id)
		)
	`)
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) CreateUser(user *models.User) error {
	query := `INSERT INTO users (id, username, email, password, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := d.db.Exec(query,
		user.ID, user.Username, user.Email, user.Password, user.CreatedAt, user.UpdatedAt)
	return err
}

func (d *Database) GetUser(id string) (*models.User, error) {
	query := `SELECT id, username, email, password, created_at, updated_at 
			  FROM users WHERE id = $1`
	user := &models.User{}
	err := d.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *Database) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, email, password, created_at, updated_at 
			  FROM users WHERE email = $1`
	user := &models.User{}
	err := d.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *Database) GetUsers() ([]*models.User, error) {
	query := `SELECT id, username, email, created_at, updated_at FROM users`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email,
			&user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (d *Database) UpdateUser(user *models.User) error {
	query := `UPDATE users SET username = $1, email = $2, password = $3, 
			  updated_at = $4 WHERE id = $5`
	_, err := d.db.Exec(query, user.Username, user.Email, user.Password, user.UpdatedAt, user.ID)
	return err
}

func (d *Database) DeleteUser(id string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := d.db.Exec(query, id)
	return err
}

func (d *Database) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, email, password, created_at, updated_at FROM users WHERE username = $1`
	user := &models.User{}
	err := d.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password,
		&user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *Database) UsernameExists(username string) (bool, error) {
	query := `SELECT 1 FROM users WHERE username = $1 LIMIT 1`
	var exists int
	err := d.db.QueryRow(query, username).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *Database) EmailExists(email string) (bool, error) {
	query := `SELECT 1 FROM users WHERE email = $1 LIMIT 1`
	var exists int
	err := d.db.QueryRow(query, email).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

// CreateUserInfo добавляет запись в user_info
func (d *Database) CreateUserInfo(info *models.UserInfo) error {
	query := `INSERT INTO user_info (user_id, full_name, age, city)
			  VALUES ($1, $2, $3, $4)`
	_, err := d.db.Exec(query, info.UserID, info.FullName, info.Age, info.City)
	return err
}

// GetUserInfoByUserID возвращает информацию о пользователе по user_id
func (d *Database) GetUserInfoByUserID(userID string) (*models.UserInfo, error) {
	query := `SELECT id, user_id, full_name, age, city FROM user_info WHERE user_id = $1`
	info := &models.UserInfo{}
	err := d.db.QueryRow(query, userID).Scan(
		&info.ID, &info.UserID, &info.FullName, &info.Age, &info.City,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return info, nil
}

// UpdateUserInfo обновляет информацию о пользователе по user_id
func (d *Database) UpdateUserInfo(info *models.UserInfo) error {
	query := `UPDATE user_info SET full_name = $1, age = $2, city = $3 WHERE user_id = $4`
	_, err := d.db.Exec(query, info.FullName, info.Age, info.City, info.UserID)
	return err
}

// DeleteUserInfo удаляет информацию о пользователе по user_id
func (d *Database) DeleteUserInfo(userID string) error {
	query := `DELETE FROM user_info WHERE user_id = $1`
	_, err := d.db.Exec(query, userID)
	return err
}

func (d *Database) UpsertUserAvatar(userID string, avatarURL string) error {
	query := `INSERT INTO user_avatar (user_id, avatar_url)
			  VALUES ($1, $2)
			  ON CONFLICT (user_id) DO UPDATE SET avatar_url = EXCLUDED.avatar_url`
	_, err := d.db.Exec(query, userID, avatarURL)
	return err
}

func (d *Database) GetUserAvatarFilename(userID string) (string, error) {
	query := `SELECT avatar_url FROM user_avatar WHERE user_id = $1`
	var filename string
	err := d.db.QueryRow(query, userID).Scan(&filename)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return filename, nil
}

func (d *Database) EnsureUserStatistics(userID string) (int, error) {
	var id int
	err := d.db.QueryRow(`
		INSERT INTO user_statistics (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
		RETURNING id
	`, userID).Scan(&id)
	return id, err
}

func (d *Database) UpdateDifficultyStats(userID, difficulty string, timeSpent int) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var statisticsID int
	err = tx.QueryRow(`SELECT id FROM user_statistics WHERE user_id = $1`, userID).Scan(&statisticsID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO user_difficulty_stats (
			user_statistics_id, difficulty, total_solved, total_time_seconds, best_time_seconds
		)
		VALUES ($1, $2, 1, $3, $3)
		ON CONFLICT (user_statistics_id, difficulty)
		DO UPDATE SET 
			total_solved = user_difficulty_stats.total_solved + 1,
			total_time_seconds = user_difficulty_stats.total_time_seconds + EXCLUDED.total_time_seconds,
			best_time_seconds = LEAST(COALESCE(user_difficulty_stats.best_time_seconds, EXCLUDED.best_time_seconds), EXCLUDED.best_time_seconds)
	`, statisticsID, difficulty, timeSpent)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *Database) GetUserStatistics(userID string) ([]models.UserDifficultyStat, error) {
	var statsID int
	err := d.db.QueryRow(`SELECT id FROM user_statistics WHERE user_id = $1`, userID).Scan(&statsID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []models.UserDifficultyStat{}, nil
		}
		return nil, err
	}

	rows, err := d.db.Query(`
		SELECT difficulty, total_solved, total_time_seconds, best_time_seconds
		FROM user_difficulty_stats
		WHERE user_statistics_id = $1
	`, statsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.UserDifficultyStat
	for rows.Next() {
		var stat models.UserDifficultyStat
		err := rows.Scan(&stat.Difficulty, &stat.TotalSolved, &stat.TotalTimeSeconds, &stat.BestTimeSeconds)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}
