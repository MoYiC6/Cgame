package game

import (
	"context"
	"fmt"

	"backend/internal/platform/database"
)

type Repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) *Repository {
	return &Repository{dbtx: dbtx}
}

func (r *Repository) CreateRoom(ctx context.Context, room *GameRoom) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO game_rooms (room_code, host_user_id, status, max_players, current_position, game_over, current_turn_user_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		room.RoomCode, room.HostUserID, room.Status, room.MaxPlayers, room.CurrentPosition, room.GameOver, room.CurrentTurnUserID,
	).Scan(&room.ID)
}

func (r *Repository) GetRoomByCode(ctx context.Context, roomCode string) (*GameRoom, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, room_code, host_user_id, status, max_players, current_position, game_over, current_turn_user_id, created_at, updated_at
		 FROM game_rooms WHERE room_code = $1`,
		roomCode,
	)
	var room GameRoom
	err := row.Scan(&room.ID, &room.RoomCode, &room.HostUserID, &room.Status, &room.MaxPlayers, &room.CurrentPosition, &room.GameOver, &room.CurrentTurnUserID, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
	}
	return &room, nil
}

func (r *Repository) GetRoomByID(ctx context.Context, id int64) (*GameRoom, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, room_code, host_user_id, status, max_players, current_position, game_over, current_turn_user_id, created_at, updated_at
		 FROM game_rooms WHERE id = $1`,
		id,
	)
	var room GameRoom
	err := row.Scan(&room.ID, &room.RoomCode, &room.HostUserID, &room.Status, &room.MaxPlayers, &room.CurrentPosition, &room.GameOver, &room.CurrentTurnUserID, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
	}
	return &room, nil
}

func (r *Repository) UpdateRoom(ctx context.Context, room *GameRoom) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE game_rooms SET status = $1, current_position = $2, game_over = $3, current_turn_user_id = $4, updated_at = NOW() WHERE id = $5`,
		room.Status, room.CurrentPosition, room.GameOver, room.CurrentTurnUserID, room.ID,
	)
	if err != nil {
		return fmt.Errorf("update room: %w", err)
	}
	return nil
}

func (r *Repository) CreateRoomPlayer(ctx context.Context, player *GameRoomPlayer) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO game_room_players (room_id, user_id, nickname, avatar, is_host, left_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		player.RoomID, player.UserID, player.Nickname, player.Avatar, player.IsHost, player.LeftAt,
	).Scan(&player.ID)
}

func (r *Repository) GetRoomPlayers(ctx context.Context, roomID int64) ([]*GameRoomPlayer, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	rows, err := exec.QueryContext(ctx,
		`SELECT id, room_id, user_id, nickname, avatar, is_host, joined_at, left_at, created_at, updated_at
		 FROM game_room_players WHERE room_id = $1 AND left_at IS NULL`,
		roomID,
	)
	if err != nil {
		return nil, fmt.Errorf("get room players: %w", err)
	}
	defer rows.Close()

	var players []*GameRoomPlayer
	for rows.Next() {
		var p GameRoomPlayer
		if err := rows.Scan(&p.ID, &p.RoomID, &p.UserID, &p.Nickname, &p.Avatar, &p.IsHost, &p.JoinedAt, &p.LeftAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan player: %w", err)
		}
		players = append(players, &p)
	}
	return players, nil
}

func (r *Repository) GetActivePlayer(ctx context.Context, roomID, userID int64) (*GameRoomPlayer, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, room_id, user_id, nickname, avatar, is_host, joined_at, left_at, created_at, updated_at
		 FROM game_room_players WHERE room_id = $1 AND user_id = $2 AND left_at IS NULL`,
		roomID, userID,
	)
	var p GameRoomPlayer
	err := row.Scan(&p.ID, &p.RoomID, &p.UserID, &p.Nickname, &p.Avatar, &p.IsHost, &p.JoinedAt, &p.LeftAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get active player: %w", err)
	}
	return &p, nil
}

func (r *Repository) CountActivePlayers(ctx context.Context, roomID int64) (int, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var count int
	err := exec.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM game_room_players WHERE room_id = $1 AND left_at IS NULL`,
		roomID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count players: %w", err)
	}
	return count, nil
}

func (r *Repository) LeaveRoom(ctx context.Context, playerID int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE game_room_players SET left_at = NOW(), updated_at = NOW() WHERE id = $1`,
		playerID,
	)
	if err != nil {
		return fmt.Errorf("leave room: %w", err)
	}
	return nil
}

func (r *Repository) CreateGameRecord(ctx context.Context, record *GameRecord) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO game_records (room_id, room_code, host_user_id, final_position, status, started_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		record.RoomID, record.RoomCode, record.HostUserID, record.FinalPosition, record.Status, record.StartedAt,
	).Scan(&record.ID)
}

func (r *Repository) FinishGameRecord(ctx context.Context, recordID int64, finalPosition int) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE game_records SET status = 'finished', final_position = $1, finished_at = NOW() WHERE id = $2`,
		finalPosition, recordID,
	)
	if err != nil {
		return fmt.Errorf("finish record: %w", err)
	}
	return nil
}

func (r *Repository) CreateGameMove(ctx context.Context, move *GameMove) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO game_moves (record_id, room_id, user_id, dice, from_position, to_position, cell_text, special_type)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		move.RecordID, move.RoomID, move.UserID, move.Dice, move.FromPosition, move.ToPosition, move.CellText, move.SpecialType,
	).Scan(&move.ID)
}

func (r *Repository) ListGameMaps(ctx context.Context, status int) ([]*GameMap, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	rows, err := exec.QueryContext(ctx,
		`SELECT id, name, description, icon, sort, status, created_at, updated_at FROM game_maps WHERE status = $1 ORDER BY sort ASC, created_at ASC`,
		status,
	)
	if err != nil {
		return nil, fmt.Errorf("list game maps: %w", err)
	}
	defer rows.Close()

	var maps []*GameMap
	for rows.Next() {
		var m GameMap
		if err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.Icon, &m.Sort, &m.Status, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan map: %w", err)
		}
		maps = append(maps, &m)
	}
	return maps, nil
}

func (r *Repository) GetGameMapByID(ctx context.Context, id int64) (*GameMap, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, name, description, icon, sort, status, created_at, updated_at FROM game_maps WHERE id = $1`,
		id,
	)
	var m GameMap
	err := row.Scan(&m.ID, &m.Name, &m.Description, &m.Icon, &m.Sort, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get map: %w", err)
	}
	return &m, nil
}

func (r *Repository) CreateGameMap(ctx context.Context, m *GameMap) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO game_maps (name, description, icon, sort, status) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		m.Name, m.Description, m.Icon, m.Sort, m.Status,
	).Scan(&m.ID)
}

func (r *Repository) UpdateGameMap(ctx context.Context, m *GameMap) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE game_maps SET name = $1, description = $2, icon = $3, sort = $4, status = $5, updated_at = NOW() WHERE id = $6`,
		m.Name, m.Description, m.Icon, m.Sort, m.Status, m.ID,
	)
	if err != nil {
		return fmt.Errorf("update map: %w", err)
	}
	return nil
}

func (r *Repository) DeleteGameMap(ctx context.Context, id int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx, `DELETE FROM game_maps WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete map: %w", err)
	}
	return nil
}

func (r *Repository) ListGameMapsPaged(ctx context.Context, page, pageSize int, keyword string, status *int) ([]*GameMap, int, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	countQuery := "SELECT COUNT(*) FROM game_maps WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if keyword != "" {
		argCount++
		countQuery += fmt.Sprintf(" AND name ILIKE $%d", argCount)
		args = append(args, "%"+keyword+"%")
	}
	if status != nil {
		argCount++
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *status)
	}

	var total int
	if err := exec.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count maps: %w", err)
	}

	dataQuery := `SELECT id, name, description, icon, sort, status, created_at, updated_at FROM game_maps WHERE 1=1`
	dataArgs := []interface{}{}
	argCount = 0

	if keyword != "" {
		argCount++
		dataQuery += fmt.Sprintf(" AND name ILIKE $%d", argCount)
		dataArgs = append(dataArgs, "%"+keyword+"%")
	}
	if status != nil {
		argCount++
		dataQuery += fmt.Sprintf(" AND status = $%d", argCount)
		dataArgs = append(dataArgs, *status)
	}

	dataQuery += " ORDER BY sort ASC, created_at DESC"
	argCount++
	dataQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	dataArgs = append(dataArgs, pageSize, (page-1)*pageSize)

	rows, err := exec.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list maps paged: %w", err)
	}
	defer rows.Close()

	var maps []*GameMap
	for rows.Next() {
		var m GameMap
		if err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.Icon, &m.Sort, &m.Status, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan map: %w", err)
		}
		maps = append(maps, &m)
	}
	return maps, total, nil
}

func (r *Repository) GetBombRankingList(ctx context.Context) ([]*BombRanking, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	rows, err := exec.QueryContext(ctx,
		`SELECT br.id, br.rank_position, br.teacher_id, u.id as teacher_user_id, u.nickname as teacher_name, u.avatar as teacher_avatar, 
		 COALESCE(t.assessment_level, '') as assessment_level, br.created_at, br.updated_at
		 FROM bomb_rankings br
		 LEFT JOIN teachers t ON br.teacher_id = t.id
		 LEFT JOIN sys_users u ON t.user_id = u.id
		 ORDER BY br.rank_position ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("get bomb ranking: %w", err)
	}
	defer rows.Close()

	var rankings []*BombRanking
	for rows.Next() {
		var br BombRanking
		if err := rows.Scan(&br.ID, &br.RankPosition, &br.TeacherID, &br.TeacherUserID, &br.TeacherName, &br.TeacherAvatar, &br.AssessmentLevel, &br.CreatedAt, &br.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan ranking: %w", err)
		}
		rankings = append(rankings, &br)
	}
	return rankings, nil
}

func (r *Repository) ReplaceBombRanking(ctx context.Context, rankings []*BombRanking) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	if _, err := exec.ExecContext(ctx, `DELETE FROM bomb_rankings`); err != nil {
		return fmt.Errorf("clear rankings: %w", err)
	}

	for _, br := range rankings {
		if _, err := exec.ExecContext(ctx,
			`INSERT INTO bomb_rankings (rank_position, teacher_id) VALUES ($1, $2)`,
			br.RankPosition, br.TeacherID,
		); err != nil {
			return fmt.Errorf("insert ranking: %w", err)
		}
	}
	return nil
}

func (r *Repository) GetTeacherMapIDs(ctx context.Context, teacherID int64) ([]int64, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	rows, err := exec.QueryContext(ctx,
		`SELECT map_id FROM teacher_map_permissions WHERE teacher_id = $1`,
		teacherID,
	)
	if err != nil {
		return nil, fmt.Errorf("get teacher maps: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan map id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *Repository) AssignTeacherMaps(ctx context.Context, teacherID int64, mapIDs []int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	if len(mapIDs) == 0 {
		return nil
	}

	for _, mapID := range mapIDs {
		if _, err := exec.ExecContext(ctx,
			`INSERT INTO teacher_map_permissions (teacher_id, map_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			teacherID, mapID,
		); err != nil {
			return fmt.Errorf("assign map: %w", err)
		}
	}
	return nil
}

func (r *Repository) RemoveTeacherMaps(ctx context.Context, teacherID int64, mapIDs []int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	if len(mapIDs) == 0 {
		return nil
	}
	query := `DELETE FROM teacher_map_permissions WHERE teacher_id = $1 AND map_id IN (`
	args := []interface{}{teacherID}
	for i, id := range mapIDs {
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}
	query += ")"

	_, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("remove maps: %w", err)
	}
	return nil
}

func (r *Repository) HasMapPermission(ctx context.Context, teacherID, mapID int64) (bool, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var count int
	err := exec.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM teacher_map_permissions WHERE teacher_id = $1 AND map_id = $2`,
		teacherID, mapID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check permission: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) GetTeacherIDsByMapID(ctx context.Context, mapID int64) ([]int64, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	rows, err := exec.QueryContext(ctx,
		`SELECT teacher_id FROM teacher_map_permissions WHERE map_id = $1`,
		mapID,
	)
	if err != nil {
		return nil, fmt.Errorf("get teacher ids: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan teacher id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *Repository) GetGoodsMapIDs(ctx context.Context, goodsID int64) ([]int64, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	rows, err := exec.QueryContext(ctx,
		`SELECT game_map_id FROM goods_maps WHERE goods_id = $1`,
		goodsID,
	)
	if err != nil {
		return nil, fmt.Errorf("get goods maps: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan map id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *Repository) ConfigureGoodsMaps(ctx context.Context, goodsID int64, mapIDs []int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	if _, err := exec.ExecContext(ctx, `DELETE FROM goods_maps WHERE goods_id = $1`, goodsID); err != nil {
		return fmt.Errorf("clear maps: %w", err)
	}

	for _, mapID := range mapIDs {
		if _, err := exec.ExecContext(ctx,
			`INSERT INTO goods_maps (goods_id, game_map_id) VALUES ($1, $2)`,
			goodsID, mapID,
		); err != nil {
			return fmt.Errorf("insert map: %w", err)
		}
	}
	return nil
}

func (r *Repository) DeleteGoodsMaps(ctx context.Context, goodsID int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx, `DELETE FROM goods_maps WHERE goods_id = $1`, goodsID)
	if err != nil {
		return fmt.Errorf("delete maps: %w", err)
	}
	return nil
}

func (r *Repository) HasGoodsMap(ctx context.Context, goodsID, mapID int64) (bool, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var count int
	err := exec.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM goods_maps WHERE goods_id = $1 AND game_map_id = $2`,
		goodsID, mapID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check goods map: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) GetGameRecord(ctx context.Context, roomID int64) (*GameRecord, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, room_id, room_code, host_user_id, final_position, status, started_at, finished_at, created_at, updated_at
		 FROM game_records WHERE room_id = $1 ORDER BY id DESC LIMIT 1`,
		roomID,
	)
	var record GameRecord
	err := row.Scan(&record.ID, &record.RoomID, &record.RoomCode, &record.HostUserID, &record.FinalPosition, &record.Status, &record.StartedAt, &record.FinishedAt, &record.CreatedAt, &record.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get record: %w", err)
	}
	return &record, nil
}
