package user

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

// Repository defines the interface for user data access
type Repository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindAll(ctx context.Context, skip int, limit int, search string, currentUserID string) ([]domain.User, error)
	Count(ctx context.Context, search string, currentUserID string) (int, error)
	Update(ctx context.Context, id string, user *domain.User) error
	Delete(ctx context.Context, id string) error
}

// repository implements the Repository interface
type repository struct {
	collection *mongo.Collection
}

// NewRepository creates a new user repository
func NewRepository(db *mongo.Database) Repository {
	return &repository{
		collection: db.Collection("users"),
	}
}

// NOTE Create inserts a new user into the database
func (r *repository) Create(ctx context.Context, user *domain.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// NOTE FindByID retrieves a user by ID
func (r *repository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var user domain.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// NOTE FindByUsername retrieves a user by username
func (r *repository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// NOTE FindByEmail retrieves a user by email
func (r *repository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// buildSearchFilter creates a filter for searching users by username or email
func (r *repository) buildSearchFilter(search string) bson.M {
	filter := bson.M{}
	if search != "" {
		filter = bson.M{
			"$or": []bson.M{
				{"username": bson.M{"$regex": search, "$options": "i"}},
				{"email": bson.M{"$regex": search, "$options": "i"}},
			},
		}
	}
	return filter
}

// NOTE FindAll retrieves all users excluding the current user
func (r *repository) FindAll(ctx context.Context, skip int, limit int, search string, currentUserID string) ([]domain.User, error) {
	opts := options.Find()
	opts.SetSkip(int64(skip))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by newest first

	// Build filter with search
	filter := r.buildSearchFilter(search)

	// Exclude current user from results
	if currentUserID != "" {
		objectID, err := primitive.ObjectIDFromHex(currentUserID)
		if err == nil {
			// If filter already has conditions, combine them with $and
			if len(filter) > 0 {
				filter = bson.M{
					"$and": []bson.M{
						filter,
						{"_id": bson.M{"$ne": objectID}},
					},
				}
			} else {
				filter["_id"] = bson.M{"$ne": objectID}
			}
		}
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

// NOTE Count returns the total number of users (excluding current user)
func (r *repository) Count(ctx context.Context, search string, currentUserID string) (int, error) {
	// Build filter with search
	filter := r.buildSearchFilter(search)

	// Exclude current user from count
	if currentUserID != "" {
		objectID, err := primitive.ObjectIDFromHex(currentUserID)
		if err == nil {
			// If filter already has conditions, combine them with $and
			if len(filter) > 0 {
				filter = bson.M{
					"$and": []bson.M{
						filter,
						{"_id": bson.M{"$ne": objectID}},
					},
				}
			} else {
				filter["_id"] = bson.M{"$ne": objectID}
			}
		}
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return int(count), nil
}

// NOTE Update updates a user by ID
func (r *repository) Update(ctx context.Context, id string, user *domain.User) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	user.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"username":        user.Username,
			"email":           user.Email,
			"role":            user.Role,
			"phone":           user.Phone,
			"first_name":      user.FirstName,
			"last_name":       user.LastName,
			"password":        user.Password,
			"department_id":   user.DepartmentID,
			"sector_id":       user.SectorID,
			"profile_picture": user.ProfilePicture,
			"updated_at":      user.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// NOTE Delete deletes a user by ID
func (r *repository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
