//go:build web

package database

import "fmt"

type DnsForwarder struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Addr    string `gorm:"not null" json:"addr"`
	Port    int    `gorm:"not null" json:"port"`
	Network string `gorm:"not null" json:"network"`
}

func (f *DnsForwarder) DialAddr() string {
	return fmt.Sprintf("%s:%d", f.Addr, f.Port)
}

func (d *Database) ListDnsForwarders() ([]DnsForwarder, error) {
	var forwarders []DnsForwarder
	result := d.db.Find(&forwarders)
	return forwarders, result.Error
}

func (d *Database) CreateDnsForwarder(addr string, port int, network string) (*DnsForwarder, error) {
	f := DnsForwarder{Addr: addr, Port: port, Network: network}
	result := d.db.Create(&f)
	if result.Error != nil {
		return nil, result.Error
	}
	return &f, nil
}

func (d *Database) GetDnsForwarder(id uint) (*DnsForwarder, error) {
	var f DnsForwarder
	result := d.db.First(&f, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &f, nil
}

func (d *Database) DeleteDnsForwarder(id uint) error {
	return d.db.Delete(&DnsForwarder{}, id).Error
}

func (d *Database) seedDefaultDnsForwarders() error {
	var count int64
	d.db.Model(&DnsForwarder{}).Count(&count)
	if count > 0 {
		return nil
	}

	defaults := []DnsForwarder{
		{Addr: "8.8.8.8", Port: 53, Network: "udp"},
		{Addr: "8.8.4.4", Port: 53, Network: "udp"},
	}

	return d.db.Create(&defaults).Error
}
