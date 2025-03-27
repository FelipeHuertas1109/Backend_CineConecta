package controllers

import (
	"cine_conecta_backend/auth/factories"
	"cine_conecta_backend/auth/models"
	"cine_conecta_backend/auth/services"
	"cine_conecta_backend/auth/utils"
	"cine_conecta_backend/config"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Registrar usuario
func Register(c *gin.Context) {
	var input models.User

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos")
		return
	}

	// Encriptar la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error al encriptar la contraseña")
		return
	}

	// Asignar rol por defecto o como admin
	role := "user"
	if input.Email == "fhuertas@unillanos.edu.co" {
		role = "admin"
	}

	// Crear el usuario con el factory
	user := factories.NewUser(input.Name, input.Email, string(hashedPassword), role)

	// Guardar el usuario con el servicio
	if err := services.SaveUser(user); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Correo ya utilizado en otra cuenta")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuario registrado correctamente"})
}

// Iniciar sesión
func Login(c *gin.Context) {
	var input models.User
	var user models.User

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Datos inválidos")
		return
	}

	// Buscar usuario
	result := config.DB.Where("email = ?", input.Email).First(&user)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Email o contraseña incorrectos")
		return
	}

	// Verificar contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Email o contraseña incorrectos")
		return
	}

	// Generar token
	token, err := utils.GenerateJWT(user.Name, user.Role)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "No se pudo generar el token")
		return
	}

	utils.SetTokenCookie(c, token)

	c.JSON(http.StatusOK, gin.H{
		"message": "Sesión iniciada correctamente",
	})
}

func Logout(c *gin.Context) {
	// Detectar si estás en producción
	isProduction := os.Getenv("ENV") == "production"

	// Expirar la cookie 'cine_token'
	c.SetCookie(
		"cine_token",
		"", // valor vacío
		-1, // duración negativa = eliminar
		"/",
		"",
		isProduction, // secure
		true,         // httpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Sesión cerrada correctamente",
	})
}

// Obtener todos los usuarios (solo admin)
func GetAllUsers(c *gin.Context) {
	var users []models.User

	result := config.DB.Find(&users)
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "No se pudieron obtener los usuarios")
		return
	}

	// Ocultar contraseñas
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, users)
}

func DeleteAllUsers(c *gin.Context) {
	result := config.DB.Exec("DELETE FROM users WHERE role != ?", "admin")
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "No se pudieron eliminar los usuarios",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todos los usuarios no admin han sido eliminados"})
}

func VerifyToken(c *gin.Context) {
	// La cookie ya está siendo verificada por el middleware de autenticación
	// Si llegamos aquí, significa que el token es válido
	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
	})
}
