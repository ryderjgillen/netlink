package netlink

import (
	"bytes"
	"reflect"
	"testing"
	"golang.org/x/sys/unix"
	"github.com/kckeiks/netlink/internal/testutils"
)

func TestNewSerializedNetlinkMessage(t *testing.T) {
	// Given: a NlMsghdr header
	h := testutils.NewTestNlMsghdr()
	// Given: length of nl msg with 4 more bytes of space 
	h.Len = 16 + 4

	// When: we serialize the header and the data
	serializedData := NewSerializedNetlinkMessage(h)

	// Then: the message was serialized with the correct data
	if h.Len != testutils.TestByteOrder.Uint32(serializedData[:4]) {
		t.Fatalf("NlMsghdr.Length = %d, expected %d", testutils.TestByteOrder.Uint32(serializedData[:4]), h.Len)
	}
	if h.Type != testutils.TestByteOrder.Uint16(serializedData[4:6]) {
		t.Fatalf("NlMsghdr.Type = %d, expected %d", testutils.TestByteOrder.Uint16(serializedData[4:6]), h.Type)
	}
	if h.Flags != testutils.TestByteOrder.Uint16(serializedData[6:8]) {
		t.Fatalf("NlMsghdr.Flags = %d, expected %d", testutils.TestByteOrder.Uint16(serializedData[6:8]), h.Flags)
	}
	if h.Seq != testutils.TestByteOrder.Uint32(serializedData[8:12]) {
		t.Fatalf("NlMsghdr.Seq = %d, expected %d", testutils.TestByteOrder.Uint32(serializedData[8:12]), h.Seq)
	}
	if h.Pid != testutils.TestByteOrder.Uint32(serializedData[12:16]) {
		t.Fatalf("NlMsghdr.Pid = %d, expected %d", testutils.TestByteOrder.Uint32(serializedData[12:16]), h.Pid)
	}
	if len(serializedData[16:]) != 4 {
		t.Fatalf("Len(serializedData) = %d, expected %d", len(serializedData[16:]), 4)
	}
}

func TestDeserializeNetlinkMsg(t *testing.T) {
	// Given: a serialized netlink message
	h := testutils.NewTestNlMsghdr()
	data := [4]byte{0xFF, 0xFF, 0xFF, 0xFF}
	h.Len = h.Len + uint32(len(data))
	serializedData := testutils.NewTestSerializedNetlinkMsg(h, data[:])
	
	// When: we deserialize the message
	nlmsg := DeserializeNetlinkMsg(serializedData)

	// Then: the struct that we get has the same values as the initial struct
	if !reflect.DeepEqual(nlmsg.Header, h) {
		t.Fatalf("Given NlMsghdr %+v and deserialized is %+v,", nlmsg.Header, h)
	}
	// Then: the extra data was returned
	if bytes.Compare(nlmsg.Data, data[:]) != 0 {
		t.Fatalf("Extra data=%d, expected %d", nlmsg.Data, data)
	}
}

func TestDeserializeNetlinkMsgWithOutData(t *testing.T) {
	// Given: a serialized netlink message without extra data
	h := testutils.NewTestNlMsghdr()
	data := []byte{} 
	h.Len = uint32(unix.SizeofNlMsghdr)
	serializedData := testutils.NewTestSerializedNetlinkMsg(h, data)
	
	// When: we deserialize the message
	nlmsg := DeserializeNetlinkMsg(serializedData)

	// Then: empty slice is returned for payload
	if len(nlmsg.Data) != 0 {
		t.Fatalf("Extra data=%d, expected [].", nlmsg.Data)
	}
}

func TestDeserializeNetlinkMsgBadLen(t *testing.T) {
	defer func() {
        if r := recover(); r == nil {
            t.Errorf("Error: did not panic.")
        }
    }()
	// Given: a serialized nl message with extra data but 
	// we do not update length in header
	h := testutils.NewTestNlMsghdr()
	data := [4]byte{0xFF, 0xFF, 0xFF, 0xFF}
	serializedData := testutils.NewTestSerializedNetlinkMsg(h, data[:])
	// When: we deserialize the message
	// Then: we panic
	DeserializeNetlinkMsg(serializedData)
}

func TestParseNetlinkMessage(t *testing.T) {
	// Given: a list of serialized netlink messages
	h1 := testutils.NewTestNlMsghdr()
	data1 := [4]byte{0xFF, 0xFF, 0xFF, 0xFF}
	h1.Len = h1.Len + uint32(len(data1))
	h2 := testutils.NewTestNlMsghdr()
	data2 := [6]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x01, 0x01}
	h2.Len = h2.Len + uint32(len(data2))

	var nlmsgs []byte
	nlmsgs = append(nlmsgs, testutils.NewTestSerializedNetlinkMsg(h1, data1[:])...)
	nlmsgs = append(nlmsgs, testutils.NewTestSerializedNetlinkMsg(h2, data2[:])...)

	// When: parse these serialized data
	result := ParseNetlinkMessage(nlmsgs)

	// Then: We get the messages as expected
	expectedNlMsg1 := NetlinkMessage{Header: h1, Data: data1[:]}
	expectedNlMsg2 := NetlinkMessage{Header: h2, Data: data2[:]}

	if !reflect.DeepEqual(result[0], expectedNlMsg1) {
		t.Fatalf("Given first Netlink Msg %+v but received %+v,", result[0], expectedNlMsg1)
	}
	if !reflect.DeepEqual(result[1], expectedNlMsg2) {
		t.Fatalf("Given second Netlink Msg %+v but received %+v,", result[1], expectedNlMsg2)
	}
}