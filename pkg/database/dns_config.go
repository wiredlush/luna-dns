//go:build web

package database

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type DnsConfig struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Addr      string `gorm:"not null" json:"addr"`
	Port      int    `gorm:"not null;default:5355" json:"port"`
	Network   string `gorm:"not null" json:"network"`
	CacheTTL  int64  `gorm:"not null;default:14400" json:"cache_ttl"`
	AutoStart bool   `gorm:"not null;default:false" json:"auto_start"`
}

func (c *DnsConfig) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.Addr, c.Port)
}

func (d *Database) GetDnsConfig() (*DnsConfig, error) {
	var cfg DnsConfig
	result := d.db.Session(&gorm.Session{Logger: d.db.Logger.LogMode(logger.Silent)}).First(&cfg, 1)
	if result.Error != nil {
		return nil, result.Error
	}

	return &cfg, nil
}

func (d *Database) seedDefaultDnsConfig() error {
	var count int64
	d.db.Model(&DnsConfig{}).Count(&count)
	if count > 0 {
		return nil
	}

	return d.db.Create(&DnsConfig{
		ID:       1,
		Addr:     "0.0.0.0",
		Port:     53,
		Network:  "udp",
		CacheTTL: 14400,
	}).Error
}

func (d *Database) SetDnsAutoStart(autoStart bool) error {
	return d.db.Model(&DnsConfig{}).Where("id = 1").Update("auto_start", autoStart).Error
}

func (d *Database) SaveDnsConfig(addr string, port int, network string, cacheTTL int64) (*DnsConfig, error) {
	cfg := DnsConfig{
		ID:       1,
		Addr:     addr,
		Port:     port,
		Network:  network,
		CacheTTL: cacheTTL,
	}

	result := d.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"addr", "port", "network", "cache_ttl"}),
	}).Create(&cfg)
	if result.Error != nil {
		return nil, result.Error
	}

	return &cfg, nil
}
