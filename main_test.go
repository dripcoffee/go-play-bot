package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestRunCode(t *testing.T) {
	code := `
package main

import "fmt"

func main() {
	fmt.Println("Hello, world")
}
	`
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
	resp, _ := runWithPlayground(code)
	fmt.Println(resp)
}

func TestFmtCode(t *testing.T) {
	code := `
package main
import "fmt"
func main() {
fmt.Println("Hello, world")
}
	`
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
	resp, _ := fmtWithPlayground(code)
	fmt.Println(resp)
}

func TestParseToken(t *testing.T) {
	msg := `run
package main
import "fmt"
func main() {
fmt.Println("Hello, world")
}
	`
	cmd, code := parseToken(msg)
	if cmd != "run" {
		t.Error(cmd)
	}
	msg = `fmt
package main
import "fmt"
func main() {
fmt.Println("Hello, world")
}
	`
	cmd, code = parseToken(msg)
	if cmd != "fmt" {
		t.Error(cmd)
	}
	if strings.HasPrefix(code, "run") {
		t.Error(code)
	}
}
