package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"matriz/internal/models"
)

type EmpresaStore interface {
	Create(ctx context.Context, e *models.Empresa) (string, error)
	Get(ctx context.Context, id string) (*models.Empresa, error)
	GetByCNPJ(ctx context.Context, cnpj string) (*models.Empresa, error)
	List(ctx context.Context) ([]models.Empresa, error)
	Update(ctx context.Context, id string, e *models.Empresa) error
	Delete(ctx context.Context, id string) error
}

type EmpresaRepo struct {
	col *mongo.Collection
}

func NewMongoEmpresaRepo(client *mongo.Client, db, collection string) (*EmpresaRepo, error) {
	col := client.Database(db).Collection(collection)
	// ensure unique index on cnpj
	_, err := col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "cnpj", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return &EmpresaRepo{col: col}, err
}

func (r *EmpresaRepo) Create(ctx context.Context, e *models.Empresa) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := r.col.InsertOne(ctx, e)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", errors.New("cnpj já cadastrado")
		}
		return "", err
	}
	id := res.InsertedID
	switch v := id.(type) {
	case primitive.ObjectID:
		return v.Hex(), nil
	case string:
		return v, nil
	default:
		return "", nil
	}
}

func (r *EmpresaRepo) Get(ctx context.Context, id string) (*models.Empresa, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var e models.Empresa
	objID, err := primitive.ObjectIDFromHex(id)
	if err == nil {
		err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&e)
	} else {
		err = r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&e)
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmpresaRepo) GetByCNPJ(ctx context.Context, cnpj string) (*models.Empresa, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var e models.Empresa
	if err := r.col.FindOne(ctx, bson.M{"cnpj": cnpj}).Decode(&e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EmpresaRepo) List(ctx context.Context) ([]models.Empresa, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cur, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var items []models.Empresa
	for cur.Next(ctx) {
		var e models.Empresa
		if err := cur.Decode(&e); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, cur.Err()
}

func (r *EmpresaRepo) Update(ctx context.Context, id string, e *models.Empresa) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var filter interface{}
	if obj, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": obj}
	} else {
		filter = bson.M{"_id": id}
	}
	_, err := r.col.UpdateOne(ctx, filter, bson.M{"$set": e})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("cnpj já cadastrado")
		}
	}
	return err
}

func (r *EmpresaRepo) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var filter interface{}
	if obj, err := primitive.ObjectIDFromHex(id); err == nil {
		filter = bson.M{"_id": obj}
	} else {
		filter = bson.M{"_id": id}
	}
	_, err := r.col.DeleteOne(ctx, filter)
	return err
}
