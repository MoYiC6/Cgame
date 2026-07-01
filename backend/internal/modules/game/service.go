package game

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"backend/internal/platform/database"
)

const (
	RoomStatusWaiting   = "waiting"
	RoomStatusPlaying   = "playing"
	RoomStatusDisbanded = "disbanded"
	RoomStatusFinished  = "finished"

	RecordStatusPlaying   = "playing"
	RecordStatusFinished  = "finished"
	RecordStatusDisbanded = "disbanded"

	MaxPlayersMin = 2
	MaxPlayersMax = 4
)

type Service struct {
	repo      *Repository
	txManager database.TxManager
}

func NewService(repo *Repository, txManager database.TxManager) *Service {
	s := &Service{repo: repo}
	if txManager != nil {
		s.txManager = txManager
	} else {
		s.txManager = database.NoopTxManager{}
	}
	return s
}

func (s *Service) CreateRoom(ctx context.Context, userID int64, maxPlayers int, nickname, avatar string) (*GameRoomVO, error) {
	if maxPlayers < MaxPlayersMin || maxPlayers > MaxPlayersMax {
		return nil, fmt.Errorf("max_players must be between %d and %d", MaxPlayersMin, MaxPlayersMax)
	}

	roomCode, err := s.generateUniqueRoomCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate room code: %w", err)
	}

	room := &GameRoom{
		RoomCode:          roomCode,
		HostUserID:        userID,
		Status:            RoomStatusPlaying,
		MaxPlayers:        maxPlayers,
		CurrentPosition:   1,
		GameOver:          false,
		CurrentTurnUserID: userID,
	}

	if err := s.repo.CreateRoom(ctx, room); err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}

	player := &GameRoomPlayer{
		RoomID:   room.ID,
		UserID:   userID,
		Nickname: nickname,
		Avatar:   avatar,
		IsHost:   true,
	}
	if err := s.repo.CreateRoomPlayer(ctx, player); err != nil {
		return nil, fmt.Errorf("create player: %w", err)
	}

	record := &GameRecord{
		RoomID:     room.ID,
		RoomCode:   room.RoomCode,
		HostUserID: room.HostUserID,
		Status:     RecordStatusPlaying,
		StartedAt:  timePtr(time.Now()),
	}
	if err := s.repo.CreateGameRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("create record: %w", err)
	}

	return s.buildRoomVO(room, []*GameRoomPlayer{player}), nil
}

func (s *Service) JoinRoom(ctx context.Context, userID int64, roomCode, nickname, avatar string) (*GameRoomVO, error) {
	room, err := s.repo.GetRoomByCode(ctx, roomCode)
	if err != nil {
		return nil, fmt.Errorf("room not found: %w", err)
	}

	if room.Status != RoomStatusPlaying || room.GameOver {
		return nil, fmt.Errorf("room is not available")
	}

	count, err := s.repo.CountActivePlayers(ctx, room.ID)
	if err != nil {
		return nil, fmt.Errorf("count players: %w", err)
	}
	if count >= room.MaxPlayers {
		return nil, fmt.Errorf("room is full")
	}

	_, err = s.repo.GetActivePlayer(ctx, room.ID, userID)
	if err == nil {
		return nil, fmt.Errorf("already in room")
	}

	player := &GameRoomPlayer{
		RoomID:   room.ID,
		UserID:   userID,
		Nickname: nickname,
		Avatar:   avatar,
		IsHost:   false,
	}
	if err := s.repo.CreateRoomPlayer(ctx, player); err != nil {
		return nil, fmt.Errorf("join room: %w", err)
	}

	players, err := s.repo.GetRoomPlayers(ctx, room.ID)
	if err != nil {
		return nil, fmt.Errorf("get players: %w", err)
	}

	return s.buildRoomVO(room, players), nil
}

func (s *Service) LeaveRoom(ctx context.Context, userID int64, roomID int64) (*GameRoomVO, error) {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("room not found: %w", err)
	}

	player, err := s.repo.GetActivePlayer(ctx, room.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("player not in room")
	}

	if err := s.repo.LeaveRoom(ctx, player.ID); err != nil {
		return nil, fmt.Errorf("leave room: %w", err)
	}

	if player.IsHost {
		room.Status = RoomStatusDisbanded
		room.GameOver = true
		if err := s.repo.UpdateRoom(ctx, room); err != nil {
			return nil, fmt.Errorf("update room: %w", err)
		}

		record, _ := s.repo.GetGameRecord(ctx, room.ID)
		if record != nil && record.Status == RecordStatusPlaying {
			_ = s.repo.FinishGameRecord(ctx, record.ID, 0)
		}
	}

	players, _ := s.repo.GetRoomPlayers(ctx, room.ID)
	return s.buildRoomVO(room, players), nil
}

func (s *Service) GetRoom(ctx context.Context, roomCode string) (*GameRoomVO, error) {
	room, err := s.repo.GetRoomByCode(ctx, roomCode)
	if err != nil {
		return nil, fmt.Errorf("room not found: %w", err)
	}

	players, err := s.repo.GetRoomPlayers(ctx, room.ID)
	if err != nil {
		return nil, fmt.Errorf("get players: %w", err)
	}

	return s.buildRoomVO(room, players), nil
}

func (s *Service) DisbandRoom(ctx context.Context, userID int64, roomID int64) (*GameRoomVO, error) {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("room not found: %w", err)
	}

	if room.HostUserID != userID {
		return nil, fmt.Errorf("only host can disband room")
	}

	room.Status = RoomStatusDisbanded
	room.GameOver = true
	if err := s.repo.UpdateRoom(ctx, room); err != nil {
		return nil, fmt.Errorf("disband room: %w", err)
	}

	record, _ := s.repo.GetGameRecord(ctx, room.ID)
	if record != nil && record.Status == RecordStatusPlaying {
		_ = s.repo.FinishGameRecord(ctx, record.ID, 0)
	}

	players, _ := s.repo.GetRoomPlayers(ctx, room.ID)
	return s.buildRoomVO(room, players), nil
}

func (s *Service) RecordMove(ctx context.Context, roomID, userID, dice, fromPos, toPos int64, cellText string, specialType *string) error {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("room not found: %w", err)
	}

	if room.Status != RoomStatusPlaying || room.GameOver {
		return fmt.Errorf("room is not playing")
	}

	record, err := s.repo.GetGameRecord(ctx, roomID)
	if err != nil {
		return fmt.Errorf("record not found: %w", err)
	}

	move := &GameMove{
		RecordID:     record.ID,
		RoomID:       roomID,
		UserID:       userID,
		Dice:         int(dice),
		FromPosition: int(fromPos),
		ToPosition:   int(toPos),
		CellText:     cellText,
		SpecialType: specialType,
	}
	if err := s.repo.CreateGameMove(ctx, move); err != nil {
		return fmt.Errorf("record move: %w", err)
	}

	room.CurrentPosition = int(toPos)
	if toPos >= 26 {
		room.GameOver = true
		room.Status = RoomStatusFinished
		_ = s.repo.FinishGameRecord(ctx, record.ID, int(toPos))
	}

	if err := s.repo.UpdateRoom(ctx, room); err != nil {
		return fmt.Errorf("update room: %w", err)
	}

	return nil
}

func (s *Service) GetEnabledMaps(ctx context.Context) ([]*GameMap, error) {
	return s.repo.ListGameMaps(ctx, 1)
}

func (s *Service) GetGoodsMaps(ctx context.Context, goodsID int64) ([]*GameMap, error) {
	mapIDs, err := s.repo.GetGoodsMapIDs(ctx, goodsID)
	if err != nil {
		return nil, fmt.Errorf("get goods maps: %w", err)
	}

	if len(mapIDs) == 0 {
		return s.repo.ListGameMaps(ctx, 1)
	}

	var maps []*GameMap
	for _, id := range mapIDs {
		m, err := s.repo.GetGameMapByID(ctx, id)
		if err == nil {
			maps = append(maps, m)
		}
	}
	return maps, nil
}

func (s *Service) CreateMap(ctx context.Context, m *GameMap) (int64, error) {
	if m.Name == "" {
		return 0, fmt.Errorf("name is required")
	}
	if err := s.repo.CreateGameMap(ctx, m); err != nil {
		return 0, fmt.Errorf("create map: %w", err)
	}
	return m.ID, nil
}

func (s *Service) UpdateMap(ctx context.Context, m *GameMap) error {
	if m.ID == 0 {
		return fmt.Errorf("id is required")
	}
	if err := s.repo.UpdateGameMap(ctx, m); err != nil {
		return fmt.Errorf("update map: %w", err)
	}
	return nil
}

func (s *Service) DeleteMap(ctx context.Context, id int64) error {
	if err := s.repo.DeleteGameMap(ctx, id); err != nil {
		return fmt.Errorf("delete map: %w", err)
	}
	return nil
}

func (s *Service) GetMapByID(ctx context.Context, id int64) (*GameMap, error) {
	m, err := s.repo.GetGameMapByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get map: %w", err)
	}
	return m, nil
}

func (s *Service) ListMaps(ctx context.Context, page, pageSize int, keyword string, status *int) ([]*GameMap, int, error) {
	maps, total, err := s.repo.ListGameMapsPaged(ctx, page, pageSize, keyword, status)
	if err != nil {
		return nil, 0, fmt.Errorf("list maps: %w", err)
	}
	return maps, total, nil
}

func (s *Service) GetBombRankingList(ctx context.Context) ([]*BombRankingVO, error) {
	rankings, err := s.repo.GetBombRankingList(ctx)
	if err != nil {
		return nil, fmt.Errorf("get bomb ranking: %w", err)
	}

	var vos []*BombRankingVO
	for _, br := range rankings {
		vos = append(vos, &BombRankingVO{
			RankPosition:    br.RankPosition,
			TeacherUserID:   br.TeacherUserID,
			TeacherNickname: br.TeacherName,
			TeacherAvatar:   br.TeacherAvatar,
			AssessmentLevel: br.AssessmentLevel,
		})
	}
	return vos, nil
}

func (s *Service) UpdateBombRanking(ctx context.Context, rankings []*BombRanking) error {
	if len(rankings) > 5 {
		return fmt.Errorf("max 5 rankings allowed")
	}

	teacherIDs := make(map[int64]bool)
	for _, br := range rankings {
		if teacherIDs[br.TeacherID] {
			return fmt.Errorf("duplicate teacher id")
		}
		teacherIDs[br.TeacherID] = true
	}

	if err := s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		return s.repo.ReplaceBombRanking(txCtx, rankings)
	}); err != nil {
		return fmt.Errorf("update bomb ranking: %w", err)
	}
	return nil
}

func (s *Service) AssignTeacherMaps(ctx context.Context, teacherID int64, mapIDs []int64) error {
	if err := s.repo.AssignTeacherMaps(ctx, teacherID, mapIDs); err != nil {
		return fmt.Errorf("assign maps: %w", err)
	}
	return nil
}

func (s *Service) RemoveTeacherMaps(ctx context.Context, teacherID int64, mapIDs []int64) error {
	if err := s.repo.RemoveTeacherMaps(ctx, teacherID, mapIDs); err != nil {
		return fmt.Errorf("remove maps: %w", err)
	}
	return nil
}

func (s *Service) GetTeacherMaps(ctx context.Context, teacherID int64) ([]*GameMap, error) {
	mapIDs, err := s.repo.GetTeacherMapIDs(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("get teacher maps: %w", err)
	}

	var maps []*GameMap
	for _, id := range mapIDs {
		m, err := s.repo.GetGameMapByID(ctx, id)
		if err == nil {
			maps = append(maps, m)
		}
	}
	return maps, nil
}

func (s *Service) GetTeacherMapIDs(ctx context.Context, teacherID int64) ([]int64, error) {
	return s.repo.GetTeacherMapIDs(ctx, teacherID)
}

func (s *Service) HasMapPermission(ctx context.Context, teacherID, mapID int64) (bool, error) {
	return s.repo.HasMapPermission(ctx, teacherID, mapID)
}

func (s *Service) CanAcceptOrderByMap(ctx context.Context, teacherID, mapID int64) (bool, error) {
	if mapID == 0 {
		return true, nil
	}
	return s.repo.HasMapPermission(ctx, teacherID, mapID)
}

func (s *Service) GetTeacherIDsByMapID(ctx context.Context, mapID int64) ([]int64, error) {
	return s.repo.GetTeacherIDsByMapID(ctx, mapID)
}

func (s *Service) ConfigureGoodsMaps(ctx context.Context, goodsID int64, mapIDs []int64) error {
	if err := s.repo.ConfigureGoodsMaps(ctx, goodsID, mapIDs); err != nil {
		return fmt.Errorf("configure maps: %w", err)
	}
	return nil
}

func (s *Service) RemoveGoodsMaps(ctx context.Context, goodsID int64) error {
	if err := s.repo.DeleteGoodsMaps(ctx, goodsID); err != nil {
		return fmt.Errorf("remove maps: %w", err)
	}
	return nil
}

func (s *Service) HasGoodsMap(ctx context.Context, goodsID, mapID int64) (bool, error) {
	return s.repo.HasGoodsMap(ctx, goodsID, mapID)
}

func (s *Service) generateUniqueRoomCode(ctx context.Context) (string, error) {
	const maxAttempts = 10
	chars := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

	for i := 0; i < maxAttempts; i++ {
		code := ""
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for j := 0; j < 6; j++ {
			code += string(chars[r.Intn(len(chars))])
		}

		_, err := s.repo.GetRoomByCode(ctx, code)
		if err != nil {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique room code")
}

func (s *Service) buildRoomVO(room *GameRoom, players []*GameRoomPlayer) *GameRoomVO {
	vo := &GameRoomVO{
		RoomID:          room.ID,
		RoomCode:        room.RoomCode,
		HostUserID:      room.HostUserID,
		Status:          room.Status,
		MaxPlayers:      room.MaxPlayers,
		CurrentPosition: room.CurrentPosition,
		GameOver:        room.GameOver,
		CurrentTurnUserID: room.CurrentTurnUserID,
		Players:         make([]GamePlayerVO, 0, len(players)),
	}

	for _, p := range players {
		vo.Players = append(vo.Players, GamePlayerVO{
			UserID:   p.UserID,
			Nickname: p.Nickname,
			Avatar:   p.Avatar,
			IsHost:   p.IsHost,
		})
	}

	return vo
}

func timePtr(t time.Time) *time.Time {
	return &t
}
