package commands

import (
    "errors"
    "fmt"
    "os"
    "regexp"
    "strings"
)

// DeleteDisk estructura que representa el comando rmdisk con su parámetro
type DeleteDisk struct {
    path string // Ruta del archivo (disco) a eliminar
}

/*
   rmdisk -path=/ruta/al/disco.mia
*/

// ParseDeleteDisk parsea los tokens y extrae el parámetro -path
func ParseDeleteDisk(tokens []string) (*DeleteDisk, error) {
    dd := &DeleteDisk{}
    // Unir tokens en una sola cadena
    args := strings.Join(tokens, " ")
    // Expresión regular para obtener el parámetro -path
    re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+`)
    matches := re.FindAllString(args, -1)
    
    for _, match := range matches {
        kv := strings.SplitN(match, "=", 2)
        if len(kv) != 2 {
            return nil, fmt.Errorf("formato de parámetro inválido: %s", match)
        }
        key, value := strings.ToLower(kv[0]), kv[1]
        // Eliminar comillas si existen
        if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
            value = strings.Trim(value, "\"")
        }
        if key != "-path" {
            return nil, fmt.Errorf("parámetro desconocido: %s", key)
        }
        dd.path = value
    }
    // Validar que se suministre el parámetro obligatorio -path
    if dd.path == "" {
        return nil, errors.New("faltan parámetros requeridos: -path")
    }
    return dd, nil
}

// ExecuteDeleteDisk elimina el archivo que representa el disco
func ExecuteDeleteDisk(dd *DeleteDisk) error {
    // Verificar si el archivo existe
    if _, err := os.Stat(dd.path); os.IsNotExist(err) {
        return fmt.Errorf("error: el archivo %s no existe", dd.path)
    }

    // Eliminar el archivo
    err := os.Remove(dd.path)
    if err != nil {
        return fmt.Errorf("error al eliminar el archivo: %v", err)
    }
    return nil
}

// RemoveDisk es la función llamada por analyzer para el comando rmdisk
// Se actualiza la firma para devolver (interface{}, error)
func RemoveDisk(tokens []string) (interface{}, error) {
    dd, err := ParseDeleteDisk(tokens)
    if err != nil {
        return nil, err
    }
    err = ExecuteDeleteDisk(dd)
    if err != nil {
        return nil, err
    }
    successMessage := fmt.Sprintf("Archivo eliminado exitosamente: %s", dd.path)
    fmt.Println("Archivo eliminado exitosamente: ", dd.path)
    return successMessage, nil
}