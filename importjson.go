package main


//run with
//
//cat main.go | pygmentize -f raw | perl -MJSON -ne '@a=split(/\s+/, $_,2);$a[1]=~ s/^u.//;chop $a[1];chop $a[1];print encode_json(\@a)."\n"' | go run importjson.go

import (
    "bufio"
    "os"
    "fmt"
    "encoding/json"
)


func parseStdin() [][]string{
    scanner := bufio.NewScanner(os.Stdin)
    all := [][]string{}
    for scanner.Scan() {
        vals := make([]string, 0)
        line := scanner.Text()
        json.Unmarshal([]byte(line), &vals)
        all = append(all, vals)
    }

    if scanner.Err() != nil {
        // handle error.
    }
    return all
}



func main() {
    tokens := parseStdin()
    fmt.Printf("%+v\n", tokens)
}
