package main

import (
	"log"
	"common"
	"io"
	"cpio"
	"strconv"
	"os"
	"strings"
)

func parseInput(in io.ReadCloser, c chan<- *Entry) {
	buf := common.NewBufferedReader(in)
	for {
		line, e := buf.ReadWholeLine()
		if e != nil && e != io.EOF {
			log.Printf("Warning: Could not read whole file: %s", e.Error())
			if !*continueFlag {
				break
			}
		}
		ent := parseLine(line)
		if ent != nil {
			c <- ent
		}

		if e == io.EOF {
			break
		}
	}
	close(c)
}

func parseLine(line string) *Entry {
	lineparts := strings.Split(line, " ")
	if len(lineparts) < 1 {
		return nil
	}

	switch(lineparts[0]) {
		case "file":
			return parseFile(lineparts[1:])
		case "dir":
			return parseDir(lineparts[1:])
		case "nod":
			return parseNod(lineparts[1:])
		case "slink":
			return parseSlink(lineparts[1:])
		case "pipe":
			return parsePipe(lineparts[1:])
		case "sock":
			return parseSock(lineparts[1:])
		default:
			log.Printf("Warning: %s is in invalid type\n", lineparts[0])
	}
	return nil
}

func parseFile(parts []string) *Entry {
	if len(parts) != 5 {
		return nil
	}
	name := parts[0]
	localname := parts[1]
	mode, uid, gid, e := parseModeUidGid(parts[2], parts[3], parts[4])
	if e != nil {
		log.Printf("Invalid permission settings: %s\n", e.Error())
		return nil
	}

	f, e := os.Open(localname)
	if e != nil {
		log.Printf("Could not open file %s: %s\n", localname, e.Error())
		return nil
	}
	finfo, e := f.Stat()
	if e != nil {
		log.Printf("Could not obtain file size of %s: %s\n")
		return nil
	}

	return &Entry {
		hdr: cpio.Header {
			Mode: mode,
			Uid: uid,
			Gid: gid,
			Size: finfo.Size,
			Type: cpio.TYPE_REG,
			Name: name,
		},
		data: f,
	}
}

func parseDir(parts []string) *Entry {
	if len(parts) != 4 {
		return nil
	}
	name := parts[0]
	mode, uid, gid, e := parseModeUidGid(parts[1], parts[2], parts[3])
	if e != nil {
		log.Printf("Invalid permission settings: %s\n", e.Error())
		return nil
	}

	return &Entry {
		hdr: cpio.Header {
			Mode: mode,
			Uid: uid,
			Gid: gid,
			Type: cpio.TYPE_DIR,
			Name: name,
		},
	}
}

func parseNod(parts []string) *Entry {
	if len(parts) != 7 {
		return nil
	}
	name := parts[0]
	mode, uid, gid, e := parseModeUidGid(parts[1], parts[2], parts[3])
	if e != nil {
		log.Printf("Invalid permission settings: %s\n", e.Error())
		return nil
	}

	var dev_type int64
	switch(parts[4]) {
		case "b":
			dev_type = cpio.TYPE_BLK
		case "c":
			dev_type = cpio.TYPE_CHAR
		default:
			log.Printf("Invalid device type: %s\n", parts[4])
			return nil
	}

	maj, e := strconv.Atoi64(parts[5])
	if e != nil {
		log.Printf("Invalid major device: %s\n", e.Error())
		return nil
	}
	min, e := strconv.Atoi64(parts[6])
	if e != nil {
		log.Printf("Invalid major device: %s\n", e.Error())
		return nil
	}

	return &Entry {
		hdr: cpio.Header {
			Mode: mode,
			Uid: uid,
			Gid: gid,
			Type: dev_type,
			Devmajor: maj,
			Devminor: min,
			Name: name,
		},
	}

}

func parseSlink(parts []string) *Entry {
	if len(parts) != 5 {
		return nil
	}

	name := parts[0]
	target := parts[1]

	mode, uid, gid, e := parseModeUidGid(parts[2], parts[3], parts[4])
	if e != nil {
		log.Printf("Invalid permission settings: %s\n", e.Error())
		return nil
	}

	return &Entry {
		hdr: cpio.Header {
			Mode: mode,
			Uid: uid,
			Gid: gid,
			Type: cpio.TYPE_SYMLINK,
			Size: int64(len(target)+1),
			Name: name,
		},
		data: strings.NewReader(target),
	}

}

func parsePipe(parts []string) *Entry {
	if len(parts) != 4 {
		return nil
	}
	name := parts[0]
	mode, uid, gid, e := parseModeUidGid(parts[1], parts[2], parts[3])
	if e != nil {
		log.Printf("Invalid permission settings: %s\n", e.Error())
		return nil
	}

	return &Entry {
		hdr: cpio.Header {
			Mode: mode,
			Uid: uid,
			Gid: gid,
			Type: cpio.TYPE_FIFO,
			Name: name,
		},
	}
}

func parseSock(parts []string) *Entry {
	if len(parts) != 4 {
		return nil
	}
	name := parts[0]
	mode, uid, gid, e := parseModeUidGid(parts[1], parts[2], parts[3])
	if e != nil {
		log.Printf("Invalid permission settings: %s\n", e.Error())
		return nil
	}

	return &Entry {
		hdr: cpio.Header {
			Mode: mode,
			Uid: uid,
			Gid: gid,
			Type: cpio.TYPE_SOCK,
			Name: name,
		},
	}
}

func parseModeUidGid(s_mode, s_uid, s_gid string) (mode int64, uid, gid int, err error) {
	mode, err = strconv.Btoi64(s_mode, 0)
	if err != nil {
		return
	}
	uid, err = strconv.Atoi(s_uid)
	if err != nil {
		return
	}
	gid, err = strconv.Atoi(s_gid)
	return mode, uid, gid, nil
}
