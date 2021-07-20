// moshtakeover, a program to kill non-parent mosh-server processes
// Copyright (C) 2020 Lars Lehtonen

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
		prc = parent
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
			if errors.Is(err, process.ErrorProcessNotRunning) {
				// process went away since we gathered pids, this is normal
				continue
			}
			log.Printf("NewProcess err:%v type:%T", err, err)
			continue
		}
		uids, err := p.Uids()
		if err != nil {
			log.Fatalf("error getting uids: %v", err)
		}
		if uidMatch(thisUid, uids) && // this process' uid matches desired uid
			pmp != pid { // process is not our parent mosh-server
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
			}
		}
	}
}
