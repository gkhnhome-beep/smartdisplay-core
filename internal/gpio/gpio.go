package gpio

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

var exportedPins = make(map[int]bool)

type GPIOPin struct {
	PinNumber int
	Direction string // "in" or "out"
	Value     int    // 0 or 1
}

func (p *GPIOPin) Export() error {
	if exportedPins[p.PinNumber] {
		log.Printf("GPIO Export: pin %d already exported", p.PinNumber)
		return fmt.Errorf("pin %d already exported", p.PinNumber)
	}
	path := "/sys/class/gpio/export"
	err := writeString(path, strconv.Itoa(p.PinNumber))
	if err == nil {
		exportedPins[p.PinNumber] = true
		log.Printf("GPIO Export: pin %d exported", p.PinNumber)
	} else {
		log.Printf("GPIO Export ERROR: pin %d: %v", p.PinNumber, err)
	}
	return err
}

func (p *GPIOPin) Unexport() error {
	if !exportedPins[p.PinNumber] {
		log.Printf("GPIO Unexport: pin %d not exported", p.PinNumber)
		return fmt.Errorf("pin %d not exported", p.PinNumber)
	}
	path := "/sys/class/gpio/unexport"
	err := writeString(path, strconv.Itoa(p.PinNumber))
	if err == nil {
		delete(exportedPins, p.PinNumber)
		log.Printf("GPIO Unexport: pin %d unexported", p.PinNumber)
	} else {
		log.Printf("GPIO Unexport ERROR: pin %d: %v", p.PinNumber, err)
	}
	return err
}

func (p *GPIOPin) SetDirection(dir string) error {
	if dir != "in" && dir != "out" {
		log.Printf("GPIO SetDirection ERROR: pin %d invalid direction %s", p.PinNumber, dir)
		return fmt.Errorf("invalid direction: %s", dir)
	}
	path := filepath.Join("/sys/class/gpio", fmt.Sprintf("gpio%d", p.PinNumber), "direction")
	err := writeString(path, dir)
	if err == nil {
		p.Direction = dir
		log.Printf("GPIO SetDirection: pin %d set to %s", p.PinNumber, dir)
	} else {
		log.Printf("GPIO SetDirection ERROR: pin %d: %v", p.PinNumber, err)
	}
	return err
}

func (p *GPIOPin) Read() (int, error) {
	if p.Direction != "in" {
		log.Printf("GPIO Read ERROR: pin %d not input", p.PinNumber)
		return 0, fmt.Errorf("cannot read from output pin %d", p.PinNumber)
	}
	path := filepath.Join("/sys/class/gpio", fmt.Sprintf("gpio%d", p.PinNumber), "value")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("GPIO Read ERROR: pin %d: %v", p.PinNumber, err)
		return 0, err
	}
	v, err := strconv.Atoi(string(b[:1]))
	if err == nil {
		p.Value = v
		log.Printf("GPIO Read: pin %d value %d", p.PinNumber, v)
	} else {
		log.Printf("GPIO Read ERROR: pin %d: %v", p.PinNumber, err)
	}
	return v, err
}

func (p *GPIOPin) Write(value int) error {
	if p.Direction != "out" {
		log.Printf("GPIO Write ERROR: pin %d not output", p.PinNumber)
		return fmt.Errorf("cannot write to input pin %d", p.PinNumber)
	}
	if value != 0 && value != 1 {
		log.Printf("GPIO Write ERROR: pin %d invalid value %d", p.PinNumber, value)
		return fmt.Errorf("invalid value: %d", value)
	}
	path := filepath.Join("/sys/class/gpio", fmt.Sprintf("gpio%d", p.PinNumber), "value")
	err := writeString(path, strconv.Itoa(value))
	if err == nil {
		p.Value = value
		log.Printf("GPIO Write: pin %d set to %d", p.PinNumber, value)
	} else {
		log.Printf("GPIO Write ERROR: pin %d: %v", p.PinNumber, err)
	}
	return err
}

func writeString(path, val string) error {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(val)
	return err
}
