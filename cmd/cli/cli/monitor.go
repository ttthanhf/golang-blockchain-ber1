package cli

import (
	"context"
	"fmt"
	"strings"
)

func printMonitorNodes() {
	fmt.Print("\033[?25l")     // Hide Cursor
	fmt.Print("\033[H\033[2J") // Clear console and cursor on top
	fmt.Print("\r")

	for _, node := range nodeOrder {
		nodeId := strings.Split(node, ":")[0]

		mu.Lock()
		value, ok := infoNodes[nodeId]
		mu.Unlock()

		if ok {
			fmt.Printf("%s: %s\n", nodeId, value)
		}
	}

}

func monitorNodesHandle(address string) {
	printMonitorNodes()

	client, err := GetClient(address)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	stream, err := client.StreamNodeInfo(context.Background(), nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	for {
		res, err := stream.Recv()
		if err != nil {
			mu.Lock()
			infoNodes[addressNodeMap[address]] = "DISCONNECT"
			mu.Unlock()

			printMonitorNodes()

			return
		}
		if res != nil {
			mu.Lock()
			infoNodes[res.NodeId] = res.NodeStatus
			mu.Unlock()

			printMonitorNodes()
		}
	}
}

func MonitorNodesCLI() {
	for i := range nodes {
		mu.Lock()
		addressNodeMap[nodes[i]] = nodeOrder[i]
		mu.Unlock()
	}
	for i := range nodeOrder {
		mu.Lock()
		infoNodes[nodeOrder[i]] = "DISCONNECT"
		mu.Unlock()
	}

	for _, address := range nodes {
		go monitorNodesHandle(address)
	}

	select {}
}
