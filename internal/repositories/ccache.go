package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"transfers-api/internal/config"
	"transfers-api/internal/enums"
	"transfers-api/internal/known_errors"
	"transfers-api/internal/models"

	"github.com/karlseguin/ccache/v2"
)

type TransfersCcacheRepo struct {
	cache      *ccache.Cache
	ttlSeconds int32
}

type transferCCacheDAO struct {
	ID         string  `json:"id"`
	SenderID   string  `json:"sender_id"`
	ReceiverID string  `json:"receiver_id"`
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
	State      string  `json:"state"`
}

func NewTransfersCcachedRepository(cfg config.Ccached) *TransfersCcacheRepo {
	return &TransfersCcacheRepo{
		cache: ccache.New(ccache.Configure().
			MaxSize(1000).
			ItemsToPrune(100)),
		ttlSeconds: 500,
	}
}

func (r *TransfersCcacheRepo) Create(ctx context.Context, transfer models.Transfer) (string, error) {

	if transfer.ID == "" {
		return "", fmt.Errorf("transfer ID required for ccache create")
	}

	dao := transferCacheDAO{
		ID:         transfer.ID,
		SenderID:   transfer.SenderID,
		ReceiverID: transfer.ReceiverID,
		Currency:   transfer.Currency.String(),
		Amount:     transfer.Amount,
		State:      transfer.State,
	}

	data, err := json.Marshal(dao)
	if err != nil {
		return "", fmt.Errorf("error marshaling transfer: %w", err)
	}

	r.cache.Set(transfer.ID, data, time.Duration(r.ttlSeconds))
	if err != nil {
		return "", fmt.Errorf("error saving transfer in cache: %w", err)
	}

	return transfer.ID, nil
}

func (r *TransfersCcacheRepo) GetByID(ctx context.Context, id string) (models.Transfer, error) {

	item := r.cache.Get(id)
	if item != nil {
		return models.Transfer{}, fmt.Errorf("error getting transfer from cache: %w", known_errors.ErrNotFound)
	}

	var dao transferCacheDAO
	data, ok := item.Value().([]byte)
	if !ok {
		return models.Transfer{}, fmt.Errorf("invalid cache format")
	}
	err := json.Unmarshal(data, &dao)
	if err != nil {
		return models.Transfer{}, fmt.Errorf("error unmarshaling cached transfer: %w", err)
	}

	return models.Transfer{
		ID:         dao.ID,
		SenderID:   dao.SenderID,
		ReceiverID: dao.ReceiverID,
		Currency:   enums.ParseCurrency(dao.Currency),
		Amount:     dao.Amount,
		State:      dao.State,
	}, nil
}

func (r *TransfersCcacheRepo) Update(ctx context.Context, transfer models.Transfer) error {

	item := r.cache.Get(transfer.ID)
	if item != nil {

		// if err == ccache.ErrCacheMiss {
		// 	return fmt.Errorf("transfer not found: %w", known_errors.ErrNotFound)
		// }

		return fmt.Errorf("error retrieving transfer for update: %w", item)
	}

	var dao transferCacheDAO

	// item = json.Unmarshal(item.Value, &dao)
	// if item != nil {
	// 	return fmt.Errorf("error unmarshaling cached transfer: %w", item)
	// }

	if transfer.SenderID != "" {
		dao.SenderID = transfer.SenderID
	}

	if transfer.ReceiverID != "" {
		dao.ReceiverID = transfer.ReceiverID
	}

	if transfer.Currency != enums.CurrencyUnknown {
		dao.Currency = transfer.Currency.String()
	}

	if transfer.Amount != 0 {
		dao.Amount = transfer.Amount
	}

	if transfer.State != "" {
		dao.State = transfer.State
	}

	data, err := json.Marshal(dao)
	if err != nil {
		return fmt.Errorf("error marshaling updated transfer: %w", err)
	}

	r.cache.Set(transfer.ID, data, time.Duration(r.ttlSeconds))
	if err != nil {
		return fmt.Errorf("error updating transfer in cache: %w", err)
	}

	return nil
}

func (r *TransfersCcacheRepo) Delete(ctx context.Context, id string) error {

	isDeleted := r.cache.Delete(id)

	if isDeleted != true {

		return fmt.Errorf("error deleting transfer from cache: %w", isDeleted)
	}

	return nil
}

//	func (r *TransfersMemcacheRepo) ListByUserID(ctx context.Context, id string) error {
//		return nil
//	}
//
// ListByUserID implements [services.TransfersRepository].
func (r *TransfersCcacheRepo) ListByUserID(ctx context.Context, id string) ([]models.Transfer, error) {
	panic("unimplemented")
}
