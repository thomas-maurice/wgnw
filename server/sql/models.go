package sql

type Network struct {
	//ID         int64  `gorm"column:id;type:bigint;unique"`
	Name       string `gorm:"column:name;type:varchar(128);unique;primary_key"`
	Address    string `gorm:"column:address;type:varchar(64)"`
	NumSubnets int32  `gorm:"column:subnets;type:integer"`
}

func (t Network) TableName() string {
	return "network"
}

type SubNetwork struct {
	ID      int64  `gorm:"column:id;auto_increment"`
	Address string `gorm:"column:address;type:varchar(64)"`
	Parent  string `gorm:"column:parent;type:varchar(128) references network(name) on delete cascade on update no action"`
	Free    int64  `gorm:"column:free;type:bigint"`
}

func (t SubNetwork) TableName() string {
	return "subnetwork"
}

type Lease struct {
	ID          int64   `gorm:"column:id;auto_increment"`
	Parent      string  `gorm:"column:parent;type:varchar(128) references network(name) on delete cascade on update no action"`
	Expires     int64   `gorm:"column:expires;type:bigint"`
	Address     string  `gorm:"column:address"`
	PublicKey   string  `gorm:"column:public_key;not null"`
	PeerAddress *string `gorm:"column:peer_address"`
	PeerPort    int32   `gorm:"column:peer_port"`
	UUID        string  `gorm:"column:lease_uuid;not null"`
}

func (t Lease) TableName() string {
	return "lease"
}
