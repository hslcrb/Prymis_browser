package gui

import (
	"encoding/binary"
	"fmt"
	"image"
	"net"
)

type X11Window struct {
	conn     net.Conn
	wid      uint32
	gcid     uint32
	width    uint16
	height   uint16
	depth    uint8
	visualID uint32
}

func NewX11Window(width, height uint16) (*X11Window, error) {
	conn, err := net.Dial("unix", "/tmp/.X11-unix/X0")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to X11: %v", err)
	}

	// 1. Handshake
	// Client Hello: l + 0 + 11.0 + auth_proto_len + auth_data_len + 0
	buf := make([]byte, 12)
	buf[0] = 'l'                                // Little endian
	binary.LittleEndian.PutUint16(buf[2:4], 11) // major
	binary.LittleEndian.PutUint16(buf[4:6], 0)  // minor
	_, err = conn.Write(buf)
	if err != nil {
		return nil, err
	}

	// Read Server Response (Setup)
	resp := make([]byte, 8)
	_, err = conn.Read(resp)
	if err != nil {
		return nil, err
	}

	if resp[0] != 1 {
		return nil, fmt.Errorf("X11 handshake failed with status: %d", resp[0])
	}

	totalLen := binary.LittleEndian.Uint16(resp[6:8]) * 4
	setupData := make([]byte, totalLen)
	_, err = conn.Read(setupData)
	if err != nil {
		return nil, err
	}

	// Extract IDs and root window
	resBase := binary.LittleEndian.Uint32(setupData[4:8])
	resMask := binary.LittleEndian.Uint32(setupData[8:12])

	// Finder: skip over vendor string and reach screens
	vendorLen := binary.LittleEndian.Uint16(setupData[16:18])
	numFormats := uint8(setupData[21])
	offset := 32 + int((vendorLen+3)&0xFFFC) + int(numFormats)*8

	rootWindow := binary.LittleEndian.Uint32(setupData[offset : offset+4])
	visualID := binary.LittleEndian.Uint32(setupData[offset+32 : offset+36])

	wid := resBase | (1 & resMask)
	gcid := resBase | (2 & resMask)

	win := &X11Window{
		conn:     conn,
		wid:      wid,
		gcid:     gcid,
		width:    width,
		height:   height,
		depth:    24, // Assuming 24-bit
		visualID: visualID,
	}

	// 2. Create Window (Opcode 1)
	// uint8 opcode, uint8 depth, uint16 length, uint32 wid, uint32 parent,
	// int16 x, int16 y, uint16 width, uint16 height, uint16 border, uint16 class,
	// uint32 visual, uint32 mask, ...
	createBuf := make([]byte, 40)
	createBuf[0] = 1                                  // CreateWindow
	createBuf[1] = 24                                 // depth
	binary.LittleEndian.PutUint16(createBuf[2:4], 10) // length
	binary.LittleEndian.PutUint32(createBuf[4:8], wid)
	binary.LittleEndian.PutUint32(createBuf[8:12], rootWindow)
	binary.LittleEndian.PutUint16(createBuf[12:14], 0) // x
	binary.LittleEndian.PutUint16(createBuf[14:16], 0) // y
	binary.LittleEndian.PutUint16(createBuf[16:18], width)
	binary.LittleEndian.PutUint16(createBuf[18:20], height)
	binary.LittleEndian.PutUint16(createBuf[20:22], 0) // border
	binary.LittleEndian.PutUint16(createBuf[22:24], 1) // InputOutput
	binary.LittleEndian.PutUint32(createBuf[24:28], visualID)
	binary.LittleEndian.PutUint32(createBuf[28:32], 0x02)     // background-pixel mask (was 0x800)
	binary.LittleEndian.PutUint32(createBuf[32:36], 0xFFFFFF) // white

	_, err = conn.Write(createBuf)
	if err != nil {
		return nil, err
	}

	// 3. Create GC (Opcode 55)
	gcBuf := make([]byte, 16)
	gcBuf[0] = 55 // CreateGC
	binary.LittleEndian.PutUint16(gcBuf[2:4], 4)
	binary.LittleEndian.PutUint32(gcBuf[4:8], gcid)
	binary.LittleEndian.PutUint32(gcBuf[8:12], wid)
	// No values
	_, err = conn.Write(gcBuf)
	if err != nil {
		return nil, err
	}

	// 4. Map Window (Opcode 8)
	mapBuf := make([]byte, 8)
	mapBuf[0] = 8 // MapWindow
	binary.LittleEndian.PutUint16(mapBuf[2:4], 2)
	binary.LittleEndian.PutUint32(mapBuf[4:8], wid)
	_, err = conn.Write(mapBuf)
	if err != nil {
		return nil, err
	}

	// 5. Sync: Send GetInputFocus (Opcode 43) and wait for reply to ensure window is ready
	syncBuf := make([]byte, 4)
	syncBuf[0] = 43 // GetInputFocus
	binary.LittleEndian.PutUint16(syncBuf[2:4], 1)
	_, err = conn.Write(syncBuf)
	if err != nil {
		return nil, err
	}

	reply := make([]byte, 32)
	_, err = conn.Read(reply)
	if err != nil {
		return nil, err
	}

	return win, nil
}

func (w *X11Window) Draw(img *image.RGBA) error {
	// PutImage (Opcode 72)
	// uint8 opcode, uint8 format, uint16 length, uint32 drawable, uint32 gc,
	// uint16 width, uint16 height, int16 dst_x, int16 dst_y, uint8 left_pad, uint8 depth,
	// data...

	bounds := img.Bounds()
	width := uint16(bounds.Dx())
	height := uint16(bounds.Dy())

	// Limit to window size if needed
	if width > w.width {
		width = w.width
	}
	if height > w.height {
		height = w.height
	}

	pix := img.Pix
	// Converter RGBA to BGRX (standard X11 32-bit format)
	bgrx := make([]byte, int(width)*int(height)*4)
	for i := 0; i < int(width)*int(height); i++ {
		r, g, b := pix[i*4], pix[i*4+1], pix[i*4+2]
		bgrx[i*4] = b
		bgrx[i*4+1] = g
		bgrx[i*4+2] = r
		bgrx[i*4+3] = 0 // padding
	}

	headerLen := 6 // words (24 bytes)
	dataLen := (len(bgrx) + 3) / 4

	header := make([]byte, 24)
	header[0] = 72 // PutImage
	header[1] = 2  // ZPixmap
	binary.LittleEndian.PutUint16(header[2:4], uint16(headerLen+dataLen))
	binary.LittleEndian.PutUint32(header[4:8], w.wid)
	binary.LittleEndian.PutUint32(header[8:12], w.gcid)
	binary.LittleEndian.PutUint16(header[12:14], width)
	binary.LittleEndian.PutUint16(header[14:16], height)
	binary.LittleEndian.PutUint16(header[16:18], 0) // dst_x
	binary.LittleEndian.PutUint16(header[18:20], 0) // dst_y
	header[20] = 0                                  // left_pad
	header[21] = 24                                 // depth

	_, err := w.conn.Write(header)
	if err != nil {
		return err
	}
	_, err = w.conn.Write(bgrx)
	return err
}

func (w *X11Window) Close() {
	if w.conn != nil {
		w.conn.Close()
	}
}
