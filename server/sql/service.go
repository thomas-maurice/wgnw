package sql

import (
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	proto "github.com/thomas-maurice/wgnw/proto"
	"github.com/thomas-maurice/wgnw/server/interfaces"
)

type SQLWireguardService struct {
	db *gorm.DB
}

func getDatabase(driver string, connString string, verbose bool) (*gorm.DB, error) {
	db, err := gorm.Open(driver, connString)
	if db != nil && verbose {
		db.LogMode(true)
	}

	if driver == "sqlite3" {
		err = db.Exec("PRAGMA foreign_keys = ON").Error
		if err != nil {
			return nil, err
		}
	}

	err = db.AutoMigrate(Network{}, SubNetwork{}, Lease{}).Error
	if err != nil {
		return nil, err
	}
	err = db.Model(&Network{}).AddUniqueIndex("idx_network_name", "name").Error
	if err != nil {
		return nil, err
	}

	return db, err
}

func NewSQLWireguardService(driver string, connString string, verbose bool) (interfaces.WireguardService, error) {
	db, err := getDatabase(driver, connString, verbose)
	if err != nil {
		return nil, err
	}

	return &SQLWireguardService{
		db: db,
	}, nil
}

func (s *SQLWireguardService) CreateNetwork(n *proto.Network) error {
	err := s.db.Create(&Network{
		Name:       n.Name,
		Address:    n.Address,
		NumSubnets: n.NumSubnets,
	}).Error

	if err != nil {
		return err
	}

	for _, sn := range n.Subnets {
		err = s.db.Create(&SubNetwork{
			Parent:  n.Name,
			Address: sn,
		}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLWireguardService) ListNetworks() ([]*proto.Network, error) {
	var networks []Network
	err := s.db.Find(&networks).Error
	if err != nil {
		return nil, err
	}

	var protoNetworks []*proto.Network
	for _, nw := range networks {
		protoNetworks = append(protoNetworks, &proto.Network{Name: nw.Name, Address: nw.Address, NumSubnets: nw.NumSubnets})
	}

	return protoNetworks, nil
}

func (s *SQLWireguardService) GetNetwork(name string) (*proto.Network, error) {
	var network Network
	err := s.db.Where(&Network{Name: name}).First(&network).Error
	if err != nil {
		return nil, err
	}

	var subnets []SubNetwork
	err = s.db.Where(&SubNetwork{Parent: name}).Find(&subnets).Error
	if err != nil {
		return nil, err
	}

	var cidrs []string
	for _, sn := range subnets {
		cidrs = append(cidrs, sn.Address)
	}

	return &proto.Network{
		Name:       network.Name,
		NumSubnets: network.NumSubnets,
		Address:    network.Address,
		Subnets:    cidrs,
	}, nil
}

func (s *SQLWireguardService) DeleteNetwork(name string) error {
	var network Network
	err := s.db.Where(&Network{Name: name}).First(&network).Error
	if err != nil {
		return err
	}

	err = s.db.Delete(&network).Error
	if err != nil {
		return err
	}

	return s.db.Where("parent = ?", network.Name).Delete(SubNetwork{}).Error
}

func (s *SQLWireguardService) AcquireLease(leaseRequest *proto.AcquireLeaseRequest) (*proto.Lease, error) {
	var network Network
	err := s.db.Where(&Network{Name: leaseRequest.NetworkName}).First(&network).Error
	if err != nil {
		return nil, err
	}

	expires := time.Now().Unix() + 600

	tx := s.db.Begin()
	var subnet SubNetwork
	err = tx.First(&subnet, "parent = ? AND free < ?", network.Name, time.Now().Unix()).Error

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Model(&subnet).Updates(&SubNetwork{Free: expires}).Error

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	lease := Lease{
		Parent:    network.Name,
		Expires:   expires,
		PublicKey: leaseRequest.PublicKey,
		Address:   subnet.Address,
		UUID:      uuid.New().String(),
	}

	if leaseRequest.Peer != nil {
		lease.PeerAddress = &leaseRequest.Peer.Address
		lease.PeerPort = leaseRequest.Peer.Port
	}

	err = tx.Create(&lease).Error
	if err != nil {
		return nil, err
	}

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit().Error

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return &proto.Lease{
		IpRange:   lease.Address,
		Network:   lease.Parent,
		Expires:   lease.Expires,
		Uuid:      lease.UUID,
		PublicKey: leaseRequest.PublicKey,
	}, nil
}

func (s *SQLWireguardService) ListLeases() ([]*proto.Lease, error) {
	var lease []Lease
	err := s.db.Find(&lease).Error
	if err != nil {
		return nil, err
	}

	var protoLeases []*proto.Lease
	for _, lease := range lease {
		protoLeases = append(protoLeases, &proto.Lease{
			Uuid:      lease.UUID,
			Expires:   lease.Expires,
			PublicKey: lease.PublicKey,
			Network:   lease.Parent,
			IpRange:   lease.Address,
		})
	}

	return protoLeases, nil
}

func (s *SQLWireguardService) GetLease(id string) (*proto.Lease, error) {
	var lease Lease
	err := s.db.Where(&Lease{UUID: id}).First(&lease).Error
	if err != nil {
		return nil, err
	}

	return &proto.Lease{
		Uuid:      lease.UUID,
		Expires:   lease.Expires,
		PublicKey: lease.PublicKey,
		Network:   lease.Parent,
		IpRange:   lease.Address,
		Expired:   lease.Expires-time.Now().Unix() < 0,
	}, nil
}

func (s *SQLWireguardService) RenewLease(id string) (*proto.Lease, error) {
	var lease Lease
	err := s.db.Where(&Lease{UUID: id}).First(&lease).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return &proto.Lease{
				Uuid:    id,
				Expired: true,
			}, nil
		}
		return nil, err
	}

	// Do not renew an expired lease, they should be GC'ed
	// The client should request a new lease
	if lease.Expires-time.Now().Unix() < 0 {
		return &proto.Lease{
			Uuid:    id,
			Expired: true,
		}, nil
	}

	var subnet SubNetwork
	err = s.db.Where(&SubNetwork{Address: lease.Address}).First(&subnet).Error
	if err != nil {
		return nil, err
	}

	expires := time.Now().Unix() + 600

	err = s.db.Model(&subnet).Updates(&SubNetwork{Free: expires}).Error
	if err != nil {
		return nil, err
	}

	err = s.db.Model(&lease).Updates(&Lease{Expires: expires}).Error
	if err != nil {
		return nil, err
	}

	return &proto.Lease{
		Uuid:      lease.UUID,
		Expires:   expires,
		PublicKey: lease.PublicKey,
		Network:   lease.Parent,
		IpRange:   lease.Address,
		Expired:   false,
	}, nil
}

func (s *SQLWireguardService) DeleteLease(id string) error {
	var lease Lease
	err := s.db.Where(&Lease{UUID: id}).First(&lease).Error
	if err != nil {
		return err
	}
	return s.db.Delete(&lease).Error
}

func (s *SQLWireguardService) PurgeLeases() error {
	return s.db.Where("expires < ?", time.Now().Unix()).Delete(&Lease{}).Error
}

func (s *SQLWireguardService) FetchConfiguration(name string) (*proto.ConfigurationResponse, error) {
	var network Network
	err := s.db.Where(&Network{Name: name}).First(&network).Error
	if err != nil {
		return nil, err
	}

	var leases []Lease
	err = s.db.Where("expires > ? AND parent = ?", time.Now().Unix(), name).Find(&leases).Error
	if err != nil {
		return nil, err
	}

	var endpoints []*proto.Endpoint
	for _, lease := range leases {
		var peer *proto.PublicPeer
		if lease.PeerAddress != nil {
			peer = &proto.PublicPeer{
				Address: *lease.PeerAddress,
				Port:    lease.PeerPort,
			}
		}

		endpoints = append(endpoints, &proto.Endpoint{
			Peer:      peer,
			PublicKey: lease.PublicKey,
			Networks:  []string{lease.Address},
		})
	}

	return &proto.ConfigurationResponse{
		Network: &proto.NetworkDefinition{
			Name:      network.Name,
			Address:   network.Address,
			Endpoints: endpoints,
		},
	}, nil
}
