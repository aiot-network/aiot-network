package param

import (
	"github.com/aiot-network/aiot-network/common/private"
	"github.com/aiot-network/aiot-network/tools/arry"
	"time"
)

const (
	// Block interval period
	BlockInterval = uint64(15)
	// Re-election interval
	CycleInterval = 60 * 60 * 24
	//CycleInterval = 60
	// Maximum number of super nodes
	SuperSize = 9

	DPosSize = SuperSize*2/3 + 1
)

const (
	// Mainnet logo
	MainNet = "mainnet"
	// Testnet logo
	TestNet = "testnet"

	APPName = "AIOT_NETWORK"
)
const (
	MaxReadBytes = 1024 * 10
	MaxReqBytes  = MaxReadBytes * 1000
)

// AtomsPerCoin is the number of atomic units in one coin.
const AtomsPerCoin = 1e8

type Param struct {
	Name              string
	Data              string
	App               string
	RollBack          uint64
	PubKeyHashAddrID  [2]byte
	PubKeyHashTokenID [2]byte
	Logging           bool
	PeerRequestChan   uint32
	*PrivateParam
	*TokenParam
	*P2pParam
	*RpcParam
	*DPosParam
	*PoolParam
	private.IPrivate
}

type TokenParam struct {
	PreCirculation        uint64
	Circulation           uint64
	CoinBaseOneDay        uint64
	EveryChangeCoinHeight uint64
	Proportion            uint64
	MinCoinCount          float64
	MaxCoinCount          float64
	MinimumTransfer       uint64
	MaximumTransfer       uint64
	Consume               uint64
	MaximumReceiver       int
	MainToken             arry.Address
	EaterAddress          arry.Address
}

type PrivateParam struct {
	PrivateFile string
	PrivatePass string
}

type P2pParam struct {
	P2pPort    string
	ExternalIp string
	NetWork    string
	CustomBoot string
}

type RpcParam struct {
	RpcIp      string
	RpcPort    string
	RpcTLS     bool
	RpcCert    string
	RpcCertKey string
	RpcPass    string
}

type Super struct {
	Address string
	P2PId   string
}

type DPosParam struct {
	BlockInterval    uint64
	CycleInterval    uint64
	SuperSize        int
	DPosSize         int
	GenesisTime      uint64
	GenesisCycle     uint64
	WorkProofAddress string
	GenesisSuperList []Super
}

type PoolParam struct {
	MsgExpiredTime     int64
	MonitorMsgInterval time.Duration
	MaxPoolMsg         int
	MaxAddressMsg      uint64
}

var TestNetParam = &Param{
	Name:              TestNet,
	Data:              "data",
	App:               "AIOT_NETWORK",
	RollBack:          0,
	PubKeyHashAddrID:  [2]byte{0xf7, 0x95},
	PubKeyHashTokenID: [2]byte{0xf7, 0xae},
	Logging:           true,
	PeerRequestChan:   1000,
	PrivateParam: &PrivateParam{
		PrivateFile: "key.json",
		PrivatePass: APPName,
	},
	TokenParam: &TokenParam{
		PreCirculation:  0,
		Circulation:     1 * 1e8 * AtomsPerCoin,
		CoinBaseOneDay:  27390 * AtomsPerCoin,
		Consume:         1e4 * AtomsPerCoin,
		MinCoinCount:    1 * 1e4,
		MaxCoinCount:    9 * 1e10,
		MinimumTransfer: 0.0001 * AtomsPerCoin,
		MaximumTransfer: 1 * 1e7 * AtomsPerCoin,
		MaximumReceiver: 1 * 1e4,
		MainToken:       arry.StringToAddress("AIOT"),
		EaterAddress:    arry.StringToAddress("aiCoinEaterAddressDontSend000000000"),
	},
	P2pParam: &P2pParam{
		NetWork:    TestNet + "AIOT_NETWORK",
		P2pPort:    "13561",
		ExternalIp: "0.0.0.0",
		//aiVrBLugvcKoXCb3YTwYfbZe2WJ3nY8ftDG
		CustomBoot: "/ip4/103.68.63.164/tcp/19563/ipfs/16Uiu2HAmKWBySkhEquPd1T2QweDYaeKHCxMfkmr7c1AoHpGSnh9x",
	},
	RpcParam: &RpcParam{
		RpcIp:      "127.0.0.1",
		RpcPort:    "13562",
		RpcTLS:     false,
		RpcCert:    "",
		RpcCertKey: "",
		RpcPass:    "",
	},
	DPosParam: &DPosParam{
		BlockInterval:    BlockInterval,
		CycleInterval:    CycleInterval,
		SuperSize:        SuperSize,
		DPosSize:         DPosSize,
		GenesisTime:      1592268410,
		GenesisCycle:     1592268410 / CycleInterval,
		WorkProofAddress: "aiGqjoXAoStn7zSYY4cETR9m4trPEEQ7CTa",
		GenesisSuperList: []Super{
			{
				Address: "aiGJW3ZFuudsLeP4YjdZrmqCBJChJH1LBoB",
				P2PId:   "16Uiu2HAm6xn4pcjJ6HviBgZeicQqVdEZMsa5uoQ1prA6o5LpAyJt",
			},
			{
				Address: "aiMnamdZsAnWy5Kw4pK9HPtgfKt7cDfLMUi",
				P2PId:   "16Uiu2HAm6xn4pcjJ6HviBgZeicQqVdEZMsa5uoQ1prA6o5LpAyJt",
			},
			{
				Address: "aiCY2y3im6MG1wTpJgAogS8NgscwLr66Ueu",
				P2PId:   "16Uiu2HAmPeL1jafz9mpEkftBpfRZRFZGKackGEzaXd5Rh1pLQbvq",
			},
			{
				Address: "aiVCdvVT6nkRTPbSBoGswxWDf1tSfqk2KkG",
				P2PId:   "16Uiu2HAm8jmcocgCELyFnuL5DGfBUn6WaCoCbs4kGwgbDpFNkM6h",
			},
			{
				Address: "aiC3aWKDnLog5Df7r1tUcZsViRTzF1AYJ5U",
				P2PId:   "16Uiu2HAmRxe38EUafwy2YqTwiRmc3CgSi6zW2kLAsQKUQgnoEB2A",
			},
			{
				Address: "aiGCRf7AHPLCjfE6VX1Zcz8RtJoXHYHe8Je",
				P2PId:   "16Uiu2HAmRTvsT7HcZJGYncTBeefxBm6b5pnyuSCuYV6Hgcmj4VuW",
			},
			{
				Address: "aiRzwTSYXYCqLmA295x2gY4BMn7khBbzLr4",
				P2PId:   "16Uiu2HAmDoWd5ioXVs6qrCmmJfwo9PctKfv4HWzmrP5Lnv3Jo5iP",
			},
			{
				Address: "aiVnGCRrVLr4ntrz7J11HkjDX9a9JHinRc1",
				P2PId:   "16Uiu2HAkufPsFP4acpAMeKh6xQvLCzShpSuMFRaEFxTSE8Mjuk5i",
			},
			{
				Address: "aiQuA6VMPT4rdVZjKMWyucRY8jSdAF2dpvh",
				P2PId:   "16Uiu2HAmHYLacSmfNQjUfew59SBo1pdqKjrgz8awgN3J4DwbkWSS",
			},
		},
	},
	PoolParam: &PoolParam{
		MaxPoolMsg:         100000,
		MsgExpiredTime:     60 * 60 * 3,
		MonitorMsgInterval: 10,
		MaxAddressMsg:      1000,
	},
}

var MainNetParam = &Param{
	Name:              MainNet,
	Data:              "data",
	App:               APPName,
	RollBack:          0,
	PubKeyHashAddrID:  [2]byte{0xea, 0x12},
	PubKeyHashTokenID: [2]byte{0x5, 0x91},
	Logging:           true,
	PeerRequestChan:   1000,
	PrivateParam: &PrivateParam{
		PrivateFile: "key.json",
		PrivatePass: "AIOT_NETWORK",
	},
	TokenParam: &TokenParam{
		PreCirculation:  0,
		Circulation:     1 * 1e8 * AtomsPerCoin,
		CoinBaseOneDay:  27390 * AtomsPerCoin,
		Consume:         1e4 * AtomsPerCoin,
		MinCoinCount:    1 * 1e4,
		MaxCoinCount:    9 * 1e10,
		MinimumTransfer: 0.0001 * AtomsPerCoin,
		MaximumTransfer: 1 * 1e7 * AtomsPerCoin,
		MaximumReceiver: 1 * 1e4,
		MainToken:       arry.StringToAddress("AIOT"),
		EaterAddress:    arry.StringToAddress("AiCoinEaterAddressDontSend000000000"),
	},
	P2pParam: &P2pParam{
		NetWork:    MainNet + "AIOT_NETWORK",
		P2pPort:    "23561",
		ExternalIp: "0.0.0.0",
		//AihsVKMf6WdmUAyhin2FH7itHiqbCcTZEsj
		CustomBoot: "/ip4/103.68.63.164/tcp/19563/ipfs/16Uiu2HAm1WsqwXYH3mvFk46zejvewcQ6wAbVBj5qniAgFYiYwKd4",
	},
	RpcParam: &RpcParam{
		RpcIp:      "127.0.0.1",
		RpcPort:    "23562",
		RpcTLS:     false,
		RpcCert:    "",
		RpcCertKey: "",
		RpcPass:    "",
	},
	DPosParam: &DPosParam{
		BlockInterval:    BlockInterval,
		CycleInterval:    CycleInterval,
		SuperSize:        SuperSize,
		DPosSize:         DPosSize,
		GenesisTime:      1592268410,
		GenesisCycle:     1592268410 / CycleInterval,
		WorkProofAddress: "AifLkzE8iMPEUy7mHwJ89YbsEDesVkRZ8Fn",
		GenesisSuperList: []Super{
			{
				Address: "AiXAojSN3EKMCkkeKQSTzqiACw37siM8CSF",
				P2PId:   "16Uiu2HAmNDdZbFgXvmqWRK65LjqxMWwExJercSL4H4Yb9zHgrHEf",
			},
			{
				Address: "Aif7s4vj1fdCf2d3bGFHwqxXSTHFLcHbhSz",
				P2PId:   "16Uiu2HAmTbWamkjotQbRdt8PdxuGq4kXCHUi33bVG4TMR7v5Vz7q",
			},
			{
				Address: "AiWPzrPd7XvzmBCD3WjM9NmUVAoYUsmcp9d",
				P2PId:   "16Uiu2HAm66iJEEDxS1t7q5Jp7R7deGFGn6B1QLX2KepHVXX2ksW4",
			},
			{
				Address: "AiSMfPthT5Uj9KmGHXFd4arSDL3mGsAF4Gu",
				P2PId:   "16Uiu2HAm5a7zJ87ZSdWr255Ez9s9am2DJFNHRQCKLkBDRp8H5APY",
			},
			{
				Address: "AiYAG3ZzeVx8Gi7fXFsprGmRQ6g2sx5vQvL",
				P2PId:   "16Uiu2HAmC6pHkE18pkRKagv1zm7HcQD5yza8qXeRTRqLBFxBpMHX",
			},
			{
				Address: "AiNP7ExrY4S8Hi6ZL9SMPKRiVwpSWUJ9L5t",
				P2PId:   "16Uiu2HAmBZqLmUptUPYDHGKiHaCNKSVezggZvAY9YvU9DdvAfDNb",
			},
			{
				Address: "AiV8UBrEJdYNVqa6bM1iXQVezwX98eUs3Dm",
				P2PId:   "16Uiu2HAmCmHMcteRpdCSYHPcWSPbdxdZwFYZzEQQ1EQNy4bsj9jp",
			},
			{
				Address: "AiU1DGBim4tNEWDwJwmHG5uEE9kMu53U7MZ",
				P2PId:   "16Uiu2HAmPtAsrpBRw26GptVxZcfKcREdqu81H5LRCVKWvtgJJhXp",
			},
			{
				Address: "AijYeC9MEF1VABWnBqra6S6j9JRjjZfFUiv",
				P2PId:   "16Uiu2HAmP9BS5UCn7rq6UXgwvJ1UeFGsodwLamSMo2KBwCoVhvwT",
			},
		},
	},
	PoolParam: &PoolParam{
		MaxPoolMsg:         100000,
		MsgExpiredTime:     60 * 60 * 3,
		MonitorMsgInterval: 10,
		MaxAddressMsg:      1000,
	},
}

type PreCirculation struct {
	Address string
	Note    string
	Amount  uint64
}

var PreCirculations = []PreCirculation{}
