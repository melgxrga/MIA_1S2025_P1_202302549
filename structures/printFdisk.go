package structures

import (
    "fmt"
    "os"
    "strings"
)

// PrintDiskInfo lee el disco ubicado en "path" y muestra información detallada,
// incluyendo el tamaño total, el MBR (153 bytes) y las particiones almacenadas.
func PrintDiskInfo(path string) error {
    // Obtener el tamaño total del disco
    fi, err := os.Stat(path)
    if err != nil {
        return fmt.Errorf("error obteniendo información del disco: %v", err)
    }
    totalSize := fi.Size()

    // Leer el MBR desde el disco
    mbr, err := ReadMBR(path)
    if err != nil {
        return fmt.Errorf("error leyendo el MBR: %v", err)
    }

    // Imprimir la información del disco
    fmt.Println("------ Información del Disco ------")
    fmt.Printf("Tamaño total del disco: %d bytes\n", totalSize)
    fmt.Println("------ MBR (153 bytes) ------")
    fmt.Printf("Tamaño del disco (según MBR): %d bytes\n", mbr.Mbr_size)
    fmt.Printf("Fecha de creación: %.0f\n", mbr.Mbr_creation_date)
    fmt.Printf("Firma del disco: %d\n", mbr.Mbr_disk_signature)
    fmt.Printf("Tipo de ajuste del MBR: %c\n", mbr.Mbr_disk_fit[0])
    fmt.Println("------ Particiones en el MBR ------")
    for i, part := range mbr.Mbr_partitions {
        // Se asume que si Part_size es -1 la partición está libre
        if part.Part_size != -1 {
            // Convertir el nombre (relleno de 16 bytes) a cadena limpia
            rawName := string(part.Part_name[:])
            name := strings.TrimRight(rawName, "\x00")
            fmt.Printf("Partición %d:\n", i+1)
            fmt.Printf("  Estado: %c\n", part.Part_status[0])
            fmt.Printf("  Tipo: %c\n", part.Part_type[0])
    
            ajuste := strings.Trim(string(part.Part_fit[:]), "\x00")
            if ajuste == "" || ajuste == "W" || ajuste == "F" || ajuste == "B" {
                ajuste = "WF"
            }
            fmt.Printf("  Ajuste: %s\n", ajuste)
            fmt.Printf("  Inicio: %d\n", part.Part_start)
            fmt.Printf("  Tamaño: %d bytes\n", part.Part_size)
            fmt.Printf("  Nombre: %s\n", name)
            fmt.Printf("  Correlativo: %d\n", part.Part_correlative)
            fmt.Printf("  ID: %s\n", strings.TrimRight(string(part.Part_id[:]), "\x00"))
        } else {
            fmt.Printf("Partición %d: Libre\n", i+1)
        }
    }
    return nil
}