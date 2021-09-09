package netlink

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/sys/unix"
	"os"
)

var OSPageSize = os.Getpagesize()

type NetlinkMessage struct {
	Header unix.NlMsghdr
	Data []byte
}

func DeserializeNetlinkMsg(data []byte) NetlinkMessage {
	if len(data) < unix.SizeofNlMsghdr {
		panic("Error: Could not deserialize. Invalid length for serialized NlMsghdr.")
	}
	serializedData := bytes.NewBuffer(data[:unix.SizeofNlMsghdr])
	h := unix.NlMsghdr{}

	err := binary.Read(serializedData, byteOrder, &h)
	if err != nil {
		panic("Error: Could not deserialize NlMsghdr.")
	}

	return NetlinkMessage{Header: h, Data: data[unix.SizeofNlMsghdr:]}
}

func ParseNetlinkMessages(data []byte) []NetlinkMessage {
	var msgs []NetlinkMessage
	for len(data) > unix.NLMSG_HDRLEN {
		l := byteOrder.Uint32(data[:4])
		nlmsg := DeserializeNetlinkMsg(data[:l])
		msgs = append(msgs, nlmsg)
		data = data[l:]
	}
	return msgs
}