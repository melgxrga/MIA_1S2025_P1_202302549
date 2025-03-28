package comandos

import (
	"bytes"
	"fmt"
	"strings"
	"regexp"  // Paquete para trabajar con expresiones regulares, útil para encontrar y manipular patrones en cadenas
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/structures"
	"github.com/melgxrga/proyecto1Archivos/functions"
	"github.com/melgxrga/proyecto1Archivos/list"
)

type ParametrosMount struct {
	Path string
	Name [16]byte
}

type Mount struct {
	Params ParametrosMount
}

func (m *Mount) Exe(parametros []string) {
	m.Params = m.SaveParams(parametros)
	if m.Mount(m.Params.Path, m.Params.Name) {
		consola.AddToConsole(fmt.Sprintf("\nparticion %s montada con exito\n\n", m.Params.Path))
	} else {
		consola.AddToConsole(fmt.Sprintf("no se logro montar la particion %s\n", m.Params.Path))
	}
}

func (m *Mount) SaveParams(parametros []string) ParametrosMount {
	var params ParametrosMount
	// Unir todos los parámetros en una sola cadena
	args := strings.Join(parametros, " ")

	// Expresión regular para capturar los parámetros
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-name="[^"]+"|-name=[^\s]+`)
	matches := re.FindAllString(args, -1)

	// Iterar sobre cada coincidencia
	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			fmt.Printf("Formato de parámetro inválido: %s\n", match)
			continue
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		// Quitar comillas si las tiene
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		// Procesar según el parámetro encontrado
		switch key {
		case "-path":
			if value == "" {
				fmt.Println("Error: el path no puede estar vacío")
				continue
			}
			params.Path = value
		case "-name":
			if value == "" {
				fmt.Println("Error: el nombre no puede estar vacío")
				continue
			}
			copy(params.Name[:], value)
		default:
			fmt.Printf("Parámetro desconocido: %s\n", key)
		}
	}

	// Validación final de los parámetros obligatorios
	if params.Path == "" {
		fmt.Println("Error: Falta el parámetro obligatorio -path")
	}
	if params.Name == [16]byte{} {
		fmt.Println("Error: Falta el parámetro obligatorio -name")
	}

	return params
}


func (m *Mount) Mount(path string, name [16]byte) bool {
	// comprobando que el parametro "path" sea diferente de ""
	if path == "" {
		consola.AddToConsole("no se encontro una ruta\n")
		return false
	}
	// comprobando que el parametro "name" sea diferente de ""
	if bytes.Equal(name[:], []byte("")) {
		consola.AddToConsole("se debe de contar con un nombre para realizar este comando\n")
		return false
	}
	master := GetMBR(path)
	partitionMounted := false
	particionEncontrada := false
	for _, particion := range master.Mbr_partitions {
		// si entro aqui es porque si leyo el MBR del disco
		if bytes.Equal(particion.Part_name[:], name[:]) {
			// comprobaremos que la particion no se haya montado previamente
			particionEncontrada = true
			if particion.Part_status == '2' {
				consola.AddToConsole("la particion ya se encuentra montada\n")
				return false
			}
			if particion.Part_type == 'e' || particion.Part_type == 'E' {
				consola.AddToConsole("no se puede montar una particion extendida\n")
				return false
			}
			particion.Part_type = '2'
			var part *datos.Partition = new(datos.Partition)
			part = &particion
			lista.ListaMount.Mount(path, 49, part, nil)
			partitionMounted = true
			// tener un metodo de MountList que agregue un texto a la consola
			lista.ListaMount.PrintList()
			break

		}
	}
	if !particionEncontrada {
		// buscaremos si existe una particion logica con ese nombre
		for _, particion := range master.Mbr_partitions {
			if particion.Part_type == 'e' || particion.Part_type == 'E' {
				partitionMounted = true
				m.MountParticionLogica(path, int(particion.Part_start), name)
				// tener un metodo de Mount List que agregue un texto a la consola
				lista.ListaMount.PrintList()
			}
		}
	}
	if !partitionMounted {
		consola.AddToConsole(fmt.Sprintf("no se encontro una particion con el nombre de %s\n", string(functions.TrimArray(name[:]))))
		return false
	}
	WriteMBR(&master, path)
	return true
}

func (m *Mount) MountParticionLogica(path string, whereToStart int, name [16]byte) {
	logicPartitionMounted := false
	temp := ReadEBR(path, int64(whereToStart))
	flag := true
	for flag {
		if bytes.Equal(temp.Part_name[:], name[:]) {
			temp.Part_status = '2'
			var partL *datos.EBR = new(datos.EBR)
			partL = &temp
			lista.ListaMount.Mount(path, 49, nil, partL)
			logicPartitionMounted = true
			flag = false
			break
		} else if temp.Part_next != -1 {
			temp = ReadEBR(path, temp.Part_next)
		} else {
			flag = false
		}
	}
	if !logicPartitionMounted {
		consola.AddToConsole(fmt.Sprintf("no se encontro una particion con el nombre %s\n", string(functions.TrimArray(name[:]))))
		return
	}
}
