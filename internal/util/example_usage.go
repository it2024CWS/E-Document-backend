package util

// Example usage of error codes in handlers
//
// ===== RECOMMENDED: Use ErrorResponse in Service Layer =====
//
// In Service Layer (return CustomError):
// func (s *service) GetUserByID(ctx context.Context, id string) (*UserResponse, error) {
//     user, err := s.repo.FindByID(ctx, id)
//     if err != nil {
//         return nil, util.ErrorResponse(
//             "User not found",
//             util.USER_NOT_FOUND,
//             404,
//             fmt.Sprintf("user with id %s not found", id),
//         )
//     }
//     return user, nil
// }
//
// In Handler Layer (use HandleError with service error):
// func (h *Handler) GetUserByID(c echo.Context) error {
//     id := c.Param("id")
//     user, err := h.service.GetUserByID(c.Request().Context(), id)
//     if err != nil {
//         return util.HandleError(c, err)  // Automatically uses CustomError info
//     }
//     return util.OKResponse(c, "User retrieved successfully", user)
// }
//
// ===== Use ErrorResponse directly in Handler =====
//
// EXAMPLE 1: Configuration error
// func SomeHandler(c echo.Context) error {
//     if jwtSecret == "" {
//         return util.HandleError(c, util.ErrorResponse(
//             "configuration error",
//             util.CONFIG_NOT_SET,
//             500,
//             "JWT_REFRESH_SECRET_EXPIRES is not configured",
//         ))
//     }
// }
//
// EXAMPLE 2: Authentication error
// func Login(c echo.Context) error {
//     if !authenticated {
//         return util.HandleError(c, util.ErrorResponse(
//             "unauthorized",
//             util.INVALID_CREDENTIALS,
//             401,
//             "Incorrect username or password",
//         ))
//     }
// }
//
// EXAMPLE 3: Validation error
// func CreateUser(c echo.Context) error {
//     if req.Email == "" {
//         return util.HandleError(c, util.ErrorResponse(
//             "Validation failed",
//             util.MISSING_REQUIRED_FIELD,
//             400,
//             "Email is required",
//         ))
//     }
// }
//
// EXAMPLE 4: Token expired
// func ProtectedRoute(c echo.Context) error {
//     if tokenExpired {
//         return util.HandleError(c, util.ErrorResponse(
//             "Token has expired",
//             util.TOKEN_EXPIRED,
//             401,
//             "Please login again to get a new token",
//         ))
//     }
// }
//
// ===== Success Responses =====
//
// EXAMPLE 1: Return OK response (200)
// func GetUser(c echo.Context) error {
//     user, err := db.FindUser(id)
//     if err != nil {
//         return util.HandleError(c, util.ErrorResponse("User not found", util.DATABASE_ERROR, 404, err.Error()))
//     }
//     return util.OKResponse(c, "User retrieved successfully", user)
// }
//
// EXAMPLE 2: Return Created response (201)
// func CreateUser(c echo.Context) error {
//     user, err := db.CreateUser(req)
//     if err != nil {
//         return util.HandleError(c, util.ErrorResponse("Failed to create user", util.DATABASE_ERROR, 500, err.Error()))
//     }
//     return util.CreatedResponse(c, "User created successfully", user)
// }
//
// Response format:
// {
//     "success": true,
//     "message": "User created successfully",
//     "data": { ... }
// }
