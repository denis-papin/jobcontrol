package main

import (
	"fmt"
	"io/ioutil"
	"time"
	"log"
	"os"
	"os/exec"
	//"io"        
	"net/http" 
	"strconv"	
	// "bufio"
	"github.com/BurntSushi/toml"
	"strings"
	"path/filepath"
)

// run --server eureka-peer-2 -- port 10011 --profile eureka-dev  --project /home/denis/CI/eureka-server --file eureka.toml
// stop --server eureka-peer-2
type Params struct {
	server string
	port string
	profile string
	project string
	file string
	bin string
	action string
}


type Config struct {
	App Application
}


type Application struct {
	Bin  		string
	Cmd			string 
	Status 	   	Status
}


type Status struct {
	Url string
	Keyword string
}


func getConf(configFile string) *Config {	
    _, err := os.Stat(configFile)
    if err != nil {
        log.Fatal("Config file is missing: ", configFile)
    }

    var config Config
    if _, err := toml.DecodeFile(configFile, &config); err != nil {
        log.Fatal(err)
    }
    //log.Print(config)
    return &config
}

func parse_param( args []string ) *Params {


	if len(args) == 1 {

		fmt.Println("\nMissing arguments for jobcontrol \n");

		fmt.Println("jobcontrol run|stop|list  --server <server-name> --port <port> --profile <profile>  --project  <path> --file <config file> ")

		fmt.Println("\nExamples \n");

		fmt.Println("jobcontrol run --server localhost --port 10010 --profile eureka-dev  --project /home/denis/official/eureka-server --file eureka.toml ")

		fmt.Println("jobcontrol stop --server localhost [--port 10010] [--profile eureka-dev] ")

		fmt.Println("jobcontrol list ")

		os.Exit(-45)

	} 

	// parse the command line params
	var param = &Params {"", "", "", "", "", "", ""} 

	// run or stop
	param.action = os.Args[1]
	
	for i, arg := range args {
		switch arg {
			case "--project":
				param.project = os.Args[i+1]
			case "--profile" :
				param.profile = os.Args[i+1]
			case "--server" :
				param.server = os.Args[i+1]
			case "--port" :
				param.port = os.Args[i+1]
			case "--file" :
				param.file = os.Args[i+1]
		}
	}

	return param
}


/**
*/
func findMatch( pattern string ) string {
	var elected = ""
	matches, err := filepath.Glob(pattern)

	if err != nil {
		log.Fatal(err)
	}

	if len(matches) > 0 {
		elected = matches[0]
	}

	return elected
}


/**
*/
func CreateDirIfNotExist(dir string) {
	// fmt.Println("Check folder", dir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
					panic(err)
			}
	}
}

func evaluate( expression string, p *Params ) string {

	expression = strings.Replace(expression, "${SERVER}", p.server, -1)
	expression = strings.Replace(expression, "${PORT}", p.port, -1)
	expression = strings.Replace(expression, "${PROJECT}", p.project, -1)
	expression = strings.Replace(expression, "${PROFILE}", p.profile, -1)

	// extra params
	expression = strings.Replace(expression, "${bin}", p.bin, -1)

	return expression
}

func readPid(server string, profile string, port string) string {

	var serviceName = server + "." + profile

	job_control_folder := getJcFolder() 

	var base_dir = job_control_folder + "/.jc"

	CreateDirIfNotExist(base_dir )

	filename := base_dir + "/" + serviceName + "." + port

	return readTextFile(filename)

}

func readTextFile( fileName string ) string {

	// fmt.Println(">> FileName", fileName)

	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer f.Close()

	//r := bufio.NewReader(f)
	//line, err := r.ReadString(100)    // line defined once

	b, err := ioutil.ReadAll(f)

	line := string(b)
	
	//fmt.Println(">> Line : ", line)
	
	line = strings.Trim(line, "\r")
	line = strings.Trim(line, "\n")
	line = strings.Trim(line, " ")

	//fmt.Println(">> Clean line : ", line )

	var parts = strings.Split(line, ":")
	var pid = strings.Trim( parts[1], " " )
	return pid
}



func writePid(pid int, server string, profile string, port string  ) {
		
	var serviceName = server + "." + profile

	job_control_folder := getJcFolder() 

	var base_dir = job_control_folder + "/.jc"

	CreateDirIfNotExist(base_dir )

    f, err := os.Create( base_dir + "/" + serviceName + "." + port )
    check(err)

    defer f.Close()

    n3, err := f.WriteString( serviceName + ":" + strconv.Itoa(pid)  )
	fmt.Printf("wrote %d bytes\n", n3)
	check(err)

    f.Sync()
}

func deletePid( server string, profile string, port string  ) {
	var serviceName = server + "." + profile
	job_control_folder := getJcFolder() 
	var base_dir = job_control_folder + "/.jc"

	filename := base_dir + "/" + serviceName + "." + port

	err := os.Remove(filename)
	if err != nil {
		panic(err)
	}
}

/**
*/
func check(e error) {
    if e != nil {
        panic(e)
    }
}


/**
	Ex : localhost.*.9999 or  *.ms-mission-dev.*
 */
func genericMatch( expression string ) []string {

	job_control_folder := getJcFolder() 

	var base_dir = job_control_folder + "/.jc"

	// Read all the pid files
	var matches []string
	var err error
	matches, err = filepath.Glob(base_dir + "/" + expression)
	if err != nil {
		log.Fatal(err)
	}
	return matches
}


/**

 */
func list(param *Params) {

	// fmt.Println(" Listing  : " , param.server)
	matches := genericMatch("*")

	showHeader()

	for _, ff := range matches {
		// fmt.Println("file:", ff)

		filename := filepath.Base(ff)
		parts := strings.Split(filename, ".")
		// fmt.Println(filename)
		pid, _ := strconv.Atoi(readTextFile(ff))
		param1 := Params{parts[0],  parts[2], parts[1], "", "", "", ""}
		showInfo(pid, &param1, "Running")
	}


}


func starIfEmpty(str string) string {

	//fmt.Println("STRING TO STAR :[", str, "]")
	if strings.Trim(str, " ") == "" {
	//	fmt.Println("EMPTY")
		return "*"
	}
	return str
}

/**

 */
func stop(param *Params) {
	
	// fmt.Println(" Stopping  : " , param.server)

	expression := starIfEmpty(param.server) + "." + starIfEmpty(param.profile) + "." + starIfEmpty(param.port)
	matches := genericMatch(expression)

	// fmt.Println("Matches  : ", matches)

	showHeader()

	for _, ff := range matches {
		// fmt.Println("file:", ff)

		filename := filepath.Base(ff)
		parts := strings.Split(filename, ".")
		// fmt.Println(filename)
		pid_str := readTextFile(ff)
		pid, _ := strconv.Atoi(pid_str)

		// fmt.Println("found pid : ", pid)

		param1 := Params{parts[0],  parts[2], parts[1], "", "", "", ""}

		cmd := exec.Command( "kill", "-9", pid_str )
		cmd.Stdout = os.Stdout
		err := cmd.Start()
		if err != nil {
			log.Fatal(err)
		} else {
			deletePid(param1.server, param1.profile, param1.port)

			showInfo(pid, &param1, "Stopped")

		}

	}

	os.Exit(0)

}


/**
*/
func run(param *Params) {

	// Read the .toml file

	conf := getConf(param.file)
	// fmt.Println(" configuration file  : " , conf)

	// Evaluate the expressions, replace the values in each conf param.

	// bin
	conf.App.Bin = evaluate( conf.App.Bin, param )
	param.bin = conf.App.Bin;

	//fmt.Println("Evaluate bin :", param.bin)

	// Found the binary (jar) that matches the pattern, ex : /home/denis/official/eureka-server/target/eureka-server-*.jar
	// Give /home/denis/official/eureka-server/target/eureka-server-0.0.1-SNAPSHOT.jar
	param.bin = findMatch( param.bin )
	
	if param.bin == "" {
		log.Fatal("No binary file found" )
	}

	//fmt.Println("FOUND MATCH :", param.bin)

	// cmd
	conf.App.Cmd = evaluate( conf.App.Cmd, param )
	//fmt.Println(" Cmd  : " , conf.App.Cmd)

	// Url
	conf.App.Status.Url = evaluate( conf.App.Status.Url, param )

	// Keyword
	conf.App.Status.Keyword = evaluate( conf.App.Status.Keyword, param )

	// run the main command
	var parts = strings.Split(conf.App.Cmd, " ")
	var clean_parts []string = []string{}

	for _, part := range parts {

		if strings.Trim(part, " ") != "" {			
			clean_parts = append( clean_parts, strings.Trim(part, " ") )

		}		
	}
	

	command := clean_parts[0]
	args := clean_parts[1:]

	log.Println("[", command, args, "]")

	cmd := exec.Command( command, args...)

	/* -- WORKS
	cmd := exec.Command(
						"java", 
						"-Dspring.profiles.active=eureka-peer-2",						
						"-jar",						
						"/home/denis/official/eureka-server/target/eureka-server-0.0.1-SNAPSHOT.jar")
						*/
				
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	//log.Printf("Just ran subprocess %d, exiting\n", cmd.Process.Pid)
	
	// URL check 
	
	var success = false
	
	// loop until the status is "UP"

	for !success {

		log.Println("Trying the server...")

		response, err := http.Get(conf.App.Status.Url)

		if err != nil {
			log.Println("No Response Yet")
		} else {
				defer response.Body.Close()
		
				_, err := ioutil.ReadAll(response.Body)
				if err != nil {						
					log.Println("No Response Yet")

				} else {
					success = true
					log.Println("!! Server is running !!")
				}
		}


		if  !success {
			time.Sleep(2000 * time.Millisecond)
		}
			
	}

	writePid( cmd.Process.Pid, param.server, param.profile, param.port )

	showHeader()
	showInfo(cmd.Process.Pid, param, "Running")
	os.Exit(0)

}


func getJcFolder() string {

	job_control_folder, _ := os.LookupEnv("HOME")
	job_control_folder = strings.Trim(job_control_folder, " ")
	if job_control_folder == "" {
		log.Fatal("Cannot find the $HOME environment variable.")
		os.Exit(-99)
	}

	// fmt.Println("Found JC folder : ", job_control_folder )
	return job_control_folder
}

func showHeader() {
	log.Print("[STATUS]", "\t\t", "[SERVER]", "\t\t", "[PORT]" , "\t\t", "[PROFILE]" , "\t\t", "[PID]")
}

func showInfo(pid int, param *Params, status string) {
	log.Print(status, "\t\t", param.server , "\t\t", param.port, "\t\t",  param.profile, "\t\t" ,  pid)
}

/**
*/
func main() {

	fmt.Println("jobcontrol v1.0\n")

	// parse the command line params
	var param = parse_param(os.Args)

	//fmt.Println(param)

	switch param.action {

		case "run" :
			run(param)
		case "stop" :
			stop(param)
		case "list" :
			list(param)
	}


}