# Fix: Create Incoming Docs with dept_id for Outgoing Documents

## Problem
When a user creates an outgoing document with recipient departments, the backend doesn't create `incoming_docs` records with the `dept_id` field set.

## Root Cause
The backend's `CreateIncomingDocs` function in the `incomingdoc` service doesn't set the `dept_id` when creating incoming documents.

## Solution

### 1. Update IncomingDoc Domain Model

**File: `internal/domain/incomingdoc.go`** (or similar)

Ensure the IncomingDoc struct has the `DeptID` field:

```go
type IncomingDoc struct {
    ID           uuid.UUID
    IncomingNo   string
    IncomingDate *time.Time
    ReceivedDate *time.Time
    Status       string
    DocDetailsID uuid.UUID
    FolderID     *uuid.UUID
    CreatedBy    *uuid.UUID
    UpdatedBy    *uuid.UUID
    ApproverID   *uuid.UUID
    ApproverDate *time.Time
    Remark       string
    UpdatedAt    time.Time
    DeptID       *uuid.UUID  // ← NEW FIELD
}
```

### 2. Update IncomingDoc Service

**File: `internal/app/incomingdoc/service.go`**

Update the `CreateIncomingDocs` method to set `DeptID`:

```go
func (s *service) CreateIncomingDocs(ctx context.Context, docID uuid.UUID, createdByID *uuid.UUID, deptIDs []uuid.UUID, remark string) error {
    for _, deptID := range deptIDs {
        incomingNo := generateIncomingNumber() // Generate unique incoming number
        
        incomingDoc := &domain.IncomingDoc{
            IncomingNo:   incomingNo,
            DocDetailsID: docID,
            Status:       "pending",
            CreatedBy:    createdByID,
            Remark:       remark,
            DeptID:       &deptID,  // ← SET DEPARTMENT ID
            IncomingDate: &now,
        }
        
        if err := s.repo.Create(ctx, incomingDoc); err != nil {
            return fmt.Errorf("failed to create incoming doc for dept %s: %w", deptID, err)
        }
        
        log.Info().
            Str("doc_id", docID.String()).
            Str("dept_id", deptID.String()).
            Str("incoming_no", incomingNo).
            Msg("Created incoming doc with department")
    }
    return nil
}
```

### 3. Update IncomingDoc Repository

**File: `internal/app/incomingdoc/repository.go`** (or `repository_postgres.go`)

Update the `Create` method to include `dept_id` in INSERT statement:

```go
func (r *Repository) Create(ctx context.Context, incomingDoc *domain.IncomingDoc) error {
    query := `
        INSERT INTO incoming_docs 
        (incoming_no, incoming_date, status, doc_details_id, created_by, remark, dept_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at
    `
    
    return r.db.QueryRowContext(ctx, query,
        incomingDoc.IncomingNo,
        incomingDoc.IncomingDate,
        incomingDoc.Status,
        incomingDoc.DocDetailsID,
        incomingDoc.CreatedBy,
        incomingDoc.Remark,
        incomingDoc.DeptID,  // ← ADD dept_id PARAMETER
    ).Scan(&incomingDoc.ID, &incomingDoc.CreatedAt)
}
```

Also update the `GetByID`, `Update`, and `Scan` methods to handle the `dept_id` field:

```go
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.IncomingDoc, error) {
    query := `
        SELECT id, incoming_no, incoming_date, received_date, status, 
               doc_details_id, folder_id, created_by, updated_by, 
               approver_id, approver_date, remark, created_at, dept_id
        FROM incoming_docs
        WHERE id = $1
    `
    
    row := r.db.QueryRowContext(ctx, query, id)
    doc := &domain.IncomingDoc{}
    
    err := row.Scan(
        &doc.ID, &doc.IncomingNo, &doc.IncomingDate, &doc.ReceivedDate,
        &doc.Status, &doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy,
        &doc.UpdatedBy, &doc.ApproverID, &doc.ApproverDate, &doc.Remark,
        &doc.CreatedAt, &doc.DeptID,  // ← ADD dept_id PARAMETER
    )
    
    return doc, err
}
```

### 4. Migration Already Created

The migration `000013_add_dept_id_to_incoming_docs` is already created. Just run:

```bash
make migrate-up
# or
go run ./cmd/migrate/main.go
```

---

## Testing

After implementing these changes:

### 1. Run Migration
```bash
make migrate-up
```

### 2. Create a Document with Recipients

Upload a document with:
- `target_module=outgoing`
- `receiver_ids=dept-id-1,dept-id-2,dept-id-3`
- `description=Test document`

### 3. Verify incoming_docs Created

```sql
SELECT 
    id, 
    incoming_no, 
    dept_id, 
    status, 
    created_at 
FROM incoming_docs 
ORDER BY created_at DESC 
LIMIT 5;
```

You should see `dept_id` populated for each incoming document.

### 4. Check Outgoing Docs API

```bash
curl http://localhost:5001/api/v1/outgoing-docs?page=1&limit=10
```

The response should include:
```json
{
  "recipients": [
    {
      "department_id": "...",
      "department_name": "IT Department",
      "status": "pending"
    }
  ],
  "status_counts": {
    "pending": 1,
    "received": 0,
    "approved": 0,
    "rejected": 0
  }
}
```

---

## Frontend Changes Already Done

✅ Changed metadata field from `recipient_department_ids` to `receiver_ids`  
✅ OutboundTab displays recipients and status counts  
✅ DocumentDetailModal shows all departments and their statuses  

---

## Data Flow (Complete)

```
1. User selects departments in Add New Document modal
2. Frontend sends: receiver_ids=dept1,dept2,dept3
3. File uploaded via TUS protocol
4. Backend:
   a. Creates doc_details record
   b. Calls handleOutgoingModule
   c. Creates outgoing_doc record
   d. Calls handleIncomingModule
   e. Creates incoming_doc records with dept_id set for each department
5. Frontend fetches outgoing docs with recipients info
6. Display shows:
   - Department names
   - Status counts (Pending/Received/Approved/Rejected)
```
