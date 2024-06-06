package protocol

import (
	"fmt"

	"github.com/kevin88886/eth_indexer/pkg/utils"
	"github.com/shopspring/decimal"
)

// ================ deploy =================

type DeployCommand struct {
	IERCTransactionBase `json:"-"`
	Tick                string          `json:"tick,omitempty"`
	MaxSupply           decimal.Decimal `json:"max_supply"`              // 最大发行量
	Decimals            int64           `json:"decimals,omitempty"`      // tick的精度
	MintLimitOfSingleTx decimal.Decimal `json:"mint_limit_of_single_tx"` // 一笔交易最大mint数量, 单笔交易mint限制
	MintLimitOfWallet   decimal.Decimal `json:"mint_limit_of_wallet"`    // 一个地址最多mint数量
	Workc               string          `json:"workc,omitempty"`         // 挖矿难度, 可选
	Nonce               string          `json:"nonce,omitempty"`
}

func (d *DeployCommand) Validate() error {
	if err := d.IERCTransactionBase.Validate(); err != nil {
		return err
	}

	// check: len(tick) <= 64
	if len(d.Tick) > TickMaxLength {
		return NewProtocolError(InvalidProtocolParams, "invalid tick. length > 64")
	}

	// check workc
	if len(d.Workc) != 0 && !WorkCRegexp.MatchString(d.Workc) {
		return NewProtocolError(InvalidProtocolParams, fmt.Sprintf("invalid workc: %s", d.Workc))
	}

	// check: 0 <= decimals <= 18
	if d.Decimals < 0 || d.Decimals > 18 {
		return NewProtocolError(InvalidProtocolParams, "invalid decimals. tick decimals required 0 ~ 18")
	}

	// check: 0 <= max_supply <= max_supply_limit
	if d.MaxSupply.LessThan(decimal.Zero) || d.MaxSupply.GreaterThan(TickMaxSupplyLimit) {
		return NewProtocolError(InvalidProtocolParams, "invalid max supply. out of range")
	}

	// check: 0 <= mint <=  wallet_limit <= max_supply
	switch {
	// TODO: z limit 是否不能为0
	case d.MintLimitOfSingleTx.LessThanOrEqual(decimal.Zero):
		return NewProtocolError(InvalidProtocolParams, "invalid mint limit. limit <= 0")

	// TODO: z 由于 wallet_limit 可能大于 max_supply. 所以这里得放在这里, 先校验
	// limit <= max_supply
	case d.MintLimitOfSingleTx.GreaterThan(d.MaxSupply):
		return NewProtocolError(InvalidProtocolParams, "invalid mint limit. limit > wallet_limit")

	// check: limit <= wallet limit
	case d.MintLimitOfSingleTx.GreaterThan(d.MintLimitOfWallet):
		return NewProtocolError(InvalidProtocolParams, "invalid mint limit. limit > wallet_limit")
	}

	return nil
}

// ================ mint =================

type MintCommand struct {
	IERCTransactionBase
	Tick   string          `json:"tick,omitempty"`  // tick
	Amount decimal.Decimal `json:"amount"`          // mint 数量
	Nonce  string          `json:"nonce,omitempty"` // nonce
}

func (m *MintCommand) Validate() error {
	if err := m.IERCTransactionBase.Validate(); err != nil {
		return err
	}

	if m.Amount.LessThan(decimal.Zero) {
		return NewProtocolError(InvalidProtocolParams, "invalid mint amount. amount < 0")
	}

	return nil
}

// ================ transfer =================

type TransferRecord struct {
	Protocol Protocol        `json:"protocol"`
	Operate  Operate         `json:"operate"`
	Tick     string          `json:"tick,omitempty"`
	From     string          `json:"from,omitempty"`
	Recv     string          `json:"recv,omitempty"`
	Amount   decimal.Decimal `json:"amount"`
}

type TransferCommand struct {
	IERCTransactionBase
	Records []*TransferRecord
}

func (t *TransferCommand) Validate() error {
	if err := t.IERCTransactionBase.Validate(); err != nil {
		return err
	}

	if len(t.Records) == 0 {
		return NewProtocolError(InvalidProtocolParams, "missing transfer target")
	}

	for _, record := range t.Records {
		// check: len(tick) <= 64
		if len(record.Tick) > TickMaxLength {
			return NewProtocolError(InvalidProtocolParams, "invalid tick. length > 64")
		}

		// from是以太饭交易发起人的地址, 不需要校验

		if !utils.IsHexAddressWith0xPrefix(record.Recv) {
			return NewProtocolError(InvalidProtocolParams, "invalid recv address")
		}

		if record.Amount.LessThan(decimal.Zero) {
			return NewProtocolError(InvalidProtocolParams, "invalid amount. amount < 0")
		}
	}

	return nil
}

// ================ freeze sell =================

type FreezeRecord struct {
	Protocol   Protocol
	Operate    Operate
	Tick       string
	Platform   string
	Seller     string
	SellerSign string
	SignNonce  string
	Amount     decimal.Decimal
	Value      decimal.Decimal
	GasPrice   decimal.Decimal
}

type EntrustMessagePayment struct {
	Tick  string          `json:"tick"`
	Value decimal.Decimal `json:"value"`
	Fee   decimal.Decimal `json:"fee"`
}

type CreateEntrustMessage struct {
	Title    string                `json:"title"`
	Version  string                `json:"version"`
	Platform string                `json:"to"`
	Tick     string                `json:"tick"`
	Amount   decimal.Decimal       `json:"amt"`
	Value    decimal.Decimal       `json:"value"`
	Nonce    string                `json:"nonce"`
	Expire   string                `json:"expire"`
	Payment  EntrustMessagePayment `json:"payment"`
}

func (record *FreezeRecord) ValidateParams() error {
	if !utils.IsHexAddressWith0xPrefix(record.Seller) {
		return NewProtocolError(InvalidProtocolParams, "invalid seller address")
	}

	// TODO: z 平台地址? 是否需要校验地址是否正确
	if !utils.IsHexAddressWith0xPrefix(record.Platform) {
		return NewProtocolError(InvalidProtocolParams, "invalid platform address")
	}

	if record.Amount.LessThanOrEqual(decimal.Zero) {
		return NewProtocolError(InvalidProtocolParams, "invalid freeze amount. amount <= 0")
	}

	if record.Value.LessThanOrEqual(decimal.Zero) {
		return NewProtocolError(InvalidProtocolParams, "invalid value. value <= 0")
	}

	if record.GasPrice.LessThan(decimal.Zero) {
		return NewProtocolError(InvalidProtocolParams, "invalid gas price. gas_price < 0")
	}

	return nil
}

func (record *FreezeRecord) ValidateSignature() error {

	signature := NewSignature(
		record.Tick,
		record.Seller,
		PlatformAddress,
		record.Amount.String(),
		record.Value.String(),
		record.SignNonce,
	)

	return signature.ValidSignature(record.SellerSign)
}

type FreezeSellCommand struct {
	IERCTransactionBase
	Records []FreezeRecord
}

func (f *FreezeSellCommand) Validate() error {
	if err := f.IERCTransactionBase.Validate(); err != nil {
		return err
	}

	return nil
}

// ================ freeze sell v4=================

type FreezeRecordV4 struct {
	Protocol   Protocol
	Operate    Operate
	Tick       string
	Platform   string
	Seller     string
	SellerSign string
	SignNonce  string
	Amount     decimal.Decimal
	Value      decimal.Decimal
	Fee        decimal.Decimal
	Expire     string
	Version    string
	Payment    EntrustMessagePayment
}

func (record *FreezeRecordV4) ValidateParamsV4() error {
	if !utils.IsHexAddressWith0xPrefix(record.Seller) {
		return NewProtocolError(InvalidProtocolParams, "invalid seller address")
	}

	// TODO: z 平台地址? 是否需要校验地址是否正确
	if !utils.IsHexAddressWith0xPrefix(record.Platform) {
		return NewProtocolError(InvalidProtocolParams, "invalid platform address")
	}

	if record.Amount.LessThanOrEqual(decimal.Zero) {
		return NewProtocolError(InvalidProtocolParams, "invalid freeze amount. amount <= 0")
	}

	if record.Value.LessThanOrEqual(decimal.Zero) {
		return NewProtocolError(InvalidProtocolParams, "invalid value. value <= 0")
	}

	return nil
}

func (record *FreezeRecordV4) ValidateSignatureV4() error {

	signature := NewSignature(
		record.Tick,
		record.Seller,
		PlatformAddress,
		record.Amount.String(),
		record.Value.String(),
		record.SignNonce,
	)

	return signature.FreezeValidSignatureV4(record)
}

type FreezeSellCommandV4 struct {
	IERCTransactionBase
	Records []FreezeRecordV4
}

func (f *FreezeSellCommandV4) Validate() error {
	if err := f.IERCTransactionBase.Validate(); err != nil {
		return err
	}

	return nil
}

// ================= unfreeze sell ================

type UnfreezeRecord struct {
	Protocol            Protocol
	Operate             Operate
	TxHash              string // 解冻条目对应的 freezeSell记录 hash
	PositionInIERC20Txs int32  //
	Sign                string // 解冻条目对应的 freezeSell记录 sign
	Msg                 string // 解冻原因
}

type UnfreezeSellCommand struct {
	IERCTransactionBase
	Records []UnfreezeRecord
}

func (s *UnfreezeSellCommand) Validate() error {
	if err := s.IERCTransactionBase.Validate(); err != nil {
		return err
	}

	return nil
}

// ================ proxy transfer =================

type ProxyTransferRecord struct {
	Protocol    Protocol
	Operate     Operate
	Tick        string
	From        string
	To          string
	Amount      decimal.Decimal
	Value       decimal.Decimal
	Sign        string
	SignerNonce string
}

func (record *ProxyTransferRecord) ValidateParams() error {
	if !utils.IsHexAddressWith0xPrefix(record.From) {
		return NewProtocolError(InvalidProtocolParams, fmt.Sprintf("invalid from address: %s", record.From))
	}

	if !utils.IsHexAddressWith0xPrefix(record.To) {
		return NewProtocolError(InvalidProtocolParams, fmt.Sprintf("invalid to address: %s", record.To))
	}

	if record.Amount.LessThan(decimal.Zero) {
		return NewProtocolError(InvalidProtocolParams, fmt.Sprintf("invalid amount. Amount(%s) < 0", record.Amount))
	}

	return nil
}

func (record *ProxyTransferRecord) ValidateSignature() error {

	signature := NewSignature(
		record.Tick,
		record.From,
		PlatformAddress,
		record.Amount.String(),
		record.Value.String(),
		record.SignerNonce,
	)

	return signature.ValidSignature(record.Sign)
}

type ProxyTransferCommand struct {
	IERCTransactionBase
	Records []ProxyTransferRecord
}

func (pt *ProxyTransferCommand) Validate() error {
	if err := pt.IERCTransactionBase.Validate(); err != nil {
		return err
	}

	if len(pt.Records) == 0 {
		return NewProtocolError(InvalidProtocolParams, "missing records")
	}

	return nil
}

// ================ proxy transfer v4=================

type ProxyTransferRecordV4 struct {
	Protocol    Protocol
	Operate     Operate
	Tick        string
	From        string
	To          string
	Amount      decimal.Decimal
	Value       decimal.Decimal
	Sign        string
	SignerNonce string
	Expire      string
	Version     string
	Payment     EntrustMessagePayment
}

func (record *ProxyTransferRecordV4) ValidateParams() error {
	if !utils.IsHexAddressWith0xPrefix(record.From) {
		return NewProtocolError(InvalidProtocolParams, fmt.Sprintf("invalid from address: %s", record.From))
	}

	if !utils.IsHexAddressWith0xPrefix(record.To) {
		return NewProtocolError(InvalidProtocolParams, fmt.Sprintf("invalid to address: %s", record.To))
	}

	if record.Amount.LessThan(decimal.Zero) {
		return NewProtocolError(InvalidProtocolParams, fmt.Sprintf("invalid amount. Amount(%s) < 0", record.Amount))
	}

	return nil
}

func (record *ProxyTransferRecordV4) ValidateSignature() error {

	signature := NewSignature(
		record.Tick,
		record.From,
		PlatformAddress,
		record.Amount.String(),
		record.Value.String(),
		record.SignerNonce,
	)

	return signature.ProxyTransferValidSignatureV4(record)
}

type ProxyTransferCommandV4 struct {
	IERCTransactionBase
	Records []ProxyTransferRecordV4
}

func (pt *ProxyTransferCommandV4) Validate() error {
	if err := pt.IERCTransactionBase.Validate(); err != nil {
		return err
	}

	if len(pt.Records) == 0 {
		return NewProtocolError(InvalidProtocolParams, "missing records")
	}

	return nil
}

// ================ about staking pool ================

type TickConfigDetail struct {
	Tick                 string          // 可质押的Tick
	RewardsRatioPerBlock decimal.Decimal // 当前Tick对应的每个块的奖励比例
	MaxAmount            decimal.Decimal // 最大质押数量
}

func (s *TickConfigDetail) String() string {
	return fmt.Sprintf("%T(tick: %s, ratio: %s, maxAmount: %s)", s, s.Tick, s.RewardsRatioPerBlock, s.MaxAmount)
}

type ConfigStakeCommand struct {
	IERCTransactionBase
	Pool      string              // 池子地址
	PoolSubID uint64              //
	Owner     string              // 所有者
	Admins    []string            // 普通管理员
	Name      string              // 池子名称
	StopBlock uint64              // 生效区块高度
	Details   []*TickConfigDetail // 更新项
}

func (c *ConfigStakeCommand) String() string {
	return fmt.Sprintf(
		`%T("%s, pool: %s, stop_block: %d, details: %v")`,
		c, c.IERCTransactionBase.String(), c.Pool, c.StopBlock, c.Details,
	)
}

func (c *ConfigStakeCommand) Validate() error {
	if err := c.IERCTransactionBase.Validate(); err != nil {
		return err
	}

	if !utils.IsHexAddressWith0xPrefix(c.Pool) {
		return NewProtocolError(InvalidProtocolParams, "invalid pool address")
	}

	if c.PoolSubID < 0 {
		return NewProtocolError(InvalidProtocolParams, "invalid pool id")
	}

	if !utils.IsHexAddressWith0xPrefix(c.Owner) {
		return NewProtocolError(InvalidProtocolParams, "invalid owner address")
	}

	for _, admin := range c.Admins {
		if !utils.IsHexAddressWith0xPrefix(admin) {
			return NewProtocolError(InvalidProtocolParams, "invalid owner address")
		}
	}

	if len(c.Details) == 0 {
		return NewProtocolError(InvalidProtocolParams, "missing config params")
	}

	var flagMap = make(map[string]struct{})

	for _, detail := range c.Details {
		if len(detail.Tick) == 0 {
			return NewProtocolError(InvalidProtocolParams, "missing tick")
		}

		if _, existed := flagMap[detail.Tick]; existed {
			return NewProtocolError(InvalidProtocolParams, "repeated tick")
		}
		flagMap[detail.Tick] = struct{}{}

		if detail.RewardsRatioPerBlock.LessThan(decimal.Zero) {
			return NewProtocolError(InvalidProtocolParams, "ratio must be greater than or equal 0")
		}
	}

	return nil
}

// ================ about staking & unstaking & proxy_unstaking ================

type StakingDetail struct {
	Protocol  Protocol
	Operate   Operate
	Staker    string
	Pool      string
	PoolSubID uint64
	Tick      string
	Amount    decimal.Decimal
}

func (s *StakingDetail) String() string {
	return fmt.Sprintf("%T(staker: %s, tick: %s, amount: %s)", s, s.Staker, s.Tick, s.Amount)
}

func (s *StakingDetail) ValidateParams() error {
	if !utils.IsHexAddressWith0xPrefix(s.Staker) {
		return NewProtocolError(InvalidProtocolParams, "invalid staker address")
	}

	if !utils.IsHexAddressWith0xPrefix(s.Pool) {
		return NewProtocolError(InvalidProtocolParams, "invalid pool address")
	}

	if s.PoolSubID < 0 {
		return NewProtocolError(InvalidProtocolParams, "invalid pool id")
	}

	if len(s.Tick) == 0 {
		return NewProtocolError(InvalidProtocolParams, "missing tick")
	}

	if s.Amount.LessThan(decimal.Zero) {
		return NewProtocolError(InvalidProtocolParams, "amount must be greater than or equal 0")
	}

	return nil
}

type StakingCommand struct {
	IERCTransactionBase
	Pool      string
	PoolSubID uint64
	Details   []*StakingDetail
}

func (s *StakingCommand) String() string {
	return fmt.Sprintf(
		"%T(%s, pool: %s, pool_id: %d, details: %v)",
		s, s.IERCTransactionBase.String(), s.Pool, s.PoolSubID, s.Details,
	)
}

func (s *StakingCommand) Validate() error {
	if err := s.IERCTransactionBase.Validate(); err != nil {
		return err
	}

	if !utils.IsHexAddressWith0xPrefix(s.Pool) {
		return NewProtocolError(InvalidProtocolParams, "invalid pool address")
	}

	if s.PoolSubID < 0 {
		return NewProtocolError(InvalidProtocolParams, "invalid pool id")
	}

	if len(s.Details) == 0 {
		return NewProtocolError(InvalidProtocolParams, "missing staking params")
	}

	// 用于标记列表中同一个用户是否存在重复的tick
	var flagMap = make(map[string]map[string]struct{})

	for _, record := range s.Details {

		// 判断是否有重复的tick
		tickFlag, existed := flagMap[record.Staker]
		if existed {
			if _, existed := tickFlag[record.Tick]; existed {
				return NewProtocolError(InvalidProtocolParams, "repeated tick")
			}
		} else {
			tickFlag = make(map[string]struct{})
			flagMap[record.Staker] = tickFlag
		}
		tickFlag[record.Tick] = struct{}{}

		// 验证参数
		if err := record.ValidateParams(); err != nil {
			return err
		}
	}

	return nil
}
