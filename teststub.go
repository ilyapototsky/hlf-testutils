package hlf_testutils

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"math/big"
	"fmt"
	"crypto/x509"
	"crypto/x509/pkix"
	"time"
	"crypto/rsa"
	"encoding/pem"
	"crypto/rand"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("TestStub")

type TestStub struct {
	*shim.MockStub

	args [][]byte

	cc shim.Chaincode

	caller string
}

func (stub *TestStub) GetArgs() [][]byte {
	return stub.args
}

func (stub *TestStub) GetStringArgs() []string {
	args := stub.GetArgs()
	stringArgs := make([]string, 0, len(args))
	for _, arg := range args {
		stringArgs = append(stringArgs, string(arg))
	}
	return stringArgs
}

func (stub *TestStub) GetFunctionAndParameters() (function string, params []string) {
	args := stub.GetStringArgs()
	function = ""
	params = []string{}
	if len(args) >= 1 {
		function = args[0]
		params = args[1:]
	}
	return
}

func (stub *TestStub) SetCaller(org string) {
	stub.caller = org
}

// Implemented to have a possibility to test privileges
func (stub *TestStub) GetCreator() ([]byte, error) {
	org := stub.caller

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fmt.Printf("Failed to generate serial number: %s", err)
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{org},
		},
		Issuer: pkix.Name{
			Organization: []string{org},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Failed to generate private key: %s", err)
		return nil, err
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		fmt.Printf("Failed to create certificate: %s", err)
		return nil, err
	}

	result := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	return result, nil
}

// Initialise this chaincode,  also starts and ends a transaction.
// NOTE: you should set caller (if it matters) before using this function
func (stub *TestStub) MockInit(uuid string, args [][]byte) pb.Response {
	stub.args = args
	stub.MockTransactionStart(uuid)
	res := stub.cc.Init(stub)
	stub.MockTransactionEnd(uuid)
	return res
}

// Invoke this chaincode, also starts and ends a transaction.
// NOTE: you should set caller (if it matters) before using this function
func (stub *TestStub) MockInvoke(uuid string, args [][]byte) pb.Response {
	stub.args = args
	stub.MockTransactionStart(uuid)
	res := stub.cc.Invoke(stub)
	stub.MockTransactionEnd(uuid)
	return res
}

func NewTestStub(name string, cc shim.Chaincode) *TestStub {
	ts := &TestStub{MockStub: shim.NewMockStub(name, cc), cc: cc}
	return ts
}