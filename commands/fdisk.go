package commands

import (
	structures "CLASE03/structures"
	utils "CLASE03/utils"
	"bytes"
	"encoding/binary"
	"errors" // Paquete para manejar errores y crear nuevos errores con mensajes personalizados
	"fmt"    // Paquete para formatear cadenas y realizar operaciones de entrada/salida
	"os"
	"regexp"  // Paquete para trabajar con expresiones regulares, útil para encontrar y manipular patrones en cadenas
	"strconv" // Paquete para convertir cadenas a otros tipos de datos, como enteros
	"strings" // Paquete para manipular cadenas, como unir, dividir, y modificar contenido de cadenas
)

// FDISK estructura que representa el comando fdisk con sus parámetros
type FDISK struct {
	size int    // Tamaño de la partición
	unit string // Unidad de medida del tamaño (K o M)
	fit  string // Tipo de ajuste (BF, FF, WF)
	path string // Ruta del archivo del disco
	typ  string // Tipo de partición (P, E, L)
	name string // Nombre de la partición
}

/*
	fdisk -size=1 -type=L -unit=M -fit=BF -name="Particion3" -path="/home/keviin/University/PRACTICAS/MIA_LAB_S2_2024/CLASEEXTRA/disks/Disco1.mia"
	fdisk -size=300 -path=/home/Disco1.mia -name=Particion1
	fdisk -type=E -path=/home/Disco2.mia -Unit=K -name=Particion2 -size=300
*/

// CommandFdisk parsea el comando fdisk y devuelve una instancia de FDISK
func ParseFdisk(tokens []string) (*FDISK, error) {
	cmd := &FDISK{} // Crea una nueva instancia de FDISK

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando fdisk
	re := regexp.MustCompile(`-size=\d+|-unit=[kKmM]|-fit=[bBfF]{2}|-path="[^"]+"|-path=[^\s]+|-type=[pPeElL]|-name="[^"]+"|-name=[^\s]+`)
	// Encuentra todas las coincidencias de la expresión regular en la cadena de argumentos
	matches := re.FindAllString(args, -1)

	// Itera sobre cada coincidencia encontrada
	for _, match := range matches {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Remove quotes from value if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Switch para manejar diferentes parámetros
		switch key {
		case "-size":
			// Convierte el valor del tamaño a un entero
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return nil, errors.New("el tamaño debe ser un número entero positivo")
			}
			cmd.size = size
		case "-unit":
			// Verifica que la unidad sea "K" o "M"
			if value != "K" && value != "M" {
				return nil, errors.New("la unidad debe ser K o M")
			}
			cmd.unit = strings.ToUpper(value)
		case "-fit":
			// Verifica que el ajuste sea "BF", "FF" o "WF"
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return nil, errors.New("el ajuste debe ser BF, FF o WF")
			}
			cmd.fit = value
		case "-path":
			// Verifica que el path no esté vacío
			if value == "" {
				return nil, errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-type":
			// Verifica que el tipo sea "P", "E" o "L"
			value = strings.ToUpper(value)
			if value != "P" && value != "E" && value != "L" {
				return nil, errors.New("el tipo debe ser P, E o L")
			}
			cmd.typ = value
		case "-name":
			// Verifica que el nombre no esté vacío
			if value == "" {
				return nil, errors.New("el nombre no puede estar vacío")
			}
			cmd.name = value
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return nil, fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que los parámetros -size, -path y -name hayan sido proporcionados
	if cmd.size == 0 {
		return nil, errors.New("faltan parámetros requeridos: -size")
	}
	if cmd.path == "" {
		return nil, errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.name == "" {
		return nil, errors.New("faltan parámetros requeridos: -name")
	}

	// Si no se proporcionó la unidad, se establece por defecto a "M"
	if cmd.unit == "" {
		cmd.unit = "M"
	}

	// Si no se proporcionó el ajuste, se establece por defecto a "FF"
	if cmd.fit == "" {
		cmd.fit = "WF"
	}

	// Si no se proporcionó el tipo, se establece por defecto a "P"
	if cmd.typ == "" {
		cmd.typ = "P"
	}

	// Crear la partición con los parámetros proporcionados
	err := commandFdisk(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return cmd, nil // Devuelve el comando FDISK creado
}

func commandFdisk(fdisk *FDISK) error {
	// Convertir el tamaño a bytes
	sizeBytes, err := utils.ConvertToBytes(fdisk.size, fdisk.unit)
	if err != nil {
		fmt.Println("Error converting size:", err)
		return err
	}

	if fdisk.typ == "P" {
		// Crear partición primaria
		err = createPrimaryPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error creando partición primaria:", err)
			return err
		}
	} else if fdisk.typ == "E" {
		// Crear partición extendida
		out, err := createExtendedPartition(fdisk, sizeBytes)
		if err != nil {
			fmt.Println("Error creando partición extendida:", err)
			return err
		}
		fmt.Println(out)
	} else if fdisk.typ == "L" {
	      // Crear partición lógica
	  	    out, err := createLogicalPartition(fdisk, sizeBytes)
	      if err != nil {
	          fmt.Println("Error creando partición lógica:", err)
	          return err
	      }
	      fmt.Println(out)
	  }

	return nil
}

func createPrimaryPartition(fdisk *FDISK, sizeBytes int) error {
	// Crear una instancia de MBR
	var mbr structures.MBR

	// Deserializar la estructura MBR desde un archivo binario
	err := mbr.DeserializeMBR(fdisk.path)
	if err != nil {
		fmt.Println("Error deserializando el MBR:", err)
		return err
	}

	fmt.Println("\nMBR original:")
	mbr.PrintMBR()

	// Obtener la primera partición disponible
	availablePartition, startPartition, indexPartition := mbr.GetFirstAvailablePartition()
	if availablePartition == nil {
		fmt.Println("No hay particiones disponibles.")
	}

	/* SOLO PARA VERIFICACIÓN */
	// Print para verificar que la partición esté disponible
	fmt.Println("\nPartición disponible:")
	availablePartition.PrintPartition()

	// Crear la partición con los parámetros proporcionados
	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.typ, fdisk.fit, fdisk.name)

	// Print para verificar que la partición se haya creado correctamente
	fmt.Println("\nPartición creada (modificada):")
	availablePartition.PrintPartition()

	// Colocar la partición en el MBR
	if availablePartition != nil {
		mbr.Mbr_partitions[indexPartition] = *availablePartition
	}

	// Imprimir las particiones del MBR
	fmt.Println("\nParticiones del MBR:")
	mbr.PrintPartitions()

	// Serializar el MBR en el archivo binario
	err = mbr.SerializeMBR(fdisk.path)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return nil
}
func createExtendedPartition(fdisk *FDISK, sizeBytes int) (string, error) {
	var mbr structures.MBR

	// Deserializar el MBR desde disco
	err := mbr.DeserializeMBR(fdisk.path)
	if err != nil {
		return "", fmt.Errorf("Error deserializando el MBR: %w", err)
	}
	// Verificar que no exista ya una partición extendida
	for _, part := range mbr.Mbr_partitions {
		if part.Part_type[0] == 'E' || part.Part_type[0] == 'e' {
			return "", fmt.Errorf("Ya existe una partición extendida. No se puede crear otra")
		}
	}

	fmt.Println("\nMBR original:")
	fmt.Printf("MBR Size: %d\n", mbr.Mbr_size)
	fmt.Printf("Disk Signature: %d\n", mbr.Mbr_disk_signature)
	fmt.Printf("Disk Fit: %c\n", mbr.Mbr_disk_fit[0])

	// Buscar una partición libre para colocar la extendida
	availablePartition, startPartition, indexPartition := mbr.GetFirstAvailablePartition()
	if availablePartition == nil {
		return "", errors.New("No hay particiones disponibles para crear la extendida")
	}

	fmt.Println("\nPartición disponible:")
	fmt.Printf("Part_start: %d\n", availablePartition.Part_start)
	fmt.Printf("Part_size: %d\n", availablePartition.Part_size)
	fmt.Printf("Part_name: %s\n", string(availablePartition.Part_name[:]))
	fmt.Printf("Part_correlative: %d\n", availablePartition.Part_correlative)
	fmt.Printf("Part_id: %s\n", string(availablePartition.Part_id[:]))

	// Crear la partición extendida con los parámetros proporcionados
	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.typ, fdisk.fit, fdisk.name)

	fmt.Println("\nPartición extendida creada:")
	fmt.Printf("Part_status: %c\n", availablePartition.Part_status[0])
	fmt.Printf("Part_type: %c\n", availablePartition.Part_type[0])
	fmt.Printf("Part_fit: %c\n", availablePartition.Part_fit[0])
	fmt.Printf("Part_start: %d\n", availablePartition.Part_start)
	fmt.Printf("Part_size: %d\n", availablePartition.Part_size)
	fmt.Printf("Part_name: %s\n", string(availablePartition.Part_name[:]))
	fmt.Printf("Part_correlative: %d\n", availablePartition.Part_correlative)
	fmt.Printf("Part_id: %s\n", string(availablePartition.Part_id[:]))

	// Calcular la posición de inicio del EBR
	ebrStart := availablePartition.Part_start

	// Inicializar el EBR en la primera posición de la partición extendida
	ebr := structures.EBR{
		Part_status: '0',
		Part_fit:    availablePartition.Part_fit[0],
		Part_start:  ebrStart,
		Part_next:   -1,
		Part_size:   0,
	}
	copy(ebr.Part_name[:], availablePartition.Part_name[:])

	// Serializar el EBR en el archivo binario
	err = ebr.SerializeEBR(fdisk.path, int64(ebrStart))
	if err != nil {
		return "", fmt.Errorf("Error serializando el EBR: %w", err)
	}

	fmt.Println("\nEBR inicializado en la partición extendida:")
	fmt.Printf("Part_mount: %c\n", ebr.Part_status)
	fmt.Printf("Part_fit: %c\n", ebr.Part_fit)
	fmt.Printf("Part_start: %d\n", ebr.Part_start)
	fmt.Printf("Part_next: %d\n", ebr.Part_next)
	fmt.Printf("Part_size: %d\n", ebr.Part_size)
	fmt.Printf("Part_name: %s\n", string(ebr.Part_name[:]))

	// Colocar la partición en el MBR
	mbr.Mbr_partitions[indexPartition] = *availablePartition

	fmt.Println("\nParticiones del MBR:")
	for i, partition := range mbr.Mbr_partitions {
		fmt.Printf("Partition %d:\n", i+1)
		fmt.Printf("  Part_status: %c\n", partition.Part_status[0])
		fmt.Printf("  Part_type: %c\n", partition.Part_type[0])
		fmt.Printf("  Part_fit: %c\n", partition.Part_fit[0])
		fmt.Printf("  Part_start: %d\n", partition.Part_start)
		fmt.Printf("  Part_size: %d\n", partition.Part_size)
		fmt.Printf("  Part_name: %s\n", string(partition.Part_name[:]))
		fmt.Printf("  Part_correlative: %d\n", partition.Part_correlative)
		fmt.Printf("  Part_id: %s\n", string(partition.Part_id[:]))
	}

	// Serializar el MBR en el archivo binario
	err = mbr.SerializeMBR(fdisk.path)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return "", nil
}

// WriteEBR escribe la estructura EBR en el archivo ubicado en 'path' en la posición 'position'.
// Se utiliza para registrar o actualizar el descriptor de la unidad lógica en disco.
func WriteEBR(ebr *structures.EBR, path string, position int64) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("no se pudo abrir el archivo para escribir el EBR %s\n", err.Error())
		return
	}
	// Posicionandonos en la posición especificada del archivo
	_, err = file.Seek(position, 0)
	if err != nil {
		fmt.Printf("no se pudo posicionar en la posición especificada del archivo %s\n", err.Error())
		return
	}
	// Escribiendo el EBR
	err = binary.Write(file, binary.LittleEndian, ebr)
	if err != nil {
		fmt.Printf("no se pudo escribir el EBR %s\n", err.Error())
		file.Close()
		return
	}
	fmt.Println("EBR escrito correctamente")
	defer file.Close()
}

func ReadEBR(path string, position int64) *structures.EBR {
	var ebr *structures.EBR
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("no se pudo Abrir el archivo que contiene el EBR %s\n", err.Error())
		return ebr
	}

	defer file.Close()

	// leyendo el mbr del archivo
	file.Seek(position, 0)
	err = binary.Read(file, binary.LittleEndian, &ebr)
	if err != nil {
		fmt.Printf("no se pudo obtener la informacion del archivo para obtener el EBR %s\n", err.Error())
		return ebr
	}
	return ebr
}

func createLogicalPartition(fdisk *FDISK, sizeBytes int) (string, error) {
    var mbr structures.MBR

    // Deserializar el MBR desde disco
    err := mbr.DeserializeMBR(fdisk.path)
    if err != nil {
        return "", fmt.Errorf("Error deserializando el MBR: %w", err)
    }

    // Buscar la partición extendida
    var extendedFound bool = false
    var whereToStart int
    var extendedFit byte
    var extendedName [16]byte
    for _, part := range mbr.Mbr_partitions {
        if part.Part_type[0] == 'E' || part.Part_type[0] == 'e' {
            extendedFound = true
            whereToStart = int(part.Part_start)
            extendedFit = part.Part_fit[0]
            copy(extendedName[:], part.Part_name[:])
            break
        }
    }
    if !extendedFound {
        return "", errors.New("No existe una partición extendida para crear una partición lógica")
    }

    fmt.Println("\nMBR original:")
    fmt.Printf("%+v\n", mbr)

    // Aquí se simula la creación de una partición lógica.
    // Se asume que existe una función que crea la partición lógica en el área extendida.
    // Por ejemplo: CreateLogicPartition(logicPartition, path, whereToStart, extendedFit, extendedName)
    // En este ejemplo, creamos un registro lógico de forma simplificada.

    // Supongamos que construimos un EBR (Extended Boot Record) de la partición lógica:
    logicPartition := structures.EBR{
        Part_status: '1',
        Part_fit:    extendedFit,
        Part_next:   -1,
        Part_size:   int32(sizeBytes),
    }
    copy(logicPartition.Part_name[:], []byte(fdisk.name))

    // Serializar el EBR en el archivo binario
    WriteEBR(&logicPartition, fdisk.path, int64(whereToStart))

    fmt.Println("\nPartición lógica creada:")
    fmt.Printf("Part_fit: %c\n", logicPartition.Part_fit)
    fmt.Printf("Part_next: %d\n", logicPartition.Part_next)
    fmt.Printf("Part_size: %d\n", logicPartition.Part_size)
    fmt.Printf("Part_status: %c\n", logicPartition.Part_status)
    fmt.Printf("Part_name: %s\n", string(logicPartition.Part_name[:]))

    // Se actualiza el disco (MBR) si fuese necesario.
    err = mbr.SerializeMBR(fdisk.path)
    if err != nil {
        fmt.Println("Error:", err)
    }

    return "", nil
}

func (fdisk  *FDISK) CreateLogicPartition(logicPartition  *structures.EBR, path string, whereToStart int, partitionSize int, extendedFit byte, extendedName [16]byte) bool {
	if extendedFit == 'f' {
		return FirstFitLogicPart(logicPartition, path, whereToStart, partitionSize, extendedName)
	} else if extendedFit == 'b' {
		return BestFitLogicPart(logicPartition, path, whereToStart, partitionSize, extendedName)
	} else if extendedFit == 'w' {
		return WorstFitLogicPart(logicPartition, path, whereToStart, partitionSize, extendedName)
	}
	return false
}

func FirstFitLogicPart(logicPartition *structures.EBR, path string, whereToStart int, partitionSize int, extendedName [16]byte) bool {
    var temp structures.EBR
    totalSize := 0
    totalSize += int(logicPartition.Part_size)
    temp = *ReadEBR(path, int64(whereToStart))
    flag := true
    for flag {
        if temp.Part_size == 0 {
            if partitionSize < int(logicPartition.Part_size) {
                fmt.Println("la particion logica es mas grande que la extendida")
                return false
            }
            logicPartition.Part_start = int32(whereToStart)
            WriteEBR(logicPartition, path, int64(whereToStart))
            flag = false
        } else if temp.Part_status == '5' {
            if temp.Part_size >= logicPartition.Part_size {
                logicPartition.Part_start = temp.Part_start
                logicPartition.Part_next = temp.Part_next
                WriteEBR(logicPartition, path, int64(temp.Part_start))
                flag = false
            }
        } else if temp.Part_next == -1 {
            totalSize += int(temp.Part_size)
            if partitionSize < totalSize {
                fmt.Println("el tamano de todas las particiones logicas unidas son mas grandes que la particion extendida, espacio insuficiente")
                return false
            }
            temp.Part_next = temp.Part_start + temp.Part_size
            logicPartition.Part_start = temp.Part_next
            WriteEBR(&temp, path, int64(temp.Part_start))
            WriteEBR(logicPartition, path, int64(temp.Part_next))
            flag = false
        } else {
            totalSize += int(temp.Part_size)
            temp = *ReadEBR(path, int64(temp.Part_next))
        }
    }
    // aquí debería ir un print a la consola
    PrintLogicPartitions(path, int64(whereToStart), int64(partitionSize), extendedName)
    return true
}

func BestFitLogicPart(logicPartition *structures.EBR, path string, whereToStart int, partitionSize int, extendedName [16]byte) bool {
    var particionesLogicas []structures.EBR
    var temp structures.EBR
    totalSize := 0
    totalSize += int(logicPartition.Part_size)
    temp = *ReadEBR(path, int64(whereToStart))
    Wrote := false
    flag := true
    for flag {
        if temp.Part_size == 0 {
            if partitionSize < int(logicPartition.Part_size) {
                fmt.Println("la particion logica es mas grande que la extendida")
                return false
            }
            logicPartition.Part_start = int32(whereToStart)
            WriteEBR(logicPartition, path, int64(whereToStart))
            flag = false
            Wrote = true
        } else if temp.Part_status == '5' {
            particionesLogicas = append(particionesLogicas, temp)
        } else if temp.Part_next == -1 {
            flag = false
        } else {
            totalSize += int(temp.Part_size)
            temp = *ReadEBR(path, int64(temp.Part_next))
        }
    }
    bestFit := 0
    tempSize := 0
    if len(particionesLogicas) != 0 {
        for i, v := range particionesLogicas {
            if tempSize == 0 || (tempSize > int(v.Part_size) && v.Part_size >= logicPartition.Part_size) {
                tempSize = int(v.Part_size)
                bestFit = i
            }
        }
        logicPartition.Part_start = particionesLogicas[bestFit].Part_start
        logicPartition.Part_next = particionesLogicas[bestFit].Part_next
        WriteEBR(logicPartition, path, int64(logicPartition.Part_start))
        Wrote = true
    }
    if !Wrote {
        totalSize = int(logicPartition.Part_size)
        temp = *ReadEBR(path, int64(whereToStart))
        flag2 := true
        for flag2 {
            if temp.Part_next == -1 {
                totalSize += int(temp.Part_size)
                if partitionSize < totalSize {
                    fmt.Println("el tamano de todas las particiones logicas unidas son mas grandes que la particion extendida, espacio insuficiente")
                    return false
                }
                temp.Part_next = temp.Part_start + temp.Part_size
                logicPartition.Part_start = temp.Part_next
                WriteEBR(&temp, path, int64(temp.Part_start))
                WriteEBR(logicPartition, path, int64(temp.Part_next))
                flag2 = false
            } else {
                totalSize += int(temp.Part_size)
                temp = *ReadEBR(path, int64(temp.Part_next))
            }
        }
    }
    // aquí debería ir un print a la consola
    PrintLogicPartitions(path, int64(whereToStart), int64(partitionSize), extendedName)
    return true
}

func WorstFitLogicPart(logicPartition *structures.EBR, path string, whereToStart int, partitionSize int, extendedName [16]byte) bool {
    var particionesLogicas []structures.EBR
    var temp structures.EBR
    totalSize := 0
    totalSize += int(logicPartition.Part_size)
    temp = *ReadEBR(path, int64(whereToStart))
    Wrote := false
    flag := true
    for flag {
        if temp.Part_size == 0 {
            if partitionSize < int(logicPartition.Part_size) {
                fmt.Println("la particion logica es mas grande que la extendida")
                return false
            }
            logicPartition.Part_start = int32(whereToStart)
            WriteEBR(logicPartition, path, int64(whereToStart))
            flag = false
            Wrote = true
        } else if temp.Part_status == '5' {
            particionesLogicas = append(particionesLogicas, temp)
        } else if temp.Part_next == -1 {
            flag = false
        } else {
            totalSize += int(temp.Part_size)
            temp = *ReadEBR(path, int64(temp.Part_next))
        }
    }
    worstFit := 0
    tempSize := 0
    if len(particionesLogicas) != 0 {
        for i, v := range particionesLogicas {
            if tempSize == 0 || (tempSize < int(v.Part_size) && v.Part_size >= logicPartition.Part_size) {
                tempSize = int(v.Part_size)
                worstFit = i
            }
        }
        logicPartition.Part_start = particionesLogicas[worstFit].Part_start
        logicPartition.Part_next = particionesLogicas[worstFit].Part_next
        WriteEBR(logicPartition, path, int64(logicPartition.Part_start))
        Wrote = true
    }
    if !Wrote {
        totalSize = int(logicPartition.Part_size)
        temp = *ReadEBR(path, int64(whereToStart))
        flag2 := true
        for flag2 {
            if temp.Part_next == -1 {
                totalSize += int(temp.Part_size)
                if partitionSize < totalSize {
                    fmt.Println("el tamano de todas las particiones logicas unidas son mas grandes que la particion extendida, espacio insuficiente")
                    return false
                }
                temp.Part_next = temp.Part_start + temp.Part_size
                logicPartition.Part_start = temp.Part_next
                WriteEBR(&temp, path, int64(temp.Part_start))
                WriteEBR(logicPartition, path, int64(temp.Part_next))
                flag2 = false
            } else {
                totalSize += int(temp.Part_size)
                temp = *ReadEBR(path, int64(temp.Part_next))
            }
        }
    }
    // aquí debería ir un print a la consola
    PrintLogicPartitions(path, int64(whereToStart), int64(partitionSize), extendedName)
    return true
}


func ExisteParticion(master *structures.MBR, name [16]byte) bool {
	for _, v := range master.Mbr_partitions {
		if bytes.Equal(v.Part_name[:], name[:]) {
			return true
		}
	}
	return false
}


// TrimArray elimina los ceros a la derecha de un array de bytes.
func TrimArray(arr []byte) []byte {
    n := len(arr)
    for n > 0 && arr[n-1] == 0 {
        n--
    }
    return arr[:n]
}

// PrintLogicPartitions imprime las particiones lógicas.
func PrintLogicPartitions(path string, whereToStart, PartitionSize int64, extendedName [16]byte) {
    str := ""
    for i := 0; i < 70; i++ {
        str += "-"
    }
    contenido := ""
    contenido += fmt.Sprintf("Partition name: %s\n", string(TrimArray(extendedName[:])))
    contenido += fmt.Sprintf("Partition size: %d\n", PartitionSize)
    contenido += fmt.Sprintf("%s\n", str)
    contenido += fmt.Sprintf("%-20s%-12s%-10s%-10s%-10s%-10s\n", "Name", "Next Part", "Fit", "Start", "Size", "Status")
    var temp structures.EBR
    Fread(&temp, path, whereToStart)
    flag := true
    for flag {
        contenido += fmt.Sprintf("%s\n", str)
        if string(TrimArray(temp.Part_name[:])) == "" {
            contenido += fmt.Sprintf("%-20s", "Disponible")
        } else {
            contenido += fmt.Sprintf("%-20s", string(TrimArray(temp.Part_name[:])))
        }
        contenido += fmt.Sprintf("%-12d", temp.Part_next)
        if temp.Part_fit == '0' {
            contenido += fmt.Sprintf("%-10c", '-')
        } else {
            contenido += fmt.Sprintf("%-10c", temp.Part_fit)
        }
        contenido += fmt.Sprintf("%-10d", temp.Part_start)
        contenido += fmt.Sprintf("%-10d", temp.Part_size)
        contenido += fmt.Sprintf("%-10c\n", temp.Part_status)
        if temp.Part_next == -1 {
            flag = false
        } else {
            Fread(&temp, path, int64(temp.Part_next))
        }
    }
    contenido += fmt.Sprintf("%s\n", str)
    fmt.Println(contenido)
}

// Fread lee la estructura EBR desde el archivo ubicado en 'path' en la posición 'position'.
func Fread(ebr *structures.EBR, path string, position int64) {
    file, err := os.Open(path)
    if err != nil {
        fmt.Printf("no se pudo abrir el archivo para leer el EBR %s\n", err.Error())
        return
    }
    defer file.Close()

    // Posicionandonos en la posición especificada del archivo
    _, err = file.Seek(position, 0)
    if err != nil {
        fmt.Printf("no se pudo posicionar en la posición especificada del archivo %s\n", err.Error())
        return
    }

    // Leyendo el EBR
    err = binary.Read(file, binary.LittleEndian, ebr)
    if err != nil {
        fmt.Printf("no se pudo leer el EBR %s\n", err.Error())
        return
    }
}