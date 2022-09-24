package d7024e

import "testing"

func TestEncodeDecode(t *testing.T) {
	test_msg := msg{Ping, "PAYLOAD"}
	expected_msg := msg{Ping, "PAYLOAD"}
	// Encode msg.
	msg_bin, _ := encodeMsg(test_msg)
	// Decode msg.
	returned_msg, _ := decodeMsg(msg_bin)
	if returned_msg != expected_msg {
		t.Errorf("Encode and decode should be inverses")
	}
}
