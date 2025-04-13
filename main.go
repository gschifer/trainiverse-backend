package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rwcarlsen/goexif/exif"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Verificar se o método é POST
	if r.Method != "POST" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Parse do arquivo enviado
	err := r.ParseMultipartForm(10 << 20) // Limite de 10MB
	if err != nil {
		http.Error(w, "Erro ao processar o arquivo", http.StatusBadRequest)
		return
	}

	// Obter o arquivo enviado
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Erro ao obter o arquivo", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Decodificar os metadados EXIF
	x, err := exif.Decode(file)
	if err != nil {
		http.Error(w, "Erro ao ler os metadados EXIF", http.StatusInternalServerError)
		return
	}

	// Tentar obter as coordenadas GPS
	x.JpegThumbnail()
	lat, long, _ := x.LatLong()
	if lat != 0 && long != 0 {
		// Imprimir latitude e longitude se disponíveis
		fmt.Fprintf(w, "Latitude: %f, Longitude: %f\n", lat, long)
		// fmt.Println(x.JpegThumbnail())
		jsonData, err := x.MarshalJSON()
		if err != nil {
			http.Error(w, "Erro ao converter metadados EXIF para JSON", http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(jsonData))
		fmt.Println(x.MarshalJSON())
	} else {
		fmt.Fprintf(w, "Sem coordenadas GPS disponíveis.\n")
	}
}

func main() {
	http.HandleFunc("/upload", uploadHandler)

	// Iniciar servidor HTTP
	fmt.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
