package structures

import (
    "encoding/binary"
    "os"
)

// EBR es el descriptor de una unidad lógica dentro de una partición extendida.
// Sólo las unidades lógicas (no las primarias) se gestionan a través del EBR.
// Cada unidad lógica contiene información sobre su montaje, el tipo de ajuste a aplicar, 
// el inicio y tamaño en bytes, y un puntero (offset) que indica el siguiente EBR en la lista enlazada.
// Esto es útil para estructurar múltiples unidades lógicas dentro de una partición extendida.
type EBR struct {
    Part_mount byte     // Indica si la unidad lógica está montada ('1') o no ('0')
    Part_fit   byte     // Tipo de ajuste: 'B' (Best), 'F' (First), o 'W' (Worst)
    Part_start int32    // Byte en el disco donde inicia la partición lógica
    Part_size  int32    // Tamaño total de la partición lógica en bytes
    Part_next  int32    // Byte donde se encuentra el siguiente EBR; -1 si no hay siguiente
    Part_name  [16]byte // Nombre de la partición lógica
}

// SerializeEBR escribe la estructura EBR en el archivo ubicado en 'path' en la posición 'pos'.
// Se utiliza para registrar o actualizar el descriptor de la unidad lógica en disco.
func (ebr *EBR) SerializeEBR(path string, pos int64) error {
    file, err := os.OpenFile(path, os.O_RDWR, 0666)
    if err != nil {
        return err
    }
    defer file.Close()

    _, err = file.Seek(pos, 0)
    if err != nil {
        return err
    }
    return binary.Write(file, binary.LittleEndian, ebr)
}