package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-redis/redis"
)

type Redis struct {
	Addr  string
	Port  int
	State string
}

func getSlaves(masterHost string, masterPort int) []Redis {
	var slaves []Redis
	c := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", masterHost, masterPort),
		Password:     options.auth,
		DialTimeout:  100 * time.Millisecond,
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
	})
	text, err := c.Do("info", "replication").String()
	if err != nil {
		log.Println(err)
	}
	lines := strings.Split(text, "\n")
	re := regexp.MustCompile("slave[0-9]*:ip=(.*),port=([0-9]+),state=([a-z]+).*")
	for _, line := range lines {
		if !strings.HasPrefix(line, "slave") {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if len(matches) < 4 {
			continue
		}
		host := matches[1]
		port, err := strconv.Atoi(matches[2])
		if err != nil {
			log.Fatalln(err)
		}
		state := matches[3]
		slaves = append(slaves, Redis{Addr: host, Port: port, State: state})
	}
	return slaves
}
func PrintTopology() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		host := fields[0]
		port, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Fatalln(err)
		}
		ips, err := net.LookupHost(host)
		if err != nil {
			log.Fatalln(err)
		}
		if ips[0] != host {
			fmt.Printf("%s:%d %s\n", host, port, ips)
		} else {
			fmt.Printf("%s:%d\n", ips[0], port)
		}
		slaves := getSlaves(host, port)
		PrintSlaves(slaves, "├──")
	}
}

func spaces(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += " "
	}
	return s
}
func PrintSlaves(slaves []Redis, prefix string) {
	for _, rds := range slaves {
		fmt.Printf("%s %s:%d %s\n", prefix, rds.Addr, rds.Port, rds.State)
		slaves := getSlaves(rds.Addr, rds.Port)
		if len(slaves) > 0 {
			PrintSlaves(slaves, "│"+spaces(utf8.RuneCountInString(prefix))+"└──")
		}
	}
}

var options = &struct {
	auth string
}{}

func main() {
	flag.StringVar(&options.auth, "a", "", "auth of redis")
	flag.Parse()

	PrintTopology()
}
