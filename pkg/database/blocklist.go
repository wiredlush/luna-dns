//go:build web

package database

import (
	"github.com/wiredlush/luna-dns/pkg/entry"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BlocklistEntry struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Domain string `gorm:"uniqueIndex;not null" json:"domain"`
}

func (d *Database) SearchBlocklistEntries(search string, limit, offset int) ([]BlocklistEntry, int64, error) {
	var entries []BlocklistEntry
	var count int64

	q := d.db.Model(&BlocklistEntry{})
	if search != "" {
		q = q.Where("domain LIKE ?", "%"+search+"%")
	}

	if err := q.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := q.Order("domain asc").Limit(limit).Offset(offset).Find(&entries).Error; err != nil {
		return nil, 0, err
	}

	return entries, count, nil
}

func (d *Database) IterateBlocklistDomains(batchSize int, fn func(domains []string) error) error {
	var offset int
	for {
		var domains []string
		if err := d.db.Model(&BlocklistEntry{}).Offset(offset).Limit(batchSize).Pluck("domain", &domains).Error; err != nil {
			return err
		}

		if len(domains) == 0 {
			return nil
		}

		if err := fn(domains); err != nil {
			return err
		}

		offset += len(domains)
	}
}

func (d *Database) CreateBlocklistEntry(domain string) (*BlocklistEntry, error) {
	if _, err := entry.NewEntry(domain, "0.0.0.0"); err != nil {
		return nil, err
	}

	e := BlocklistEntry{Domain: domain}
	if err := d.db.Create(&e).Error; err != nil {
		return nil, err
	}

	return &e, nil
}

func (d *Database) DeleteBlocklistEntry(id uint) error {
	return d.db.Delete(&BlocklistEntry{}, id).Error
}

func (d *Database) DeleteAllBlocklistEntries() error {
	return d.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&BlocklistEntry{}).Error
}

func (d *Database) GetBlocklistEntry(id uint) (*BlocklistEntry, error) {
	var e BlocklistEntry
	if err := d.db.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (d *Database) BulkCreateBlocklistEntries(domains []string) (int, error) {
	var valid []BlocklistEntry
	seen := make(map[string]bool)
	for _, domain := range domains {
		if seen[domain] {
			continue
		}

		if _, err := entry.NewEntry(domain, "0.0.0.0"); err != nil {
			continue
		}

		seen[domain] = true
		valid = append(valid, BlocklistEntry{Domain: domain})
	}

	if len(valid) == 0 {
		return 0, nil
	}

	const batchSize = 100
	var totalAffected int64
	for i := 0; i < len(valid); i += batchSize {
		end := min(i+batchSize, len(valid))
		batch := valid[i:end]
		result := d.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&batch)
		if result.Error != nil {
			return int(totalAffected), result.Error
		}
		totalAffected += result.RowsAffected
	}

	return int(totalAffected), nil
}
