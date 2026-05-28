package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	// Define constants that wont change after compilation
	const (
		// Read Deadline for 5 seconds
		deadline = 15 * time.Second
		// HeartBeat Interval
		interval = 8 * time.Second
		// Timeout to connect to the server
		timeout_duration = 4 * time.Second
		server_ip        = "127.0.0.1"
		server_port      = 8080
	)

	fmt.Println("Trying to connect to the server")
	// try and wait for 3 seconds to connect to remote server
	c, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%v", server_ip, server_port), 3*time.Second)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			fmt.Println("Connection time out")
			return
		} else {
			fmt.Printf("Error: %v", netErr)
			return
		}
	}
	defer c.Close()
	fmt.Println("Connected to the server")
	// Make a channel to stop go routines when the server no longer responds within alloted time
	con_dead := make(chan struct{})
	// This go routine will keep checking for server messages and print it
	go func() {
		scanner := bufio.NewScanner(c)
		for {
			// extend the deadline before every read
			c.SetReadDeadline(time.Now().Add(deadline))
			// IF scanner read method runs in a timeout error or EOF than scanner.Scan will be false
			if !scanner.Scan() {
				err := scanner.Err()
				if err != nil {
					var netErr net.Error
					// Check if timeout error
					if errors.As(err, &netErr) && netErr.Timeout() {
						fmt.Println("\n[heartbeat] No response from server — connection dead")
					} else {
						// If not timeout, than a read error
						fmt.Printf("\n[read] Error: %v\n", err)
					}
				} else {
					// If no error but the scanner.Read false or empty, means the server closed connection
					fmt.Println("Server closed the connection")
				}
				// Close the channel so that other goroutines dependent on con_dead will also exit
				close(con_dead)
				// exit routine
				return
			}
			message := scanner.Text()
			if message == "PONG" {
				fmt.Println("[HEARTBEAT] - Server Alive")
			} else {
				fmt.Printf("Server message: %v", message)
			}

		}
	}()
	// This goroutine will periodically send PING messages to the server when the intervial fires
	go func() {
		ticker := time.NewTicker(interval)
		// Cleanup step where the ticker will be discarded at the end of the routine
		defer ticker.Stop()

		for {
			select {
			// Simply check if the ticker fired, no value storing, if so send ping to server
			case <-ticker.C:
				_, err := fmt.Fprintln(c, "PING")
				if err != nil {
					fmt.Println("Error sending the heartbeat")
				}
				// If the channel con_dead is closed, stop sending the heartbeat by exiting the routine
			case <-con_dead:
				fmt.Println("Connection dead, type anything and press enter to exit")
				return
			}
		}
	}()
	// Use a scanner to accept the user input via the standard input
	uscan := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter your message, type exit to disconnect from server")
	// Keep looping to send messages to the server
	for uscan.Scan() {
		// Check if the con_dead is closed meaning connection is dead and client script must notify and exit
		select {
		case <-con_dead:
			return
		default:
		}
		utext := uscan.Text()

		if utext == "PING" || utext == "PONG" {
			fmt.Println("[!] Reserved keyword")
			continue
		} else if utext == "exit" {
			_, err := fmt.Fprintln(c, utext)
			if err != nil {
				fmt.Println("Error sending message")
			}
			return
		}
		_, er := fmt.Fprintln(c, utext)
		if er != nil {
			fmt.Println("Error sending message")
		}
		fmt.Println("Enter your message")
	}
	e := uscan.Err()
	if e != nil {
		fmt.Println("Error in user input")
	}

}
