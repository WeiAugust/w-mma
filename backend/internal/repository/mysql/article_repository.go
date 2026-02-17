package mysqlrepo

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/bajiaozhi/w-mma/backend/internal/model"
	"github.com/bajiaozhi/w-mma/backend/internal/review"
)

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) GetPending(ctx context.Context, pendingID int64) (review.PendingArticle, error) {
	var pending model.PendingArticle
	err := r.db.WithContext(ctx).
		Where("id = ?", pendingID).
		Where("status = ?", "pending").
		Take(&pending).Error
	if err != nil {
		return review.PendingArticle{}, err
	}

	return review.PendingArticle{
		ID:        pending.ID,
		Title:     pending.Title,
		Summary:   pending.Summary,
		SourceURL: pending.SourceURL,
	}, nil
}

func (r *ArticleRepository) PublishArticle(ctx context.Context, rec review.PendingArticle) error {
	article := model.Article{
		Title:       rec.Title,
		Content:     rec.Summary,
		SourceURL:   rec.SourceURL,
		Status:      "published",
		PublishedAt: time.Now(),
	}
	return r.db.WithContext(ctx).Create(&article).Error
}

func (r *ArticleRepository) MarkApproved(ctx context.Context, pendingID int64, reviewerID int64) error {
	reviewedAt := time.Now()
	result := r.db.WithContext(ctx).Model(&model.PendingArticle{}).
		Where("id = ?", pendingID).
		Updates(map[string]any{
			"status":      "approved",
			"reviewer_id": reviewerID,
			"reviewed_at": reviewedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("pending article not found")
	}
	return nil
}

func (r *ArticleRepository) ListPending(ctx context.Context) ([]review.PendingArticle, error) {
	var rows []model.PendingArticle
	if err := r.db.WithContext(ctx).
		Where("status = ?", "pending").
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]review.PendingArticle, 0, len(rows))
	for _, row := range rows {
		items = append(items, review.PendingArticle{
			ID:        row.ID,
			Title:     row.Title,
			Summary:   row.Summary,
			SourceURL: row.SourceURL,
		})
	}
	return items, nil
}

func (r *ArticleRepository) CreatePending(ctx context.Context, item review.PendingArticle) (review.PendingArticle, error) {
	row := model.PendingArticle{
		Title:     item.Title,
		Summary:   item.Summary,
		SourceURL: item.SourceURL,
		Status:    "pending",
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return review.PendingArticle{}, err
	}
	item.ID = row.ID
	return item, nil
}

func (r *ArticleRepository) ListPublished(ctx context.Context) ([]review.PendingArticle, error) {
	var rows []model.Article
	if err := r.db.WithContext(ctx).
		Where("status = ?", "published").
		Order("published_at DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]review.PendingArticle, 0, len(rows))
	for _, row := range rows {
		items = append(items, review.PendingArticle{
			ID:        row.ID,
			Title:     row.Title,
			Summary:   row.Content,
			SourceURL: row.SourceURL,
		})
	}
	return items, nil
}
