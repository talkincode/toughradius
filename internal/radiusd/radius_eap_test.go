package radiusd

import (
	"fmt"
	"testing"

	"layeh.com/radius"
	"layeh.com/radius/rfc2869"
)

func TestEAPMSCHAPv2Challeng(t *testing.T) {
	// Create a new EAP-MSCHAPv2 Challenge packet
	eapChallenge := NewEAPMSCHAPv2Challenge(0x01, "RADIUS Server")

	// Serialize the packet to bytes
	packetBytes := eapChallenge.Serialize()

	// Print the serialized packet in hexadecimal
	fmt.Printf("EAP-MSCHAPv2 Challenge Packet: %x\n", packetBytes)
}

func TestEAPMSCHAPv2SuccessFailure(t *testing.T) {
	// Example usage of creating an EAP-MSCHAPv2 Success packet
	successMessage := "S=0123456789ABCDEF0123456789ABCDEF"
	eapSuccess := NewEAPMSCHAPv2SuccessFailure(EAPCodeRequest, 0x01, MSCHAPv2Success, successMessage)

	// Serialize the Success packet to bytes
	successPacketBytes := eapSuccess.Serialize()

	// Print the serialized Success packet in hexadecimal
	fmt.Printf("EAP-MSCHAPv2 Success Packet: %x\n", successPacketBytes)

	// Example usage of creating an EAP-MSCHAPv2 Failure packet
	failureMessage := "E=691 R=0 C=0123456789ABCDEF V=3 M=Authentication failed"
	eapFailure := NewEAPMSCHAPv2SuccessFailure(EAPCodeRequest, 0x01, MSCHAPv2Failure, failureMessage)

	// Serialize the Failure packet to bytes
	failurePacketBytes := eapFailure.Serialize()

	// Print the serialized Failure packet in hexadecimal
	fmt.Printf("EAP-MSCHAPv2 Failure Packet: %x\n", failurePacketBytes)
}

func TestEAPMSCHAPv2Response(t *testing.T) {
	// Example usage of creating and parsing an EAP-MSCHAPv2 Response packet
	peerChallenge := make([]byte, 16) // This should be generated or provided
	response := make([]byte, 24)      // This should be calculated based on the challenge and user's password
	name := "User"

	// Create a new EAP-MSCHAPv2 Response packet
	eapResponse := NewEAPMSCHAPv2Response(0x01, peerChallenge, response, name)

	// Serialize the packet to bytes
	packetBytes := eapResponse.Serialize()
	packet := radius.New(radius.CodeAccessRequest, []byte("secret"))
	rfc2869.EAPMessage_Set(packet, packetBytes)

	// Print the serialized packet in hexadecimal
	fmt.Printf("EAP-MSCHAPv2 Response Packet: %x\n", packetBytes)

	// Parse the serialized packet back into a struct
	parsedEapResponse, err := ParseEAPMSCHAPv2Response(packet)
	if err != nil {
		fmt.Println("Error parsing EAP-MSCHAPv2 Response:", err)
		return
	}

	// Print the parsed packet
	fmt.Printf("Parsed EAP-MSCHAPv2 Response: %+v\n", parsedEapResponse)
}
