package game

import (
	"net/http"
	"strconv"
	"time"

	"backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	service        *Service
	authMiddleware gin.HandlerFunc
}

func NewHandler(service *Service, authMiddleware gin.HandlerFunc) *Handler {
	return &Handler{service: service, authMiddleware: authMiddleware}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	client := group.Group("/client/game")
	if h.authMiddleware != nil {
		client.Use(h.authMiddleware)
	}
	{
		client.POST("/room/create", h.CreateRoom)
		client.POST("/room/join", h.JoinRoom)
		client.POST("/room/leave/:roomId", h.LeaveRoom)
		client.GET("/room/:roomCode", h.GetRoom)
		client.POST("/room/disband/:roomId", h.DisbandRoom)
		client.GET("/maps", h.GetEnabledMaps)
		client.GET("/maps/goods/:goodsId", h.GetGoodsMaps)
	}

	admin := group.Group("/admin/game")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("/maps", h.ListMaps)
		admin.GET("/maps/:id", h.GetMap)
		admin.POST("/maps", h.CreateMap)
		admin.PUT("/maps/:id", h.UpdateMap)
		admin.DELETE("/maps/:id", h.DeleteMap)
		admin.GET("/maps/enabled", h.GetEnabledMaps)
		admin.GET("/bomb-ranking", h.GetBombRankingList)
		admin.PUT("/bomb-ranking", h.UpdateBombRanking)
	}
}

func (h *Handler) RegisterWebSocket(engine *gin.Engine, authMiddleware gin.HandlerFunc) {
	ws := engine.Group("/ws")
	{
		ws.GET("/game", h.HandleGameWebSocket)
	}
}

func (h *Handler) CreateRoom(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	maxPlayers, _ := strconv.Atoi(c.DefaultQuery("maxPlayers", "4"))
	
	var req struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}

	room, err := h.service.CreateRoom(c.Request.Context(), userID, maxPlayers, req.Nickname, req.Avatar)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, room)
}

func (h *Handler) JoinRoom(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	
	var req struct {
		RoomCode string `json:"roomCode"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}

	room, err := h.service.JoinRoom(c.Request.Context(), userID, req.RoomCode, req.Nickname, req.Avatar)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, room)
}

func (h *Handler) LeaveRoom(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	roomID, _ := strconv.ParseInt(c.Param("roomId"), 10, 64)

	room, err := h.service.LeaveRoom(c.Request.Context(), userID, roomID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, room)
}

func (h *Handler) GetRoom(c *gin.Context) {
	roomCode := c.Param("roomCode")
	room, err := h.service.GetRoom(c.Request.Context(), roomCode)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, room)
}

func (h *Handler) DisbandRoom(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	roomID, _ := strconv.ParseInt(c.Param("roomId"), 10, 64)

	room, err := h.service.DisbandRoom(c.Request.Context(), userID, roomID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, room)
}

func (h *Handler) GetEnabledMaps(c *gin.Context) {
	maps, err := h.service.GetEnabledMaps(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, maps)
}

func (h *Handler) GetGoodsMaps(c *gin.Context) {
	goodsID, _ := strconv.ParseInt(c.Param("goodsId"), 10, 64)
	maps, err := h.service.GetGoodsMaps(c.Request.Context(), goodsID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, maps)
}

func (h *Handler) ListMaps(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	
	var statusPtr *int
	if s := c.Query("status"); s != "" {
		sVal, _ := strconv.Atoi(s)
		statusPtr = &sVal
	}

	maps, total, err := h.service.ListMaps(c.Request.Context(), page, pageSize, keyword, statusPtr)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"total": total, "rows": maps})
}

func (h *Handler) GetMap(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	m, err := h.service.GetMapByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, m)
}

func (h *Handler) CreateMap(c *gin.Context) {
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Icon        *string `json:"icon"`
		Sort        int     `json:"sort"`
		Status      int     `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}

	m := &GameMap{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		Sort:       req.Sort,
		Status:     req.Status,
	}
	id, err := h.service.CreateMap(c.Request.Context(), m)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *Handler) UpdateMap(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Icon        *string `json:"icon"`
		Sort        *int    `json:"sort"`
		Status      *int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}

	m := &GameMap{ID: id}
	if req.Name != nil {
		m.Name = *req.Name
	}
	if req.Description != nil {
		m.Description = req.Description
	}
	if req.Icon != nil {
		m.Icon = req.Icon
	}
	if req.Sort != nil {
		m.Sort = *req.Sort
	}
	if req.Status != nil {
		m.Status = *req.Status
	}

	if err := h.service.UpdateMap(c.Request.Context(), m); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) DeleteMap(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.service.DeleteMap(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) GetBombRankingList(c *gin.Context) {
	rankings, err := h.service.GetBombRankingList(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, rankings)
}

func (h *Handler) UpdateBombRanking(c *gin.Context) {
	var req struct {
		Rankings []*BombRanking `json:"rankings"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}

	if err := h.service.UpdateBombRanking(c.Request.Context(), req.Rankings); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"message": "updated"})
}

func (h *Handler) HandleGameWebSocket(c *gin.Context) {
	userIDStr := c.Query("user_id")
	roomCode := c.Query("room_code")
	
	if userIDStr == "" || roomCode == "" {
		c.String(http.StatusBadRequest, "missing user_id or room_code")
		return
	}

	userID, _ := strconv.ParseInt(userIDStr, 10, 64)
	room, err := h.service.GetRoom(c.Request.Context(), roomCode)
	if err != nil {
		c.String(http.StatusNotFound, "room not found")
		return
	}

	if room.Status != RoomStatusPlaying || room.GameOver {
		c.String(http.StatusBadRequest, "room is not available")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	_ = room
	_ = userID
	_ = time.Now()
}
