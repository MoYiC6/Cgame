package feedback

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"backend/internal/platform/database"
)

type repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) Repository {
	return &repository{dbtx: dbtx}
}

func (r *repository) CreateFeedback(ctx context.Context, feedback *Feedback) error {
	images, _ := json.Marshal(feedback.Images)
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO feedback (user_id, ticket_no, content, images, status)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		feedback.UserID, feedback.TicketNo, feedback.Content, string(images), feedback.Status,
	).Scan(&feedback.ID)
}

func (r *repository) ListUserFeedback(ctx context.Context, userID int64, page, pageSize int) (*FeedbackPage, error) {
	return r.listFeedback(ctx, "WHERE f.user_id = $1", []interface{}{userID}, page, pageSize)
}

func (r *repository) ListAdminFeedback(ctx context.Context, page, pageSize int, status *int, keyword string) (*FeedbackPage, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	if status != nil {
		where += fmt.Sprintf(" AND f.status = $%d", len(args)+1)
		args = append(args, *status)
	}
	if strings.TrimSpace(keyword) != "" {
		where += fmt.Sprintf(" AND (f.ticket_no ILIKE $%d OR f.content ILIKE $%d)", len(args)+1, len(args)+1)
		args = append(args, "%"+strings.TrimSpace(keyword)+"%")
	}
	return r.listFeedback(ctx, where, args, page, pageSize)
}

func (r *repository) GetFeedbackDetail(ctx context.Context, userID *int64, id int64) (*FeedbackDetailVO, error) {
	where := "WHERE id = $1"
	args := []interface{}{id}
	if userID != nil {
		where += " AND user_id = $2"
		args = append(args, *userID)
	}
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, ticket_no, content, images, status, created_at, updated_at FROM feedback `+where,
		args...,
	)
	var rawImages []byte
	detail := &FeedbackDetailVO{}
	if err := row.Scan(&detail.ID, &detail.TicketNo, &detail.Content, &rawImages, &detail.Status, &detail.CreatedAt, &detail.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get feedback detail: %w", err)
	}
	detail.Images = decodeImages(rawImages)
	detail.StatusText = statusText(detail.Status)
	replies, err := r.listReplies(ctx, id)
	if err != nil {
		return nil, err
	}
	detail.Replies = replies
	return detail, nil
}

func (r *repository) CreateReply(ctx context.Context, reply *FeedbackReply) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO feedback_reply (feedback_id, reply_user_id, reply_type, content)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		reply.FeedbackID, reply.ReplyUserID, reply.ReplyType, reply.Content,
	).Scan(&reply.ID)
}

func (r *repository) UpdateStatus(ctx context.Context, id int64, status int) error {
	_, err := r.dbtx.ExecContext(ctx, `UPDATE feedback SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	if err != nil {
		return fmt.Errorf("update feedback status: %w", err)
	}
	return nil
}

func (r *repository) DeleteFeedback(ctx context.Context, id int64) error {
	_, err := r.dbtx.ExecContext(ctx, `DELETE FROM feedback WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete feedback: %w", err)
	}
	return nil
}

func (r *repository) listFeedback(ctx context.Context, where string, args []interface{}, page, pageSize int) (*FeedbackPage, error) {
	var total int64
	if err := r.dbtx.QueryRowContext(ctx, "SELECT COUNT(*) FROM feedback f "+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count feedback: %w", err)
	}

	queryArgs := append([]interface{}{}, args...)
	limitIndex := len(queryArgs) + 1
	offsetIndex := len(queryArgs) + 2
	queryArgs = append(queryArgs, pageSize, (page-1)*pageSize)
	rows, err := r.dbtx.QueryContext(ctx,
		fmt.Sprintf(`SELECT f.id, f.ticket_no, f.user_id, f.content, f.images, f.status, f.created_at, f.updated_at,
		                    COALESCE((SELECT COUNT(*) FROM feedback_reply fr WHERE fr.feedback_id = f.id), 0)
		             FROM feedback f %s
		             ORDER BY f.created_at DESC
		             LIMIT $%d OFFSET $%d`, where, limitIndex, offsetIndex),
		queryArgs...,
	)
	if err != nil {
		return nil, fmt.Errorf("list feedback: %w", err)
	}
	defer rows.Close()

	result := &FeedbackPage{Total: total, Records: []FeedbackVO{}}
	for rows.Next() {
		var item FeedbackVO
		var rawImages []byte
		if err := rows.Scan(&item.ID, &item.TicketNo, &item.UserID, &item.Content, &rawImages, &item.Status, &item.CreatedAt, &item.UpdatedAt, &item.ReplyCount); err != nil {
			return nil, fmt.Errorf("scan feedback: %w", err)
		}
		item.Images = decodeImages(rawImages)
		item.StatusText = statusText(item.Status)
		result.Records = append(result.Records, item)
	}
	return result, nil
}

func (r *repository) listReplies(ctx context.Context, feedbackID int64) ([]FeedbackReplyVO, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, reply_type, content, created_at FROM feedback_reply WHERE feedback_id = $1 ORDER BY created_at ASC`,
		feedbackID,
	)
	if err != nil {
		return nil, fmt.Errorf("list feedback replies: %w", err)
	}
	defer rows.Close()

	replies := []FeedbackReplyVO{}
	for rows.Next() {
		var reply FeedbackReplyVO
		if err := rows.Scan(&reply.ID, &reply.ReplyType, &reply.Content, &reply.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan feedback reply: %w", err)
		}
		reply.ReplyTypeText = replyTypeText(reply.ReplyType)
		replies = append(replies, reply)
	}
	return replies, nil
}

func decodeImages(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var images []string
	if err := json.Unmarshal(raw, &images); err != nil {
		return []string{}
	}
	return images
}
