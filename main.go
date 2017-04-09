package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"strings"
	"time"
)

var Global = make(map[string]paxos)

type paxos struct {
	value string
	// key
	key string
	// seq = N+ IP
	requestor seq
	// keeps track of commands recived
	command []string
}
type seq struct {
	Num       int
	currentIP string
}

// ex req = newseq(420, "WATY")
//pax := paxos{requestor: *req}
//db["test"] = pax
// var tmp = db["test"]
// tmp.requestor.Num = 320
// db["test"] = tmp

func newseq(x int, ip string) *seq {
	return &seq{
		Num:       x,
		currentIP: ip,
	}
}

var (
	n                     = 0
	numNodes      float64 = 1
	ports         []string
	myport        = 3410
	numofCommands int
	count         = 0
	sleeptime     = 1000
)

func main() {

	// n for chatty, 0=stand by, 1 = receving or proposing and 2 for accepted
	// var n int

	// ports for current nodes 3410, 3400,3420 ..etc

	// your default port is 3410.

	// var paxo []*paxos

	prompt := "Paxos> "
	// Enter the repl
	help("")
	fmt.Println(``)
	for {

		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error: %q\n", err)
		}
		cmmd := strings.Split(strings.TrimSpace(line), " ")
		fmt.Printf(prompt)
		// input, err := getInput()
		// if err != nil {
		// 	fmt.Println("There was an error with your command")
		// 	continue
		// }
		if _, ok := commands[cmmd[0]]; !ok {
			fmt.Println("Command not found.")
			continue
		}
		err = commands[cmmd[0]](cmmd[1:]...)
		if err != nil {
			fmt.Println(err)
		}
	}
}

var commands = map[string]func(args ...string) error{
	"help":   help,
	"put":    putFn,
	"get":    getFn,
	"delete": deleteFn,
	"quit":   quitFn,
	"port":   portFn,
	"dump":   dumpFn,
	"ports":  portsFn,
}

type PutArgs struct {
	Key, Val string
}

func (n *paxos) Put(args PutArgs, success *bool) error {
	n.Data[args.Key] = args.Val
	*success = true
	return nil
}
func putFn(args ...string) error {
	key := args[0]
	value := args[1]
	propose(key, args)
	return nil
}

func help(args ...string) error {
	fmt.Println("Commands \n help: this displays a list of recognized commands.\n")
	fmt.Println("put <key> <value>: inserts a key and a value into Paxos DB. \n quit: shut down.")
	fmt.Println("get <key>: returns a value assoicated with that key.\n delete <key>: delete a value or sets to 0 \n dump: display information about the current node")
	fmt.Println("Ping <address> :  a ping request to an address")
	fmt.Println("get, put, and delete + <address> keyboard commands that require an address to be specified")
	return nil
}
func Compare(a, b seq) int {
	if a.Num < b.Num {
		return -1
	}
	if a.Num > b.Num {
		return 1
	}

	return 0

}

// sets sets paxoes to a key.
func updateseq(key string, elt paxos) {
	tmp := Global[key]
	tmp = elt
	Global[key] = tmp
}
func perpare(key string, seq *seq) bool {
	//sets chatty to 1
	n = 1
	if Global[key].requestor == nil {
		tmppaxos := paxos{requestor: *seq}
		updateseq(key, tmppaxos)
		// Global[key].requestor = seq
		fmt.Println("promised" + key + " to" + Global[key].requestor)
		time.Sleep(sleeptime)
		return true
	}
	if seq > Global[key].requestor {
		tmppaxos := paxos{requestor: *seq}
		updateseq(key, tmppaxos)
		// Global[key].requestor = seq
		fmt.Println("promised" + key + " to" + Global[key].requestor)
		time.Sleep(sleeptime)
		return true
	} else {
		time.Sleep(sleeptime)
		return false
		fmt.Println(" failed to promise" + key + " to" + seq)
	}

}
func accept(key string, seq seq, command string) bool {
	n = 2 // chatting in accepting status
	if seq == Global[key].requestor && checkIP(seq, Global[key].requestor) {
		//Applycommand(command) // applies requested commands
		fmt.Println("Command:" + command + "has been applied to" + key + "by " + seq)
		time.Sleep(sleeptime)
		return true
	}
	return false
}

// bool of sucesss or failed to insert.
func propose(key string, command []string) {
	n = 1 // chatty
	var minvotes float64
	if math.Remainder(numNodes, 2) == 0 {
		minvotes = numNodes / 2

	} else {
		minvotes = (numNodes / 2) + 0.5
	}
	if Global[key].requestor.Num == 0 {
		count++
		currentPropsal := newseq(count, getLocalAddress())
		// for each replica call out perpare
		// perpare(key,currentPropsal)
		//for each true Yesvote++
		// for each false  Novote++
		// if minvote == Yesvotes { break, accept (key, command)}
		// maybe launch a go routine to check if minvotes met.

	} else if Global[key].requestor != nil && count != db[key].requestor.Num {
		count = Global[key].requestor.Num
		count++
		currentPropsal := newseq(count, getLocalAddress())
		// for each replica
		// perpare(key, currentPropsal)
		// checks if minvotes met! if then break

	}

}

// checks if both requestors have the same IP
func checkIP(a, b seq) bool {
	if a.currentIP == b.currentIP {
		return true
	}
	// if a.currentIP != b.currentIP {
	// 	return false
	// }
	return false
}

func getLocalAddress() string {
	var localaddress string

	ifaces, err := net.Interfaces()
	if err != nil {
		panic("init: failed to find network interfaces")
	}

	// find the first non-loopback interface with an IP address
	for _, elt := range ifaces {
		if elt.Flags&net.FlagLoopback == 0 && elt.Flags&net.FlagUp != 0 {
			addrs, err := elt.Addrs()
			if err != nil {
				panic("init: failed to get addresses for network interface")
			}

			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ip4 := ipnet.IP.To4(); len(ip4) == net.IPv4len {
						localaddress = ip4.String()
						break
					}
				}
			}
		}
	}
	if localaddress == "" {
		panic("init: failed to find non-loopback interface with valid address on this node")
	}

	return localaddress
}
