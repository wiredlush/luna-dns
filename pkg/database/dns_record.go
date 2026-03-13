//go:build web

package database

type DnsRecord struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Host string `gorm:"not null" json:"host"`
	IP   string `gorm:"not null" json:"ip"`
}

func (d *Database) ListDnsRecords() ([]DnsRecord, error) {
	var records []DnsRecord
	result := d.db.Find(&records)
	return records, result.Error
}

func (d *Database) CreateDnsRecord(host, ip string) (*DnsRecord, error) {
	r := DnsRecord{Host: host, IP: ip}
	result := d.db.Create(&r)
	if result.Error != nil {
		return nil, result.Error
	}
	return &r, nil
}

func (d *Database) GetDnsRecord(id uint) (*DnsRecord, error) {
	var r DnsRecord
	result := d.db.First(&r, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &r, nil
}

func (d *Database) DeleteDnsRecord(id uint) error {
	return d.db.Delete(&DnsRecord{}, id).Error
}
