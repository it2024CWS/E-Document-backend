package seed

import (
	"context"
	"e-document-backend/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedMasterData seeds all reference / master data tables.
// Each statement uses ON CONFLICT DO NOTHING so it is safe to run multiple times.
func SeedMasterData(ctx context.Context, pool *pgxpool.Pool) error {
	if err := seedRoles(ctx, pool); err != nil {
		return err
	}
	if err := seedDepartments(ctx, pool); err != nil {
		return err
	}
	if err := seedSectors(ctx, pool); err != nil {
		return err
	}
	if err := seedDocTypes(ctx, pool); err != nil {
		return err
	}
	return nil
}

// ─── Roles ────────────────────────────────────────────────────────────────────

func seedRoles(ctx context.Context, pool *pgxpool.Pool) error {
	roles := []struct {
		name        string
		description string
	}{
		{"Director", "Head of the organization with full authority"},
		{"Manager", "Department manager with approval authority"},
		{"Secretary", "Handles document reception and routing"},
		{"Staff", "Regular employee with basic document access"},
		{"Admin", "System administrator with full system access"},
	}

	for _, r := range roles {
		_, err := pool.Exec(ctx, `
			INSERT INTO user_roles (role_name, description)
			VALUES ($1, $2)
			ON CONFLICT (role_name) DO NOTHING
		`, r.name, r.description)
		if err != nil {
			return err
		}
	}

	logger.Info("✓ Roles seeded")
	return nil
}

// ─── Departments ──────────────────────────────────────────────────────────────

func seedDepartments(ctx context.Context, pool *pgxpool.Pool) error {
	departments := []struct {
		name        string
		description string
	}{
		{"General Administration", "Handles general administrative tasks and document management"},
		{"Human Resources", "Manages employee affairs and personnel documents"},
		{"Finance & Accounting", "Handles financial documents and budget management"},
		{"Information Technology", "Manages IT infrastructure and digital systems"},
		{"Legal & Compliance", "Handles legal documents and regulatory compliance"},
		{"Operations", "Manages day-to-day operational activities"},
		{"Procurement", "Handles purchasing and vendor documents"},
	}

	for _, d := range departments {
		_, err := pool.Exec(ctx, `
			INSERT INTO departments (dept_name, description)
			VALUES ($1, $2)
			ON CONFLICT (dept_name) DO NOTHING
		`, d.name, d.description)
		if err != nil {
			return err
		}
	}

	logger.Info("✓ Departments seeded")
	return nil
}

// ─── Sectors ──────────────────────────────────────────────────────────────────

func seedSectors(ctx context.Context, pool *pgxpool.Pool) error {
	// map: sector name → parent department name
	sectors := []struct {
		name     string
		deptName string
	}{
		// General Administration
		{"Document Control", "General Administration"},
		{"Records Management", "General Administration"},
		// Human Resources
		{"Recruitment", "Human Resources"},
		{"Payroll", "Human Resources"},
		{"Training & Development", "Human Resources"},
		// Finance & Accounting
		{"Budgeting", "Finance & Accounting"},
		{"Accounts Payable", "Finance & Accounting"},
		{"Accounts Receivable", "Finance & Accounting"},
		// Information Technology
		{"Software Development", "Information Technology"},
		{"Network & Infrastructure", "Information Technology"},
		{"IT Support", "Information Technology"},
		// Legal & Compliance
		{"Contract Management", "Legal & Compliance"},
		{"Regulatory Affairs", "Legal & Compliance"},
		// Operations
		{"Logistics", "Operations"},
		{"Facility Management", "Operations"},
		// Procurement
		{"Vendor Management", "Procurement"},
		{"Asset Management", "Procurement"},
	}

	for _, s := range sectors {
		var deptID *string
		err := pool.QueryRow(ctx,
			`SELECT id::text FROM departments WHERE dept_name = $1`, s.deptName,
		).Scan(&deptID)
		if err != nil || deptID == nil {
			logger.Infof("⚠ Department '%s' not found, skipping sector '%s'", s.deptName, s.name)
			continue
		}

		_, err = pool.Exec(ctx, `
			INSERT INTO sectors (name, dept_id)
			VALUES ($1, $2::uuid)
			ON CONFLICT (name, dept_id) DO NOTHING
		`, s.name, *deptID)
		if err != nil {
			return err
		}
	}

	logger.Info("✓ Sectors seeded")
	return nil
}

// ─── Document Types ───────────────────────────────────────────────────────────

func seedDocTypes(ctx context.Context, pool *pgxpool.Pool) error {
	docTypes := []struct {
		name        string
		description string
	}{
		{"Memo", "Internal memorandum for communicating within the organization"},
		{"Letter", "Formal correspondence sent to external parties"},
		{"Report", "Detailed document summarizing findings or progress"},
		{"Circular", "Official notice distributed to multiple recipients"},
		{"Minutes", "Record of proceedings and decisions from a meeting"},
		{"Proposal", "Document proposing a project or initiative for approval"},
		{"Contract", "Legal agreement between two or more parties"},
		{"Invoice", "Commercial document requesting payment for goods or services"},
		{"Certificate", "Official document attesting to a fact or achievement"},
		{"Form", "Structured document for collecting standardized information"},
	}

	for _, dt := range docTypes {
		_, err := pool.Exec(ctx, `
			INSERT INTO doc_types (type_name, description)
			VALUES ($1, $2)
			ON CONFLICT (type_name) DO NOTHING
		`, dt.name, dt.description)
		if err != nil {
			return err
		}
	}

	logger.Info("✓ Document types seeded")
	return nil
}
