# Database Migrations

ระบบ migration สำหรับจัดการการเปลี่ยนแปลง schema และข้อมูลใน MongoDB

## การใช้งาน

### 1. ดูสถานะ Migration

```bash
make migrate-status
```

แสดงรายการ migrations ทั้งหมดและสถานะว่า applied แล้วหรือยัง

### 2. รัน Migrations

```bash
make migrate-up
```

รัน migrations ทั้งหมดที่ยังไม่ได้ apply

### 3. Rollback Migration

```bash
make migrate-down
```

Rollback migration ล่าสุดที่ apply ไปแล้ว

## การสร้าง Migration ใหม่

### ขั้นตอนการสร้าง:

1. **สร้างไฟล์ migration ใหม่** ใน `migrations/` directory:

   ตั้งชื่อตามรูปแบบ: `XXX_description.go` (เช่น `003_add_role_field.go`)

2. **เขียน migration function**:

```go
package migrations

import (
	"context"
	"e-document-backend/internal/migration"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func Migration003_AddRoleField() migration.MigrationDefinition {
	return migration.MigrationDefinition{
		Version:     "003",
		Name:        "add_role_field",
		Description: "Add role field to users with default value 'user'",
		Up:          migration003Up,
		Down:        migration003Down,
	}
}

func migration003Up(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	filter := bson.M{"role": bson.M{"$exists": false}}
	update := bson.M{"$set": bson.M{"role": "user"}}

	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	log.Printf("  Added role field to %d users", result.ModifiedCount)
	return nil
}

func migration003Down(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	filter := bson.M{}
	update := bson.M{"$unset": bson.M{"role": ""}}

	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	log.Printf("  Removed role field from %d users", result.ModifiedCount)
	return nil
}
```

3. **Register migration** ใน `registry.go`:

```go
func GetAll() []migration.MigrationDefinition {
	return []migration.MigrationDefinition{
		Migration001_AddPhoneFieldToUsers(),
		Migration002_CreateEmailIndex(),
		Migration003_AddRoleField(), // เพิ่มตรงนี้
	}
}
```

4. **ทดสอบ migration**:

```bash
# ดูสถานะก่อน
make migrate-status

# รัน migration
make migrate-up

# ตรวจสอบผลลัพธ์
make migrate-status

# ถ้ามีปัญหา rollback ได้
make migrate-down
```

## ตัวอย่างการใช้งาน

### เพิ่ม Field ใหม่

```go
func migration004Up(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	// เพิ่ม field ให้กับ documents ที่ยังไม่มี field นี้
	filter := bson.M{"new_field": bson.M{"$exists": false}}
	update := bson.M{"$set": bson.M{"new_field": "default_value"}}

	_, err := collection.UpdateMany(ctx, filter, update)
	return err
}
```

### สร้าง Index

```go
func migration005Up(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("username_idx"),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	return err
}
```

### Rename Field

```go
func migration006Up(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	// Rename field จาก old_name เป็น new_name
	filter := bson.M{}
	update := bson.M{"$rename": bson.M{"old_name": "new_name"}}

	_, err := collection.UpdateMany(ctx, filter, update)
	return err
}
```

### Update Field Type

```go
func migration007Up(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("users")

	// แปลง string เป็น int
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		// แปลงค่าและ update
		if strValue, ok := doc["age"].(string); ok {
			intValue, _ := strconv.Atoi(strValue)
			filter := bson.M{"_id": doc["_id"]}
			update := bson.M{"$set": bson.M{"age": intValue}}
			collection.UpdateOne(ctx, filter, update)
		}
	}

	return nil
}
```

## Best Practices

1. **Version Number**: ใช้ตัวเลข 3 หลักเรียงตามลำดับ (001, 002, 003)
2. **Descriptive Names**: ตั้งชื่อให้บอกถึงสิ่งที่ migration ทำ
3. **Always Test**: ทดสอบทั้ง up และ down functions
4. **Idempotent**: Migration ควรรันได้หลายครั้งโดยไม่เกิดปัญหา
5. **Backup**: Backup database ก่อนรัน migration ใน production
6. **Down Function**: ควรมี down function เสมอเพื่อ rollback ได้

## ข้อควรระวัง

- MongoDB เป็น schema-less database ดังนั้น field ใหม่จะถูกเพิ่มใน documents ที่สร้างหลังจากนี้โดยอัตโนมัติ
- Migration จำเป็นเมื่อ:
  - ต้องการ update ข้อมูลเก่าให้มี field ใหม่
  - สร้าง index เพื่อ performance
  - Rename หรือ migrate ข้อมูล
  - แปลงชนิดข้อมูล

## โครงสร้างไฟล์

```
migrations/
├── README.md                          # เอกสารนี้
├── registry.go                         # Registry สำหรับ register migrations
├── 001_example_add_phone_field.go     # ตัวอย่าง: เพิ่ม field
└── 002_example_create_email_index.go  # ตัวอย่าง: สร้าง index
```
