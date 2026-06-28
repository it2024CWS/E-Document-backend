user_roles {
  id          : UUID                                  // PK, default gen_random_uuid()
  role_name   : VARCHAR(100)                          // UQ
  description : TEXT          = ''
  created_at  : TIMESTAMPTZ   = NOW()
  updated_at  : TIMESTAMPTZ   = NOW()
  deleted_at  : TIMESTAMPTZ?                          // soft-delete (mig 000022)
}

departments {
  id          : UUID                                  // PK
  dept_name   : VARCHAR(150)                          // UQ
  description : TEXT          = ''
  created_at  : TIMESTAMPTZ   = NOW()
  updated_at  : TIMESTAMPTZ   = NOW()
  deleted_at  : TIMESTAMPTZ?                          // soft-delete (mig 000022)
}

sectors {
  id          : UUID                                  // PK
  name        : VARCHAR(150)
  dept_id     : UUID                                  // FK→departments.id (CASCADE), IDX
  created_at  : TIMESTAMPTZ   = NOW()
  updated_at  : TIMESTAMPTZ   = NOW()
  deleted_at  : TIMESTAMPTZ?                          // soft-delete (mig 000022)
  // UQ (name, dept_id)
}

users {
  id              : UUID                              // PK
  username        : VARCHAR(100)                      // UQ
  email           : VARCHAR(255)                      // UQ
  phone           : VARCHAR(30)   = ''
  firstname       : VARCHAR(100)  = ''
  lastname        : VARCHAR(100)  = ''
  nickname        : VARCHAR(100)  = ''
  password        : TEXT
  role_id         : UUID?                             // FK→user_roles.id  (SET NULL), IDX
  department_id   : UUID?                             // FK→departments.id (SET NULL), IDX
  sector_id       : UUID?                             // FK→sectors.id     (SET NULL), IDX
  is_active       : BOOLEAN       = TRUE
  profile_picture : TEXT          = ''
  created_at      : TIMESTAMPTZ   = NOW()
  updated_at      : TIMESTAMPTZ   = NOW()
  deleted_at      : TIMESTAMPTZ?                      // soft-delete (mig 000022)
}

doc_types {
  id          : UUID                                  // PK
  type_name   : VARCHAR(150)                          // UQ
  description : TEXT          = ''
  created_at  : TIMESTAMPTZ   = NOW()
  updated_at  : TIMESTAMPTZ   = NOW()
  deleted_at  : TIMESTAMPTZ?                          // soft-delete (mig 000022)
}

folders {
  id               : UUID                             // PK
  folder_name      : VARCHAR(255)
  folder_path      : TEXT
  user_id          : UUID                             // FK→users.id   (CASCADE),  IDX
  parent_folder_id : UUID?                            // FK→folders.id (SET NULL), IDX  (self-ref)
  created_at       : TIMESTAMPTZ = NOW()
  deleted_at       : TIMESTAMPTZ?                     // soft-delete (mig 000012)
}

doc_details {
  id             : UUID                               // PK
  doc_no         : VARCHAR(100)                       // UQ partial WHERE deleted_at IS NULL (mig 000018)
  doc_name       : VARCHAR(255)
  description    : TEXT?
  version_number : INT             = 1
  status         : document_status = 'none'           // IDX
  doc_type_id    : UUID?                              // FK→doc_types.id (SET NULL), IDX
  user_id        : UUID?                              // FK→users.id     (SET NULL), IDX
  created_at     : TIMESTAMPTZ     = NOW()
  updated_at     : TIMESTAMPTZ     = NOW()
  deleted_at     : TIMESTAMPTZ?                       // soft-delete
}

versions {
  id             : UUID                               // PK
  doc_details_id : UUID                               // FK→doc_details.id (CASCADE),  IDX
  folder_id      : UUID?                              // FK→folders.id     (SET NULL), IDX
  version_number : INT
  doc_path       : TEXT
  file_type      : VARCHAR(50)?                       // IDX (mig 000016)
  created_at     : TIMESTAMPTZ = NOW()
  deleted_at     : TIMESTAMPTZ?                       // soft-delete
  // UQ (doc_details_id, version_number)
}

outgoing_docs {
  id             : UUID                               // PK
  doc_details_id : UUID                               // FK→doc_details.id (CASCADE),  IDX
  folder_id      : UUID?                              // FK→folders.id     (SET NULL), IDX
  created_by     : UUID?                              // FK→users.id       (SET NULL), IDX
  updated_by     : UUID?                              // FK→users.id       (SET NULL), IDX
  status         : VARCHAR(20)  = 'pending'           // IDX  values: pending|approved|rejected  (mig 000019)
  owner_dept_id  : UUID?                              // FK→departments.id (SET NULL), IDX  (mig 000021)
  created_at     : TIMESTAMPTZ  = NOW()
  deleted_at     : TIMESTAMPTZ?                       // soft-delete (mig 000022)
  // outgoing_no column DROPPED in mig 000020 — use doc_details.doc_no
}

incoming_docs {
  id              : UUID                              // PK
  incoming_date   : TIMESTAMPTZ?
  received_date   : TIMESTAMPTZ?
  status          : incoming_doc_status = 'pending'   // IDX
  doc_details_id  : UUID                              // FK→doc_details.id   (CASCADE),  IDX
  folder_id       : UUID?                             // FK→folders.id       (SET NULL), IDX
  created_by      : UUID?                             // FK→users.id         (SET NULL), IDX
  updated_by      : UUID?                             // FK→users.id         (SET NULL), IDX
  approver_id     : UUID?                             // FK→users.id         (SET NULL), IDX
  approver_date   : TIMESTAMPTZ?
  remark          : TEXT          = ''
  dept_id         : UUID?                             // FK→departments.id   (SET NULL), IDX  (mig 000013)
  outgoing_doc_id : UUID?                             // FK→outgoing_docs.id (CASCADE),  IDX  (mig 000015)
  updated_at      : TIMESTAMPTZ   = NOW()
  deleted_at      : TIMESTAMPTZ?                      // soft-delete (mig 000022)
  // incoming_no column DROPPED in mig 000020 — use doc_details.doc_no
}

outgoing_doc_routes {
  id              : UUID                              // PK
  outgoing_doc_id : UUID                              // FK→outgoing_docs.id (CASCADE),  IDX
  dept_id         : UUID                              // FK→departments.id   (CASCADE),  IDX
  sequence_order  : INT                               //   1 = owner dept, then recipients in order
  incoming_doc_id : UUID?                             // FK→incoming_docs.id (SET NULL), IDX  (filled lazily per step)
  created_at      : TIMESTAMPTZ = NOW()
  // UQ (outgoing_doc_id, sequence_order)
}