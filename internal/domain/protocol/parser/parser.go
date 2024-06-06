package parser

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/kevin88886/eth_indexer/internal/domain"
	"github.com/kevin88886/eth_indexer/internal/domain/protocol"
)

type Parser interface {
	// 检查格式
	CheckFormat(data []byte) error
	// 解析
	Parse(tx *domain.Transaction) (protocol.IERCTransaction, error)
}

type parser struct {
	header       string
	headerLength int
	parsers      map[protocol.Protocol]Parser
}

func (p *parser) CheckFormat(data []byte) error {
	if len(data) == 0 {
		return protocol.NewProtocolError(protocol.NotProtocolData, "not protocol data")
	}

	dataStr := string(data)
	// 不是协议数据
	if !utf8.ValidString(dataStr) || !strings.HasPrefix(dataStr, p.header) {
		return protocol.NewProtocolError(protocol.NotProtocolData, "not protocol data")
	}

	var base struct {
		Protocol string `json:"p"`  // 协议名称
		Operate  string `json:"op"` // 操作
	}

	err := json.Unmarshal([]byte(dataStr[len(p.header):]), &base)
	if err != nil {
		return protocol.NewProtocolError(protocol.InvalidProtocolFormat, "invalid protocol format")
	}

	if len(base.Protocol) == 0 {
		return protocol.NewProtocolError(protocol.UnknownProtocol, "unknown protocol")
	}

	return nil
}

func (p *parser) Parse(tx *domain.Transaction) (protocol.IERCTransaction, error) {

	var base struct {
		Protocol protocol.Protocol `json:"p"`  // 协议名称
		Operate  protocol.Operate  `json:"op"` // 操作
	}
	// 由于落库的时候已经检查过格式了, 因此这里可以直接对切片进行操作而不担心越界
	if err := json.Unmarshal([]byte(tx.TxData[p.headerLength:]), &base); err != nil {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolFormat, "invalid protocol format")
	}

	fmt.Println("Parse")
	fmt.Println(tx.TxData[p.headerLength:])
	// 根据协议获取解析器, 获取不到说明协议不支持
	parser, existed := p.parsers[base.Protocol]
	if !existed {
		return nil, protocol.NewProtocolError(protocol.UnknownProtocol, "unknown protocol")
	}

	return parser.Parse(tx)
}

func NewParser() Parser {

	parsers := make(map[protocol.Protocol]Parser)

	ierc20Parser := NewIERC20Parser(protocol.ProtocolHeader, protocol.TickETHI)
	parsers[protocol.ProtocolTERC20] = ierc20Parser
	parsers[protocol.ProtocolIERC20] = ierc20Parser
	parsers[protocol.ProtocolIERCPoW] = newIERC20PoWParser(protocol.ProtocolHeader)

	return &parser{
		header:       protocol.ProtocolHeader,
		headerLength: len(protocol.ProtocolHeader),
		parsers:      parsers,
	}
}
