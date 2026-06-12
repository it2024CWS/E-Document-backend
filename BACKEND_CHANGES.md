# Backend Implementation Guide for Outgoing Documents with Department Recipients

## 📋 Database Schema Changes

### Migration: `000013_add_dept_id_to_incoming_docs`

Added `dept_id` column to `incoming_docs` table to track which department received each document.

```sql
ALTER TABLE incoming_docs
ADD COLUMN dept_id UUID REFERENCES departments(id) ON DELETE SET NULL;
```

**Run Migration:**
```bash
make migrate-up
# or
go run ./cmd/migrate/main.go
```

---

## 🔧 Backend Service Changes Required

### 1. Update Outgoing Document Model

**File: `internal/app/outgoingdoc/model.go`** (or similar)

Add struct fields to support recipients:

```go
type OutgoingDoc struct {
    ID              string    `json:"id"`
    OutgoingNo      string    `json:"outgoing_no"`
    DocDetailsID    string    `json:"doc_details_id"`
    DocNo           string    `json:"doc_no"`
    DocName         string    `json:"doc_name"`
    CreatedBy       string    `json:"created_by"`
    CreatorName     string    `json:"creator_name"`
    CreatedAt       time.Time `json:"created_at"`
    // NEW FIELDS
    Recipients      []RecipientDept `json:"recipients,omitempty"`
    StatusCounts    StatusCounts    `json:"status_counts,omitempty"`
}

type RecipientDept struct {
    DepartmentID   string        `json:"department_id"`
    DepartmentName string        `json:"department_name"`
    Status         string        `json:"status"` // from incoming_doc status
    IncomingDoc    *IncomingDocDetail `json:"incoming_doc,omitempty"`
}

type IncomingDocDetail struct {
    ID           string `json:"id"`
    IncomingNo   string `json:"incoming_no"`
    Status       string `json:"status"`
    ReceivedDate *time.Time `json:"received_date,omitempty"`
}

type StatusCounts struct {
    Pending  int `json:"pending"`
    Received int `json:"received"`
    Approved int `json:"approved"`
    Rejected int `json:"rejected"`
}
```

### 2. Update Create Outgoing Document Handler

**File: `internal/app/outgoingdoc/handler.go`** (CreateHandler method)

When a document is uploaded with `recipient_department_ids`, create incoming_docs for each:

```go
func (h *Handler) CreateOutgoingDoc(c echo.Context) error {
    // ... existing code to create outgoing_doc ...
    
    outgoingDocID := createdDoc.ID
    docDetailsID := createdDoc.DocDetailsID
    userID := c.Get("user_id").(string)
    
    // Get recipient department IDs from metadata
    // These come from the upload form as: recipient_department_ids
    recipientDeptIDs := strings.Split(c.FormValue("recipient_department_ids"), ",")
    
    // Create incoming_docs for each recipient department
    for _, deptID := range recipientDeptIDs {
        if deptID == "" {
            continue
        }
        
        incomingNo := h.generateIncomingNo() // Generate unique incoming number
        
        incomingDoc := &IncomingDoc{
            IncomingNo:   incomingNo,
            DocDetailsID: docDetailsID,
            Status:       "pending",
            CreatedBy:    userID,
            DeptID:       deptID, // ← NEW: Set the department ID
            IncomingDate: time.Now(),
        }
        
        if err := h.service.CreateIncomingDoc(ctx, incomingDoc); err != nil {
            // Handle error - log but continue with other departments
            h.logger.Error("Failed to create incoming doc", "dept_id", deptID, "error", err)
        }
    }
    
    return c.JSON(http.StatusCreated, map[string]interface{}{
        "success": true,
        "data":    createdDoc,
    })
}
```

### 3. Update Get Outgoing Docs (List) Handler

**File: `internal/app/outgoingdoc/handler.go`** (GetAllHandler method)

Fetch outgoing docs with their recipients and status counts:

```go
func (h *Handler) GetAllOutgoingDocs(c echo.Context) error {
    page := c.QueryParam("page")
    limit := c.QueryParam("limit")
    deptID := c.QueryParam("department_id") // Filter by department
    startDate := c.QueryParam("start_date")
    endDate := c.QueryParam("end_date")
    docNo := c.QueryParam("doc_no")
    
    // ... existing pagination and filtering logic ...
    
    docs, err := h.service.GetAllOutgoingDocs(ctx, filters)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, ErrorResponse{...})
    }
    
    // Enrich documents with recipient info
    enrichedDocs := h.enrichOutgoingDocsWithRecipients(ctx, docs)
    
    return c.JSON(http.StatusOK, map[string]interface{}{
        "success": true,
        "data": map[string]interface{}{
            "items": enrichedDocs,
        },
        "pagination": pagination,
    })
}

func (h *Handler) enrichOutgoingDocsWithRecipients(ctx context.Context, docs []OutgoingDoc) []OutgoingDoc {
    for i := range docs {
        // Get all incoming docs for this outgoing doc
        incomingDocs, _ := h.service.GetIncomingDocsByDocDetailsID(ctx, docs[i].DocDetailsID)
        
        recipients := make([]RecipientDept, 0)
        statusCounts := StatusCounts{}
        
        for _, inc := range incomingDocs {
            // Get department info
            dept, _ := h.deptService.GetDepartmentByID(ctx, inc.DeptID)
            
            recipient := RecipientDept{
                DepartmentID:   inc.DeptID,
                DepartmentName: dept.Name,
                Status:         inc.Status,
            }
            
            // Count statuses
            switch strings.ToLower(inc.Status) {
            case "pending":
                statusCounts.Pending++
            case "received":
                statusCounts.Received++
            case "approved":
                statusCounts.Approved++
            case "rejected":
                statusCounts.Rejected++
            }
            
            recipients = append(recipients, recipient)
        }
        
        docs[i].Recipients = recipients
        docs[i].StatusCounts = statusCounts
    }
    
    return docs
}
```

### 4. Update Get Outgoing Doc by ID Handler

**File: `internal/app/outgoingdoc/handler.go`** (GetByIDHandler method)

Return full recipient details with incoming doc info:

```go
func (h *Handler) GetOutgoingDocByID(c echo.Context) error {
    id := c.Param("id")
    
    doc, err := h.service.GetOutgoingDocByID(ctx, id)
    if err != nil {
        return c.JSON(http.StatusNotFound, ErrorResponse{...})
    }
    
    // Get all incoming docs for this outgoing doc
    incomingDocs, _ := h.service.GetIncomingDocsByDocDetailsID(ctx, doc.DocDetailsID)
    
    // Enrich with recipient details
    recipients := make([]RecipientDept, 0)
    
    for _, inc := range incomingDocs {
        dept, _ := h.deptService.GetDepartmentByID(ctx, inc.DeptID)
        
        recipient := RecipientDept{
            DepartmentID:   inc.DeptID,
            DepartmentName: dept.Name,
            Status:         inc.Status,
            IncomingDoc: &IncomingDocDetail{
                ID:           inc.ID,
                IncomingNo:   inc.IncomingNo,
                Status:       inc.Status,
                ReceivedDate: inc.ReceivedDate,
            },
        }
        
        recipients = append(recipients, recipient)
    }
    
    doc.Recipients = recipients
    
    return c.JSON(http.StatusOK, map[string]interface{}{
        "success": true,
        "data":    doc,
    })
}
```

### 5. Update Repository Layer

**File: `internal/app/incomingdoc/repository.go`**

Add method to get incoming docs by doc_details_id:

```go
func (r *Repository) GetByDocDetailsID(ctx context.Context, docDetailsID string) ([]IncomingDoc, error) {
    query := `
        SELECT id, incoming_no, incoming_date, received_date, status, 
               doc_details_id, folder_id, created_by, updated_by, 
               approver_id, approver_date, remark, dept_id
        FROM incoming_docs
        WHERE doc_details_id = $1
        ORDER BY created_at DESC
    `
    
    rows, err := r.db.QueryContext(ctx, query, docDetailsID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var docs []IncomingDoc
    for rows.Next() {
        var doc IncomingDoc
        if err := rows.Scan(
            &doc.ID, &doc.IncomingNo, &doc.IncomingDate, &doc.ReceivedDate,
            &doc.Status, &doc.DocDetailsID, &doc.FolderID, &doc.CreatedBy,
            &doc.UpdatedBy, &doc.ApproverID, &doc.ApproverDate, &doc.Remark,
            &doc.DeptID, // ← NEW FIELD
        ); err != nil {
            return nil, err
        }
        docs = append(docs, doc)
    }
    
    return docs, rows.Err()
}

// Insert with dept_id
func (r *Repository) Create(ctx context.Context, doc *IncomingDoc) error {
    query := `
        INSERT INTO incoming_docs 
        (incoming_no, incoming_date, status, doc_details_id, 
         created_by, dept_id)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at
    `
    
    return r.db.QueryRowContext(ctx, query,
        doc.IncomingNo, doc.IncomingDate, doc.Status,
        doc.DocDetailsID, doc.CreatedBy, doc.DeptID,
    ).Scan(&doc.ID, &doc.CreatedAt)
}
```

---

## 📝 Summary of Changes

| Component | Change | Purpose |
|-----------|--------|---------|
| Database | Add `dept_id` to `incoming_docs` | Track which department received the doc |
| Models | Add `Recipients`, `StatusCounts` fields | Return enriched document data |
| Handler | Create incoming_docs for each recipient | Associate docs with departments |
| Service | Enrich outgoing docs with recipients | Populate recipient info on response |
| Repository | Query incoming docs by doc_details_id | Fetch department-specific data |

---

## 🚀 Testing

After implementing these changes:

1. **Run migration:**
   ```bash
   make migrate-up
   ```

2. **Create a document:**
   - Upload a file via `/api/v1/upload/files`
   - Include `target_module=outgoing` and `recipient_department_ids=dept-1,dept-2,dept-3`

3. **Verify incoming_docs created:**
   ```sql
   SELECT id, incoming_no, dept_id, status FROM incoming_docs 
   WHERE doc_details_id = (SELECT id FROM doc_details ORDER BY created_at DESC LIMIT 1);
   ```

4. **Test list endpoint:**
   ```bash
   GET /api/v1/outgoing-docs
   ```
   Should include `recipients` and `status_counts` in response

5. **Test detail endpoint:**
   ```bash
   GET /api/v1/outgoing-docs/{id}
   ```
   Should show all recipient departments with their statuses
