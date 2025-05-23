package db

import (
	"fmt"
	"log"
	"template-services/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type YugabyteDB struct {
	*gorm.DB
}

func NewYugabyteDB(HOST, PORT, USER, PASSWORD, DB string) *YugabyteDB {

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=10",
		HOST,
		PORT,
		USER,
		PASSWORD,
		DB,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to YugabyteDB: %v", err)
	}

	// Auto migrate the schema
	if err := db.AutoMigrate(&models.Template{}); err != nil {
		log.Fatalf("Failed to migrate schema: %v", err)
	}

	return &YugabyteDB{db}
}

func (y *YugabyteDB) Model(value interface{}) DBModeler {
	return &YugabyteDB{y.DB.Model(value)}
}

func (y *YugabyteDB) Create(value interface{}) DBModeler {
	return &YugabyteDB{y.DB.Create(value)}
}

func (y *YugabyteDB) Where(query interface{}, args ...interface{}) DBModeler {
	return &YugabyteDB{y.DB.Where(query, args...)}
}

func (y *YugabyteDB) Updates(values interface{}) DBModeler {
	return &YugabyteDB{y.DB.Updates(values)}
}

func (y *YugabyteDB) Update(column string, value interface{}) DBModeler {
	return &YugabyteDB{y.DB.Update(column, value)}
}

func (y *YugabyteDB) First(dest interface{}, conds ...interface{}) DBModeler {
	return &YugabyteDB{y.DB.First(dest, conds...)}
}

func (y *YugabyteDB) Find(dest interface{}, conds ...interface{}) DBModeler {
	return &YugabyteDB{y.DB.Find(dest, conds...)}
}

func (y *YugabyteDB) Delete(value interface{}, conds ...interface{}) DBModeler {
	return &YugabyteDB{y.DB.Delete(value, conds...)}
}

func (y *YugabyteDB) Error() error {
	return y.DB.Error
}
