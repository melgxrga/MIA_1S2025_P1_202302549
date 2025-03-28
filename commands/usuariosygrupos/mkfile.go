package usuariosygrupos

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/melgxrga/proyecto1Archivos/bitmap"
	"github.com/melgxrga/proyecto1Archivos/commands"
	"github.com/melgxrga/proyecto1Archivos/consola"
	"github.com/melgxrga/proyecto1Archivos/structures"
	"github.com/melgxrga/proyecto1Archivos/list"
	"github.com/melgxrga/proyecto1Archivos/logger"
)

type ParametrosMkfile struct {
	Path string
	R    bool
	Size int
	Cont string
}

type Mkfile struct {
	Params ParametrosMkfile
}

func (m *Mkfile) Exe(parametros []string) {
	m.Params = m.SaveParams(parametros)
	if m.Mkfile(m.Params.Path, m.Params.R, m.Params.Size, m.Params.Cont) {
		consola.AddToConsole(fmt.Sprintf("\nel archivo con ruta %s se creo correctamente\n\n", m.Params.Path))
	} else {
		consola.AddToConsole(fmt.Sprintf("\nel archivo con ruta %s no se pudo crear\n\n", m.Params.Path))
	}
}

func (m *Mkfile) SaveParams(parametros []string) ParametrosMkfile {
	for _, v := range parametros {
		// fmt.Println(v)
		v = strings.TrimSpace(v)
		v = strings.TrimRight(v, " ")
		v = strings.ReplaceAll(v, "\"", "")
		if strings.Contains(v, "path") {
			v = strings.ReplaceAll(v, "path=", "")
			m.Params.Path = v
		} else if v == "r" {
			// v = strings.ReplaceAll(v, "r", "")
			m.Params.R = true
		} else if strings.Contains(v, "cont") {
			v = strings.ReplaceAll(v, "cont=", "")
			m.Params.Cont = v
		} else if strings.Contains(v, "size") {
			v = strings.ReplaceAll(v, "size=", "")
			v = strings.ReplaceAll(v, " ", "")
			num, err := strconv.Atoi(v)
			if err != nil {
				fmt.Println("hubo un error al convertir a int", err.Error())
			}
			m.Params.Size = num
		}
	}
	return m.Params
}

func (m *Mkfile) Mkfile(path string, r bool, size int, cont string) bool {
	if path == "" {
		consola.AddToConsole("no se encontro una ruta\n")
		return false
	}
	path = strings.Replace(path, "/", "", 1)
	if !logger.Log.IsLoggedIn() {
		consola.AddToConsole("no se encuentra un usuario loggeado para crear un archivo\n")
		return false
	}
	if size < 0 {
		consola.AddToConsole(("el size no puede ser negativo"))
		return false
	}

	if lista.ListaMount.GetNodeById(logger.Log.GetUserId()).Value != nil {
		return createFile(logger.Log.GetUserName(), lista.ListaMount.GetNodeById(logger.Log.GetUserId()).Ruta, path, lista.ListaMount.GetNodeById(logger.Log.GetUserId()).Value.Part_start, r, size, cont)
	} else if lista.ListaMount.GetNodeById(logger.Log.GetUserId()).ValueL != nil {
		return createFile(logger.Log.GetUserName(), lista.ListaMount.GetNodeById(logger.Log.GetUserId()).Ruta, path, lista.ListaMount.GetNodeById(logger.Log.GetUserId()).ValueL.Part_start+int64(unsafe.Sizeof(datos.EBR{})), r, size, cont)
	}
	return false
}

func createFile(name [10]byte, path, ruta string, whereToStart int64, r bool, size int, cont string) bool {
	// superbloque de la particion
	var superbloque datos.SuperBloque
	comandos.Fread(&superbloque, path, whereToStart)
	var tablaInodoRoot datos.TablaInodo
	comandos.Fread(&tablaInodoRoot, path, superbloque.S_inode_start)

	// obteniendo el contenido de la tabla de inodos de Users.txt
	var tablaInodoUsers datos.TablaInodo
	comandos.Fread(&tablaInodoUsers, path, superbloque.S_inode_start+superbloque.S_inode_size)
	contenido := ReadFile(&tablaInodoUsers, path, &superbloque)
	// obteniendo el user id y el group id
	userId := GetUserId(contenido, string(TrimArray(name[:])))
	groupdId := GetGroupId(contenido, string(TrimArray(name[:])))
	if r {
		// Create directories
		FindAndCreateDirectories(&tablaInodoRoot, path, ruta, &superbloque, 0, userId, groupdId)
	}
	content := ""
	if cont != "" {
		// Retrieve information
		// fmt.Println("Retrieve information")
		content = getContent(cont)
		// fmt.Println(content)
	}
	if size != 0 && size > 0 {
		// Create content
		if StrlenBytes([]byte(content)) != 0 && StrlenBytes([]byte(content)) < size {
			// fmt.Println("entra aqui a ver lo del content")
			contador := 0
			for i := StrlenBytes([]byte(content)); i < size; i++ {
				if contador != 9 {
					contador++
				} else {
					contador = 0
				}
				content += strconv.Itoa(contador)

			}
		} else {
			contador := 0
			for i := 0; i < size; i++ {
				if contador != 9 {
					contador++
				} else {
					contador = 0
				}
				content += strconv.Itoa(contador)
			}
		}
		// fmt.Println("Create content")
	}
	// fmt.Println("el content despues de agregarle numeros")
	consola.AddToConsole("-------CONTENIDO DEL ARCHIVO-------\n")
	consola.AddToConsole("\"" + content + "\"\n")
	num := NewInodeFile(&superbloque, path, userId, groupdId, content)
	FindDirectories(num, &tablaInodoRoot, path, ruta, &superbloque, 0)
	comandos.Fwrite(&tablaInodoRoot, path, superbloque.S_inode_start)
	comandos.Fwrite(&superbloque, path, whereToStart)
	fmt.Println(" Depuración: Superbloque actualizado correctamente en disco.")
	PrintTree(&tablaInodoRoot, &superbloque, path)
	return true
}

func GetGroupId(contenido, name string) int64 {
	lineas := strings.Split(contenido, "\n")
	lineas = lineas[:len(lineas)-1]
	groupName := ""
	for _, linea := range lineas {
		linea = strings.ReplaceAll(linea, "\x00", "")
		parametros := strings.Split(linea, ",")
		if parametros[1] != "U" {
			continue
		}
		if parametros[3] == name {
			groupName = parametros[2]
		}
	}
	if groupName == "" {
		return -1
	}
	for _, linea := range lineas {
		linea = strings.ReplaceAll(linea, "\x00", "")
		parametros := strings.Split(linea, ",")
		if parametros[1] != "G" {
			continue
		}
		if parametros[2] == groupName {
			num, _ := strconv.Atoi(parametros[0])
			return int64(num)
		}
	}
	return -1
}

func GetUserId(contenido, name string) int64 {
	lineas := strings.Split(contenido, "\n")
	lineas = lineas[:len(lineas)-1]
	for _, linea := range lineas {
		linea = strings.ReplaceAll(linea, "\x00", "")
		parametros := strings.Split(linea, ",")
		if parametros[1] != "U" {
			continue
		}
		if parametros[3] == name {
			num, _ := strconv.Atoi(parametros[0])
			return int64(num)
		}
	}
	return -1
}

func NewInodeFile(superbloque *datos.SuperBloque, path string, userId, groupId int64, contenido string) int64 {
	var nuevaTabla datos.TablaInodo
	nuevaPosicion := bitmap.WriteInBitmapInode(path, superbloque)

	if nuevaPosicion == -1 {
		fmt.Println("❌ Error: No se pudo asignar un nuevo inodo.")
		return -1
	}

	fmt.Printf("Depuración: Creando archivo con inodo en posición %d\n", nuevaPosicion)

	// Aquí llenaremos la nueva tabla de inodos
	nuevaTabla.I_uid = userId
	nuevaTabla.I_gid = groupId
	nuevaTabla.I_size = int64(StrlenBytes([]byte(contenido)))
	nuevaTabla.I_type = '1'
	nuevaTabla.I_perm = 664

	// llenando las fechas
	atime := time.Now()
	copy(nuevaTabla.I_atime[:], atime.String())
	ctime := time.Now()
	copy(nuevaTabla.I_ctime[:], ctime.String())
	mtime := time.Now()
	copy(nuevaTabla.I_mtime[:], mtime.String())

	// llenando a todos los bloques no utilizados
	for i := 0; i < len(nuevaTabla.I_block); i++ {
		nuevaTabla.I_block[i] = -1
	}

	// Escribir contenido en bloques
	llenarTablaDeInodoDeArchivos(&nuevaTabla, superbloque, path, contenido)

	// Escribiendo la tabla de inodos en disco
	fmt.Printf(" Depuración: Escribiendo inodo en superbloque en posición %d\n", superbloque.S_inode_start+nuevaPosicion*superbloque.S_inode_size)
	comandos.Fwrite(&nuevaTabla, path, superbloque.S_inode_start+nuevaPosicion*superbloque.S_inode_size)

	return nuevaPosicion
}


func getContent(cont string) string {
	// aqui hay que leer el archivo y ejecutarlo
	file, err := os.Open(cont)
	if err != nil {
		consola.AddToConsole(fmt.Sprintf("Error al intentar abrir el archivo: %s\n", cont))
		return ""
	}

	defer file.Close()

	// Crear un scanner para luego leer linea por linea el archivo de entrada
	scanner := bufio.NewScanner(file)
	content := ""
	// Leyendo linea por linea
	for scanner.Scan() {
		// obteniendo la linea actual
		content += scanner.Text() + "\n"
	}

	// comprobar que no hubo error al leer el archivo
	if err := scanner.Err(); err != nil {
		consola.AddToConsole(fmt.Sprintf("Error al leer el archivo: %s\n", err))
		return ""
	}
	return content
}

func llenarTablaDeInodoDeArchivos(tablaInodo *datos.TablaInodo, superbloque *datos.SuperBloque, path, contenido string) {
	for i := 0; i < len(tablaInodo.I_block); i++ {
		if tablaInodo.I_block[i] == -1 {
			var bloqueArchivo datos.BloqueDeArchivos
			posicionBloqueDeArchivo := bitmap.WriteInBitmapBlock(path, superbloque)

			if posicionBloqueDeArchivo == -1 {
				fmt.Println("Error: No se pudo asignar un nuevo bloque de datos.")
				return
			}

			fmt.Printf(" Depuración: Bloque de archivo asignado en posición %d\n", posicionBloqueDeArchivo)

			tablaInodo.I_block[i] = posicionBloqueDeArchivo
			if StrlenBytes([]byte(contenido)) > 63 {
				copy(bloqueArchivo.B_content[:], []byte(contenido[:63]))
				comandos.Fwrite(&bloqueArchivo, path, superbloque.S_block_start+posicionBloqueDeArchivo*superbloque.S_block_size)
				llenarTablaDeInodoDeArchivos(tablaInodo, superbloque, path, contenido[63:])
			} else {
				copy(bloqueArchivo.B_content[:], []byte(contenido[:]))
				comandos.Fwrite(&bloqueArchivo, path, superbloque.S_block_start+posicionBloqueDeArchivo*superbloque.S_block_size)
			}
			return
		}
	}
}


