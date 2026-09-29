package service

import (
	"errors"

	"gorm.io/gorm"
)

// maxSlugRetryAttempts slug 唯一键冲突重试的最大次数。
// 冲突仅来自「软删除墓碑占位」或并发写入，正常路径不会触发，
// 3 次（含首次尝试）足以覆盖绝大多数情况。
const maxSlugRetryAttempts = 3

// isDuplicateKey 判断错误是否为数据库唯一键冲突（slug 已存在）。
// 依赖 gorm.Config.TranslateError 将驱动错误统一翻译为 gorm.ErrDuplicatedKey
// （MySQL 1062 / PostgreSQL 23505 / SQLite SQLITE_CONSTRAINT_UNIQUE）。
func isDuplicateKey(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}

// withUniqueSlugRetry 执行 attempt；若返回唯一键冲突（典型场景：
// 「软删除墓碑仍占用唯一索引」或并发写入占位导致服务层查重漏判），
// 调用 onConflict 重新生成 slug 后重试，最多 maxAttempts 次；
// 非冲突错误原样返回，不重试。
func withUniqueSlugRetry(maxAttempts int, attempt func() error, onConflict func()) error {
	var err error
	for i := 0; i < maxAttempts; i++ {
		err = attempt()
		if err == nil || !isDuplicateKey(err) {
			return err
		}
		onConflict()
	}
	return err
}
