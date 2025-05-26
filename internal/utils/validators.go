package utils

import (
	"regexp"
	"strconv"
	"strings"
)

// ValidateClientKey valida que la clave del cliente sea un número válido
func ValidateClientKey(clave string) (bool, string) {
	clave = strings.TrimSpace(clave)

	if clave == "" {
		return false, "La clave no puede estar vacía"
	}

	// Verificar que sea un número
	if _, err := strconv.Atoi(clave); err != nil {
		return false, "La clave debe ser un número válido"
	}

	return true, ""
}

// ValidateClientName valida que el nombre solo contenga letras y espacios
func ValidateClientName(nombre string) (bool, string) {
	nombre = strings.TrimSpace(nombre)

	if nombre == "" {
		return false, "El nombre no puede estar vacío"
	}

	// Verificar que solo contenga letras, espacios y algunos caracteres especiales permitidos
	match, _ := regexp.MatchString(`^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s\.'-]+$`, nombre)
	if !match {
		return false, "El nombre solo puede contener letras, espacios y caracteres especiales básicos"
	}

	// Verificar que no contenga números
	if regexp.MustCompile(`\d`).MatchString(nombre) {
		return false, "El nombre no puede contener números"
	}

	return true, ""
}

// ValidateEmail valida que el correo tenga un formato válido y dominio permitido
func ValidateEmail(correo string) (bool, string) {
	correo = strings.TrimSpace(strings.ToLower(correo))

	if correo == "" {
		return false, "El correo no puede estar vacío"
	}

	// Dominios permitidos
	allowedDomains := []string{
		"@gmail.com",
		"@hotmail.com",
		"@outlook.com",
		"@yahoo.com",
		"@live.com",
		"@icloud.com",
		"@msn.com",
	}

	// Verificar formato básico de email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(correo) {
		return false, "El formato del correo electrónico no es válido"
	}

	// Verificar dominio permitido
	validDomain := false
	for _, domain := range allowedDomains {
		if strings.HasSuffix(correo, domain) {
			validDomain = true
			break
		}
	}

	if !validDomain {
		return false, "El dominio del correo no está permitido. Use: gmail.com, hotmail.com, outlook.com, yahoo.com, live.com, icloud.com, msn.com"
	}

	return true, ""
}

// ValidatePhone valida que el teléfono tenga una lada permitida
func ValidatePhone(telefono string) (bool, string) {
	telefono = strings.TrimSpace(telefono)

	if telefono == "" {
		return false, "El teléfono no puede estar vacío"
	}

	// Limpiar el teléfono (quitar espacios, guiones, paréntesis)
	cleanPhone := regexp.MustCompile(`[^\d]`).ReplaceAllString(telefono, "")

	// Verificar que solo contenga números
	if !regexp.MustCompile(`^\d+$`).MatchString(cleanPhone) {
		return false, "El teléfono solo puede contener números"
	}

	// Verificar longitud (debe tener al menos 10 dígitos)
	if len(cleanPhone) < 10 {
		return false, "El teléfono debe tener al menos 10 dígitos"
	}

	// Ladas permitidas de Chiapas
	allowedAreaCodes := []string{
		"916", "917", "918", "919", "932", "934",
		"961", "962", "963", "964", "965", "966",
		"967", "968", "992", "994",
	}

	// Verificar lada (primeros 3 dígitos)
	if len(cleanPhone) >= 3 {
		areaCode := cleanPhone[:3]
		validAreaCode := false
		for _, code := range allowedAreaCodes {
			if areaCode == code {
				validAreaCode = true
				break
			}
		}

		if !validAreaCode {
			return false, "La lada del teléfono no es válida para Chiapas. Ladas permitidas: " + strings.Join(allowedAreaCodes, ", ")
		}
	}

	return true, ""
}

// CleanString limpia cadenas eliminando espacios extra
func CleanString(s string) string {
	// Eliminar espacios al inicio y final
	s = strings.TrimSpace(s)

	// Reemplazar múltiples espacios por uno solo
	re := regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, " ")

	return s
}

// IsEmpty verifica si una cadena está vacía después de limpiarla
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}
