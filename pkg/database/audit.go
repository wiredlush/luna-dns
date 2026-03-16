//go:build web

package database

import "time"

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
	Username  string    `gorm:"not null" json:"username"`
	Action    string    `gorm:"not null" json:"action"`
	Detail    string    `json:"detail"`
	IP        string    `json:"ip"`
}

const auditRetention = 30 * 24 * time.Hour

func (d *Database) LogAudit(username, action, detail, ip string) {
	d.db.Create(&AuditLog{
		Username: username,
		Action:   action,
		Detail:   detail,
		IP:       ip,
	})
	d.db.Where("timestamp < ?", time.Now().Add(-auditRetention)).Delete(&AuditLog{})
}

func (d *Database) ListAuditLogs() ([]AuditLog, error) {
	var logs []AuditLog
	if err := d.db.Order("timestamp desc").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
