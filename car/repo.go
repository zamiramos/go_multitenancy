package car

import (
	"context"
	"github.com/go-kit/kit/log"
	"go_multitenancy/db"
	"go_multitenancy/middlewares"
	"gorm.io/gorm"
)

type repo struct {
	db     db.Handler
	logger log.Logger
}

func NewRepo(db db.Handler, logger log.Logger) Repository {
	return &repo{
		db:     db,
		logger: log.With(logger, "repo", "gorm"),
	}
}
func (repo *repo) getDB(ctx context.Context) *gorm.DB {
	realmId := ctx.Value(middlewares.RealmIdContext)
	if realmId == nil {
		return nil
	}
	db, isInit := repo.db.GetDBHandler(&db.Context{RealmId: realmId.(string)})
	if !isInit {
		db.AutoMigrate(&Car{})
	}

	return db
}

func (repo *repo) CreateCar(ctx context.Context, car Car) error {
	db := repo.getDB(ctx)
	result := db.Create(car)

	return result.Error
}

func (repo *repo) GetCar(ctx context.Context, id string) (Car, error) {
	db := repo.getDB(ctx)

	var car Car
	result := db.First(&car, id)

	return car, result.Error
}

func (repo *repo) DeleteCar(ctx context.Context, id string) error {
	db := repo.getDB(ctx)

	var car Car
	result := db.First(&car, id)
	if result.Error != nil {
		return result.Error
	}

	result = db.Delete(car)

	return result.Error
}
