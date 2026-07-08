package coupon

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"backend/internal/platform/database"
)

type repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) Repository {
	return &repository{dbtx: dbtx}
}

func (r *repository) ListAvailableCoupons(ctx context.Context, userID int64) ([]CouponVO, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	rows, err := exec.QueryContext(ctx,
		`SELECT c.id, c.name, c.type, COALESCE(c.face_value, 0), COALESCE(c.min_order_amount, 0),
		        COALESCE(c.total_quantity, 0), COALESCE(c.claimed_quantity, 0), COALESCE(c.per_user_limit, 1),
		        COALESCE(c.valid_days, 0), c.start_time, c.end_time, COALESCE(c.applicable_scope, '["all"]'),
		        COALESCE((SELECT COUNT(*) FROM user_coupon uc WHERE uc.coupon_id = c.id AND uc.user_id = $1), 0)
		   FROM coupon c
		  WHERE COALESCE(c.enabled, TRUE) = TRUE
		    AND COALESCE(c.distribution_mode, 2) = 2
		    AND (c.start_time IS NULL OR c.start_time <= NOW())
		    AND (c.end_time IS NULL OR c.end_time >= NOW())
		    AND (COALESCE(c.total_quantity, 0) = 0 OR COALESCE(c.claimed_quantity, 0) < COALESCE(c.total_quantity, 0))
		    AND COALESCE(c.deleted, 0) = 0
		  ORDER BY c.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list available coupons: %w", err)
	}
	defer rows.Close()

	result := []CouponVO{}
	for rows.Next() {
		var item CouponVO
		var rawScope string
		var claimedCount int
		if err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.FaceValue, &item.MinOrderAmount, &item.TotalQuantity, &item.ClaimedQuantity, &item.PerUserLimit, &item.ValidDays, &item.StartTime, &item.EndTime, &rawScope, &claimedCount); err != nil {
			return nil, fmt.Errorf("scan available coupon: %w", err)
		}
		item.TypeDesc = typeDesc(item.Type)
		item.ApplicableScope = stringListFromJSON(rawScope)
		item.RemainingQuantity = remainingQuantity(item.TotalQuantity, item.ClaimedQuantity)
		item.Claimed = claimedCount > 0
		item.SoldOut = item.TotalQuantity > 0 && item.ClaimedQuantity >= item.TotalQuantity
		item.Claimable = !item.SoldOut && claimedCount < item.PerUserLimit
		item.ConditionDesc = buildConditionDesc(item.MinOrderAmount)
		result = append(result, item)
	}
	return result, nil
}

func (r *repository) ListUserCoupons(ctx context.Context, userID int64, status *int) ([]UserCouponVO, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where := "WHERE uc.user_id = $1"
	args := []any{userID}
	if status != nil {
		where += " AND uc.status = $2"
		args = append(args, *status)
	}
	rows, err := exec.QueryContext(ctx,
		`SELECT uc.id, uc.coupon_id, c.name, c.type, COALESCE(c.face_value, 0), c.max_discount_amount,
		        COALESCE(c.min_order_amount, 0), COALESCE(c.applicable_scope, '["all"]'), COALESCE(uc.status, 0),
		        COALESCE(uc.source, ''), uc.order_id, uc.claimed_at, uc.used_at, uc.expire_at
		   FROM user_coupon uc
		   JOIN coupon c ON c.id = uc.coupon_id `+where+`
		  ORDER BY uc.claimed_at DESC NULLS LAST, uc.created_at DESC`,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("list user coupons: %w", err)
	}
	defer rows.Close()

	now := time.Now()
	result := []UserCouponVO{}
	for rows.Next() {
		var item UserCouponVO
		var rawScope string
		if err := rows.Scan(&item.ID, &item.CouponID, &item.Name, &item.Type, &item.FaceValue, &item.MaxDiscountAmount, &item.MinOrderAmount, &rawScope, &item.Status, &item.Source, &item.OrderID, &item.ClaimedAt, &item.UsedAt, &item.ExpireAt); err != nil {
			return nil, fmt.Errorf("scan user coupon: %w", err)
		}
		item.TypeDesc = typeDesc(item.Type)
		item.ApplicableScope = stringListFromJSON(rawScope)
		item.StatusDesc = userCouponStatusDesc(item.Status)
		item.SourceDesc = sourceDesc(item.Source)
		item.ConditionDesc = buildConditionDesc(item.MinOrderAmount)
		item.UsableForOrder = item.Status == UserCouponStatusAvailable
		item.ExpiringSoon = item.ExpireAt != nil && item.Status == UserCouponStatusAvailable && item.ExpireAt.After(now) && item.ExpireAt.Before(now.Add(72*time.Hour))
		result = append(result, item)
	}
	return result, nil
}

func (r *repository) ClaimCoupon(ctx context.Context, userID, couponID int64) (int64, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var totalQuantity, claimedQuantity, perUserLimit, validDays int
	var enabled bool
	var startTime, endTime sql.NullTime
	if err := exec.QueryRowContext(ctx,
		`SELECT COALESCE(total_quantity, 0), COALESCE(claimed_quantity, 0), COALESCE(per_user_limit, 1),
		        COALESCE(valid_days, 0), COALESCE(enabled, TRUE), start_time, end_time
		   FROM coupon
		  WHERE id = $1 AND COALESCE(deleted, 0) = 0
		  FOR UPDATE`,
		couponID,
	).Scan(&totalQuantity, &claimedQuantity, &perUserLimit, &validDays, &enabled, &startTime, &endTime); err != nil {
		return 0, fmt.Errorf("get coupon for claim: %w", err)
	}
	now := time.Now()
	if !enabled {
		return 0, fmt.Errorf("coupon is disabled")
	}
	if startTime.Valid && now.Before(startTime.Time) {
		return 0, fmt.Errorf("coupon is not started")
	}
	if endTime.Valid && now.After(endTime.Time) {
		return 0, fmt.Errorf("coupon is expired")
	}
	if totalQuantity > 0 && claimedQuantity >= totalQuantity {
		return 0, fmt.Errorf("coupon is sold out")
	}
	var claimedByUser int
	if err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_coupon WHERE user_id = $1 AND coupon_id = $2`, userID, couponID).Scan(&claimedByUser); err != nil {
		return 0, fmt.Errorf("count user coupon claims: %w", err)
	}
	if perUserLimit > 0 && claimedByUser >= perUserLimit {
		return 0, fmt.Errorf("coupon claim limit exceeded")
	}
	expireAt := now.AddDate(0, 0, validDays)
	var userCouponID int64
	if err := exec.QueryRowContext(ctx,
		`INSERT INTO user_coupon (user_id, coupon_id, status, source, claimed_at, expire_at, created_at)
		 VALUES ($1, $2, $3, 'claim', NOW(), $4, NOW()) RETURNING id`,
		userID, couponID, UserCouponStatusAvailable, expireAt,
	).Scan(&userCouponID); err != nil {
		return 0, fmt.Errorf("insert user coupon: %w", err)
	}
	if _, err := exec.ExecContext(ctx, `UPDATE coupon SET claimed_quantity = COALESCE(claimed_quantity, 0) + 1, updated_at = NOW() WHERE id = $1`, couponID); err != nil {
		return 0, fmt.Errorf("increment coupon claimed quantity: %w", err)
	}
	return userCouponID, nil
}

func (r *repository) ListAdminCoupons(ctx context.Context, query CouponQuery) (CouponPageResult, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where, args := buildAdminWhere(query)
	var total int64
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM coupon c "+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count coupons: %w", err)
	}

	queryArgs := append([]any{}, args...)
	limitIndex := len(queryArgs) + 1
	offsetIndex := len(queryArgs) + 2
	queryArgs = append(queryArgs, query.PageSize, (query.PageNum-1)*query.PageSize)
	rows, err := exec.QueryContext(ctx,
		fmt.Sprintf(`SELECT c.id, c.name, c.type, COALESCE(c.face_value, 0), c.max_discount_amount,
		                    COALESCE(c.min_order_amount, 0), COALESCE(c.total_quantity, 0), COALESCE(c.claimed_quantity, 0),
		                    COALESCE((SELECT COUNT(*) FROM user_coupon uc WHERE uc.coupon_id = c.id AND uc.status = 1), 0),
		                    COALESCE(c.per_user_limit, 1), COALESCE(c.valid_days, 0), c.start_time, c.end_time,
		                    COALESCE(c.applicable_scope, '["all"]'), COALESCE(c.distribution_mode, 2), COALESCE(c.target_level_ids, ''),
		                    COALESCE(c.enabled, TRUE), COALESCE(c.is_permanent, FALSE),
		                    COALESCE(c.restricted_goods_ids, ''), COALESCE(c.restricted_category_ids, ''),
		                    COALESCE(c.created_at, NOW()), COALESCE(c.updated_at, NOW())
		               FROM coupon c %s
		              ORDER BY c.created_at DESC
		              LIMIT $%d OFFSET $%d`, where, limitIndex, offsetIndex),
		queryArgs...,
	)
	if err != nil {
		return nil, fmt.Errorf("list admin coupons: %w", err)
	}
	defer rows.Close()

	now := time.Now()
	page := &CouponPage{Total: total, Current: query.PageNum, Size: query.PageSize, Records: []AdminCouponVO{}}
	for rows.Next() {
		var coupon Coupon
		var rawGoodsIDs, rawCategoryIDs string
		if err := rows.Scan(&coupon.ID, &coupon.Name, &coupon.Type, &coupon.FaceValue, &coupon.MaxDiscountAmount, &coupon.MinOrderAmount, &coupon.TotalQuantity, &coupon.ClaimedQuantity, &coupon.UsedQuantity, &coupon.PerUserLimit, &coupon.ValidDays, &coupon.StartTime, &coupon.EndTime, &coupon.ApplicableScope, &coupon.DistributionMode, &coupon.TargetLevelIDs, &coupon.Enabled, &coupon.IsPermanent, &rawGoodsIDs, &rawCategoryIDs, &coupon.CreatedAt, &coupon.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan admin coupon: %w", err)
		}
		coupon.RestrictedGoodsIDs = int64ListFromJSON(rawGoodsIDs)
		coupon.RestrictedCategoryIDs = int64ListFromJSON(rawCategoryIDs)
		page.Records = append(page.Records, adminVOFromCoupon(coupon, now))
	}
	return page, nil
}

func (r *repository) CreateCoupon(ctx context.Context, req CouponCreateRequest) (int64, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var id int64
	if err := exec.QueryRowContext(ctx,
		`INSERT INTO coupon (
		    name, type, face_value, max_discount_amount, min_order_amount, total_quantity, claimed_quantity,
		    per_user_limit, valid_days, start_time, end_time, applicable_scope, distribution_mode, target_level_ids,
		    enabled, is_permanent, restricted_goods_ids, restricted_category_ids, created_at, updated_at
		 ) VALUES ($1, $2, $3, $4, $5, $6, 0, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW(), NOW())
		 RETURNING id`,
		req.Name, req.Type, req.FaceValue, req.MaxDiscountAmount, req.MinOrderAmount, req.TotalQuantity, req.PerUserLimit,
		req.ValidDays, req.StartTime, req.EndTime, req.ApplicableScope, req.DistributionMode, int64ListToJSON(req.TargetLevelIDs),
		*req.Enabled, *req.IsPermanent, int64ListToJSON(req.RestrictedGoodsIDs), int64ListToJSON(req.RestrictedCategoryIDs),
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("create coupon: %w", err)
	}
	return id, nil
}

func (r *repository) UpdateCoupon(ctx context.Context, id int64, req CouponUpdateRequest) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	sets := []string{}
	args := []any{}
	add := func(column string, value any) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	if req.Name != nil {
		add("name", *req.Name)
	}
	if req.Type != nil {
		add("type", *req.Type)
	}
	if req.FaceValue != nil {
		add("face_value", *req.FaceValue)
	}
	if req.MaxDiscountAmount != nil {
		add("max_discount_amount", *req.MaxDiscountAmount)
	}
	if req.MinOrderAmount != nil {
		add("min_order_amount", *req.MinOrderAmount)
	}
	if req.TotalQuantity != nil {
		add("total_quantity", *req.TotalQuantity)
	}
	if req.PerUserLimit != nil {
		add("per_user_limit", *req.PerUserLimit)
	}
	if req.ValidDays != nil {
		add("valid_days", *req.ValidDays)
	}
	if req.StartTime != nil {
		add("start_time", *req.StartTime)
	}
	if req.EndTime != nil {
		add("end_time", *req.EndTime)
	}
	if req.ApplicableScope != nil {
		add("applicable_scope", *req.ApplicableScope)
	}
	if req.DistributionMode != nil {
		add("distribution_mode", *req.DistributionMode)
	}
	if req.TargetLevelIDsSet {
		add("target_level_ids", int64ListToJSON(req.TargetLevelIDs))
	}
	if req.Enabled != nil {
		add("enabled", *req.Enabled)
	}
	if req.IsPermanent != nil {
		add("is_permanent", *req.IsPermanent)
	}
	if req.RestrictedGoodsIDsSet {
		add("restricted_goods_ids", int64ListToJSON(req.RestrictedGoodsIDs))
	}
	if req.RestrictedCategorySet {
		add("restricted_category_ids", int64ListToJSON(req.RestrictedCategoryIDs))
	}
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE coupon SET %s, updated_at = NOW() WHERE id = $%d", strings.Join(sets, ", "), len(args))
	if _, err := exec.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("update coupon: %w", err)
	}
	return nil
}

func (r *repository) DeleteCoupon(ctx context.Context, id int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx, `DELETE FROM coupon WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete coupon: %w", err)
	}
	return nil
}

func (r *repository) GetStats(ctx context.Context) (*CouponStats, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	stats := &CouponStats{}
	if err := exec.QueryRowContext(ctx,
		`SELECT COUNT(*),
		        COALESCE(SUM(CASE WHEN COALESCE(enabled, TRUE) THEN 1 ELSE 0 END), 0),
		        COALESCE(SUM(COALESCE(claimed_quantity, 0)), 0),
		        COALESCE((SELECT COUNT(*) FROM user_coupon WHERE status = 1), 0)
		   FROM coupon
		  WHERE COALESCE(deleted, 0) = 0`,
	).Scan(&stats.TotalCoupons, &stats.EnabledCoupons, &stats.ClaimedCoupons, &stats.UsedCoupons); err != nil {
		return nil, fmt.Errorf("get coupon stats: %w", err)
	}
	return stats, nil
}

func buildAdminWhere(query CouponQuery) (string, []any) {
	where := "WHERE COALESCE(c.deleted, 0) = 0"
	args := []any{}
	if query.Name != "" {
		args = append(args, "%"+query.Name+"%")
		where += fmt.Sprintf(" AND c.name ILIKE $%d", len(args))
	}
	if query.Type != nil {
		args = append(args, *query.Type)
		where += fmt.Sprintf(" AND c.type = $%d", len(args))
	}
	if query.Enabled != nil {
		args = append(args, *query.Enabled)
		where += fmt.Sprintf(" AND COALESCE(c.enabled, TRUE) = $%d", len(args))
	}
	if query.IsPermanent != nil {
		args = append(args, *query.IsPermanent)
		where += fmt.Sprintf(" AND COALESCE(c.is_permanent, FALSE) = $%d", len(args))
	}
	switch query.Status {
	case CouponStatusAvailable:
		where += " AND COALESCE(c.enabled, TRUE) = TRUE AND (c.start_time IS NULL OR c.start_time <= NOW()) AND (c.end_time IS NULL OR c.end_time >= NOW()) AND (COALESCE(c.total_quantity, 0) = 0 OR COALESCE(c.claimed_quantity, 0) < COALESCE(c.total_quantity, 0))"
	case CouponStatusExpired:
		where += " AND c.end_time IS NOT NULL AND c.end_time < NOW()"
	case CouponStatusSoldOut:
		where += " AND COALESCE(c.total_quantity, 0) > 0 AND COALESCE(c.claimed_quantity, 0) >= COALESCE(c.total_quantity, 0)"
	case CouponStatusNotStarted:
		where += " AND c.start_time IS NOT NULL AND c.start_time > NOW()"
	case CouponStatusDisabled:
		where += " AND COALESCE(c.enabled, TRUE) = FALSE"
	}
	return where, args
}

func adminVOFromCoupon(c Coupon, now time.Time) AdminCouponVO {
	status, statusDesc := couponStatus(c, now)
	restriction, restrictionDesc := restrictionType(c.RestrictedGoodsIDs, c.RestrictedCategoryIDs)
	return AdminCouponVO{
		ID:                    c.ID,
		Name:                  c.Name,
		Type:                  c.Type,
		TypeDesc:              typeDesc(c.Type),
		FaceValue:             c.FaceValue,
		FaceValueDesc:         buildFaceValueDesc(c.Type, c.FaceValue, c.MinOrderAmount),
		MaxDiscountAmount:     c.MaxDiscountAmount,
		MinOrderAmount:        c.MinOrderAmount,
		TotalQuantity:         c.TotalQuantity,
		ClaimedQuantity:       c.ClaimedQuantity,
		RemainingQuantity:     remainingQuantity(c.TotalQuantity, c.ClaimedQuantity),
		UsedQuantity:          c.UsedQuantity,
		PerUserLimit:          c.PerUserLimit,
		ValidDays:             c.ValidDays,
		StartTime:             c.StartTime,
		EndTime:               c.EndTime,
		ApplicableScope:       c.ApplicableScope,
		DistributionMode:      c.DistributionMode,
		TargetLevelIDs:        c.TargetLevelIDs,
		Enabled:               c.Enabled,
		IsPermanent:           c.IsPermanent,
		RestrictedGoodsIDs:    c.RestrictedGoodsIDs,
		RestrictedCategoryIDs: c.RestrictedCategoryIDs,
		RestrictionType:       restriction,
		RestrictionTypeDesc:   restrictionDesc,
		Status:                status,
		StatusDesc:            statusDesc,
		CreatedAt:             c.CreatedAt,
		UpdatedAt:             c.UpdatedAt,
	}
}

func remainingQuantity(total, claimed int) int {
	if total <= 0 {
		return 0
	}
	remaining := total - claimed
	if remaining < 0 {
		return 0
	}
	return remaining
}
