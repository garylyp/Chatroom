package main

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"golang.org/x/net/websocket"
)

var (
	cmd   *exec.Cmd
	name1 = "c1"
	// time.Date(year, month, day, hour, min, sec, nsec)
	t1 = time.Date(2020, 1, 1, 1, 1, 1, 1, time.UTC)
	t2 = time.Date(2020, 1, 1, 2, 2, 2, 2, time.UTC)
	t3 = time.Date(2020, 1, 1, 3, 3, 3, 3, time.UTC)
	t4 = time.Date(2020, 1, 1, 4, 4, 4, 4, time.UTC)
	t5 = time.Date(2020, 1, 1, 5, 5, 5, 5, time.UTC)
	m  = []Message{
		{
			Text:      "m1",
			Timestamp: t1,
			SenderID:  name1,
			Code:      "OK",
		},
		{
			Text:      "m2",
			Timestamp: t2,
			SenderID:  name1,
			Code:      "OK",
		},
		{
			Text:      "m3",
			Timestamp: t3,
			SenderID:  name1,
			Code:      "OK",
		},
		{
			Text:      "m4",
			Timestamp: t4,
			SenderID:  name1,
			Code:      "OK",
		},
		{
			Text:      "m5",
			Timestamp: t5,
			SenderID:  name1,
			Code:      "OK",
		},
	}
)

func setup() {
	cmd = exec.Command("../run_server.exe")
	cmd.Start()
	time.Sleep(time.Second) // Allow server to setup
}

func teardown() {
	cmd.Process.Kill()
}

func TestSendAndReceive(t *testing.T) {

	setup()

	client1, err := connect("9000")
	if err != nil {
		fmt.Println("Error during connection:", err.Error())
		cmd.Process.Kill()
		return
	}
	defer client1.Close()

	var currMsg Message
	// Receive initial ack from server
	websocket.JSON.Receive(client1, &currMsg)

	// Send test message to server
	sendMsg(client1, m[0].Text, m[0].Timestamp, m[0].SenderID, m[0].Code)
	// Receive broadcast message from server
	websocket.JSON.Receive(client1, &currMsg)

	if !cmp.Equal(currMsg, m[0]) {
		t.Errorf("Expected %v but received %v", m[0], currMsg)
	}

	teardown()
}

func TestMessageOrder(t *testing.T) {

	setup()

	client1, err := connect("9000")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer client1.Close()

	client2, err := connect("9000")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer client2.Close()
	time.Sleep(time.Microsecond)

	var currMsg Message
	// Receive initial ack from server
	websocket.JSON.Receive(client2, &currMsg)

	// Client1 send test messages to server in order
	for i := 0; i < 3; i++ {
		// Added sleep to simulate time difference between consecutive
		// messages from same client
		time.Sleep(400 * time.Microsecond)
		sendMsg(client1, m[i].Text, m[i].Timestamp, m[i].SenderID, m[i].Code)
	}

	// client2 receive broadcast messages from server
	var expM = m[:3]
	for i := 0; i < 3; i++ {
		websocket.JSON.Receive(client2, &currMsg)
		if !cmp.Equal(currMsg, expM[i]) {
			t.Errorf("\nExpected\t%v\nReceived\t%v\n", expM[i], currMsg)
		}
	}
	teardown()
}

// Test that new user can see last N messages sent by other clients
func TestPersistency(t *testing.T) {

	setup()

	client1, err := connect("9000")
	if err != nil {
		fmt.Println(err.Error())
	}

	// client1 sends test messages to server in order
	for i := 0; i < 5; i++ {
		time.Sleep(400 * time.Microsecond)
		sendMsg(client1, m[i].Text, m[i].Timestamp, m[i].SenderID, m[i].Code)
	}
	client1.Close()

	// client2 connects after client1 finished sending
	client2, err := connect("9000")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer client2.Close()
	time.Sleep(time.Microsecond)

	var currMsg Message
	// Receive initial ack from server
	websocket.JSON.Receive(client2, &currMsg)
	// client2 receive broadcast messages from server
	var expM = m[:5]
	for i := 0; i < 5; i++ {
		websocket.JSON.Receive(client2, &currMsg)
		if !cmp.Equal(currMsg, expM[i]) {
			t.Errorf("\nExpected\t%v\nReceived\t%v\n", expM[i], currMsg)
		}
	}
	teardown()
}
