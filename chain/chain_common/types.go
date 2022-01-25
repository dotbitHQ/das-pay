package chain_common

type Block struct {
	BlockHeader
	Transactions []Transaction `json:"transactions"`
}

type BlockHeader struct {
	Author           string   `json:"author"`
	Difficulty       string   `json:"difficulty"`
	ExtraData        string   `json:"extraData"`
	GasLimit         string   `json:"gasLimit"`
	GasUsed          string   `json:"gasUsed"`
	Hash             string   `json:"hash"`
	LogsBloom        string   `json:"logsBloom"`
	Miner            string   `json:"miner"`
	MixHash          string   `json:"mixHash"`
	Nonce            string   `json:"nonce"`
	Number           string   `json:"number"`
	ParentHash       string   `json:"parentHash"`
	ReceiptsRoot     string   `json:"receiptsRoot"`
	SealFields       []string `json:"sealFields"`
	Sha3Uncles       string   `json:"sha3Uncles"`
	Size             string   `json:"size"`
	StateRoot        string   `json:"stateRoot"`
	Timestamp        string   `json:"timestamp"`
	TotalDifficulty  string   `json:"totalDifficulty"`
	TransactionsRoot string   `json:"transactionsRoot"`
	Uncles           []string `json:"uncles"`
}

type Transaction struct {
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	ChainId          string `json:"chainId"`
	Condition        string `json:"condition"`
	Creates          string `json:"creates"`
	From             string `json:"from"`
	To               string `json:"to"`
	Gas              string `json:"gas"`
	GasUsed          string `json:"gas_used"`
	GasPrice         string `json:"gasPrice"`
	Hash             string `json:"hash"`
	Input            string `json:"input"`
	Nonce            string `json:"nonce"`
	TransactionIndex string `json:"transactionIndex"`
	PublicKey        string `json:"publicKey"`
	Value            string `json:"value"`
	Raw              string `json:"raw"`
	StandardV        string `json:"standardV"`
	R                string `json:"r"`
	S                string `json:"s"`
	V                string `json:"v"`
	Status           string `json:"status"`
}
