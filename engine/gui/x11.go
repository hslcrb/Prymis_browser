package gui

import (
	"encoding/binary"
	"fmt"
	"image"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

type X11Window struct {
	conn     net.Conn
	wid      uint32
	gcid     uint32
	width    uint16
	height   uint16
	depth    uint8
	visualID uint32
	root     uint32
}

type Event struct {
	Type int
	Key  byte
}

const (
	KeyPress        = 2
	Expose          = 12
	ConfigureNotify = 22
)

func NewX11Window(width, height uint16) (*X11Window, error) {
	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":0"
	}
	// Simplified display parsing
	socketPath := "/tmp/.X11-unix/X0"
	if strings.Contains(display, ":") {
		parts := strings.Split(display, ":")
		suffix := strings.Split(parts[1], ".")[0]
		socketPath = "/tmp/.X11-unix/X" + suffix
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to X11: %v", err)
	}

	// 1. Handshake
	buf := make([]byte, 12)
	buf[0] = 'l'
	binary.LittleEndian.PutUint16(buf[2:4], 11)
	_, err = conn.Write(buf)
	if err != nil {
		return nil, err
	}

	resp := make([]byte, 8)
	_, err = io.ReadFull(conn, resp)
	if err != nil {
		return nil, err
	}
	if resp[0] != 1 {
		return nil, fmt.Errorf("X11 handshake failed: %d", resp[0])
	}

	setupLen := binary.LittleEndian.Uint16(resp[6:8]) * 4
	setupData := make([]byte, setupLen)
	_, err = io.ReadFull(conn, setupData)
	if err != nil {
		return nil, err
	}

	// Extract IDs and root window
	resBase := binary.LittleEndian.Uint32(setupData[4:8])
	resMask := binary.LittleEndian.Uint32(setupData[8:12])
	vendorLen := binary.LittleEndian.Uint16(setupData[16:18])
	numFormats := uint8(setupData[21])

	// Skip formats
	offset := 32 + int((vendorLen+3)&0xFFFC) + int(numFormats)*8

	// Screen info
	rootWindow := binary.LittleEndian.Uint32(setupData[offset : offset+4])
	// Try to find a standard 24-bit or 32-bit depth
	rootDepth := uint8(setupData[offset+21])
	visualID := binary.LittleEndian.Uint32(setupData[offset+32 : offset+36])

	wid := resBase | (1 & resMask)
	gcid := resBase | (2 & resMask)

	win := &X11Window{
		conn:     conn,
		wid:      wid,
		gcid:     gcid,
		width:    width,
		height:   height,
		depth:    rootDepth,
		visualID: visualID,
		root:     rootWindow,
	}

	// 2. Create Window
	// Mask: BackPixel (0x2) | EventMask (0x800) | BorderPixel (0x4)
	createBuf := make([]byte, 48)
	createBuf[0] = 1 // CreateWindow
	createBuf[1] = rootDepth
	binary.LittleEndian.PutUint16(createBuf[2:4], 12)
	binary.LittleEndian.PutUint32(createBuf[4:8], wid)
	binary.LittleEndian.PutUint32(createBuf[8:12], rootWindow)
	binary.LittleEndian.PutUint16(createBuf[12:14], 0) // x
	binary.LittleEndian.PutUint16(createBuf[14:16], 0) // y
	binary.LittleEndian.PutUint16(createBuf[16:18], width)
	binary.LittleEndian.PutUint16(createBuf[18:20], height)
	binary.LittleEndian.PutUint16(createBuf[20:22], 0) // border
	binary.LittleEndian.PutUint16(createBuf[22:24], 1) // InputOutput
	binary.LittleEndian.PutUint32(createBuf[24:28], visualID)

	// Value Mask: BackgroundPixel (0x2) | BorderPixel (0x4) | EventMask (0x800)
	binary.LittleEndian.PutUint32(createBuf[28:32], 0x806)
	binary.LittleEndian.PutUint32(createBuf[32:36], 0xFFFFFF) // background white
	binary.LittleEndian.PutUint32(createBuf[36:40], 0x000000) // border black
	binary.LittleEndian.PutUint32(createBuf[40:44], 0x8001)   // KeyPress(1) | Exposure(0x8000)

	conn.Write(createBuf)

	// 3. Create GC
	gcBuf := make([]byte, 16)
	gcBuf[0] = 55
	binary.LittleEndian.PutUint16(gcBuf[2:4], 4)
	binary.LittleEndian.PutUint32(gcBuf[4:8], gcid)
	binary.LittleEndian.PutUint32(gcBuf[8:12], wid)
	conn.Write(gcBuf)

	// 4. Set Window Title & Class (ICCCM)
	win.SetTitle("Prymis Browser")
	win.setProperty("WM_CLASS", "prymis\x00Prymis\x00")

	// 5. Map Window
	mapBuf := make([]byte, 8)
	mapBuf[0] = 8
	binary.LittleEndian.PutUint16(mapBuf[2:4], 2)
	binary.LittleEndian.PutUint32(mapBuf[4:8], wid)
	conn.Write(mapBuf)

	return win, nil
}

func (w *X11Window) setProperty(propName string, value string) {
	// Simplified property setting (requires atom intern but we hardcode common ones)
	// WM_NAME=39, WM_CLASS=67
	propID := uint32(0)
	if propName == "WM_NAME" {
		propID = 39
	}
	if propName == "WM_CLASS" {
		propID = 67
	}
	if propID == 0 {
		return
	}

	data := []byte(value)
	dataLen := len(data)
	paddedLen := (dataLen + 3) & ^3
	totalLen := 6 + paddedLen/4

	buf := make([]byte, 24+paddedLen)
	buf[0] = 18 // ChangeProperty
	buf[1] = 0  // Replace
	binary.LittleEndian.PutUint16(buf[2:4], uint16(totalLen))
	binary.LittleEndian.PutUint32(buf[4:8], w.wid)
	binary.LittleEndian.PutUint32(buf[8:12], propID)
	binary.LittleEndian.PutUint32(buf[12:16], 31) // STRING
	buf[16] = 8
	binary.LittleEndian.PutUint32(buf[20:24], uint32(dataLen))
	copy(buf[24:], data)
	w.conn.Write(buf)
}

func (w *X11Window) SetTitle(title string) {
	w.setProperty("WM_NAME", title)
}

func (w *X11Window) PollEvent() *Event {
	w.conn.SetReadDeadline(time.Now().Add(time.Millisecond))
	buf := make([]byte, 32)
	n, _ := w.conn.Read(buf)
	if n < 32 {
		return nil
	}

	evType := int(buf[0] & 0x7f)
	switch evType {
	case KeyPress:
		keycode := buf[1]
		// Incredibly primitive keycode to ASCII (US layout approximation)
		// 10=1, 11=2... 24=q, 38=a, 52=z, 65=space
		char := byte(0)
		if keycode >= 24 && keycode <= 33 {
			char = "qwertyuiop"[keycode-24]
		}
		if keycode >= 38 && keycode <= 46 {
			char = "asdfghjkl"[keycode-38]
		}
		if keycode >= 52 && keycode <= 58 {
			char = "zxcvbnm"[keycode-52]
		}
		if keycode == 65 {
			char = ' '
		}
		if keycode == 22 {
			char = 8
		} // Backspace
		if keycode == 36 {
			char = 13
		} // Enter
		if keycode == 47 {
			char = ';'
		}
		if keycode == 48 {
			char = '\''
		}
		if keycode == 51 {
			char = '\\'
		}
		if keycode == 59 {
			char = ','
		}
		if keycode == 60 {
			char = '.'
		}
		if keycode == 61 {
			char = '/'
		}
		// Numbers
		if keycode >= 10 && keycode <= 19 {
			char = "1234567890"[keycode-10]
		}

		return &Event{Type: KeyPress, Key: char}
	case Expose:
		return &Event{Type: Expose}
	}
	return nil
}

func (w *X11Window) Draw(img *image.RGBA) error {
	bounds := img.Bounds()
	width := uint16(bounds.Dx())
	height := uint16(bounds.Dy())

	pix := img.Pix
	bgrx := make([]byte, int(width)*int(height)*4)
	for i := 0; i < int(width)*int(height); i++ {
		r, g, b := pix[i*4], pix[i*4+1], pix[i*4+2]
		bgrx[i*4] = b
		bgrx[i*4+1] = g
		bgrx[i*4+2] = r
	}

	headerLen := 6
	dataLen := (len(bgrx) + 3) / 4
	header := make([]byte, 24)
	header[0] = 72 // PutImage
	header[1] = 2  // ZPixmap
	binary.LittleEndian.PutUint16(header[2:4], uint16(headerLen+dataLen))
	binary.LittleEndian.PutUint32(header[4:8], w.wid)
	binary.LittleEndian.PutUint32(header[8:12], w.gcid)
	binary.LittleEndian.PutUint16(header[12:14], width)
	binary.LittleEndian.PutUint16(header[14:16], height)
	header[21] = 24
	w.conn.Write(header)
	w.conn.Write(bgrx)
	return nil
}

func (w *X11Window) Close() {
	if w.conn != nil {
		w.conn.Close()
	}
}
