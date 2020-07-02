package net

import (
	"encoding/binary"
	"math/rand"
	"testing"
)

func TestLengthFieldBasedFrameCodecWith1(t *testing.T) {
	encoderConfig := EncoderConfig{
		ByteOrder:                       binary.BigEndian,
		LengthFieldLength:               1,
		LengthAdjustment:                0,
		LengthIncludesLengthFieldLength: false,
	}
	decoderConfig := DecoderConfig{
		ByteOrder:           binary.BigEndian,
		LengthFieldOffset:   0,
		LengthFieldLength:   1,
		LengthAdjustment:    0,
		InitialBytesToStrip: 1,
	}
	codec := NewLengthFieldBasedFrameCodec(encoderConfig, decoderConfig)

	sz := 256
	data := make([]byte, sz)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	if _, err := codec.Encode(nil, data); err == nil {
		panic("should have a error of exceeding bytes.")
	}

	sz = 255
	data = make([]byte, sz)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	out, _ := codec.Encode(nil, data)
	if string(out[1:]) != string(data) {
		t.Fatalf("data don't match with big endian, raw data: %s, encoded data: %s\n", string(data), string(out))
	}

	encoderConfig.ByteOrder = binary.LittleEndian
	decoderConfig.ByteOrder = binary.LittleEndian
	codec = NewLengthFieldBasedFrameCodec(encoderConfig, decoderConfig)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	out, _ = codec.Encode(nil, data)
	if string(out[1:]) != string(data) {
		t.Fatalf("data don't match with little endian, raw data: %s, encoded data: %s\n", string(data), string(out))
	}
}

func TestLengthFieldBasedFrameCodecWith2(t *testing.T) {
	encoderConfig := EncoderConfig{
		ByteOrder:                       binary.BigEndian,
		LengthFieldLength:               2,
		LengthAdjustment:                0,
		LengthIncludesLengthFieldLength: false,
	}
	decoderConfig := DecoderConfig{
		ByteOrder:           binary.BigEndian,
		LengthFieldOffset:   0,
		LengthFieldLength:   2,
		LengthAdjustment:    0,
		InitialBytesToStrip: 2,
	}
	codec := NewLengthFieldBasedFrameCodec(encoderConfig, decoderConfig)

	sz := 65536
	data := make([]byte, sz)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	if _, err := codec.Encode(nil, data); err == nil {
		panic("should have a error of exceeding bytes.")
	}

	sz = rand.Intn(10) * 1024
	data = make([]byte, sz)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	out, _ := codec.Encode(nil, data)
	if string(out[2:]) != string(data) {
		t.Fatalf("data don't match with big endian, raw data: %s, encoded data: %s\n", string(data), string(out))
	}

	encoderConfig.ByteOrder = binary.LittleEndian
	decoderConfig.ByteOrder = binary.LittleEndian
	codec = NewLengthFieldBasedFrameCodec(encoderConfig, decoderConfig)
	sz = rand.Intn(10) * 1024
	data = make([]byte, sz)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	out, _ = codec.Encode(nil, data)
	if string(out[2:]) != string(data) {
		t.Fatalf("data don't match with little endian, raw data: %s, encoded data: %s\n", string(data), string(out))
	}
}

func TestLengthFieldBasedFrameCodecWith3(t *testing.T) {
	encoderConfig := EncoderConfig{
		ByteOrder:                       binary.BigEndian,
		LengthFieldLength:               3,
		LengthAdjustment:                0,
		LengthIncludesLengthFieldLength: false,
	}
	decoderConfig := DecoderConfig{
		ByteOrder:           binary.BigEndian,
		LengthFieldOffset:   0,
		LengthFieldLength:   3,
		LengthAdjustment:    0,
		InitialBytesToStrip: 3,
	}
	codec := NewLengthFieldBasedFrameCodec(encoderConfig, decoderConfig)

	sz := 16777216
	data := make([]byte, sz)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	if _, err := codec.Encode(nil, data); err == nil {
		panic("should have a error of exceeding bytes.")
	}

	sz = rand.Intn(10) * 64 * 1024
	data = make([]byte, sz)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	out, _ := codec.Encode(nil, data)
	if string(out[3:]) != string(data) {
		t.Fatalf("data don't match with big endian, raw data: %s, encoded data: %s\n", string(data), string(out))
	}

	encoderConfig.ByteOrder = binary.LittleEndian
	decoderConfig.ByteOrder = binary.LittleEndian
	codec = NewLengthFieldBasedFrameCodec(encoderConfig, decoderConfig)
	sz = rand.Intn(10) * 64 * 1024
	data = make([]byte, sz)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	out, _ = codec.Encode(nil, data)
	if string(out[3:]) != string(data) {
		t.Fatalf("data don't match with little endian, raw data: %s, encoded data: %s\n", string(data), string(out))
	}

	buf := make([]byte, 3)
	rand.Read(buf)
	bNum := readUint24(binary.BigEndian, buf)
	p := writeUint24(binary.BigEndian, int(bNum))
	if string(buf) != string(p) {
		t.Fatalf("data don't match with big endian, raw data: %s, recovered data: %s\n", string(buf), string(p))
	}

	rand.Read(buf)
	bNum = readUint24(binary.LittleEndian, buf)
	p = writeUint24(binary.LittleEndian, int(bNum))
	if string(buf) != string(p) {
		t.Fatalf("data don't match with little endian, raw data: %s, recovered data: %s\n", string(buf), string(p))
	}
}

func TestLengthFieldBasedFrameCodecWith8(t *testing.T) {
	encoderConfig := EncoderConfig{
		ByteOrder:                       binary.BigEndian,
		LengthFieldLength:               8,
		LengthAdjustment:                0,
		LengthIncludesLengthFieldLength: false,
	}
	decoderConfig := DecoderConfig{
		ByteOrder:           binary.BigEndian,
		LengthFieldOffset:   0,
		LengthFieldLength:   8,
		LengthAdjustment:    0,
		InitialBytesToStrip: 8,
	}
	codec := NewLengthFieldBasedFrameCodec(encoderConfig, decoderConfig)
	sz := rand.Intn(10) * 1024 * 1024
	data := make([]byte, sz)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	out, _ := codec.Encode(nil, data)
	if string(out[8:]) != string(data) {
		t.Fatalf("data don't match with big endian, raw data: %s, encoded data: %s\n", string(data), string(out))
	}

	encoderConfig.ByteOrder = binary.LittleEndian
	decoderConfig.ByteOrder = binary.LittleEndian
	codec = NewLengthFieldBasedFrameCodec(encoderConfig, decoderConfig)
	sz = rand.Intn(10) * 1024 * 1024
	data = make([]byte, sz)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	out, _ = codec.Encode(nil, data)
	if string(out[8:]) != string(data) {
		t.Fatalf("data don't match with little endian, raw data: %s, encoded data: %s\n", string(data), string(out))
	}
}

func TestFixedLengthFrameCodec_Encode(t *testing.T) {
	codec := NewFixedLengthFrameCodec(8)
	if data, err := codec.Encode(nil, make([]byte, 15)); data != nil || err != errInvalidFixedLength {
		panic("should have a error of invalid fixed length")
	}
}

func TestInnerBufferReadN(t *testing.T) {
	var in innerBuffer
	data := make([]byte, 10)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	in = data
	if _, err := in.readN(-1); err == nil {
		t.Fatal("error missing")
	}
	if _, err := in.readN(11); err == nil {
		t.Fatal("error missing")
	}
	if _, err := in.readN(1); err != nil {
		t.Fatal("unexpected error")
	}
	if len(in) != 9 {
		t.Fatal("wrong length of leftover bytes")
	}
}
