package role

import (
	"context"
	"e-document-backend/internal/domain"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository defines the interface for role data access
type Repository interface {
	Create(ctx context.Context, role *domain.Role) error
	FindByID(ctx context.Context, id string) (*domain.Role, error)
	FindByName(ctx context.Context, name string) (*domain.Role, error)
	FindAll(ctx context.Context, skip int, limit int, search string) ([]domain.Role, error)
	Count(ctx context.Context, search string) (int, error)
	Update(ctx context.Context, id string, role *domain.Role) error
	Delete(ctx context.Context, id string) error
}

// repository implements the Repository interface
type repository struct {
	collection *mongo.Collection
}

// NewRepository creates a new role repository
func NewRepository(db *mongo.Database) Repository {
	return &repository{
		collection: db.Collection("roles"),
	}
}

//NOTE Create inserts a new role into the database
func (r *repository) Create(ctx context.Context, role *domain.Role) error {
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, role)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	role.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

//NOTE FindByID retrieves a role by ID
func (r *repository) FindByID(ctx context.Context, id string) (*domain.Role, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	var role domain.Role
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&role)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to find role: %w", err)
	}

	return &role, nil
}

//NOTE FindByName retrieves a role by name
func (r *repository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&role)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to find role: %w", err)
	}

	return &role, nil
}

// buildSearchFilter creates a filter for searching users by username or email
func (r *repository) buildSearchFilter(search string) bson.M {
	filter := bson.M{}
	if search != "" {
		filter = bson.M{"name": bson.M{"$regex": search, "$options": "i"}}
	}

	return filter
}

//NOTE FindAll retrieves all roles
func (r *repository) FindAll(ctx context.Context, skip int, limit int, search string) ([]domain.Role, error) {
	opts := options.Find()
	opts.SetSkip(int64(skip))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by newest first

	filter := r.buildSearchFilter(search)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find roles: %w", err)
	}

	defer cursor.Close(ctx)

	var roles []domain.Role
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, fmt.Errorf("failed to decode roles: %w", err)
	}

	return roles, nil
}

// NOTE Count returns the total number of roles
func (r *repository) Count(ctx context.Context, search string) (int, error) {

	// Build filter with search
	filter := r.buildSearchFilter(search)
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count roles: %w", err)
	}

	return int(count), nil
}

//NOTE Update updates a role by ID
func (r *repository) Update(ctx context.Context, id string, role *domain.Role) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	role.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"name":       role.Name,
			"updated_at": role.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}

//NOTE Delete deletes a role by ID
func (r *repository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}
