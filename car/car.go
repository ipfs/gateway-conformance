package car

import (
	"fmt"
	"testing"
)

func GetCid(t *testing.T, carPath string, nodePath string) string {
	t.Helper()
	nodes, err := extractCar(carPath)
	if err != nil {
		t.Fatal(err)
	}
	node := findNode(nodes, nodePath)
	if node == nil {
		t.Fatal(fmt.Errorf("node not found: %s", nodePath))
	}
	return node.Cid.String()
}

func GetRawBlock(t *testing.T, carPath string, nodePath string) []byte {
	t.Helper()
	nodes, err := extractCar(carPath)
	if err != nil {
		t.Fatal(err)
	}
	node := findNode(nodes, nodePath)
	if node == nil {
		t.Fatal(fmt.Errorf("node not found: %s", nodePath))
	}
	return node.Raw
}
