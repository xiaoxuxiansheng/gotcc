package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/xiaoxuxiansheng/gotcc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TXRecordPO struct {
	gorm.Model
	Status               string `gorm:"status"`
	ComponentTryStatuses string `gorm:"component_try_statuses"`
}

func (t TXRecordPO) TableName() string {
	return "tx_record"
}

type ComponentTryStatus struct {
	ComponentID string `json:"componentID"`
	TryStatus   string `json:"tryStatus"`
}

type TXRecordDAO struct {
	db *gorm.DB
}

func NewTXRecordDAO(db *gorm.DB) *TXRecordDAO {
	return &TXRecordDAO{
		db: db,
	}
}

func (t *TXRecordDAO) GetTXRecords(ctx context.Context, opts ...QueryOption) ([]*TXRecordPO, error) {
	db := t.db.WithContext(ctx).Model(&TXRecordPO{})
	for _, opt := range opts {
		db = opt(db)
	}

	var records []*TXRecordPO
	return records, db.Scan(&records).Error
}

func (t *TXRecordDAO) CreateTXRecord(ctx context.Context, record *TXRecordPO) (uint, error) {
	return record.ID, t.db.WithContext(ctx).Model(&TXRecordPO{}).Create(record).Error
}

func (t *TXRecordDAO) UpdateComponentStatus(ctx context.Context, id uint, componentID string, status string) error {
	return t.LockAndDo(ctx, id, func(ctx context.Context, dao *TXRecordDAO, record *TXRecordPO) error {
		var statuses map[string]*ComponentTryStatus
		if err := json.Unmarshal([]byte(record.ComponentTryStatuses), &statuses); err != nil {
			return err
		}

		componentStatus, ok := statuses[componentID]
		if !ok {
			return fmt.Errorf("invalid component: %s in txid: %d", componentID, id)
		}
		if componentStatus.TryStatus == status {
			return nil
		}

		if componentStatus.TryStatus == gotcc.TryHanging.String() {
			componentStatus.TryStatus = status
			body, _ := json.Marshal(statuses)
			record.ComponentTryStatuses = string(body)
			return dao.UpdateTXRecord(ctx, record)
		}

		return fmt.Errorf("invalid status: %s of component: %s, txid: %d", statuses[componentID].TryStatus, componentID, id)
	})
}

func (t *TXRecordDAO) UpdateTXRecord(ctx context.Context, record *TXRecordPO) error {
	return t.db.WithContext(ctx).Updates(record).Error
}

func (t *TXRecordDAO) LockAndDo(ctx context.Context, id uint, do func(ctx context.Context, dao *TXRecordDAO, record *TXRecordPO) error) error {
	return t.db.Transaction(func(tx *gorm.DB) error {
		// 加写锁
		var record TXRecordPO

		if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&record, id).Error; err != nil {
			return err
		}

		txDAO := NewTXRecordDAO(tx)
		return do(ctx, txDAO, &record)
	})
}
