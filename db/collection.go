package db

import (
	"context"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection mongo-driver collection
type Collection struct {
	*mongo.Collection
}

// Col returns the collection.
func Col(dbc *mongo.Database, name string) *Collection {
	return &Collection{dbc.Collection(name)}
}

// All returns all results from the cursor
func (c *Collection) All(filter interface{}, opts *options.FindOptions, result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cur, err := c.Find(ctx, filter, opts)
	defer cur.Close(ctx)
	if err != nil {
		return err
	}
	if err = cur.Err(); err != nil {
		return err
	}

	resultv := reflect.ValueOf(result)
	slicev := resultv.Elem()
	if slicev.Kind() == reflect.Interface {
		slicev = slicev.Elem()
	}
	slicev = slicev.Slice(0, slicev.Cap())
	elemt := slicev.Type().Elem()
	i := 0

	for {
		elemp := reflect.New(elemt)
		if !cur.Next(nil) {
			break
		}
		err := cur.Decode(elemp.Interface())
		if err != nil {
			return err
		}
		slicev = reflect.Append(slicev, elemp.Elem())
		i++
	}
	resultv.Elem().Set(slicev.Slice(0, i))
	return nil
}

// One returns one result from the cursor
func (c *Collection) One(filter interface{}, opts *options.FindOneOptions, result interface{}) error {
	if opts == nil {
		opts = options.FindOne()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.FindOne(ctx, filter, opts).Decode(result); err != nil {
		return err
	}

	return nil
}

// Insert inserts a single document into the collection and returns insert one result.
func (c *Collection) Insert(document interface{}) (result *mongo.InsertOneResult, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err = c.InsertOne(ctx, document)
	return
}

// InsertAll inserts the provided documents and returns insert many result.
func (c *Collection) InsertAll(documents []interface{}) (result *mongo.InsertManyResult, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err = c.InsertMany(ctx, documents)
	return
}

// Update updates a single document in the collection.
func (c *Collection) Update(selector interface{}, update interface{}, upsert ...bool) error {
	if selector == nil {
		selector = primitive.D{}
	}

	var err error

	opt := options.Update()
	for _, arg := range upsert {
		if arg {
			opt.SetUpsert(arg)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err = c.UpdateOne(ctx, selector, update, opt); err != nil {
		return err
	}
	return nil
}

// UpdateID updates a single document in the collection by id
func (c *Collection) UpdateID(id interface{}, update interface{}) error {
	return c.Update(primitive.M{"_id": id}, update)
}

// UpdateAll updates multiple documents in the collection.
func (c *Collection) UpdateAll(selector interface{}, update interface{}, upsert ...bool) (*mongo.UpdateResult, error) {
	if selector == nil {
		selector = primitive.D{}
	}

	var err error

	opt := options.Update()
	for _, arg := range upsert {
		if arg {
			opt.SetUpsert(arg)
		}
	}

	var updateResult *mongo.UpdateResult
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if updateResult, err = c.UpdateMany(ctx, selector, update, opt); err != nil {
		return updateResult, err
	}
	return updateResult, nil
}

// Remove deletes a single document from the collection.
func (c *Collection) Remove(selector interface{}) error {
	if selector == nil {
		selector = primitive.D{}
	}
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err = c.DeleteOne(ctx, selector); err != nil {
		return err
	}
	return nil
}

// RemoveID deletes a single document from the collection by id.
func (c *Collection) RemoveID(id interface{}) error {
	return c.Remove(id)
}

// RemoveAll deletes multiple documents from the collection.
func (c *Collection) RemoveAll(selector interface{}) error {
	if selector == nil {
		selector = primitive.D{}
	}
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err = c.DeleteMany(ctx, selector); err != nil {
		return err
	}
	return nil
}

// Count gets the number of documents matching the filter.
func (c *Collection) Count(selector interface{}) (int64, error) {
	if selector == nil {
		selector = primitive.D{}
	}
	var err error
	var count int64
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	count, err = c.CountDocuments(ctx, selector)
	return count, err
}

func (c *Collection) Create(data interface{}, result interface{}) error {
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	filter := primitive.M{"_id": primitive.NewObjectIDFromTimestamp(time.Now())}
	update := primitive.M{
		"$set":         data,
		"$currentDate": primitive.M{"updated_at": true, "created_at": true},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r := c.FindOneAndUpdate(ctx, filter, update, opts)
	if err := r.Err(); err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if result != nil {
		if err := r.Decode(result); err != nil {
			return err
		}
	}
	return nil
}

// AggregatePipe process data records and return computed results
func (c *Collection) AggregatePipe(pipe mongo.Pipeline, result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cur, err := c.Aggregate(ctx, pipe)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	return cur.All(ctx, result)
}

// FindDistinct finds the distinct values for a specified field across a single collection
func (c *Collection) FindDistinct(filter interface{}, fieldName string, opts *options.DistinctOptions) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return c.Distinct(ctx, fieldName, filter, opts)
}

// Modify uses $set to modify matching records
func (c *Collection) Modify(filter interface{}, update interface{}, result interface{}) error {
	after := options.After
	opts := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	updateQ := primitive.M{
		"$set":         update,
		"$currentDate": primitive.M{"updated_at": true, "created_at": true},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	r := c.FindOneAndUpdate(ctx, filter, updateQ, &opts)
	if err := r.Err(); err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if result != nil {
		if err := r.Decode(result); err != nil {
			return err
		}
	}
	return nil
}
