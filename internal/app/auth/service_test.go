package auth

import (
	"strings"
	"testing"
	"time"

	"e-document-backend/internal/config"
	"e-document-backend/internal/domain"

	"github.com/google/uuid"
)

// mockUser returns a realistic user matching the "ter" account
func mockUser(roleName string) *domain.User {
	id := uuid.MustParse("61eac1f7-52f2-4090-8836-91af39e4710b")
	roleID := uuid.MustParse("9a3f483c-cf32-4043-b18a-bdb29919be23")
	deptID := uuid.MustParse("f473c89f-8d89-449f-93ed-196f8f57b533")
	return &domain.User{
		ID:           id,
		Username:     "ter",
		Email:        "ter@gmail.com",
		RoleID:       &roleID,
		RoleName:     roleName,
		DepartmentID: &deptID,
	}
}

func newTestService() *service {
	cfg := &config.Config{}
	cfg.JWT.AccessTokenSecret = "test-secret"
	cfg.JWT.AccessTokenExpiry = 3600
	cfg.JWT.RefreshTokenSecret = "test-refresh-secret"
	cfg.JWT.RefreshTokenExpiry = 3600
	return &service{cfg: cfg}
}

// TestBuildUserClaims_RoleNameInToken verifies that role_name is embedded in the JWT claims
func TestBuildUserClaims_RoleNameInToken(t *testing.T) {
	svc := newTestService()
	user := mockUser("Secretary")

	claims := svc.buildUserClaims(user, "access", 3600)

	roleName, ok := claims["role_name"].(string)
	if !ok {
		t.Fatal("role_name claim is missing or not a string")
	}
	if roleName != "Secretary" {
		t.Errorf("expected role_name='Secretary', got '%s'", roleName)
	}
}

// TestBuildUserClaims_EmptyRoleWhenMissing verifies the old bug: empty string when RoleName not set
func TestBuildUserClaims_EmptyRoleWhenMissing(t *testing.T) {
	svc := newTestService()
	user := mockUser("") // simulate old FindByUsername (no JOIN)

	claims := svc.buildUserClaims(user, "access", 3600)

	roleName, _ := claims["role_name"].(string)
	if roleName != "" {
		t.Errorf("expected empty role_name from user without role, got '%s'", roleName)
	}
}

// TestGenerateAndParseToken_PreservesRoleName does a round-trip: sign token → parse → check role_name
func TestGenerateAndParseToken_PreservesRoleName(t *testing.T) {
	svc := newTestService()
	user := mockUser("Secretary")

	tokenStr, err := svc.generateAccessToken(user)
	if err != nil {
		t.Fatalf("generateAccessToken failed: %v", err)
	}

	claims, err := svc.ValidateAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("ValidateAccessToken failed: %v", err)
	}

	if claims.RoleName != "Secretary" {
		t.Errorf("expected RoleName='Secretary' after round-trip, got '%s'", claims.RoleName)
	}
}

// TestRequireRole_SecretaryAllowed simulates the middleware check
func TestRequireRole_SecretaryAllowed(t *testing.T) {
	allowed := []string{"Secretary"}
	userRole := "Secretary"

	permitted := false
	for _, r := range allowed {
		if strings.EqualFold(userRole, r) {
			permitted = true
			break
		}
	}

	if !permitted {
		t.Error("Secretary should be permitted but was denied")
	}
}

// TestRequireRole_WrongRoleDenied verifies non-Secretary roles are blocked
func TestRequireRole_WrongRoleDenied(t *testing.T) {
	allowed := []string{"Secretary"}
	userRole := "" // what the old bug produced

	permitted := false
	for _, r := range allowed {
		if strings.EqualFold(userRole, r) {
			permitted = true
			break
		}
	}

	if permitted {
		t.Error("empty role_name should be denied but was permitted")
	}
}

// TestTokenClaims_DeptID verifies department_id survives the round-trip too
func TestTokenClaims_DeptID(t *testing.T) {
	svc := newTestService()
	user := mockUser("Secretary")

	tokenStr, _ := svc.generateAccessToken(user)
	claims, err := svc.ValidateAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("ValidateAccessToken failed: %v", err)
	}

	if claims.DepartmentID == nil {
		t.Fatal("DepartmentID should not be nil after round-trip")
	}
	if claims.DepartmentID.String() != user.DepartmentID.String() {
		t.Errorf("DepartmentID mismatch: expected %s, got %s",
			user.DepartmentID, claims.DepartmentID)
	}
}

// TestTokenExpiry verifies exp claim is set in the future
func TestTokenExpiry(t *testing.T) {
	svc := newTestService()
	user := mockUser("Secretary")

	claims := svc.buildUserClaims(user, "access", 3600)

	exp, ok := claims["exp"].(int64)
	if !ok {
		t.Fatal("exp claim missing or wrong type")
	}
	if exp <= time.Now().Unix() {
		t.Error("token should expire in the future")
	}
}
