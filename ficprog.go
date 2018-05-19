//-----------------------------------------------------------------------------
// nyacom FiC FPGA programmer (golang version) (C) 2018.05
// <kzh@nyacom.net>
// Xilinx SelectMAP x16 interface
// References
// https://japan.xilinx.com/support/documentation/application_notes/j_xapp583-fpga-configuration.pdf
// https://japan.xilinx.com/support/documentation/user_guides/j_ug570-ultrascale-configuration.pdf
//-----------------------------------------------------------------------------
package main

import (
	"fmt"
//	"flag"
	"log"
	"os"
	"time"
//	"unsafe"
//	"reflect"
//	"syscall"
	"./gpio"
)

const (
	BUFSIZE = (1024*1024)
)

// PRi PINS
var PIN = map[string] uint {
	"RP_INIT" : 4,
	"RP_RDWR" : 27,
	"RP_PROG" : 5,
	"RP_DONE" : 6,
	"RP_CCLK" : 7,

	"RP_CD0" : 8,
	"RP_CD1" : 9,
	"RP_CD2" : 10,
	"RP_CD3" : 11,

	"RP_PWOK" : 24,
	"RP_G_CKSEL" : 25,
	"RP_CSI" : 26,
}

var PIN_BIT = map[string] uint {
	"RP_INIT" : (1 << PIN["RP_INIT"]),
	"RP_RDWR" : (1 << PIN["RP_RDWR"]),
	"RP_PROG" : (1 << PIN["RP_PROG"]),
	"RP_DONE" : (1 << PIN["RP_DONE"]),
	"RP_CCLK" : (1 << PIN["RP_CCLK"]),

	"RP_CD0" : (1 << PIN["RP_CD0"]),
	"RP_CD1" : (1 << PIN["RP_CD1"]),
	"RP_CD2" : (1 << PIN["RP_CD2"]),
	"RP_CD3" : (1 << PIN["RP_CD3"]),

	"RP_PWOK" : (1 << PIN["RP_PWOK"]),
	"RP_G_CKSEL" : (1 << PIN["RP_G_CKSEL"]),
	"RP_CSI" : (1 << PIN["RP_CSI"]),
}

//-----------------------------------------------------------------------------
func setup() {
	gpio.Setup()

	for k, v := range PIN {
		switch k {
			case "RP_PWOK", "RP_INIT", "RP_DONE", "RP_G_CKSEL": {
				gpio.Set_input(v)
			}
			default: {
				gpio.Set_output(v)
				gpio.Clr_pin(v)
			}
		}
	}

	// Check power ok
	fmt.Println("CHECK: PW_OK:", gpio.Get_pin(PIN["RP_PWOK"]))
}

func prog(infile string) {
	fmt.Println("PROG: Entering Xilinx SelectMap x16 configuration mode...")

	// Invoke configuration
	gpio.Set_bus(uint32(PIN_BIT["RP_PROG"]))
	gpio.Clr_bus(uint32(PIN_BIT["RP_PROG"]|PIN_BIT["RP_CSI"]|PIN_BIT["RP_RDWR"]))
	gpio.Set_bus(uint32(PIN_BIT["RP_PROG"]))

	for gpio.Get_pin(PIN["RP_INIT"]) == 0 {
		time.Sleep(1 * time.Second)
	}

	fmt.Println("PROG: Ready to program")

	// Open bin file
	f, err := os.Open(infile)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	f_info, err := f.Stat()
	if err != nil{
		log.Fatal(err)
	}

	file_size_kb := int(f_info.Size() / 1024)
	fmt.Println("PROG: File size : ", file_size_kb, " KB")

	gpio.Set_bus(uint32(PIN["RP_CCLK"]))

	fmt.Println("PROG: Programming...")
	buf := make([]byte, BUFSIZE)

	read_byte := 0

	for ;; {
		n, err := f.Read(buf)

		if n == 0 {
			// read byte is 0
			break
		}

		if err != nil {
			// something happened
			log.Fatal(err)
		}

		read_byte += (BUFSIZE / 1024)

		for i := 0; i < len(buf); i = i+2 {
			gpio.Clr_bus(0x00ffff00 | uint32(PIN_BIT["RP_CCLK"]))
			//gpio.Set_bus(((uint32(buf[i]) << 8 | uint32(buf[i+1])) << 8) & 0x00ffff00)
			gpio.Set_bus(((uint32(buf[i+1]) << 8 | uint32(buf[i])) << 8) & 0x00ffff00)
			gpio.Set_bus(uint32(PIN_BIT["RP_CCLK"]))	// Assert CLK

			//fmt.Printf("%x\n", gpio.Get_bus())

			if gpio.Get_pin(PIN["RP_INIT"]) == 0 {
				gpio.Clr_bus(0x00ffff00 | uint32(PIN_BIT["RP_CCLK"]))
				log.Fatal("Configuraion Error (while prog)")
			}
		}

		fmt.Printf("PROG: %d / %d (%.2f %%)\n",
			read_byte, file_size_kb, float32(read_byte) / float32(file_size_kb) * 100)
	}

	gpio.Clr_bus(uint32(PIN_BIT["RP_CCLK"]))	// Negate CLK

	fmt.Println("PROG: Waiting FPGA done")

	for gpio.Get_pin(PIN["RP_DONE"]) == 0 {		// Wait until RP_DONE asserted
		if gpio.Get_pin(PIN["RP_INIT"]) == 0 {
			gpio.Clr_bus(0x00ffff00 | uint32(PIN_BIT["RP_CCLK"]))
			log.Fatal("Configuration Error (while waiting)")
		}
		gpio.Set_bus(uint32(PIN_BIT["RP_CCLK"]))
		gpio.Clr_bus(uint32(PIN_BIT["RP_CCLK"]))
	}

	gpio.Clr_bus(0x00ffff00 | uint32(PIN_BIT["RP_CCLK"]))
	fmt.Println("PROG: FPGA program done")
}

//-----------------------------------------------------------------------------
func init() {
	fmt.Println("")
	fmt.Println("FiC FPGA Configurator (golang ver)")
	fmt.Println("nyacom (C) 2018.05 <kzh@nyacom.net>")

	//f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	//fileOpt := f.String(" ", "default", "*.mcs file for FPGA")

	//f.Parse(os.Args[1:])
	//for 0 < f.NArg() {
	//	f.Parse(f.Args()[1:])
	//}

	// Check arguments
	if len(os.Args) < 2 {
		help_str()
		log.Fatal("Insufficient argument")
	}
}

func help_str() {
	fmt.Println("Usage: ficprog.go INPUT_FILE.bin")
}

func main() {
	infile := os.Args[1]
	fmt.Println("Filename: ", infile)

	setup()
	prog(infile)
}

