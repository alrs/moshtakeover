package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/shirou/gopsutil/process"
)

func parentMoshPid() (int32, error) {
	prc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return 0, err
	}
	for {
		rawName, err := prc.Name()
		name := strings.TrimSpace(rawName)
		if name == "mosh-server" {
			return prc.Pid, nil
		}
		parent, err := prc.Parent()
		if err != nil {
			return 0, err
		}
		if parent.Pid == 1 {
			return 0, errors.New("not running under mosh-server")
		}
		prc, err = prc.Parent()
		if err != nil {
			return 0, err
		}

	}
}

func uidMatch(thisUID int32, uids []int32) bool {
	for _, uid := range uids {
		if thisUID == uid {
			return true
		}
	}
	return false
}

func main() {
	pmp, err := parentMoshPid()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("parent mosh-server: %d", pmp)

	thisUid := int32(os.Getuid())
	pids, _ := process.Pids()
	for _, pid := range pids {
		p, err := process.NewProcess(pid)
		if err != nil {
			log.Printf("NewProcess: %v", err)
		}
		uids, err := p.Uids()
		if err != nil {
			log.Fatalf("error getting uids: %v", err)
		}
		if uidMatch(thisUid, uids) && // this process username matches desired uid
			pmp != pid { // process is not the parent process
			nameWithNil, err := p.Name()
			if err != nil {
				log.Fatal(err)
			}
			name := strings.TrimSpace(nameWithNil)
			if name == "mosh-server" {
				log.Printf("killing mosh-server: %d", pid)
				err = p.Kill()
				if err != nil {
					log.Fatal(err)
				}
				os.Exit(0)
			}
		}
	}
}
