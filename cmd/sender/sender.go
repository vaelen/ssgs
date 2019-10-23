/*
 * Copyright 2019 Andrew C. Young
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
)

/***************
 * Main Method *
 ***************/

var ByteOrder = binary.LittleEndian

func main() {
	if len(os.Args) < 3 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Printf("Sends files to an ssgs port as packets.\n\n")
		fmt.Printf("Usage: sender <address> <filename> [filename...]\n")
	}

	address := os.Args[1]

	// Connect
	log.Printf("Connecting to %v...\n", address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	defer conn.Close()

	// Send files
	fileList := os.Args[2:]
	fileCount := len(fileList)
	for i, filename := range fileList {
		fileNumber := i + 1
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("Couldn't load file %v/%v.  Filename: %v, Error: %v\n",
				fileNumber, fileCount, filename, err)
		}
		var dataLength uint32 = uint32(len(data))
		err = binary.Write(conn, ByteOrder, &dataLength)
		if err != nil {
			log.Fatalf("Couldn't send file %v/%v.  Filename: %v, Error: %v\n",
				fileNumber, fileCount, filename, err)
		}
		err = binary.Write(conn, ByteOrder, data)
		if err != nil {
			log.Fatalf("Couldn't send file %v/%v.  Filename: %v, Error: %v\n",
				fileNumber, fileCount, filename, err)
		}
		log.Printf("Sent file %v/%v. Filename: %v, Length: %v\n",
			fileNumber, fileCount, filename, dataLength)
	}

	log.Printf("Sent %v files.\n", fileCount)
}
