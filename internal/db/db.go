package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/supakornemchananon/go-llm-proxy-server/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type DB interface {
	SaveConnection(ctx context.Context, conn *models.Connection) error
	GetConnection(ctx context.Context, id string) (*models.Connection, error)
	ListConnections(ctx context.Context) ([]models.Connection, error)
	DeleteConnection(ctx context.Context, id string) error

	SaveVirtualKey(ctx context.Context, vk *models.VirtualKey) error
	GetVirtualKey(ctx context.Context, key string) (*models.VirtualKey, error)
	ListVirtualKeys(ctx context.Context) ([]models.VirtualKey, error)
	DeleteVirtualKey(ctx context.Context, id string) error
}

type SQLDB struct {
	db *gorm.DB
}

func (s *SQLDB) SaveConnection(ctx context.Context, conn *models.Connection) error {
	return s.db.WithContext(ctx).Save(conn).Error
}

func (s *SQLDB) GetConnection(ctx context.Context, id string) (*models.Connection, error) {
	var conn models.Connection
	err := s.db.WithContext(ctx).First(&conn, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (s *SQLDB) ListConnections(ctx context.Context) ([]models.Connection, error) {
	var conns []models.Connection
	err := s.db.WithContext(ctx).Find(&conns).Error
	return conns, err
}

func (s *SQLDB) DeleteConnection(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.Connection{}, "id = ?", id).Error
}

func (s *SQLDB) SaveVirtualKey(ctx context.Context, vk *models.VirtualKey) error {
	return s.db.WithContext(ctx).Save(vk).Error
}

func (s *SQLDB) GetVirtualKey(ctx context.Context, key string) (*models.VirtualKey, error) {
	var vk models.VirtualKey
	err := s.db.WithContext(ctx).Where("key = ?", key).First(&vk).Error
	if err != nil {
		return nil, err
	}
	return &vk, nil
}

func (s *SQLDB) ListVirtualKeys(ctx context.Context) ([]models.VirtualKey, error) {
	var vks []models.VirtualKey
	err := s.db.WithContext(ctx).Find(&vks).Error
	return vks, err
}

func (s *SQLDB) DeleteVirtualKey(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.VirtualKey{}, "id = ?", id).Error
}

type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
}

func (m *MongoDB) SaveConnection(ctx context.Context, conn *models.Connection) error {
	coll := m.db.Collection("connections")
	_, err := coll.UpdateOne(ctx, bson.M{"_id": conn.ID}, bson.M{"$set": conn}, options.UpdateOne().SetUpsert(true))
	return err
}

func (m *MongoDB) GetConnection(ctx context.Context, id string) (*models.Connection, error) {
	coll := m.db.Collection("connections")
	var conn models.Connection
	err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&conn)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (m *MongoDB) ListConnections(ctx context.Context) ([]models.Connection, error) {
	coll := m.db.Collection("connections")
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var conns []models.Connection
	err = cursor.All(ctx, &conns)
	return conns, err
}

func (m *MongoDB) DeleteConnection(ctx context.Context, id string) error {
	coll := m.db.Collection("connections")
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (m *MongoDB) SaveVirtualKey(ctx context.Context, vk *models.VirtualKey) error {
	coll := m.db.Collection("virtual_keys")
	_, err := coll.UpdateOne(ctx, bson.M{"_id": vk.ID}, bson.M{"$set": vk}, options.UpdateOne().SetUpsert(true))
	return err
}

func (m *MongoDB) GetVirtualKey(ctx context.Context, key string) (*models.VirtualKey, error) {
	coll := m.db.Collection("virtual_keys")
	var vk models.VirtualKey
	err := coll.FindOne(ctx, bson.M{"key": key}).Decode(&vk)
	if err != nil {
		return nil, err
	}
	return &vk, nil
}

func (m *MongoDB) ListVirtualKeys(ctx context.Context) ([]models.VirtualKey, error) {
	coll := m.db.Collection("virtual_keys")
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var vks []models.VirtualKey
	err = cursor.All(ctx, &vks)
	return vks, err
}

func (m *MongoDB) DeleteVirtualKey(ctx context.Context, id string) error {
	coll := m.db.Collection("virtual_keys")
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func InitDB(dbType, dsn string) (DB, error) {
	switch strings.ToLower(dbType) {
	case "sqlite":
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		db.AutoMigrate(&models.Connection{}, &models.VirtualKey{})
		return &SQLDB{db: db}, nil
	case "postgres":
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		db.AutoMigrate(&models.Connection{}, &models.VirtualKey{})
		return &SQLDB{db: db}, nil
	case "mssql":
		db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		db.AutoMigrate(&models.Connection{}, &models.VirtualKey{})
		return &SQLDB{db: db}, nil
	case "mongodb":
		client, err := mongo.Connect(options.Client().ApplyURI(dsn))
		if err != nil {
			return nil, err
		}
		dbName := "llm_proxy"
		return &MongoDB{client: client, db: client.Database(dbName)}, nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}
