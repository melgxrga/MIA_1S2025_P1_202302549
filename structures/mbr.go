package structures

import (
    "bytes"           // Paquete para manipulación de buffers
    "encoding/binary" // Paquete para codificación y decodificación de datos binarios
    "fmt"             // Paquete para formateo de E/S
    "os"              // Paquete para funciones del sistema operativo
)

// MBR representa el Master Boot Record, ocupando 153 bytes.
type MBR struct {
    Mbr_size           int32        // Tamaño del MBR en bytes
    Mbr_creation_date  float32      // Fecha y hora de creación del MBR
    Mbr_disk_signature int32        // Firma del disco
    Mbr_disk_fit       [1]byte      // Tipo de ajuste
    Mbr_partitions     [4]PARTITION // Particiones del MBR (definidas en partition.go)
}

// SerializeMBR escribe la estructura MBR al inicio de un archivo binario.
func (mbr *MBR) SerializeMBR(path string) error {
    file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    // Serializar la estructura MBR directamente en el archivo.
    if err = binary.Write(file, binary.LittleEndian, mbr); err != nil {
        return err
    }

    return nil
}

// DeserializeMBR lee la estructura MBR desde el inicio de un archivo binario.
func (mbr *MBR) DeserializeMBR(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    mbrSize := binary.Size(mbr)
    if mbrSize <= 0 {
        return fmt.Errorf("invalid MBR size: %d", mbrSize)
    }

    buffer := make([]byte, mbrSize)
    _, err = file.Read(buffer)
    if err != nil {
        return err
    }

    reader := bytes.NewReader(buffer)
    if err = binary.Read(reader, binary.LittleEndian, mbr); err != nil {
        return err
    }

    return nil
}

// ReadMBR lee y deserializa el MBR desde el archivo ubicado en "path".
func ReadMBR(path string) (*MBR, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("error abriendo el archivo: %v", err)
    }
    defer file.Close()

    var mbr MBR
    if err = binary.Read(file, binary.LittleEndian, &mbr); err != nil {
        return nil, fmt.Errorf("error leyendo el MBR: %v", err)
    }

    return &mbr, nil
}