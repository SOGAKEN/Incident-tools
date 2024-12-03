package migrations

import (
	"dbpilot/logger"
	"dbpilot/models"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Migrator はマイグレーションを管理する構造体です
type Migrator struct {
	db *gorm.DB
}

// NewMigrator は新しいMigratorインスタンスを作成します
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// RunMigrations はすべてのマイグレーションを順番に実行します
func (m *Migrator) RunMigrations() error {
	logger.Logger.Info("マイグレーションを開始します")

	// トランザクションでマイグレーションを実行
	return m.db.Transaction(func(tx *gorm.DB) error {
		// 1. 構造的なマイグレーション
		if err := m.runStructuralMigrations(tx); err != nil {
			return fmt.Errorf("構造的マイグレーションに失敗: %v", err)
		}

		// 2. インシデントステータスのマイグレーション
		if err := m.migrateIncidentStatuses(tx); err != nil {
			return fmt.Errorf("インシデントステータスのマイグレーションに失敗: %v", err)
		}

		return nil
	})
}

// runStructuralMigrations はデータベースの構造的な変更を行います
func (m *Migrator) runStructuralMigrations(tx *gorm.DB) error {
	logger.Logger.Info("構造的マイグレーションを実行します")

	// まず、IncidentStatusテーブルとIncidentStatusTransitionテーブルを作成
	if err := tx.AutoMigrate(&models.IncidentStatus{}, &models.IncidentStatusTransition{}); err != nil {
		return fmt.Errorf("ステータス関連テーブルの作成に失敗: %v", err)
	}

	// 次に、それ以外のテーブルをマイグレーション
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
}

// migrateIncidentStatuses はインシデントステータスのマイグレーションを実行します
func (m *Migrator) migrateIncidentStatuses(tx *gorm.DB) error {
	logger.Logger.Info("インシデントステータスのマイグレーションを開始します")

	// 既存のstatus_id_newカラムの存在確認
	hasColumn := tx.Migrator().HasColumn(&models.Incident{}, "status_id_new")

	if !hasColumn {
		logger.Logger.Info("status_id_newカラムを追加します")
		if err := tx.Exec(`ALTER TABLE incidents ADD COLUMN status_id_new INTEGER`).Error; err != nil {
			return fmt.Errorf("status_id_new カラムの追加に失敗: %v", err)
		}
	}

	// 初期ステータスデータの作成
	if err := m.createInitialStatuses(tx); err != nil {
		return fmt.Errorf("初期ステータスの作成に失敗: %v", err)
	}

	// 既存データの移行
	if err := m.migrateExistingData(tx); err != nil {
		return fmt.Errorf("既存データの移行に失敗: %v", err)
	}

	// カラムの入れ替え
	if err := m.finalizeColumnChanges(tx); err != nil {
		return fmt.Errorf("カラム変更の完了に失敗: %v", err)
	}

	logger.Logger.Info("インシデントステータスのマイグレーションが完了しました")
	return nil
}

// createInitialStatuses は初期ステータスデータを作成します
func (m *Migrator) createInitialStatuses(tx *gorm.DB) error {
	logger.Logger.Info("初期ステータスデータを作成します")

	statusMappings := []struct {
		Code        int
		Name        string
		Description string
		Order       int
	}{
		{0, "未着手", "新規登録されたインシデント", 10},
		{1, "調査中", "担当者が調査を実施中", 20},
		{2, "解決済み", "解決済み", 30},
		{99, "解決済み", "解決済み", 100},
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
}

// migrateExistingData は既存のインシデントデータを新しいステータス形式に移行します
func (m *Migrator) migrateExistingData(tx *gorm.DB) error {
	logger.Logger.Info("既存データの移行を開始します")

	// まず、statusカラムの存在確認を行う
	hasStatusColumn := tx.Migrator().HasColumn(&models.Incident{}, "status")
	if !hasStatusColumn {
		logger.Logger.Info("statusカラムが存在しないため、データ移行をスキップします")
		return nil // エラーとせずに正常終了
	}

	statusMapping := map[string]int{
		"未着手":  0,
		"調査中":  1,
		"解決済み": 2,
	}

	var processedCount, errorCount int64

	// statusカラムが存在する場合のみデータ移行を実行
	rows, err := tx.Table("incidents").Select("id, status").Rows()
	if err != nil {
		return fmt.Errorf("インシデントデータの取得に失敗: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id     uint
			status string
		)
		if err := rows.Scan(&id, &status); err != nil {
			return fmt.Errorf("データのスキャンに失敗: %v", err)
		}

		code, exists := statusMapping[status]
		if !exists {
			errorCount++
			logger.Logger.Warn("未知のステータス値を検出しました",
				zap.Uint("incident_id", id),
				zap.String("status", status))
			code = 0 // デフォルトステータスとして未着手を設定
		}

		// ステータスIDを取得
		var statusID uint
		if err := tx.Model(&models.IncidentStatus{}).
			Where("code = ?", code).
			Select("id").
			Scan(&statusID).Error; err != nil {
			return fmt.Errorf("ステータスIDの取得に失敗(id=%d): %v", id, err)
		}

		// status_id_newを更新
		if err := tx.Exec(
			`UPDATE incidents SET status_id_new = ? WHERE id = ?`,
			statusID,
			id,
		).Error; err != nil {
			return fmt.Errorf("ステータスの更新に失敗(id=%d): %v", id, err)
		}

		processedCount++
		logger.Logger.Debug("インシデントのステータスを更新しました",
			zap.Uint("incident_id", id),
			zap.String("old_status", status),
			zap.Uint("new_status_id", statusID))
	}

	logger.Logger.Info("ステータス移行の結果",
		zap.Int64("processed_count", processedCount),
		zap.Int64("error_count", errorCount))

	return nil
}

// finalizeColumnChanges はカラムの変更を完了させます
func (m *Migrator) finalizeColumnChanges(tx *gorm.DB) error {
	logger.Logger.Info("カラム変更の最終処理を開始します")

	// 現在のカラムの状態を確認
	hasStatusID := tx.Migrator().HasColumn(&models.Incident{}, "status_id")
	hasStatusIDNew := tx.Migrator().HasColumn(&models.Incident{}, "status_id_new")
	hasStatus := tx.Migrator().HasColumn(&models.Incident{}, "status")

	logger.Logger.Debug("現在のカラム状態",
		zap.Bool("has_status_id", hasStatusID),
		zap.Bool("has_status_id_new", hasStatusIDNew),
		zap.Bool("has_status", hasStatus))

	statements := []struct {
		sql  string
		desc string
	}{
		{
			sql:  `ALTER TABLE incidents DROP COLUMN IF EXISTS status`,
			desc: "古いstatusカラムの削除",
		},
		{
			sql:  `ALTER TABLE incidents DROP COLUMN IF EXISTS status_id CASCADE`,
			desc: "既存のstatus_idカラムの削除",
		},
		{
			sql:  `ALTER TABLE incidents RENAME COLUMN status_id_new TO status_id`,
			desc: "status_id_newのリネーム",
		},
		{
			sql:  `ALTER TABLE incidents ALTER COLUMN status_id SET NOT NULL`,
			desc: "NOT NULL制約の追加",
		},
		{
			sql: `ALTER TABLE incidents ADD CONSTRAINT fk_incident_status 
                 FOREIGN KEY (status_id) REFERENCES incident_statuses(id)`,
			desc: "外部キー制約の追加",
		},
	}

	for _, stmt := range statements {
		logger.Logger.Debug("SQLステートメントを実行します",
			zap.String("description", stmt.desc),
			zap.String("sql", stmt.sql))

		if err := tx.Exec(stmt.sql).Error; err != nil {
			return fmt.Errorf("カラム変更の実行に失敗(%s): %v", stmt.desc, err)
		}
	}

	logger.Logger.Info("カラム変更の最終処理が完了しました")
	return nil
}
