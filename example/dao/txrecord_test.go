package dao

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/xiaoxuxiansheng/gotcc"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Test_GetTXRecod(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()
	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"VERSION"}).AddRow("1"))

	gdb, err := gorm.Open(mysql.New(mysql.Config{
		Conn: db,
	}), &gorm.Config{
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Error(err)
		return
	}

	ctx := context.Background()
	txRecordDAO := NewTXRecordDAO(gdb)

	tests := []struct {
		name string
		f    func()
	}{
		{
			name: "GetTXRecords",
			f: func() {
				componentStatus := map[string]*ComponentTryStatus{
					"component_a": {
						ComponentID: "component_a",
						TryStatus:   gotcc.TryHanging.String(),
					},
				}
				body, _ := json.Marshal(componentStatus)

				now := time.Now()
				rows := sqlmock.NewRows([]string{"id", "create_at", "deleted_at", "updated_at", "status", "component_try_statuses"}).AddRow(1, now, now, now, gotcc.TXHanging.String(), string(body))
				mock.ExpectQuery("SELECT \\* FROM `tx_record` WHERE id = \\? AND status = \\? AND `tx_record`.`deleted_at` IS NULL").WithArgs(1, gotcc.TXHanging.String()).WillReturnRows(rows)
				txRecords, err := txRecordDAO.GetTXRecords(ctx, WithID(1), WithStatus(gotcc.ComponentTryStatus(gotcc.TXHanging.String())))
				if err != nil {
					t.Error(err)
					return
				}
				assert.Equal(t, 1, len(txRecords))
				assert.Equal(t, uint(1), txRecords[0].ID)
				assert.Equal(t, gotcc.TXHanging.String(), txRecords[0].Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f()
		})
	}
}

func Test_CreateTXRecod(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()
	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"VERSION"}).AddRow("1"))

	gdb, err := gorm.Open(mysql.New(mysql.Config{
		Conn: db,
	}), &gorm.Config{
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Error(err)
		return
	}

	ctx := context.Background()
	txRecordDAO := NewTXRecordDAO(gdb)

	tests := []struct {
		name string
		f    func()
	}{
		{
			name: "CreateTXRecord",
			f: func() {
				componentStatus := map[string]*ComponentTryStatus{
					"component_a": {
						ComponentID: "component_a",
						TryStatus:   gotcc.TryHanging.String(),
					},
				}
				body, _ := json.Marshal(componentStatus)

				now := time.Now()
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `tx_record`").WithArgs(now, now, nil, gotcc.TXHanging.String(), string(body)).WillReturnResult(driver.ResultNoRows)
				mock.ExpectCommit()
				_, err := txRecordDAO.CreateTXRecord(ctx, &TXRecordPO{
					Status:               gotcc.TXHanging.String(),
					ComponentTryStatuses: string(body),
					Model: gorm.Model{
						CreatedAt: now,
						UpdatedAt: now,
					},
				})
				assert.Equal(t, nil, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f()
		})
	}
}

func Test_UpdateComponentStatus(t *testing.T) {
	now := time.Now()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()
	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"VERSION"}).AddRow("1"))

	gdb, err := gorm.Open(mysql.New(mysql.Config{
		Conn: db,
	}), &gorm.Config{
		DisableAutomaticPing: true,
		NowFunc: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	ctx := context.Background()
	txRecordDAO := NewTXRecordDAO(gdb)

	tests := []struct {
		name string
		f    func()
	}{
		{
			name: "UpdateComponentStatusSuccess",
			f: func() {
				componentStatus := map[string]*ComponentTryStatus{
					"component_a": {
						ComponentID: "component_a",
						TryStatus:   gotcc.TryHanging.String(),
					},
				}
				body, _ := json.Marshal(componentStatus)

				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id", "create_at", "deleted_at", "updated_at", "status", "component_try_statuses"}).AddRow(1, now, nil, now, gotcc.TXHanging.String(), string(body))
				mock.ExpectQuery("SELECT \\* FROM `tx_record` WHERE `tx_record`.`id` = \\? AND `tx_record`.`deleted_at` IS NULL ORDER BY `tx_record`.`id` LIMIT 1 FOR UPDATE").WithArgs(1).WillReturnRows(rows)

				newComponentStatus := map[string]*ComponentTryStatus{
					"component_a": {
						ComponentID: "component_a",
						TryStatus:   gotcc.TrySucceesful.String(),
					},
				}
				newBody, _ := json.Marshal(newComponentStatus)
				mock.ExpectExec("UPDATE `tx_record` SET `updated_at`=\\?,`status`=\\?,`component_try_statuses`=\\? WHERE `tx_record`.`deleted_at` IS NULL AND `id` = \\?").WithArgs(now, gotcc.TXHanging.String(), string(newBody), 1).WillReturnResult(driver.ResultNoRows)
				mock.ExpectCommit()
				err := txRecordDAO.UpdateComponentStatus(ctx, 1, "component_a", gotcc.TrySucceesful.String())
				assert.Equal(t, nil, err)
			},
		},

		{
			name: "UpdateComponentStatusMissComponent",
			f: func() {
				componentStatus := map[string]*ComponentTryStatus{
					"component_a": {
						ComponentID: "component_a",
						TryStatus:   gotcc.TryHanging.String(),
					},
				}
				body, _ := json.Marshal(componentStatus)

				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id", "create_at", "deleted_at", "updated_at", "status", "component_try_statuses"}).AddRow(1, now, nil, now, gotcc.TXHanging.String(), string(body))
				mock.ExpectQuery("SELECT \\* FROM `tx_record` WHERE `tx_record`.`id` = \\? AND `tx_record`.`deleted_at` IS NULL ORDER BY `tx_record`.`id` LIMIT 1 FOR UPDATE").WithArgs(1).WillReturnRows(rows)

				mock.ExpectRollback()
				err := txRecordDAO.UpdateComponentStatus(ctx, 1, "component_b", gotcc.TrySucceesful.String())
				assert.Equal(t, true, err != nil)
			},
		},
		{
			name: "UpdateComponentStatusStatusFail",
			f: func() {
				componentStatus := map[string]*ComponentTryStatus{
					"component_a": {
						ComponentID: "component_a",
						TryStatus:   gotcc.TryFailure.String(),
					},
				}
				body, _ := json.Marshal(componentStatus)

				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id", "create_at", "deleted_at", "updated_at", "status", "component_try_statuses"}).AddRow(1, now, nil, now, gotcc.TXHanging.String(), string(body))
				mock.ExpectQuery("SELECT \\* FROM `tx_record` WHERE `tx_record`.`id` = \\? AND `tx_record`.`deleted_at` IS NULL ORDER BY `tx_record`.`id` LIMIT 1 FOR UPDATE").WithArgs(1).WillReturnRows(rows)

				mock.ExpectCommit()
				err := txRecordDAO.UpdateComponentStatus(ctx, 1, "component_a", gotcc.TrySucceesful.String())
				assert.Equal(t, true, err != nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f()
		})
	}
}

func Test_UpdateTXRecord(t *testing.T) {
	now := time.Now()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()
	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"VERSION"}).AddRow("1"))

	gdb, err := gorm.Open(mysql.New(mysql.Config{
		Conn: db,
	}), &gorm.Config{
		DisableAutomaticPing: true,
		NowFunc: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	ctx := context.Background()
	txRecordDAO := NewTXRecordDAO(gdb)

	tests := []struct {
		name string
		f    func()
	}{
		{
			name: "UpdateTXRecord",
			f: func() {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `tx_record` SET `updated_at`=\\?,`status`=\\?,`component_try_statuses`=\\? WHERE `tx_record`.`deleted_at` IS NULL AND `id` = \\?").WithArgs(now, gotcc.TXHanging.String(), "{}", 1).WillReturnResult(driver.ResultNoRows)
				mock.ExpectCommit()
				err := txRecordDAO.UpdateTXRecord(ctx, &TXRecordPO{
					Model: gorm.Model{
						ID: 1,
					},
					Status:               gotcc.TXHanging.String(),
					ComponentTryStatuses: "{}",
				})
				assert.Equal(t, nil, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f()
		})
	}
}
