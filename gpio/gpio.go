package gpio
//package main

import (
	"fmt"
//	"log"
	"errors"
	"os"
	"time"
	"unsafe"
	"reflect"
	"syscall"
//	"runtime"
)

//-----------------------------------------------------------------------------
const (
	LOCKFILE                = "/tmp/gpio.lock"
	LOCKTIMEOUT             = 120
	LOCKEXPIRE              = 300
	BCM2708_PERI_BASE	= 0x3F000000
	GPIO_BASE		= (BCM2708_PERI_BASE + 0x200000)
	BLOCK_SIZE		= (4 * 1024)
)

var (
	mem32 []uint32
	mem8 []byte
)

// Note: Pin number is BCM number
//-----------------------------------------------------------------------------
func Setup() (err error){
	var f *os.File

	// Open /dev/mem
	f, err = os.OpenFile("/dev/gpiomem",
		os.O_RDWR | os.O_SYNC,
		0644)

	if err != nil {
		fmt.Println("Can't open gpio")
		return
	}

	// mmap GPIO
	mem8, err = syscall.Mmap(int(f.Fd()),
		GPIO_BASE, BLOCK_SIZE,
		syscall.PROT_READ | syscall.PROT_WRITE,
		syscall.MAP_SHARED)

	if err != nil {
		fmt.Println("Can't mmap gpio")
		return
	}

	// no need f handler anymore
	if err = f.Close(); err != nil {
		return
	}

	header := *(*reflect.SliceHeader)(unsafe.Pointer(&mem8))
	header.Len /= (32 / 8)
	header.Cap /= (32 / 8)

	mem32 = *(*[]uint32)(unsafe.Pointer(&header))

	return
}

func Close() {
	syscall.Munmap(mem8)
}

//-----------------------------------------------------------------------------
func Gpio_lock()(error) {
	//pc, file, line, _ := runtime.Caller(1)

	t1 := time.Now()
	for stat, err := os.Stat(LOCKFILE); !os.IsNotExist(err); {
		// Check if the lockfile is too old -> bug?
		if (time.Now().Sub(stat.ModTime())).Seconds() > LOCKEXPIRE {
			//break
		}

		t2 := time.Now()
		if (t2.Sub(t1)).Seconds() > LOCKTIMEOUT {
			return errors.New("gpio lock timeout")
		}

		time.Sleep(1 * time.Second)
		//fmt.Println("DEBUG: GPIO_LOCK at ", pc, file, line, "is waiting")
		stat, err = os.Stat(LOCKFILE)
	}

	fd, _ := syscall.Open(LOCKFILE, syscall.O_CREAT | syscall.O_RDONLY, 0666)
	defer syscall.Close(fd)

	if err := syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return err
	}

	//fmt.Println("DEBUG: GPIO_LOCK at ", pc, file, line)

	return nil
}

func Gpio_unlock()(error) {
	//pc, file, line, _ := runtime.Caller(1)

	if err := os.Remove(LOCKFILE); err != nil {
		return errors.New("gpio unlock failed")
	}

	//fmt.Println("DEBUG: GPIO_UNLOCK at ", pc, file, line)
	return nil
}

//-----------------------------------------------------------------------------
func Set_all_input() {
	mem32[0] = 0x00	// GPFSEL0
	mem32[1] = 0x00	// GPFSEL1
	mem32[2] = 0x00 // GPFSEL2
}

func Set_input(pin uint32) {
	mem32[(pin/10)] &= ^(7 << ((pin % 10) * 3))
}

func Set_output(pin uint32) {
	Set_input(pin)
	mem32[(pin/10)] |= (1 << ((pin % 10) * 3))
}
//-----------------------------------------------------------------------------
func Set_pin(pin uint32) {
	mem32[7] = Get_bus() | (1 << pin)
}

func Clr_pin(pin uint32) {
	mem32[10] = Get_bus() | (1 << pin)
}

func Set_bus(v uint32) {
	mem32[7] = v
}

func Clr_bus(v uint32) {
	mem32[10] = v
}

func Get_pin(pin uint32) uint32 {
	return (mem32[13] & (1 << pin)) >> pin
}

func Get_bus() uint32 {
	return mem32[13]
}

//func main() {
//	gpiomem, gpiomem8, _ = Setup()
//	//set_input(gp, 1)
//	//set_input(gp, 20)
//	//fmt.Printf("%x\n", Get_pin(gp, 0))
//	Close()
//}

