package dsmr5

func CRC16(data []byte) uint16 {
	crc := uint16(0)
	for _, b := range data {
		crc ^= uint16(b)
		for range 8 {
			if crc&0x0001 != 0 {
				crc = (crc >> 1) ^ 0xA001 // Reversed form of 0x8005
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}
