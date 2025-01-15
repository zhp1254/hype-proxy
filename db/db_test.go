package db

import (
	"fmt"
	"testing"
)

func TestInit(t *testing.T)  {
	fmt.Println("ok")
}

func TestAddMainChain(t *testing.T)  {
	fmt.Println(AddMainChain(
		9069910,
		98626530,
		"0695d93f715dab9caa1c26fac68bc98246025b4498e54f6db6dfc0f32ed9b754",))
}

func TestFindMainChain(t *testing.T)  {
	fmt.Println(FindMainChain(
		9069910))
}

func TestFindLast(t *testing.T)  {
	fmt.Println(FindLastMainChain())
}