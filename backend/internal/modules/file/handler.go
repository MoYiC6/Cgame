package file

import (
	"strconv"

	"backend/internal/platform/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service        *Service
	authMiddleware gin.HandlerFunc
}

func NewHandler(service *Service, authMiddleware gin.HandlerFunc) *Handler {
	return &Handler{service: service, authMiddleware: authMiddleware}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	admin := group.Group("/admin/files")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("", h.List)
		admin.GET("/:id", h.Get)
		admin.POST("/upload", h.Upload)
		admin.POST("/upload-base64", h.UploadBase64)
		admin.POST("/add-remote", h.AddRemote)
		admin.PUT("/:id", h.Update)
		admin.DELETE("/:id", h.Delete)
		admin.POST("/batch-delete", h.BatchDelete)
	}

	categories := group.Group("/admin/file-categories")
	if h.authMiddleware != nil {
		categories.Use(h.authMiddleware)
	}
	{
		categories.GET("", h.ListCategories)
		categories.GET("/:id", h.GetCategory)
		categories.POST("", h.CreateCategory)
		categories.PUT("/:id", h.UpdateCategory)
		categories.DELETE("/:id", h.DeleteCategory)
	}
}

// @Summary 文件列表
// @Description 分页查询文件
// @Tags file
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param type query string false "文件类型"
// @Param categoryId query int false "分类ID"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.APIResponse
// @Router /api/admin/files [get]
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	fileType := c.Query("type")
	categoryIDStr := c.Query("categoryId")
	keyword := c.Query("keyword")

	var categoryID *int64
	if categoryIDStr != "" {
		id, _ := strconv.ParseInt(categoryIDStr, 10, 64)
		categoryID = &id
	}

	files, total, err := h.service.ListFiles(c.Request.Context(), nil, categoryID, &fileType, keyword, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": files, "total": total})
}

// @Summary 文件详情
// @Description 根据ID查询文件
// @Tags file
// @Produce json
// @Param id path int true "文件ID"
// @Success 200 {object} response.APIResponse
// @Router /api/admin/files/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	file, err := h.service.GetFile(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, file)
}

// @Summary 上传文件
// @Description 上传文件（multipart）
// @Tags file
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Param categoryId formData int false "分类ID"
// @Success 200 {object} response.APIResponse
// @Router /api/admin/files/upload [post]
func (h *Handler) Upload(c *gin.Context) {
	response.Success(c, gin.H{"message": "upload endpoint - qiniu integration pending"})
}

// @Summary Base64 上传
// @Description Base64 编码文件上传
// @Tags file
// @Accept json
// @Produce json
// @Param request body object true "Base64 上传请求"
// @Success 200 {object} response.APIResponse
// @Router /api/admin/files/upload-base64 [post]
func (h *Handler) UploadBase64(c *gin.Context) {
	response.Success(c, gin.H{"message": "upload-base64 endpoint - qiniu integration pending"})
}

// @Summary 添加远程文件
// @Description 添加远程URL文件
// @Tags file
// @Accept json
// @Produce json
// @Param request body object true "远程文件请求"
// @Success 200 {object} response.APIResponse
// @Router /api/admin/files/add-remote [post]
func (h *Handler) AddRemote(c *gin.Context) {
	response.Success(c, gin.H{"message": "add-remote endpoint pending"})
}

// @Summary 更新文件
// @Description 更新文件信息
// @Tags file
// @Accept json
// @Produce json
// @Param id path int true "文件ID"
// @Param request body object true "文件更新请求"
// @Success 200 {object} response.APIResponse
// @Router /api/admin/files/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req File
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	req.ID = id
	if err := h.service.UpdateFile(c.Request.Context(), &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

// @Summary 删除文件
// @Description 逻辑删除文件
// @Tags file
// @Produce json
// @Param id path int true "文件ID"
// @Success 200 {object} response.APIResponse
// @Router /api/admin/files/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.DeleteFile(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

// @Summary 批量删除
// @Description 批量删除文件
// @Tags file
// @Accept json
// @Produce json
// @Param request body []int64 true "文件ID列表"
// @Success 200 {object} response.APIResponse
// @Router /api/admin/files/batch-delete [post]
func (h *Handler) BatchDelete(c *gin.Context) {
	var ids []int64
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Fail(c, err)
		return
	}
	for _, id := range ids {
		if err := h.service.DeleteFile(c.Request.Context(), id); err != nil {
			response.Fail(c, err)
			return
		}
	}
	response.Success(c, nil)
}

// @Summary 分类列表
// @Description 获取所有文件分类
// @Tags file
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /api/admin/file-categories [get]
func (h *Handler) ListCategories(c *gin.Context) {
	categories, err := h.service.ListCategories(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, categories)
}

// @Summary 分类详情
// @Description 根据ID查询分类
// @Tags file
// @Produce json
// @Param id path int true "分类ID"
// @Success 200 {object} response.APIResponse
// @Router /api/admin/file-categories/{id} [get]
func (h *Handler) GetCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	category, err := h.service.GetCategory(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, category)
}

// @Summary 创建分类
// @Description 创建文件分类
// @Tags file
// @Accept json
// @Produce json
// @Param request body FileCategory true "分类信息"
// @Success 200 {object} response.APIResponse
// @Router /api/admin/file-categories [post]
func (h *Handler) CreateCategory(c *gin.Context) {
	var category FileCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		response.Fail(c, err)
		return
	}
	id, err := h.service.CreateCategory(c.Request.Context(), &category)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

// @Summary 更新分类
// @Description 更新文件分类
// @Tags file
// @Accept json
// @Produce json
// @Param id path int true "分类ID"
// @Param request body FileCategory true "分类信息"
// @Success 200 {object} response.APIResponse
// @Router /api/admin/file-categories/{id} [put]
func (h *Handler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var category FileCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		response.Fail(c, err)
		return
	}
	category.ID = id
	if err := h.service.UpdateCategory(c.Request.Context(), &category); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

// @Summary 删除分类
// @Description 删除文件分类
// @Tags file
// @Produce json
// @Param id path int true "分类ID"
// @Success 200 {object} response.APIResponse
// @Router /api/admin/file-categories/{id} [delete]
func (h *Handler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.DeleteCategory(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}
