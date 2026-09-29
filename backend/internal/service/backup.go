package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"blog/internal/model"

	"go.uber.org/zap"
)

// 需要备份的数据表（顺序即导出顺序；含软删除数据，备份更完整）
var backupTables = []struct {
	name  string
	model any
}{
	{"users", &model.User{}},
	{"categories", &model.Category{}},
	{"tags", &model.Tag{}},
	{"post_tags", &model.PostTag{}},
	{"posts", &model.Post{}},
	{"comments", &model.Comment{}},
	{"site_configs", &model.SiteConfig{}},
	{"friends", &model.Friend{}},
	{"files", &model.File{}},
	{"musics", &model.Music{}},
}

// BackupService 数据库备份任务：导出全部数据为 CSV，压缩后经 storage 保存
type BackupService struct {
	*Service
}

// NewBackupService 创建备份服务
func NewBackupService(base *Service) *BackupService {
	return &BackupService{Service: base}
}

// RunBackup 执行一次完整备份：各表 → CSV（敏感字段脱敏） → zip → storage.Save（backup.dir 目录，默认 backups/），
// 返回保存的相对路径。
// 说明：实际落盘文件名由 storage.generateKey 生成（时间戳 + crypto/rand 24 位随机十六进制，
// 见 pkg/storage/key.go），不可预测，防止备份文件被遍历/撞库猜中；此处 name 仅用于决定扩展名。
func (s *BackupService) RunBackup() (string, error) {
	s.Log().Info("开始执行数据库备份")
	data, err := s.createBackupZip()
	if err != nil {
		return "", err
	}

	// 备份目录可配置（backup.dir），未配置时默认 backups
	dir := strings.Trim(s.App.Config.Backup.Dir, "/")
	if dir == "" {
		dir = "backups"
	}

	ctx := context.Background()
	name := fmt.Sprintf("db-backup-%s.zip", time.Now().Format("20060102-150405"))
	path, err := s.Storage().Save(ctx, name, data, dir)
	if err != nil {
		return "", fmt.Errorf("保存备份文件失败: %w", err)
	}
	s.Log().Info("数据库备份完成", zap.Int("bytes", len(data)))
	return path, nil
}

// createBackupZip 导出全部表为 CSV 并打成 zip 压缩包
func (s *BackupService) createBackupZip() ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	for _, t := range backupTables {
		csvData, err := s.dumpTableCSV(t.name, t.model)
		if err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("导出表 %s 失败: %w", t.name, err)
		}
		w, err := zw.Create(t.name + ".csv")
		if err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("创建压缩条目失败: %w", err)
		}
		if _, err := w.Write(csvData); err != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("写入压缩条目失败: %w", err)
		}
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("关闭压缩包失败: %w", err)
	}
	return buf.Bytes(), nil
}

// dumpTableCSV 将一张表的数据导出为 CSV（含表头；包含软删除行）
func (s *BackupService) dumpTableCSV(table string, dst any) ([]byte, error) {
	// 表头列名（按模型 schema 顺序，空表也能输出正确表头）
	var cols []string
	if colTypes, err := s.DB().Migrator().ColumnTypes(dst); err == nil {
		for _, ct := range colTypes {
			cols = append(cols, ct.Name())
		}
	}

	// 读取全部行（含软删除）
	var rows []map[string]any
	if err := s.DB().Unscoped().Model(dst).Find(&rows).Error; err != nil {
		return nil, err
	}
	// 兜底：从行数据收集列名并排序，保证输出稳定
	if len(cols) == 0 {
		seen := map[string]bool{}
		for _, row := range rows {
			for k := range row {
				if !seen[k] {
					seen[k] = true
					cols = append(cols, k)
				}
			}
		}
		sort.Strings(cols)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write(cols)
	for _, row := range rows {
		// 导出前对敏感字段脱敏（密码哈希、邮箱、手机号、IP、UA），
		// 降低备份文件泄露时的数据风险
		s.maskRow(table, row)
		record := make([]string, len(cols))
		for i, col := range cols {
			record[i] = csvValue(row[col])
		}
		if err := w.Write(record); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// maskRow 按表对敏感字段执行脱敏。仅处理导出涉及隐私的列，其余数据保持不变。
func (s *BackupService) maskRow(table string, row map[string]any) {
	switch table {
	case "users":
		// 密码哈希必须整体脱敏，防止泄露后离线爆破
		maskField(row, "password", maskSecret)
		maskField(row, "email", maskEmail)
		maskField(row, "phone", maskPhone)
		maskField(row, "last_ip", maskIP)
	case "comments":
		// 评论者联系方式与设备信息
		maskField(row, "email", maskEmail)
		maskField(row, "phone", maskPhone)
		maskField(row, "ip", maskIP)
		maskField(row, "user_agent", maskSecret)
	}
}

// maskField 将 row 中 col 字段的值替换为 mask 处理后的字符串（空值/不存在跳过）
func maskField(row map[string]any, col string, mask func(string) string) {
	v, ok := row[col]
	if !ok || v == nil {
		return
	}
	row[col] = mask(csvValue(v))
}

// maskSecret 整体脱敏为 ***
func maskSecret(s string) string {
	if s == "" {
		return ""
	}
	return "***"
}

// maskEmail 邮箱部分脱敏：保留本地部分首字符与完整域名，如 m***@example.com
func maskEmail(s string) string {
	if s == "" {
		return ""
	}
	at := strings.Index(s, "@")
	if at <= 0 {
		return maskSecret(s)
	}
	local := s[:at]
	domain := s[at:]
	if local == "" {
		return maskSecret(s)
	}
	return local[:1] + "***" + domain
}

// maskPhone 手机号部分脱敏：保留前 3 位与后 4 位，如 138****5678
func maskPhone(s string) string {
	s = strings.TrimSpace(s)
	if len(s) < 7 {
		return maskSecret(s)
	}
	return s[:3] + "****" + s[len(s)-4:]
}

// maskIP IP 部分脱敏：IPv4 保留前两段（1.2.*.*），其余整体脱敏
func maskIP(s string) string {
	s = strings.TrimSpace(s)
	if strings.Contains(s, ":") {
		return maskSecret(s)
	}
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return maskSecret(s)
	}
	return parts[0] + "." + parts[1] + ".*.*"
}

// csvValue 将数据库值格式化为 CSV 字符串
func csvValue(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case time.Time:
		return t.Format(time.RFC3339)
	case []byte:
		return string(t)
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", t)
	}
}
