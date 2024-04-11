package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

//go:embed app/dist
var embeddedFS embed.FS

type Data struct {
	IP   net.IP `json:"ip"`
	Host string `json:"host"`
}

type Certs struct {
	PrivateKey  string `json:"key"`
	Certificate string `json:"cert"`
}

func main() {

	mux := http.NewServeMux()

	serverRoot, err := fs.Sub(embeddedFS, "app/dist")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", http.FileServer(http.FS(serverRoot)))

	mux.HandleFunc("/api/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var data Data

		dec := json.NewDecoder(r.Body)
		defer r.Body.Close()

		err := dec.Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		if data.IP == nil || !data.IP.IsPrivate() {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing/invalid ip param"))
			return
		}
		if data.Host == "" {
			data.Host = "homeassistant.local"
		}

		certs, err := gencert(data.IP.String(), data.Host)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Add("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.Encode(certs)

	})

	http.ListenAndServe(":3000", mux)

}

var uid int
var uidMu sync.Mutex

func gencert(ip string, host string) (Certs, error) {

	var certs Certs

	uidMu.Lock()
	path := filepath.Join(os.TempDir(), fmt.Sprintf("openssl.%d", uid))
	uid++
	uidMu.Unlock()

	args := []string{
		"req",
		"-subj", "/C=GB/CN=HomeAssistant",
		"-sha256",
		"-addext", `subjectAltName = IP:` + ip + `, DNS:` + host,
		"-newkey", "rsa:4096", "-nodes",
		"-keyout", path + ".key",
		"-x509", "-days", "24855",
		"-out", path + ".chain",
	}

	log.Println("openssl", args)

	defer os.Remove(path + ".key")
	defer os.Remove(path + ".chain")

	cmd := exec.Command("openssl", args...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Println(string(output))
		return certs, err
	}

	b, err := os.ReadFile(path + ".key")
	if err != nil {
		return certs, err
	}
	certs.PrivateKey = string(b)

	b, err = os.ReadFile(path + ".chain")
	if err != nil {
		return certs, err
	}
	certs.Certificate = string(b)

	return certs, nil
}
