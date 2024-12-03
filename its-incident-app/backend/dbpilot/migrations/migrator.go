package migrations

import (
	"dbpilot/logger"
	"dbpilot/models"
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Migrator struct {
	db *gorm.DB
}

func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) RunMigrations() error {
	logger.Logger.Info("マイグレーションを開始します")

	// マイグレーションの定義
	migrations := []*gormigrate.Migration{
		// マイグレーションIDは一意である必要があります
		{
			ID: "20241201_create_incident_status_tables",
			Migrate: func(tx *gorm.DB) error {
				logger.Logger.Info("IncidentStatus関連のテーブルを作成します")

				// IncidentStatusとIncidentStatusTransitionのテーブルを作成
				if err := tx.AutoMigrate(&models.IncidentStatus{}, &models.IncidentStatusTransition{}); err != nil {
					return fmt.Errorf("ステータス関連テーブルの作成に失敗: %v", err)
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				logger.Logger.Info("IncidentStatus関連のテーブルを削除します")
				if err := tx.Migrator().DropTable("incident_status_transitions", "incident_statuses"); err != nil {
					return fmt.Errorf("ステータス関連テーブルの削除に失敗: %v", err)
				}
				return nil
			},
		},
		{
			ID: "20241202_create_other_tables",
			Migrate: func(tx *gorm.DB) error {
				logger.Logger.Info("その他のテーブルを作成します")

				// その他のテーブルをマイグレーション
				if err := tx.AutoMigrate(
					&models.User{},
					&models.Profile{},
					&models.LoginToken{},
					&models.LoginSession{},
					&models.Response{},
					&models.IncidentRelation{},
					&models.APIResponseData{},
					&models.ErrorLog{},
					&models.EmailData{},
					&models.ProcessingStatus{},
					&models.TokenAccess{},
				); err != nil {
					return fmt.Errorf("その他のテーブルのマイグレーションに失敗: %v", err)
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				logger.Logger.Info("その他のテーブルを削除します")
				if err := tx.Migrator().DropTable(
					"users",
					"profiles",
					"login_tokens",
					"login_sessions",
					"responses",
					"incident_relations",
					"api_response_data",
					"error_logs",
					"email_data",
					"processing_statuses",
					"token_accesses",
				); err != nil {
					return fmt.Errorf("その他のテーブルの削除に失敗: %v", err)
				}
				return nil
			},
		},
		{
			ID: "20241203_insert_initial_incident_statuses",
			Migrate: func(tx *gorm.DB) error {
				logger.Logger.Info("初期ステータスデータを挿入します")

				statusMappings := []struct {
					Code        int
					Name        string
					Description string
					Order       int
				}{
					{0, "未着手", "新規登録されたインシデント", 10},
					{1, "調査中", "担当者が調査を実施中", 20},
					{2, "解決済み", "解決済み", 30},
				}

				for _, sm := range statusMappings {
					status := models.IncidentStatus{
						Code:         sm.Code,
						Name:         sm.Name,
						Description:  sm.Description,
						DisplayOrder: sm.Order,
						IsActive:     true,
					}

					// 既存のステータスがなければ作成
					result := tx.Where("code = ?", status.Code).FirstOrCreate(&status)
					if result.Error != nil {
						return fmt.Errorf("ステータスの作成に失敗(code=%d): %v", status.Code, result.Error)
					}

					logger.Logger.Debug("ステータスを作成しました",
						zap.Int("code", status.Code),
						zap.String("name", status.Name))
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				logger.Logger.Info("初期ステータスデータを削除します")
				if err := tx.Where("code IN ?", []int{0, 1, 2}).Delete(&models.IncidentStatus{}).Error; err != nil {
					return fmt.Errorf("ステータスの削除に失敗: %v", err)
				}
				return nil
			},
		},
		// 追加のマイグレーションがあれば、ここに定義します
		{
			ID: "20241204_migrate_incident_status_column",
			Migrate: func(tx *gorm.DB) error {
				logger.Logger.Info("インシデントテーブルのステータスカラムを移行します")

				// statusカラムの存在確認
				if !tx.Migrator().HasColumn(&models.Incident{}, "status") {
					logger.Logger.Info("statusカラムが存在しないため、データ移行をスキップします")
					// もしstatus_idカラムが存在しなければ、status_idカラムを追加します
					if !tx.Migrator().HasColumn(&models.Incident{}, "status_id") {
						if err := tx.AutoMigrate(&models.Incident{}); err != nil {
							return fmt.Errorf("Incidentテーブルのマイグレーションに失敗: %v", err)
						}
					}
					return nil // エラーとせずに正常終了
				}

				// 以下、statusカラムが存在する場合の処理

				// status_id_newカラムの追加（存在しない場合のみ）
				if !tx.Migrator().HasColumn(&models.Incident{}, "status_id_new") {
					if err := tx.Exec(`ALTER TABLE incidents ADD COLUMN status_id_new INTEGER`).Error; err != nil {
						return fmt.Errorf("status_id_new カラムの追加に失敗: %v", err)
					}
				}

				// 既存データの移行
				type Incident struct {
					ID     uint
					Status string
				}

				var incidents []Incident
				if err := tx.Table("incidents").Select("id, status").Find(&incidents).Error; err != nil {
					return fmt.Errorf("インシデントデータの取得に失敗: %v", err)
				}

				statusMapping := map[string]int{
					"未着手":  0,
					"調査中":  1,
					"解決済み": 2,
				}

				for _, incident := range incidents {
					code, exists := statusMapping[incident.Status]
					if !exists {
						code = 0 // デフォルトで未着手に設定
					}

					// ステータスIDの取得
					var statusID uint
					if err := tx.Model(&models.IncidentStatus{}).
						Where("code = ?", code).
						Select("id").
						Scan(&statusID).Error; err != nil {
						return fmt.Errorf("ステータスIDの取得に失敗(id=%d): %v", incident.ID, err)
					}

					// status_id_newを更新
					if err := tx.Exec(
						`UPDATE incidents SET status_id_new = ? WHERE id = ?`,
						statusID,
						incident.ID,
					).Error; err != nil {
						return fmt.Errorf("ステータスの更新に失敗(id=%d): %v", incident.ID, err)
					}
				}

				// カラムの入れ替え
				if tx.Migrator().HasColumn(&models.Incident{}, "status") {
					if err := tx.Migrator().DropColumn(&models.Incident{}, "status"); err != nil {
						return fmt.Errorf("statusカラムの削除に失敗: %v", err)
					}
				}

				if tx.Migrator().HasColumn(&models.Incident{}, "status_id") {
					if err := tx.Migrator().DropColumn(&models.Incident{}, "status_id"); err != nil {
						return fmt.Errorf("status_idカラムの削除に失敗: %v", err)
					}
				}

				if err := tx.Migrator().RenameColumn(&models.Incident{}, "status_id_new", "status_id"); err != nil {
					return fmt.Errorf("status_id_newのリネームに失敗: %v", err)
				}

				if err := tx.Exec(
					`ALTER TABLE incidents ALTER COLUMN status_id SET NOT NULL`,
				).Error; err != nil {
					return fmt.Errorf("NOT NULL制約の追加に失敗: %v", err)
				}

				if err := tx.Exec(
					`ALTER TABLE incidents ADD CONSTRAINT fk_incident_status FOREIGN KEY (status_id) REFERENCES incident_statuses(id)`,
				).Error; err != nil {
					return fmt.Errorf("外部キー制約の追加に失敗: %v", err)
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				logger.Logger.Info("インシデントテーブルのステータスカラム移行をロールバックします")
				// ロールバック処理を実装（必要に応じて）
				return nil
			},
		},
	}

	// gormigrateのオプション設定（デフォルトを使用）
	options := gormigrate.DefaultOptions

	// マイグレーションの実行
	migrator := gormigrate.New(m.db, options, migrations)
	if err := migrator.Migrate(); err != nil {
		logger.Logger.Error("マイグレーションに失敗しました", zap.Error(err))
		return fmt.Errorf("マイグレーションに失敗しました: %v", err)
	}

	logger.Logger.Info("マイグレーションが正常に完了しました")
	return nil
}
