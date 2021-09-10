package exchange_runner

var exMethods = map[string]*MethodInfo{
	"Methods": &MethodInfo{
		Name:   "Methods",
		Params: nil,
		Returns: []Value{
			{
				Name: "Open methods",
				Type: "json",
			},
		},
	},
	"MethodExist": &MethodInfo{
		Name: "MethodExist",
		Params: []Value{
			{
				Name: "method",
				Type: "string",
			},
		},
		Returns: []Value{
			{
				Name: "exist",
				Type: "bool",
			},
		},
	},
	"Pairs": &MethodInfo{
		Name:   "Pairs",
		Params: nil,
		Returns: []Value{
			{
				Name: "pair list",
				Type: "json",
			},
		},
	},
	"ExchangeRouter": &MethodInfo{
		Name: "ExchangeRouter",
		Params: []Value{
			{
				Name: "tokenA",
				Type: "string",
			},
			{
				Name: "tokenB",
				Type: "string",
			},
		},
		Returns: []Value{
			{
				Name: "paths",
				Type: "json",
			},
		},
	},

	"ExchangeRouterWithAmount": &MethodInfo{
		Name: "ExchangeRouterWithAmount",
		Params: []Value{
			{
				Name: "tokenA",
				Type: "string",
			},
			{
				Name: "tokenB",
				Type: "string",
			},
			{
				Name: "amountIn",
				Type: "float64",
			},
		},
		Returns: []Value{
			{
				Name: "path and amount",
				Type: "json",
			},
		},
	},

	"ExchangeOptimalRouter": &MethodInfo{
		Name: "ExchangeOptimalRouter",
		Params: []Value{
			{
				Name: "tokenA",
				Type: "string",
			},
			{
				Name: "tokenB",
				Type: "string",
			},
			{
				Name: "amountIn",
				Type: "float64",
			},
		},
		Returns: []Value{
			{
				Name: "optimal path",
				Type: "json",
			},
		},
	},

	"AmountOut": &MethodInfo{
		Name: "AmountOut",
		Params: []Value{
			{
				Name: "paths(tokenA,tokenB,tokenC)",
				Type: "string",
			},
			{
				Name: "amountIn",
				Type: "float64",
			},
		},
		Returns: []Value{
			{
				Name: "amountOut",
				Type: "float64",
			},
		},
	},
	"AmountIn": &MethodInfo{
		Name: "AmountIn",
		Params: []Value{
			{
				Name: "paths(tokenA,tokenB,tokenC)",
				Type: "string",
			},
			{
				Name: "amountOut",
				Type: "float64",
			},
		},
		Returns: []Value{
			{
				Name: "amountIn",
				Type: "float64",
			},
		},
	},
	"LegalPair": &MethodInfo{
		Name: "LegalPair",
		Params: []Value{
			{
				Name: "tokenA",
				Type: "string",
			},
			{
				Name: "tokenB",
				Type: "string",
			},
		},
		Returns: []Value{
			{
				Name: "is legal",
				Type: "bool",
			},
		},
	},
	"PairAddress": &MethodInfo{
		Name: "PairAddress",
		Params: []Value{
			{
				Name: "tokenA",
				Type: "string",
			},
			{
				Name: "tokenB",
				Type: "string",
			},
		},
		Returns: []Value{
			{
				Name: "pair address",
				Type: "string",
			},
		},
	},
}

var pairMethods = map[string]*MethodInfo{
	"Methods": &MethodInfo{
		Name:   "Methods",
		Params: nil,
		Returns: []Value{
			{
				Name: "Open methods",
				Type: "json",
			},
		},
	},
	"MethodExist": &MethodInfo{
		Name: "MethodExist",
		Params: []Value{
			{
				Name: "method",
				Type: "string",
			},
		},
		Returns: []Value{
			{
				Name: "exist",
				Type: "bool",
			},
		},
	},
	"QuoteAmountB": &MethodInfo{
		Name: "QuoteAmountB",
		Params: []Value{
			{
				Name: "TokenA",
				Type: "string",
			},
			{
				Name: "AmountA",
				Type: "float64",
			},
		},
		Returns: []Value{
			{
				Name: "AmountB",
				Type: "float64",
			},
		},
	},

	"TotalValue": &MethodInfo{
		Name: "TotalValue",
		Params: []Value{
			{
				Name: "liquidity",
				Type: "float64",
			},
		},
		Returns: []Value{
			{
				Name: "value",
				Type: "json",
			},
		},
	},
}
