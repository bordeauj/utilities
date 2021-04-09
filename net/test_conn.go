package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "strings"
    "time"
    "net"
    "flag"
)

var argFilename = flag.String("f", "hosts.txt", "Filename with list of hostname:portto check")

func Reader(filename string, threads int) []chan string {
    var chans = []chan string{}

    for i:=0; i<threads; i++ {
      chans = append(chans, make(chan string))
    } 

    go func() {
        file, err := os.Open(filename)
        if err != nil {
            log.Fatal(err)
        }
        defer file.Close()

        scanner := bufio.NewScanner(file)

        j := threads
        for scanner.Scan() {
            //fmt.Println(scanner.Text())
            chans[j % threads] <- scanner.Text()
            j++
        }

        if err := scanner.Err(); err != nil {
            log.Fatal(err)
        }

        for _, c := range chans { close(c) } 
    }()

    return chans
}

func Checker(chin chan string, chout chan string) bool {
  var host string
  timeout := time.Duration(10000 * time.Millisecond)
  var conn net.Conn
  var err  error
  var m = map[bool]string{
	  true:  "Connected!",
	  false: "Timeout!",
	}

  for line := range chin {
      ftp := strings.Split(line, "/")
      if len(ftp) > 1 {
          host = ftp[1]
      } else { host = ftp[0] }

      //fmt.Println(host)
      conn, err = net.DialTimeout("tcp", host, timeout)

      if (err != nil) {
        //chout <- string("Res of " + line + " (" + host + ") is " + m[err == nil])
        //fmt.Printf("%s - %s\n", host, err)
        switch err := err.(type) {
          case net.Error:
              if err.Timeout() {
                  chout <- string("Res of " + line + " (" + host + ") is " + m[err == nil])
              }
          }
      }

      if (err == nil) { conn.Close() }
  }

  chout <- "Done"
  return false
}

func main() {
   workers := 25

   flag.Parse()

   chins := Reader(*argFilename, workers)

   chout := make(chan string, 1000)

   for _, ch := range chins {
     go Checker(ch, chout)
   }

   for out := range chout {
     if (out == "Done") {workers--
     } else {
        fmt.Printf("%s\n", out)
     }
     if (workers == 0) {break}
   }

}
