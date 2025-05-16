package db

import "gorm.io/gorm"

// DBModeler defines the interface for database operations
type DBModeler interface {
	Model(value interface{}) DBModeler
	Create(value interface{}) DBModeler
	Where(query interface{}, args ...interface{}) DBModeler
	Updates(values interface{}) DBModeler
	Update(column string, value interface{}) DBModeler
	First(dest interface{}, conds ...interface{}) DBModeler
	Find(dest interface{}, conds ...interface{}) DBModeler
	Delete(value interface{}, conds ...interface{}) DBModeler
	Error() error
}


// DBQueryer defines a chainable interface for database operations
// This interface is useful for repositories and testing
type DBQueryer interface {
	Create(interface{}) *gorm.DB
	Where(query string, args ...interface{}) DBQueryer
	Updates(value interface{}) DBQueryer
	Update(column string, value interface{}) DBQueryer
	First(dest interface{}) DBQueryer
	Find(dest interface{}) DBQueryer
	Error() error
}


type GormDBModeler struct {
	db *gorm.DB
}

func NewGormDBModeler(db *gorm.DB) *GormDBModeler {
	return &GormDBModeler{db: db}
}

func (g *GormDBModeler) Model(value interface{}) DBModeler {
	return &GormDBModeler{db: g.db.Model(value)}
}
func (g *GormDBModeler) Create(value interface{}) DBModeler {
	return &GormDBModeler{db: g.db.Create(value)}
}
func (g *GormDBModeler) Where(query interface{}, args ...interface{}) DBModeler {
	return &GormDBModeler{db: g.db.Where(query, args...)}
}
func (g *GormDBModeler) Updates(values interface{}) DBModeler {
	return &GormDBModeler{db: g.db.Updates(values)}
}
func (g *GormDBModeler) Update(column string, value interface{}) DBModeler {
	return &GormDBModeler{db: g.db.Update(column, value)}
}
func (g *GormDBModeler) First(dest interface{}, conds ...interface{}) DBModeler {
	return &GormDBModeler{db: g.db.First(dest, conds...)}
}
func (g *GormDBModeler) Find(dest interface{}, conds ...interface{}) DBModeler {
	return &GormDBModeler{db: g.db.Find(dest, conds...)}
}
func (g *GormDBModeler) Delete(value interface{}, conds ...interface{}) DBModeler {
	return &GormDBModeler{db: g.db.Delete(value, conds...)}
}
func (g *GormDBModeler) Error() error {
	return g.db.Error
}