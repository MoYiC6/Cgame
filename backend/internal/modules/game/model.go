package game

import "time"

type GameRoom struct {
	ID                int64
	RoomCode          string
	HostUserID        int64
	Status            string
	MaxPlayers        int
	CurrentPosition   int
	GameOver          bool
	CurrentTurnUserID int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type GameRoomPlayer struct {
	ID        int64
	RoomID    int64
	UserID    int64
	Nickname  string
	Avatar    string
	IsHost    bool
	JoinedAt  time.Time
	LeftAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GameMap struct {
	ID          int64
	Name        string
	Description *string
	Icon        *string
	Sort        int
	Status      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GameMove struct {
	ID          int64
	RecordID    int64
	RoomID      int64
	UserID      int64
	Dice        int
	FromPosition int
	ToPosition  int
	CellText    string
	SpecialType *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GameRecord struct {
	ID         int64
	RoomID     int64
	RoomCode   string
	HostUserID int64
	FinalPosition int
	Status     string
	StartedAt  *time.Time
	FinishedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type BombRanking struct {
	ID            int64
	RankPosition  int
	TeacherID     int64
	TeacherUserID int64
	TeacherName   string
	TeacherAvatar string
	AssessmentLevel string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TeacherMapPermission struct {
	ID        int64
	TeacherID int64
	MapID     int64
	CreatedAt time.Time
}

type GoodsMap struct {
	ID        int64
	GoodsID   int64
	GameMapID int64
	CreatedAt time.Time
}

type GamePlayerVO struct {
	UserID   int64  `json:"userId"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	IsHost   bool   `json:"isHost"`
}

type GameRoomVO struct {
	RoomID          int64         `json:"roomId"`
	RoomCode        string        `json:"roomCode"`
	HostUserID      int64         `json:"hostUserId"`
	Status          string        `json:"status"`
	MaxPlayers      int           `json:"maxPlayers"`
	CurrentPosition int           `json:"currentPosition"`
	GameOver        bool          `json:"gameOver"`
	CurrentTurnUserID int64       `json:"currentTurnUserId"`
	Players         []GamePlayerVO `json:"players"`
}

type GameMapVO struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Icon        *string   `json:"icon"`
	Sort        int       `json:"sort"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

type GameMapListVO struct {
	Total int64        `json:"total"`
	Rows  []GameMapVO  `json:"rows"`
}

type BombRankingVO struct {
	RankPosition    int    `json:"rankPosition"`
	TeacherUserID   int64  `json:"teacherUserId"`
	TeacherNickname string `json:"teacherNickname"`
	TeacherAvatar   string `json:"teacherAvatar"`
	AssessmentLevel string `json:"assessmentLevel"`
}
