package main

import (
	"fmt"
	"regexp"
	"testing"
)

func TestValidationTrue(t *testing.T) {
	if !ValidateUrl("https://www.google.de") {
		t.Fatalf("Unexpected return value! - 1")
	}
	if !ValidateUrl("http://www.duckduckgo.org") {
		t.Fatalf("Unexpected return value! - 2")
	}
	if !ValidateUrl("http://localhost:8080") {
		t.Fatalf("Unexpected return value! - 3")
	}
	if !ValidateUrl("https://192.168.0.2/pugrest.de") {
		t.Fatalf("Unexpected return value! - 4")
	}
	if !ValidateUrl("https://3en.m.wikipedia.org:3333/wiki/Euler–Lagrange_equation/index.html?bla=blub#sidemarker") {
		t.Fatalf("Unexpected return value! - 5")
	}
}

func TestValidationFalse(t *testing.T) {
	if ValidateUrl("ht tps://w.google") {
		t.Fatalf("Unexpected return value! - 1")
	}
	if ValidateUrl("http:://www.duckduckgo.org") {
		t.Fatalf("Unexpected return value! - 2")
	}
	if ValidateUrl("hhh://localhost:8080") {
		t.Fatalf("Unexpected return value! - 3")
	}
	if ValidateUrl("pugrest") {
		t.Fatalf("Unexpected return value! - 4")
	}
	if ValidateUrl("https://3en.m.wikiped fsdfsd fm  l?bla=blub#sidemarker") {
		t.Fatalf("Unexpected return value! - 5")
	}
}

func TestSnippets(t *testing.T) {
    // same regex but with group-names
	paramsMap := getParams(`^(?:(?P<protocol>http|https):){1}?(?:\/{0,3})(?P<domain>[0-9.\-A-Za-z]+)(?::(?P<port>\d+))?(?:\/(?P<path>[^?#]*))?(?:\?([^#]*))?(?:#(?P<anchor>.*))?$`, "http://www.google.de:234/aa/bb/cc/index.html&param=laram#section")

	fmt.Printf("map: %v", paramsMap)
	fmt.Printf("len: %v", len(paramsMap))

}

/**
 * Parses url with the given regular expression and returns the
 * group values defined in the expression.
 */
func getParams(regEx, url string) (paramsMap map[string]string) {

	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(url)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return paramsMap
}
