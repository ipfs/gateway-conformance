package car

import (
	"fmt"
	"os"
)

func GetCid(carPath string, nodePath string) string {
	nodes, err := extractCar(carPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	node := findNode(nodes, nodePath)
	if node == nil {
		fmt.Println(fmt.Errorf("node not found: %s", nodePath))
		os.Exit(1)
	}
	return node.Cid.String()
}

func GetRawBlock(carPath string, nodePath string) []byte {
	nodes, err := extractCar(carPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	node := findNode(nodes, nodePath)
	if node == nil {
		fmt.Println(fmt.Errorf("node not found: %s", nodePath))
		os.Exit(1)
	}
	return node.Raw
}
