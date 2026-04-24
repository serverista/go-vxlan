package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/vishvananda/netlink"
)

func main() {
	vxlanIDStr := os.Getenv("VXLAN_ID")
	localIPStr := os.Getenv("LOCAL_IP")
	remoteIPStr := os.Getenv("REMOTE_IP")
	vxlanIPStr := os.Getenv("VXLAN_IP")

	if vxlanIDStr == "" || localIPStr == "" || remoteIPStr == "" || vxlanIPStr == "" {
		log.Fatal("All environment variables VXLAN_ID, LOCAL_IP, REMOTE_IP, VXLAN_IP are required")
	}

	vxlanID, err := strconv.Atoi(vxlanIDStr)
	if err != nil {
		log.Fatalf("Invalid VXLAN_ID: %v", err)
	}

	localIP := net.ParseIP(localIPStr)
	if localIP == nil {
		log.Fatalf("Invalid LOCAL_IP: %s", localIPStr)
	}

	remoteIP := net.ParseIP(remoteIPStr)
	if remoteIP == nil {
		log.Fatalf("Invalid REMOTE_IP: %s", remoteIPStr)
	}

	vxlanIPAddr, err := netlink.ParseAddr(vxlanIPStr)
	if err != nil {
		log.Fatalf("Invalid VXLAN_IP form (expected IP/mask): %v", err)
	}

	// Find device index of the local interface that owns localIP
	var vtepDevIndex int
	links, err := netlink.LinkList()
	if err != nil {
		log.Fatalf("Failed to list interfaces: %v", err)
	}
	
	for _, link := range links {
		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if addr.IP.Equal(localIP) {
				vtepDevIndex = link.Attrs().Index
				log.Printf("Found local interface %s with index %d", link.Attrs().Name, vtepDevIndex)
				break
			}
		}
		if vtepDevIndex != 0 {
			break
		}
	}

	if vtepDevIndex == 0 {
		log.Fatalf("Could not find physical interface for LOCAL_IP: %s", localIPStr)
	}

	vxlanName := fmt.Sprintf("vxlan%d", vxlanID)

	// Build the VXLAN link configuration
	vxlan := &netlink.Vxlan{
		LinkAttrs: netlink.LinkAttrs{
			Name: vxlanName,
		},
		VxlanId:      vxlanID,
		VtepDevIndex: vtepDevIndex,
		SrcAddr:      localIP,
		Group:        remoteIP, 
		Port:         4789,
	}

	// Create the link
	if err := netlink.LinkAdd(vxlan); err != nil {
		if os.IsExist(err) {
			log.Printf("Interface %s already exists, ignoring creation", vxlanName)
			link, linkErr := netlink.LinkByName(vxlanName)
			if linkErr != nil {
				log.Fatalf("Failed to get existing link %s: %v", vxlanName, linkErr)
			}
			vxlan = link.(*netlink.Vxlan)
		} else {
			log.Fatalf("Failed to create VXLAN interface: %v", err)
		}
	} else {
		log.Printf("Created VXLAN interface %s", vxlanName)
	}

	// Assign the IP address
	if err := netlink.AddrAdd(vxlan, vxlanIPAddr); err != nil {
		if !os.IsExist(err) {
			log.Fatalf("Failed to assign IP %s to %s: %v", vxlanIPStr, vxlanName, err)
		}
	}
	log.Printf("Assigned IP %s to %s", vxlanIPStr, vxlanName)

	// Bring the interface up
	if err := netlink.LinkSetUp(vxlan); err != nil {
		log.Fatalf("Failed to set interface %s up: %v", vxlanName, err)
	}
	log.Printf("Interface %s is up", vxlanName)

	log.Printf("VXLAN successfully configured. Waiting indefinitely...")
	// Block forever
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down")
}
