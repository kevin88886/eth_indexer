package protocol

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/shopspring/decimal"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

type Signature struct {
	Title  string `json:"title"`
	Signer string `json:"-"`
	To     string `json:"to"`
	Tick   string `json:"tick"`
	Amt    string `json:"amt"`
	Value  string `json:"value"`
	Nonce  string `json:"nonce"`
}

func NewSignature(tick, signer, to, amount, value, nonce string) *Signature {
	return &Signature{
		Title:  SignatureTitle,
		Signer: signer,
		To:     to,
		Tick:   tick,
		Amt:    amount,
		Value:  value,
		Nonce:  nonce,
	}
}

func (s *Signature) ValidSignature(signature string) error {
	if len(signature) == 0 || !strings.HasPrefix(strings.ToLower(signature), "0x") {
		return NewProtocolError(InvalidSignature, "invalid sign format")
	}
	sig, err := hex.DecodeString(strings.TrimPrefix(signature, "0x"))
	if err != nil {
		return NewProtocolError(InvalidSignature, "ValidateEOASignature, signature is an invalid hex string")
	}
	if len(sig) != 65 {
		return NewProtocolError(InvalidSignature, "ValidateEOASignature, signature is not of proper length")
	}
	if sig[64] > 1 {
		sig[64] -= 27 // recovery ID
	}

	message, _ := json.MarshalIndent(s, "", "    ")
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v%s", len(message), message)))

	pubKey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return NewProtocolError(InvalidSignature, err.Error())
	}

	key := crypto.PubkeyToAddress(*pubKey)
	if strings.ToLower(key.Hex()) != s.Signer {
		return NewProtocolError(SignatureNotMatch, "signature not match")
	}

	return nil
}

func (s *Signature) FreezeValidSignatureV4(rec *FreezeRecordV4) error {
	if len(rec.SellerSign) == 0 || !strings.HasPrefix(strings.ToLower(rec.SellerSign), "0x") {
		return NewProtocolError(InvalidSignature, "invalid sign format")
	}
	sig, err := hex.DecodeString(strings.TrimPrefix(rec.SellerSign, "0x"))
	if err != nil {
		return NewProtocolError(InvalidSignature, "ValidateEOASignature, signature is an invalid hex string")
	}
	if len(sig) != 65 {
		return NewProtocolError(InvalidSignature, "ValidateEOASignature, signature is not of proper length")
	}
	typedData := s.formatFreezeTypedData(rec)
	jsonData, _ := json.Marshal(typedData)
	fmt.Println(string(jsonData))
	LogOutput(string(jsonData))
	address, err := verifyAuthTokenAddress(typedData, rec.SellerSign)
	if err != nil {
		return err
	}
	if strings.ToLower(address) != rec.Seller {
		return NewProtocolError(InvalidSignature, "signature not match")
	}

	return nil
}

func (s *Signature) ProxyTransferValidSignatureV4(rec *ProxyTransferRecordV4) error {
	if len(rec.Sign) == 0 || !strings.HasPrefix(strings.ToLower(rec.Sign), "0x") {
		return NewProtocolError(InvalidSignature, "invalid sign format")
	}
	sig, err := hex.DecodeString(strings.TrimPrefix(rec.Sign, "0x"))
	if err != nil {
		return NewProtocolError(InvalidSignature, "ValidateEOASignature, signature is an invalid hex string")
	}
	if len(sig) != 65 {
		return NewProtocolError(InvalidSignature, "ValidateEOASignature, signature is not of proper length")
	}
	typedData := s.formatProxyTransferTypedData(rec)
	jsonData, _ := json.Marshal(typedData)
	fmt.Println(string(jsonData))
	LogOutput(string(jsonData))
	address, err := verifyAuthTokenAddress(typedData, rec.Sign)
	LogOutput(strings.ToLower(address))
	if err != nil {
		return err
	}
	if strings.ToLower(address) != s.Signer {
		return NewProtocolError(InvalidSignature, "signature not match")
	}

	return nil
}

func LogOutput(str_content string) {
	fd, _ := os.OpenFile("./logOutput.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	fd_time := time.Now().Format("2006-01-02 15:04:05")
	fd_content := strings.Join([]string{"======", fd_time, "=====\n", str_content, "\n"}, "")
	buf := []byte(fd_content)
	fd.Write(buf)
	fd.Close()
}

func (s *Signature) formatFreezeTypedData(rec *FreezeRecordV4) apitypes.TypedData {
	typedDataDomain := apitypes.TypedDataDomain{
		Name:              CreateOrderSignatureTitle,
		Version:           "1",
		ChainId:           math.NewHexOrDecimal256(1),
		VerifyingContract: "0x33302dbff493ed81ba2e7e35e2e8e833db023333",
	}
	types := make(apitypes.Types)
	types["EIP712Domain"] = []apitypes.Type{
		{"name", "string"},
		{"version", "string"},
		{"chainId", "uint256"},
		{"verifyingContract", "address"},
	}
	types["Transaction"] = []apitypes.Type{
		{"title", "string"},
		{"version", "string"},
		{"to", "address"},
		{"tick", "string"},
		{"amt", "string"},
		{"value", "string"},
		{"nonce", "string"},
		{"expire", "string"},
	}
	if !decimal.Decimal.IsZero(rec.Payment.Value) {
		types["Transaction"] = append(types["Transaction"], apitypes.Type{Name: "payment", Type: "Payment"})
		types["Payment"] = []apitypes.Type{
			{"tick", "string"},
			{"value", "string"},
			{"fee", "string"},
		}
	}
	msg := make(apitypes.TypedDataMessage)
	msg["title"] = CreateOrderSignatureTitle
	msg["version"] = rec.Version
	msg["to"] = rec.Platform
	msg["tick"] = rec.Tick
	msg["amt"] = rec.Amount
	msg["value"] = rec.Value
	msg["nonce"] = s.Nonce
	msg["expire"] = rec.Expire
	if !decimal.Decimal.IsZero(rec.Payment.Value) {
		msg["payment"] = rec.Payment
	}

	AuthData := apitypes.TypedData{
		Types:       types,
		PrimaryType: "Transaction",
		Domain:      typedDataDomain,
		Message:     msg,
	}
	return AuthData
}

func (s *Signature) formatProxyTransferTypedData(rec *ProxyTransferRecordV4) apitypes.TypedData {
	typedDataDomain := apitypes.TypedDataDomain{
		Name:              CreateOrderSignatureTitle,
		Version:           "1",
		ChainId:           math.NewHexOrDecimal256(1),
		VerifyingContract: "0x33302dbff493ed81ba2e7e35e2e8e833db023333",
	}
	types := make(apitypes.Types)
	types["EIP712Domain"] = []apitypes.Type{
		{"name", "string"},
		{"version", "string"},
		{"chainId", "uint256"},
		{"verifyingContract", "address"},
	}
	types["Transaction"] = []apitypes.Type{
		{"title", "string"},
		{"version", "string"},
		{"to", "address"},
		{"tick", "string"},
		{"amt", "string"},
		{"value", "string"},
		{"nonce", "string"},
		{"expire", "string"},
	}
	if !decimal.Decimal.IsZero(rec.Payment.Value) {
		types["Transaction"] = append(types["Transaction"], apitypes.Type{Name: "payment", Type: "Payment"})
		types["Payment"] = []apitypes.Type{
			{"tick", "string"},
			{"value", "string"},
			{"fee", "string"},
		}
	}
	msg := make(apitypes.TypedDataMessage)
	msg["title"] = CreateOrderSignatureTitle
	msg["version"] = rec.Version
	msg["to"] = s.To
	msg["tick"] = rec.Tick
	msg["amt"] = rec.Amount
	msg["value"] = rec.Value
	msg["nonce"] = s.Nonce
	msg["expire"] = rec.Expire
	if !decimal.Decimal.IsZero(rec.Payment.Value) {
		msg["payment"] = rec.Payment
	}

	AuthData := apitypes.TypedData{
		Types:       types,
		PrimaryType: "Transaction",
		Domain:      typedDataDomain,
		Message:     msg,
	}
	return AuthData
}

func verifyAuthTokenAddress(authToken apitypes.TypedData, sign string) (string, error) {
	signature, err := hexutil.Decode(sign)
	if err != nil {
		return "", fmt.Errorf("decode signature: %w", err)
	}

	typedDataBytes, _ := json.MarshalIndent(authToken, "", "    ")
	typedData := apitypes.TypedData{}
	if err := json.Unmarshal(typedDataBytes, &typedData); err != nil {
		return "", fmt.Errorf("unmarshal typed data: %w", err)
	}

	// EIP-712 typed data marshalling
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return "", fmt.Errorf("eip712domain hash struct: %w", err)
	}
	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return "", fmt.Errorf("primary type hash struct: %w", err)
	}

	// add magic string prefix
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	sighash := crypto.Keccak256(rawData)

	signature[64] -= 27

	// get the pubkey used to sign this signature
	sigPubkey, err := crypto.Ecrecover(sighash, signature)
	if err != nil {
		return "", fmt.Errorf("ecrecover: %w", err)
	}

	// get the address to confirm it's the same one in the auth token
	pubkey, err := crypto.UnmarshalPubkey(sigPubkey)
	if err != nil {
		return "", fmt.Errorf("unmarshal Pubkey: %w", err)
	}
	address := crypto.PubkeyToAddress(*pubkey)

	return address.Hex(), nil
}
