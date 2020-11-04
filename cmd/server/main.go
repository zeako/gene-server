package main

import (
	"fmt"
	"os"

	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/zeako/gene-server/pkg/genefinder"
)

// Health serves as healthcheck endpoint.
func Health(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, http.StatusText(http.StatusOK))
}

// FindGene searches for gene sequence in service
//
// Returns the following:
//	* 200 - found sequence
//	* 404 - sequence wasn't found
//	* 400 - input validation errors
//	* 500 - internal server error
func FindGene(gf *genefinder.GeneFinder) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		gene := ps.ByName("gene")
		found, err := gf.Find(gene)

		if err != nil {
			switch err.(type) {
			case *genefinder.ValidationError:
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, err.Error())
			default:
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, http.StatusText(http.StatusInternalServerError))
			}
			return
		}

		if !found {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, http.StatusText(http.StatusNotFound))
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, http.StatusText(http.StatusOK))
	}
}

func main() {
	path := os.Getenv("DNA_FILE_PATH")
	if path == "" {
		log.Fatal("missing DNA_FILE_PATH environment variable")
	}
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	gf, err := genefinder.New(f)
	if err != nil {
		log.Fatal(err)
	}

	r := httprouter.New()
	{
		r.GET("/health", Health)
		r.GET("/genes/find/:gene", FindGene(gf))
	}

	log.Println("Starting gene-server")
	log.Fatal(http.ListenAndServe(":8080", r))
}
