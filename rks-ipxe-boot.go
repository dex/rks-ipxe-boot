package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func usage() {
	fmt.Printf(`Usage: %s <kernel> <initrd> [<port>] [<cmdline>...]
This utility will host the <kernel> and <initrd> at port <port> for iPXE boot
loader to boot from. <port> is default to 8080 if it is not
speicifed. Other arguments after <port> will be used as additional kernel command line parameters.

<kernel> and <initrd> could also be URLs.

Here is an example usage:
run this command
>%s vmlinuz initrd.img
and inside iPXE, you should boot from this url:
http://<your-ip>:8080/

Another example adding additional kernel command line:
>%s vmlinuz initrd.img 8080 console=ttyS0,115200
You might need the additional console parameter when you are installing a
SCG-100.
`, os.Args[0], os.Args[0], os.Args[0])
}

func is_url(path string) bool {
	schemes := []string{"http", "tftp", "https"}
	for _, s := range schemes {
		prefix := fmt.Sprintf("%s://", s)
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func check_existence(files []string) {
	for _, f := range files {
		if !is_url(f) {
			if _, err := os.Stat(f); os.IsNotExist(err) {
				log.Fatalf("file %s does not exists", f)
			}
		}
	}
}

type Ipxe struct {
	kernel, initrd, cmd string
}

func (i *Ipxe) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		fallthrough
	case "/boot.txt":
		fmt.Fprintf(w, `#!ipxe
kernel %s %s
initrd %s
boot
`, i.kernel, i.cmd, i.initrd)
	default:
		http.ServeFile(w, r, r.URL.Path[1:])
	}
}

func main() {
	argc := len(os.Args)
	port := 8080
	cmd := []string{"stage2=initrd:"}

	if argc < 3 {
		usage()
		os.Exit(1)
	}
	check_existence(os.Args[1:3])

	if argc >= 4 {
		var err error
		port, err = strconv.Atoi(os.Args[3])
		if err != nil {
			usage()
			os.Exit(1)
		}
		cmd = append(cmd, os.Args[4:]...)
	}

	fmt.Println("serving at port", port)
	addr := fmt.Sprintf(":%d", port)
	handler := &Ipxe{os.Args[1], os.Args[2], strings.Join(cmd, " ")}
	log.Fatal(http.ListenAndServe(addr, handler))
}
