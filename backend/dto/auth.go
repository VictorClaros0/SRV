package dto

// LoginRequest credenciales de acceso.
type LoginRequest struct {
	Usuario    string `json:"usuario" binding:"required"`
	Contrasena string `json:"contrasena" binding:"required"`
}

// LoginResponse token JWT.
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expiresIn"`
	EsAdmin   bool   `json:"esAdmin"`
}

// RegisterRequest alta de usuario (solo admin).
type RegisterRequest struct {
	Nombre          string `json:"nombre" binding:"required"`
	Apellido        string `json:"apellido" binding:"required"`
	SegundoApellido string `json:"segundoApellido"`
	Usuario         string `json:"usuario" binding:"required"`
	Contrasena      string `json:"contrasena" binding:"required,min=6"`
	EsAdmin         bool   `json:"esAdmin"`
}
