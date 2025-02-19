package commands

import (
    "CLASE03/structures"
    "CLASE03/utils"
    "errors"
    "fmt"
    "os"
    "regexp"
    "strconv"
    "strings"
)


type FDISK struct {
    size    int    // Tamaño de la partición a crear
    unit    string // Unidad (B, K o M) – opcional, default K
    path    string // Ruta del disco (debe existir)
    partitionType string // Tipo de partición (P, E o L) – default P
    fit     string // Tipo de ajuste (BF, FF o WF) – default WF
    name    string // Nombre de la partición (no se puede repetir)
}

// ParseMkpart procesa los tokens del comando partición
// Ejemplo de comando:
// mkpart -size=500 -unit=K -path="/ruta/del/disco.mia" -type=P -fit=WF -name=Part1
func ParseMkpart(tokens []string) (*FDISK, error) {
    cmd := &FDISK{}

    // Unir tokens en una sola cadena
    args := strings.Join(tokens, " ")
    // Expresión regular para extraer parámetros
    re := regexp.MustCompile(`-size=\d+|-unit=[bBkKmM]|-path="[^"]+"|-path=[^\s]+|-type=[pPeElL]|-fit=[bBfFwW]{2}|-name="[^"]+"|-name=[^\s]+`)
    matches := re.FindAllString(args, -1)

    for _, match := range matches {
        kv := strings.SplitN(match, "=", 2)
        if len(kv) != 2 {
            return nil, fmt.Errorf("formato de parámetro inválido: %s", match)
        }
        key, value := strings.ToLower(kv[0]), kv[1]
        if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
            value = strings.Trim(value, "\"")
        }
        switch key {
        case "-size":
            size, err := strconv.Atoi(value)
            if err != nil || size <= 0 {
                return nil, errors.New("el tamaño de la partición debe ser un número entero positivo mayor a 0")
            }
            cmd.size = size
        case "-unit":
            // Solo se aceptan B, K y M (default K)
            upperUnit := strings.ToUpper(value)
            if upperUnit != "B" && upperUnit != "K" && upperUnit != "M" {
                return nil, errors.New("la unidad debe ser B, K o M")
            }
            cmd.unit = upperUnit
        case "-path":
            if value == "" {
                return nil, errors.New("el path no puede estar vacío")
            }
            cmd.path = value
        case "-type":
            upperType := strings.ToUpper(value)
            if upperType != "P" && upperType != "E" && upperType != "L" {
                return nil, errors.New("el tipo debe ser P, E o L")
            }
            cmd.partitionType = upperType
        case "-fit":
            upperFit := strings.ToUpper(value)
            if upperFit != "BF" && upperFit != "FF" && upperFit != "WF" {
                return nil, errors.New("el ajuste debe ser BF, FF o WF")
            }
            cmd.fit = upperFit
        case "-name":
            if value == "" {
                return nil, errors.New("el nombre no puede estar vacío")
            }
            cmd.name = value
        default:
            return nil, fmt.Errorf("parámetro desconocido: %s", key)
        }
    }

    // Validación de parámetros obligatorios
    if cmd.size == 0 {
        return nil, errors.New("faltan parámetros requeridos: -size")
    }
    if cmd.path == "" {
        return nil, errors.New("faltan parámetros requeridos: -path")
    }
    if cmd.name == "" {
        return nil, errors.New("faltan parámetros requeridos: -name")
    }
    // Valores por defecto
    if cmd.unit == "" {
        cmd.unit = "K"
    }
    if cmd.partitionType == "" {
        cmd.partitionType = "P"
    }
    if cmd.fit == "" {
        cmd.fit = "WF"
    }

    // Ejecutar creación de partición
    err := commandMkpart(cmd)
    if err != nil {
        fmt.Println("Error:", err)
        return nil, err
    }

    return cmd, nil
}

func commandMkpart(mkpart *FDISK) error {
    // Verificar que el disco exista
    if _, err := os.Stat(mkpart.path); os.IsNotExist(err) {
        return errors.New("el disco no existe en la ruta especificada")
    }

    // Convertir tamaño de partición a bytes
    sizeBytes, err := utils.ConvertToBytes(mkpart.size, mkpart.unit)
    if err != nil {
        return err
    }

    // Abrir archivo del disco
    file, err := os.OpenFile(mkpart.path, os.O_RDWR, 0666)
    if err != nil {
        return fmt.Errorf("error abriendo el disco: %v", err)
    }
    defer file.Close()

    // Leer MBR
    mbr, err := structures.ReadMBR(mkpart.path)
    if err != nil {
        return fmt.Errorf("error leyendo MBR: %v", err)
    }

    // Cumplir restricciones:
    // ● Máximo de 4 particiones primarias/extendidas para comandos tipo P o E.
    // ● Una única extendida.
    // ● Las lógicas (L) solo pueden crearse dentro de la extendida.
    primaryCount := 0
    extendedIndex := -1
    for i, part := range mbr.Mbr_partitions {
        // Suponemos que si Part_size es -1 la partición está libre (no creada)
        if part.Part_size != -1 {
            upperType := string(part.Part_type[:])
            if upperType == "E" {
                extendedIndex = i
            }
            primaryCount++
            // Verificar nombres repetidos
            currName := strings.TrimRight(string(part.Part_name[:]), "\x00")
            if currName == mkpart.name {
                return errors.New("ya existe una partición con el mismo nombre")
            }
        }
    }

    // Si se desea crear una partición primaria o extendida
    if mkpart.partitionType == "P" || mkpart.partitionType == "E" {
        if primaryCount >= 4 {
            return errors.New("ya se ha alcanzado el límite de 4 particiones primarias/extendidas")
        }
        if mkpart.partitionType == "E" && extendedIndex != -1 {
            return errors.New("ya existe una partición extendida en este disco")
        }
        // Buscar espacio libre en las 4 entradas de MBR
        slot := -1
        for i, part := range mbr.Mbr_partitions {
            if part.Part_size == -1 {
                slot = i
                break
            }
        }
        if slot == -1 {
            return errors.New("no hay espacio para crear una nueva partición")
        }

        // Calcular el inicio: si es la primera, justo después del MBR (153 bytes)
        var start int32 = 153
        for _, part := range mbr.Mbr_partitions {
            if part.Part_size != -1 && part.Part_start >= start {
                end := part.Part_start + part.Part_size
                if end > start {
                    start = end
                }
            }
        }

        // Preparar la partición a insertar de 35 bytes de desglose:
        // status(1), type(1), fit(1), start(4), size(4), name(16), correlative(4), id(4)
        var newPart structures.PARTITION
        newPart.Part_status = [1]byte{'1'} // Indicador de activa
        newPart.Part_type = [1]byte{mkpart.partitionType[0]}
        newPart.Part_fit = [1]byte{mkpart.fit[0]}
        newPart.Part_start = start
        newPart.Part_size = int32(sizeBytes)
        // Rellenar el nombre en un array de 16 bytes
        var nameArr [16]byte
        copy(nameArr[:], mkpart.name)
        newPart.Part_name = nameArr
        // Asignar número incremental para correlative e id basado en primaryCount + 1
        correlative := int32(primaryCount + 1)
        newPart.Part_correlative = correlative
        idStr := fmt.Sprintf("%04d", correlative)
        var idArr [4]byte
        copy(idArr[:], idStr)
        newPart.Part_id = idArr

        // Insertar la nueva partición en el slot identificado
        mbr.Mbr_partitions[slot] = newPart

        // Escribir el MBR actualizado en el disco
        err = mbr.SerializeMBR(mkpart.path)
        if err != nil {
            return fmt.Errorf("error al escribir el MBR actualizado: %v", err)
        }

        fmt.Println("Partición creada exitosamente.")
        // Imprimir la información detallada del disco inmediatamente
        if err := structures.PrintDiskInfo(mkpart.path); err != nil {
            fmt.Println("Error al mostrar información del disco:", err)
        }

        return nil
    }

    return errors.New("tipo de partición desconocido")
}