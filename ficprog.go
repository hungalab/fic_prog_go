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
	"flag"
	"log"
	"os"
	"time"
//	"unsafe"
//	"reflect"
//	"syscall"
	"./gpio"
)

const (
	LOCKFILE = "/tmp/gpio.lock"
	TIMEOUT = 5
	LOCKEXPIRE = 120
	BUFSIZE = (2*1024*1024)
)

// PRi PINS
var PIN = map[string] uint32 {
	"RP_INIT" : 4,
	"RP_PROG" : 5,
	"RP_DONE" : 6,
	"RP_CCLK" : 7,

	"RP_CD0" : 8,
	"RP_CD1" : 9,
	"RP_CD2" : 10,
	"RP_CD3" : 11,
	"RP_CD4" : 12,
	"RP_CD5" : 13,
	"RP_CD6" : 14,
	"RP_CD7" : 15,
	"RP_CD8" : 16,
	"RP_CD9" : 17,
	"RP_CD10" : 18,
	"RP_CD11" : 19,
	"RP_CD12" : 20,
	"RP_CD13" : 21,
	"RP_CD14" : 22,
	"RP_CD15" : 23,
	"RP_CD16" : 24,
	"RP_CD17" : 25,

	"RP_PWOK" : 24,
	"RP_G_CKSEL" : 25,
	"RP_CSI" : 26,
	"RP_RDWR" : 27,
}

var PIN_BIT = map[string] uint32 {
	"RP_PWOK" : (1 << PIN["RP_PWOK"]),	// Input
	"RP_INIT" : (1 << PIN["RP_INIT"]),
	"RP_DONE" : (1 << PIN["RP_DONE"]),
	"RP_G_CKSEL" : (1 << PIN["RP_G_CKSEL"]),

	"RP_CD0" : (1 << PIN["RP_CD0"]),	// Output
	"RP_CD1" : (1 << PIN["RP_CD1"]),
	"RP_CD2" : (1 << PIN["RP_CD2"]),
	"RP_CD3" : (1 << PIN["RP_CD3"]),
	"RP_CD4" : (1 << PIN["RP_CD4"]),
	"RP_CD5" : (1 << PIN["RP_CD5"]),
	"RP_CD6" : (1 << PIN["RP_CD6"]),
	"RP_CD7" : (1 << PIN["RP_CD7"]),
	"RP_CD8" : (1 << PIN["RP_CD8"]),
	"RP_CD9" : (1 << PIN["RP_CD9"]),
	"RP_CD10" : (1 << PIN["RP_CD10"]),
	"RP_CD11" : (1 << PIN["RP_CD11"]),
	"RP_CD12" : (1 << PIN["RP_CD12"]),
	"RP_CD13" : (1 << PIN["RP_CD13"]),
	"RP_CD14" : (1 << PIN["RP_CD14"]),
	"RP_CD15" : (1 << PIN["RP_CD15"]),
	"RP_CD16" : (1 << PIN["RP_CD16"]),
	"RP_CD17" : (1 << PIN["RP_CD17"]),

	"RP_PROG" : (1 << PIN["RP_PROG"]),
	"RP_CCLK" : (1 << PIN["RP_CCLK"]),
	"RP_CSI" : (1 << PIN["RP_CSI"]),
	"RP_RDWR" : (1 << PIN["RP_RDWR"]),
}

//-----------------------------------------------------------------------------
// Create lockfile before GPIO operation
//-----------------------------------------------------------------------------
func gpio_lock() {
	t1 := time.Now()
	for stat, err := os.Stat(LOCKFILE); !os.IsNotExist(err); {
		// Check if the lockfile is too old -> bug?
		if (time.Now().Sub(stat.ModTime())).Seconds() > LOCKEXPIRE {
			break
		}

		time.Sleep(1 * time.Second)

		t2 := time.Now()
		if (t2.Sub(t1)).Seconds() > TIMEOUT {
			log.Fatal("timeout")
		}
	}

	fd, err := os.OpenFile(LOCKFILE, os.O_CREATE, 0666);
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()
}

func gpio_unlock() {
	if err:= os.Remove(LOCKFILE); err != nil {
		log.Fatal(err)
	}
}

//-----------------------------------------------------------------------------
func init_pin16() {
	gpio.Set_all_input()
	for _, v := range PIN {
		switch v {
			// Set input
			case PIN["RP_PWOK"], PIN["RP_INIT"], PIN["RP_DONE"], PIN["RP_G_CKSEL"]: {
				gpio.Set_input(v)
			}
			// Set output
			default: {
				gpio.Set_output(v)
				gpio.Clr_bus(1<<v)
			}
		}
	}
}

//-----------------------------------------------------------------------------
func init_pin8() {
	gpio.Set_all_input()
	for _, v := range PIN {
		switch v {
			// Set input
			case PIN["RP_PWOK"], PIN["RP_INIT"], PIN["RP_DONE"], PIN["RP_G_CKSEL"]: {
				gpio.Set_input(v)
			}
			// Set output
			case PIN["RP_PROG"], PIN["RP_CCLK"], PIN["RP_CSI"], PIN["RP_RDWR"],
				PIN["RP_CD0"], PIN["RP_CD1"], PIN["RP_CD2"], PIN["RP_CD3"],
				PIN["RP_CD4"], PIN["RP_CD5"], PIN["RP_CD6"], PIN["RP_CD7"] : {
				gpio.Set_output(v)
				gpio.Clr_bus(1<<v)
			}
		}
	}
}

//-----------------------------------------------------------------------------
func setup() {
	gpio.Setup()
	gpio.Set_all_input()

	// Check power ok
	//fmt.Println("CHECK: PW_OK:", gpio.Get_pin(PIN["RP_PWOK"]))
}

// prog with Selectmap 8 method
func prog8(infile string) {
	init_pin8()
	fmt.Println("PROG: Entering Xilinx SelectMap x8 configuration mode...")

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

	file_size := f_info.Size()
	fmt.Println("PROG: File size : ", file_size, " B")

	gpio.Clr_bus(uint32(PIN["RP_CCLK"]))

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

		read_byte += n

		for i := 0; i < n; i = i + 1 {
			data := (uint32(buf[i]) << 8)
			gpio.Clr_bus((^data & 0x0000ff00) | uint32(PIN_BIT["RP_CCLK"]))
			gpio.Set_bus((data & 0x0000ff00))
			gpio.Set_bus(uint32(PIN_BIT["RP_CCLK"]))

			//if data != 0x00 {
			//	fmt.Fprintf(os.Stderr,  "%d DEBUG: data = %08x\n", j, gpio.Get_bus())
			//	j++
			//}
			//fmt.Printf("%x\n", gpio.Get_bus())

			if gpio.Get_pin(PIN["RP_INIT"]) == 0 {
				log.Fatal("Configuraion Error (while prog)")
			}
		}

		fmt.Printf("PROG: %d / %d (%.2f %%)\n",
			read_byte, file_size, float32(read_byte) / float32(file_size) * 100)
	}

	//gpio.Clr_bus(uint32(PIN_BIT["RP_CCLK"]))	// Negate CLK

	fmt.Println("PROG: Waiting FPGA done")

	for gpio.Get_pin(PIN["RP_DONE"]) == 0 {		// Wait until RP_DONE asserted
		if gpio.Get_pin(PIN["RP_INIT"]) == 0 {
			log.Fatal("Configuration Error (while waiting)")
		}
		gpio.Set_bus(uint32(PIN_BIT["RP_CCLK"]))
		gpio.Clr_bus(uint32(PIN_BIT["RP_CCLK"]))
	}

	gpio.Clr_bus(0x0000ff00 | uint32(PIN_BIT["RP_CCLK"]))
	fmt.Println("PROG: FPGA program done")

	defer gpio.Set_all_input()
}

// prog with Selectmap 16 method
func prog16(infile string) {
	init_pin16()
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

	file_size := f_info.Size()
	fmt.Println("PROG: File size : ", file_size, " B")

	gpio.Clr_bus(uint32(PIN["RP_CCLK"]))

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

		read_byte += n

		for i := 0; i < n; i = i + 2 {
			data := (uint32(buf[i+1]) << 8 | uint32(buf[i])) << 8
			gpio.Clr_bus((^data & 0x00ffff00) | uint32(PIN_BIT["RP_CCLK"]))
			gpio.Set_bus((data & 0x00ffff00))
			gpio.Set_bus(uint32(PIN_BIT["RP_CCLK"]))

			//if data != 0x00 {
			//	fmt.Fprintf(os.Stderr,  "%d DEBUG: data = %08x\n", j, gpio.Get_bus())
			//	j++
			//}
			//fmt.Printf("%x\n", gpio.Get_bus())

			if gpio.Get_pin(PIN["RP_INIT"]) == 0 {
				log.Fatal("Configuraion Error (while prog)")
			}
		}

		fmt.Printf("PROG: %d / %d (%.2f %%)\n",
			read_byte, file_size, float32(read_byte) / float32(file_size) * 100)
	}

	//gpio.Clr_bus(uint32(PIN_BIT["RP_CCLK"]))	// Negate CLK

	fmt.Println("PROG: Waiting FPGA done")

	for gpio.Get_pin(PIN["RP_DONE"]) == 0 {		// Wait until RP_DONE asserted
		if gpio.Get_pin(PIN["RP_INIT"]) == 0 {
			log.Fatal("Configuration Error (while waiting)")
		}
		gpio.Set_bus(uint32(PIN_BIT["RP_CCLK"]))
		gpio.Clr_bus(uint32(PIN_BIT["RP_CCLK"]))
	}

	gpio.Clr_bus(0x00ffff00 | uint32(PIN_BIT["RP_CCLK"]))
	fmt.Println("PROG: FPGA program done")

	defer gpio.Set_all_input()
}

//-----------------------------------------------------------------------------
func init() {
	fmt.Println("")
	fmt.Println("FiC FPGA Configurator (golang ver)")
	fmt.Println("nyacom (C) 2018.05 <kzh@nyacom.net>")
}

func help_str() {
	fmt.Println("Usage: ficprog.go INPUT_FILE.bin [-m {8, 16}]")
}

func main() {
	// Check arguments
	if len(os.Args) < 2 {
		help_str()
		log.Fatal("Insufficient argument")
	}

	if os.Args[1] == "help" {
	}

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var (
		confMode = fs.Int("m", 16, "Selectmap mode (default 16)")
	)

	noopt := os.Args[1]
	fs.Parse(os.Args[1:])

	left_args := fs.Args()
	if len(left_args) > 0 {
		noopt = left_args[0]
	}
	fs.Parse(left_args[1:])

	// Check options
	if *confMode != 8 && *confMode != 16 {
		log.Fatal("Error: Invalid configurtaion mode")
	}

	// Help
	if noopt == "help" {
		help_str()
		return
	}

	infile := noopt

	// Create GPIO lockfile
	gpio_lock();
	defer gpio_unlock();

	// GPIO setup
	setup()

	// Select configuration mode
	switch *confMode {
	case 8:
		prog8(infile)

	case 16:
		prog16(infile)
	}
}

