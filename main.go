package main

import (
	"code.google.com/p/go-sqlite/go1/sqlite3"
	"dns-stats/collector"
	"dns-stats/report"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var (
	dbname        string
	collectorPort string
	reportPort    string
	storeInterval string
)

func init() {
	flag.StringVar(&dbname, "db", os.Getenv("HOME")+"/dns.sqlite3", "Absolute path to SQLite3 database")
	flag.StringVar(&collectorPort, "collector-port", ":1514", "Address for syslog collector to listen to")
	flag.StringVar(&reportPort, "report-port", ":8514", "Address for report server to listen to")
	flag.StringVar(&storeInterval, "store-interval", "1m", "Defines the interval for cached queries storage")
}

func main() {
	flag.Parse()

	fmt.Printf("Configuration parameters: \n  db -> %s\n  collector-port -> %s\n  report-port -> %s\n  store-interval -> %s\n",
		dbname, collectorPort, reportPort, storeInterval)

	initializeDB()

	collector.CollectorPort = collectorPort
	collector.DBName = dbname
	collector.StoreInterval = storeInterval
	report.ReportPort = reportPort
	report.DBName = dbname

	go report.Run()
	s := collector.Run()

	sc := make(chan os.Signal, 2)
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT)
	<-sc

	fmt.Println("Storing...")
	collector.Store()
	fmt.Println("Shutdown the server...")
	s.Shutdown()
	fmt.Println("Server is down")
}

func initializeDB() {
	var db *sqlite3.Conn
	var err error
	if db, err = sqlite3.Open(dbname); err != nil {
		fmt.Println("Error opening database:", err)
		return
	}

	if err = db.Exec("CREATE TABLE IF NOT EXISTS queries (date DATE, origin INTEGER, destination INTEGER)"); err != nil {
		fmt.Println("Error creating table:", err)
		return
	}

	if err = db.Exec("CREATE TABLE IF NOT EXISTS hosts (id INTEGER PRIMARY KEY, fqdn TEXT)"); err != nil {
		fmt.Println("Error creating table:", err)
		return
	}
	db.Close()
}
