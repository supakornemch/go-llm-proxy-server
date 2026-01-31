package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/supakornemchananon/go-llm-proxy-server/internal/cryptoutil"
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

	SaveProviderModel(ctx context.Context, pm *models.ProviderModel) error
	GetProviderModel(ctx context.Context, id string) (*models.ProviderModel, error)
	ListProviderModels(ctx context.Context, connectionID string) ([]models.ProviderModel, error)
	DeleteProviderModel(ctx context.Context, id string) error

	SaveVirtualKey(ctx context.Context, vk *models.VirtualKey) error
	GetVirtualKey(ctx context.Context, key string) (*models.VirtualKey, error)
	ListVirtualKeys(ctx context.Context) ([]models.VirtualKey, error)
	DeleteVirtualKey(ctx context.Context, id string) error

	SaveVirtualKeyAssignment(ctx context.Context, vka *models.VirtualKeyAssignment) error
	GetVirtualKeyAssignment(ctx context.Context, virtualKeyID, modelAlias string) (*models.VirtualKeyAssignment, error)
	ListVirtualKeyAssignments(ctx context.Context, virtualKeyID string) ([]models.VirtualKeyAssignment, error)
	DeleteVirtualKeyAssignment(ctx context.Context, id string) error
}

type SQLDB struct {
	db *gorm.DB
}

func (s *SQLDB) SaveConnection(ctx context.Context, conn *models.Connection) error {
	if conn.APIKey != "" {
		encrypted, err := cryptoutil.Encrypt(conn.APIKey)
		if err == nil {
			conn.APIKey = encrypted
		}
	}
	return s.db.WithContext(ctx).Save(conn).Error
}

func (s *SQLDB) GetConnection(ctx context.Context, id string) (*models.Connection, error) {
	var conn models.Connection
	err := s.db.WithContext(ctx).First(&conn, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	if conn.APIKey != "" {
		decrypted, err := cryptoutil.Decrypt(conn.APIKey)
		if err == nil {
			conn.APIKey = decrypted
		}
	}
	return &conn, nil
}

func (s *SQLDB) ListConnections(ctx context.Context) ([]models.Connection, error) {
	var conns []models.Connection
	err := s.db.WithContext(ctx).Find(&conns).Error
	if err == nil {
		for i := range conns {
			if conns[i].APIKey != "" {
				decrypted, _ := cryptoutil.Decrypt(conns[i].APIKey)
				conns[i].APIKey = decrypted
			}
		}
	}
	return conns, err
}

func (s *SQLDB) DeleteConnection(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.Connection{}, "id = ?", id).Error
}

func (s *SQLDB) SaveProviderModel(ctx context.Context, pm *models.ProviderModel) error {
	return s.db.WithContext(ctx).Save(pm).Error
}

func (s *SQLDB) GetProviderModel(ctx context.Context, id string) (*models.ProviderModel, error) {
	var pm models.ProviderModel
	err := s.db.WithContext(ctx).First(&pm, "id = ?", id).Error
	return &pm, err
}

func (s *SQLDB) ListProviderModels(ctx context.Context, connectionID string) ([]models.ProviderModel, error) {
	var pms []models.ProviderModel
	q := s.db.WithContext(ctx)
	if connectionID != "" {
		q = q.Where("connection_id = ?", connectionID)
	}
	err := q.Find(&pms).Error
	return pms, err
}

func (s *SQLDB) DeleteProviderModel(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.ProviderModel{}, "id = ?", id).Error
}

func (s *SQLDB) SaveVirtualKey(ctx context.Context, vk *models.VirtualKey) error {
	if vk.Key != "" {
		vk.KeyHash = cryptoutil.HashKey(vk.Key)
		encrypted, err := cryptoutil.Encrypt(vk.Key)
		if err == nil {
			vk.Key = encrypted
		}
	}
	return s.db.WithContext(ctx).Save(vk).Error
}

func (s *SQLDB) GetVirtualKey(ctx context.Context, key string) (*models.VirtualKey, error) {
	var vk models.VirtualKey
	hash := cryptoutil.HashKey(key)
	err := s.db.WithContext(ctx).Where("key_hash = ?", hash).First(&vk).Error
	if err != nil {
		return nil, err
	}
	if vk.Key != "" {
		decrypted, err := cryptoutil.Decrypt(vk.Key)
		if err == nil {
			vk.Key = decrypted
		}
	}
	return &vk, nil
}

func (s *SQLDB) ListVirtualKeys(ctx context.Context) ([]models.VirtualKey, error) {
	var vks []models.VirtualKey
	err := s.db.WithContext(ctx).Find(&vks).Error
	if err == nil {
		for i := range vks {
			if vks[i].Key != "" {
				decrypted, _ := cryptoutil.Decrypt(vks[i].Key)
				vks[i].Key = decrypted
			}
		}
	}
	return vks, err
}

func (s *SQLDB) DeleteVirtualKey(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.VirtualKey{}, "id = ?", id).Error
}

func (s *SQLDB) SaveVirtualKeyAssignment(ctx context.Context, vka *models.VirtualKeyAssignment) error {
	return s.db.WithContext(ctx).Save(vka).Error
}

func (s *SQLDB) GetVirtualKeyAssignment(ctx context.Context, virtualKeyID, modelAlias string) (*models.VirtualKeyAssignment, error) {
	var vka models.VirtualKeyAssignment
	err := s.db.WithContext(ctx).Where("virtual_key_id = ? AND model_alias = ?", virtualKeyID, modelAlias).First(&vka).Error
	return &vka, err
}

func (s *SQLDB) ListVirtualKeyAssignments(ctx context.Context, virtualKeyID string) ([]models.VirtualKeyAssignment, error) {
	var vkas []models.VirtualKeyAssignment
	q := s.db.WithContext(ctx)
	if virtualKeyID != "" {
		q = q.Where("virtual_key_id = ?", virtualKeyID)
	}
	err := q.Find(&vkas).Error
	return vkas, err
}

func (s *SQLDB) DeleteVirtualKeyAssignment(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.VirtualKeyAssignment{}, "id = ?", id).Error
}

type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
}

func (m *MongoDB) SaveConnection(ctx context.Context, conn *models.Connection) error {
	coll := m.db.Collection("connections")
	if conn.APIKey != "" {
		encrypted, err := cryptoutil.Encrypt(conn.APIKey)
		if err == nil {
			conn.APIKey = encrypted
		}
	}
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
	if conn.APIKey != "" {
		decrypted, err := cryptoutil.Decrypt(conn.APIKey)
		if err == nil {
			conn.APIKey = decrypted
		}
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
	if err == nil {
		for i := range conns {
			if conns[i].APIKey != "" {
				decrypted, _ := cryptoutil.Decrypt(conns[i].APIKey)
				conns[i].APIKey = decrypted
			}
		}
	}
	return conns, err
}

func (m *MongoDB) DeleteConnection(ctx context.Context, id string) error {
	coll := m.db.Collection("connections")
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (m *MongoDB) SaveProviderModel(ctx context.Context, pm *models.ProviderModel) error {
	coll := m.db.Collection("provider_models")
	_, err := coll.UpdateOne(ctx, bson.M{"_id": pm.ID}, bson.M{"$set": pm}, options.UpdateOne().SetUpsert(true))
	return err
}

func (m *MongoDB) GetProviderModel(ctx context.Context, id string) (*models.ProviderModel, error) {
	coll := m.db.Collection("provider_models")
	var pm models.ProviderModel
	err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&pm)
	return &pm, err
}

func (m *MongoDB) ListProviderModels(ctx context.Context, connectionID string) ([]models.ProviderModel, error) {
	coll := m.db.Collection("provider_models")
	filter := bson.M{}
	if connectionID != "" {
		filter["connection_id"] = connectionID
	}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var pms []models.ProviderModel
	err = cursor.All(ctx, &pms)
	return pms, err
}

func (m *MongoDB) DeleteProviderModel(ctx context.Context, id string) error {
	coll := m.db.Collection("provider_models")
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (m *MongoDB) SaveVirtualKey(ctx context.Context, vk *models.VirtualKey) error {
	coll := m.db.Collection("virtual_keys")
	if vk.Key != "" {
		vk.KeyHash = cryptoutil.HashKey(vk.Key)
		encrypted, err := cryptoutil.Encrypt(vk.Key)
		if err == nil {
			vk.Key = encrypted
		}
	}
	_, err := coll.UpdateOne(ctx, bson.M{"_id": vk.ID}, bson.M{"$set": vk}, options.UpdateOne().SetUpsert(true))
	return err
}

func (m *MongoDB) GetVirtualKey(ctx context.Context, key string) (*models.VirtualKey, error) {
	coll := m.db.Collection("virtual_keys")
	var vk models.VirtualKey
	hash := cryptoutil.HashKey(key)
	err := coll.FindOne(ctx, bson.M{"key_hash": hash}).Decode(&vk)
	if err != nil {
		return nil, err
	}
	if vk.Key != "" {
		decrypted, err := cryptoutil.Decrypt(vk.Key)
		if err == nil {
			vk.Key = decrypted
		}
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
	if err == nil {
		for i := range vks {
			if vks[i].Key != "" {
				decrypted, _ := cryptoutil.Decrypt(vks[i].Key)
				vks[i].Key = decrypted
			}
		}
	}
	return vks, err
}

func (m *MongoDB) DeleteVirtualKey(ctx context.Context, id string) error {
	coll := m.db.Collection("virtual_keys")
	_, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (m *MongoDB) SaveVirtualKeyAssignment(ctx context.Context, vka *models.VirtualKeyAssignment) error {
	coll := m.db.Collection("virtual_key_assignments")
	_, err := coll.UpdateOne(ctx, bson.M{"_id": vka.ID}, bson.M{"$set": vka}, options.UpdateOne().SetUpsert(true))
	return err
}

func (m *MongoDB) GetVirtualKeyAssignment(ctx context.Context, virtualKeyID, modelAlias string) (*models.VirtualKeyAssignment, error) {
	coll := m.db.Collection("virtual_key_assignments")
	var vka models.VirtualKeyAssignment
	err := coll.FindOne(ctx, bson.M{"virtual_key_id": virtualKeyID, "model_alias": modelAlias}).Decode(&vka)
	return &vka, err
}

func (m *MongoDB) ListVirtualKeyAssignments(ctx context.Context, virtualKeyID string) ([]models.VirtualKeyAssignment, error) {
	coll := m.db.Collection("virtual_key_assignments")
	filter := bson.M{}
	if virtualKeyID != "" {
		filter["virtual_key_id"] = virtualKeyID
	}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var vkas []models.VirtualKeyAssignment
	err = cursor.All(ctx, &vkas)
	return vkas, err
}

func (m *MongoDB) DeleteVirtualKeyAssignment(ctx context.Context, id string) error {
	coll := m.db.Collection("virtual_key_assignments")
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
		db.AutoMigrate(&models.Connection{}, &models.ProviderModel{}, &models.VirtualKey{}, &models.VirtualKeyAssignment{})
		return &SQLDB{db: db}, nil
	case "postgres":
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		db.AutoMigrate(&models.Connection{}, &models.ProviderModel{}, &models.VirtualKey{}, &models.VirtualKeyAssignment{})
		return &SQLDB{db: db}, nil
	case "mssql":
		db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		db.AutoMigrate(&models.Connection{}, &models.ProviderModel{}, &models.VirtualKey{}, &models.VirtualKeyAssignment{})
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
