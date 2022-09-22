package out

import (
        "context"
        "gfx.cafe/open/jrpc"
       )


type GoOpenRPCHandler struct {
	Srv GoOpenRPCService
}


func (h *GoOpenRPCHandler) RouteRPC(r jrpc.Router) {
    // Returns an RLP-encoded header.

            r.Route("debug", func(r2 jrpc.Router) {
                r.RegisterFunc("getRawHeader", h.Srv.DebugGetRawHeader)
            })

    // Returns an RLP-encoded block.

            r.Route("debug", func(r2 jrpc.Router) {
                r.RegisterFunc("getRawBlock", h.Srv.DebugGetRawBlock)
            })

    // Returns an array of EIP-2718 binary-encoded transactions.

            r.Route("debug", func(r2 jrpc.Router) {
                r.RegisterFunc("getRawTransaction", h.Srv.DebugGetRawTransaction)
            })

    // Returns an array of EIP-2718 binary-encoded receipts.

            r.Route("debug", func(r2 jrpc.Router) {
                r.RegisterFunc("getRawReceipts", h.Srv.DebugGetRawReceipts)
            })

    // Returns an array of recent bad blocks that the client has seen on the network.

            r.Route("debug", func(r2 jrpc.Router) {
                r.RegisterFunc("getBadBlocks", h.Srv.DebugGetBadBlocks)
            })

    // Returns information about a block by hash.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getBlockByHash", h.Srv.EthGetBlockByHash)
            })

    // Returns information about a block by number.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getBlockByNumber", h.Srv.EthGetBlockByNumber)
            })

    // Returns the number of transactions in a block from a block matching the given block hash.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getBlockTransactionCountByHash", h.Srv.EthGetBlockTransactionCountByHash)
            })

    // Returns the number of transactions in a block matching the given block number.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getBlockTransactionCountByNumber", h.Srv.EthGetBlockTransactionCountByNumber)
            })

    // Returns the number of uncles in a block from a block matching the given block hash.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getUncleCountByBlockHash", h.Srv.EthGetUncleCountByBlockHash)
            })

    // Returns the number of transactions in a block matching the given block number.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getUncleCountByBlockNumber", h.Srv.EthGetUncleCountByBlockNumber)
            })

    // Returns the chain ID of the current network.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("chainId", h.Srv.EthChainId)
            })

    // Returns an object with data about the sync status or false.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("syncing", h.Srv.EthSyncing)
            })

    // Returns the client coinbase address.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("coinbase", h.Srv.EthCoinbase)
            })

    // Returns a list of addresses owned by client.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("accounts", h.Srv.EthAccounts)
            })

    // Returns the number of most recent block.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("blockNumber", h.Srv.EthBlockNumber)
            })

    // Executes a new message call immediately without creating a transaction on the block chain.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("call", h.Srv.EthCall)
            })

    // Generates and returns an estimate of how much gas is necessary to allow the transaction to complete.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("estimateGas", h.Srv.EthEstimateGas)
            })

    // Generates an access list for a transaction.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("createAccessList", h.Srv.EthCreateAccessList)
            })

    // Returns the current price per gas in wei.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("gasPrice", h.Srv.EthGasPrice)
            })

    // Returns the current maxPriorityFeePerGas per gas in wei.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("maxPriorityFeePerGas", h.Srv.EthMaxPriorityFeePerGas)
            })

    // Transaction fee history

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("feeHistory", h.Srv.EthFeeHistory)
            })

    // Creates a filter object, based on filter options, to notify when the state changes (logs).

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("newFilter", h.Srv.EthNewFilter)
            })

    // Creates a filter in the node, to notify when a new block arrives.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("newBlockFilter", h.Srv.EthNewBlockFilter)
            })

    // Creates a filter in the node, to notify when new pending transactions arrive.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("newPendingTransactionFilter", h.Srv.EthNewPendingTransactionFilter)
            })

    // Uninstalls a filter with given id.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("uninstallFilter", h.Srv.EthUninstallFilter)
            })

    // Polling method for a filter, which returns an array of logs which occurred since last poll.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getFilterChanges", h.Srv.EthGetFilterChanges)
            })

    // Returns an array of all logs matching filter with given id.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getFilterLogs", h.Srv.EthGetFilterLogs)
            })

    // Returns an array of all logs matching filter with given id.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getLogs", h.Srv.EthGetLogs)
            })

    // Returns whether the client is actively mining new blocks.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("mining", h.Srv.EthMining)
            })

    // Returns the number of hashes per second that the node is mining with.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("hashrate", h.Srv.EthHashrate)
            })

    // Returns the hash of the current block, the seedHash, and the boundary condition to be met (“target”).

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getWork", h.Srv.EthGetWork)
            })

    // Used for submitting a proof-of-work solution.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("submitWork", h.Srv.EthSubmitWork)
            })

    // Used for submitting mining hashrate.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("submitHashrate", h.Srv.EthSubmitHashrate)
            })

    // Returns an EIP-191 signature over the provided data.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("sign", h.Srv.EthSign)
            })

    // Returns an RLP encoded transaction signed by the specified account.

            r.Ro Generates and returns an estimate of how much gas is necessary to allow the transaction to complete.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("estimateGas", h.Srv.EthEstimateGas)
            })

    // Generates an access list for a transaction.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("createAccessList", h.Srv.EthCreateAccessList)
            })

    // Returns the current price per gas in wei.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("gasPrice", h.Srv.EthGasPrice)
            })

    // Returns the current maxPriorityFeePerGas per gas in wei.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("maxPriorityFeePerGas", h.Srv.EthMaxPriorityFeePerGas)
            })

    // Transaction fee history

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("feeHistory", h.Srv.EthFeeHistory)
            })

    // Creates a filter object, based on filter options, to notify when the state changes (logs).

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("newFilter", h.Srv.EthNewFilter)
            })

    // Creates a filter in the node, to notify when a new block arrives.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("newBlockFilter", h.Srv.EthNewBlockFilter)
            })

    // Creates a filter in the node, to notify when new pending transactions arrive.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("newPendingTransactionFilter", h.Srv.EthNewPendingTransactionFilter)
            })

    // Uninstalls a filter with given id.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("uninstallFilter", h.Srv.EthUninstallFilter)
            })

    // Polling method for a filter, which returns an array of logs which occurred since last poll.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getFilterChanges", h.Srv.EthGetFilterChanges)
            })

    // Returns an array of all logs matching filter with given id.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getFilterLogs", h.Srv.EthGetFilterLogs)
            })

    // Returns an array of all logs matching filter with given id.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getLogs", h.Srv.EthGetLogs)
            })

    // Returns whether the client is actively mining new blocks.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("mining", h.Srv.EthMining)
            })

    // Returns the number of hashes per second that the node is mining with.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("hashrate", h.Srv.EthHashrate)
            })

    // Returns the hash of the current block, the seedHash, and the boundary condition to be met (“target”).

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getWork", h.Srv.EthGetWork)
            })

    // Used for submitting a proof-of-work solution.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("submitWork", h.Srv.EthSubmitWork)
            })

    // Used for submitting mining hashrate.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("submitHashrate", h.Srv.EthSubmitHashrate)
            })

    // Returns an EIP-191 signature over the provided data.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("sign", h.Srv.EthSign)
            })

    // Returns an RLP encoded transaction signed by the specified account.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("signTransaction", h.Srv.EthSignTransaction)
            })

    // Returns the balance of the account of given address.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getBalance", h.Srv.EthGetBalance)
            })

    // Returns the value from a storage position at a given address.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getStorageAt", h.Srv.EthGetStorageAt)
            })

    // Returns the number of transactions sent from an address.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getTransactionCount", h.Srv.EthGetTransactionCount)
            })

    // Returns code at a given address.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getCode", h.Srv.EthGetCode)
            })

    // Returns the merkle proof for a given account and optionally some storage keys.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getProof", h.Srv.EthGetProof)
            })

    // Signs and submits a transaction.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("sendTransaction", h.Srv.EthSendTransaction)
            })

    // Submits a raw transaction.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("sendRawTransaction", h.Srv.EthSendRawTransaction)
            })

    // Returns the information about a transaction requested by transaction hash.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getTransactionByHash", h.Srv.EthGetTransactionByHash)
            })

    // Returns information about a transaction by block hash and transaction index position.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getTransactionByBlockHashAndIndex", h.Srv.EthGetTransactionByBlockHashAndIndex)
            })

    // Returns information about a transaction by block number and transaction index position.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getTransactionByBlockNumberAndIndex", h.Srv.EthGetTransactionByBlockNumberAndIndex)
            })

    // Returns the receipt of a transaction by transaction hash.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getTransactionReceipt", h.Srv.EthGetTransactionReceipt)
            })

    }

type RpcHandler struct {
// Returns an RLP-encoded header.
  FnDebugGetRawHeader func(ctx context.Context,
        Block BlockNumberOrTag,
        )(HeaderRLP Bytes, err error)

// Returns an RLP-encoded block.
  FnDebugGetRawBlock func(ctx context.Context,
        Block BlockNumberOrTag,
        )(BlockRLP Bytes, err error)

// Returns an array of EIP-2718 binary-encoded transactions.
  FnDebugGetRawTransaction func(ctx context.Context,
        TransactionHash Hash32,
        )(EIP2718BinaryEncodedTransaction Bytes, err error)

// Returns an array of EIP-2718 binary-encoded receipts.
  FnDebugGetRawReceipts func(ctx context.Context,
        Block BlockNumberOrTag,
        )(Receipts []Bytes, err error)

// Returns an array of recent bad blocks that the client has seen on the network.
  FnDebugGetBadBlocks func(ctx context.Context,
        )(Blocks []BadBlock, err error)

// Returns information about a block by hash.
  FnEthGetBlockByHash func(ctx context.Context,
        BlockHash Hash32,
        HydratedTransactions bool,
        )(BlockInformation Block, err error)

// Returns information about a block by number.
  FnEthGetBlockByNumber func(ctx context.Context,
        Block BlockNumberOrTag,
        HydratedTransactions bool,
        )(BlockInformation Block, err error)

// Returns the number of transactions in a block from a block matching the given block hash.
  FnEthGetBlockTraute("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("signTransaction", h.Srv.EthSignTransaction)
            })

    // Returns the balance of the account of given address.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getBalance", h.Srv.EthGetBalance)
            })

    // Returns the value from a storage position at a given address.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getStorageAt", h.Srv.EthGetStorageAt)
            })

    // Returns the number of transactions sent from an address.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getTransactionCount", h.Srv.EthGetTransactionCount)
            })

    // Returns code at a given address.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getCode", h.Srv.EthGetCode)
            })

    // Returns the merkle proof for a given account and optionally some storage keys.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getProof", h.Srv.EthGetProof)
            })

    // Signs and submits a transaction.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("sendTransaction", h.Srv.EthSendTransaction)
            })

    // Submits a raw transaction.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("sendRawTransaction", h.Srv.EthSendRawTransaction)
            })

    // Returns the information about a transaction requested by transaction hash.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getTransactionByHash", h.Srv.EthGetTransactionByHash)
            })

    // Returns information about a transaction by block hash and transaction index position.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getTransactionByBlockHashAndIndex", h.Srv.EthGetTransactionByBlockHashAndIndex)
            })

    // Returns information about a transaction by block number and transaction index position.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getTransactionByBlockNumberAndIndex", h.Srv.EthGetTransactionByBlockNumberAndIndex)
            })

    // Returns the receipt of a transaction by transaction hash.

            r.Route("eth", func(r2 jrpc.Router) {
                r.RegisterFunc("getTransactionReceipt", h.Srv.EthGetTransactionReceipt)
            })

    }

type RpcHandler struct {
// Returns an RLP-encoded header.
  FnDebugGetRawHeader func(ctx context.Context,
        Block BlockNumberOrTag,
        )(HeaderRLP Bytes, err error)

// Returns an RLP-encoded block.
  FnDebugGetRawBlock func(ctx context.Context,
        Block BlockNumberOrTag,
        )(BlockRLP Bytes, err error)

// Returns an array of EIP-2718 binary-encoded transactions.
  FnDebugGetRawTransaction func(ctx context.Context,
        TransactionHash Hash32,
        )(EIP2718BinaryEncodedTransaction Bytes, err error)

// Returns an array of EIP-2718 binary-encoded receipts.
  FnDebugGetRawReceipts func(ctx context.Context,
        Block BlockNumberOrTag,
        )(Receipts []Bytes, err error)

// Returns an array of recent bad blocks that the client has seen on the network.
  FnDebugGetBadBlocks func(ctx context.Context,
        )(Blocks []BadBlock, err error)

// Returns information about a block by hash.
  FnEthGetBlockByHash func(ctx context.Context,
        BlockHash Hash32,
        HydratedTransactions bool,
        )(BlockInformation Block, err error)

// Returns information about a block by number.
  FnEthGetBlockByNumber func(ctx context.Context,
        Block BlockNumberOrTag,
        HydratedTransactions bool,
        )(BlockInformation Block, err error)

// Returns the number of transactions in a block from a block matching the given block hash.
  FnEthGetBlockTransactionCountByHash func(ctx context.Context,
        BlockHash *Hash32,
        )(TransactionCount Uint, err error)

// Returns the number of transactions in a block matching the given block number.
  FnEthGetBlockTransactionCountByNumber func(ctx context.Context,
        Block *BlockNumberOrTag,
        )(TransactionCount Uint, err error)

// Returns the number of uncles in a block from a block matching the given block hash.
  FnEthGetUncleCountByBlockHash func(ctx context.Context,
        BlockHash *Hash32,
        )(UncleCount Uint, err error)

// Returns the number of transactions in a block matching the given block number.
  FnEthGetUncleCountByBlockNumber func(ctx context.Context,
        Block *BlockNumberOrTag,
        )(UncleCount Uint, err error)

// Returns the chain ID of the current network.
  FnEthChainId func(ctx context.Context,
        )(ChainID Uint, err error)

// Returns an object with data about the sync status or false.
  FnEthSyncing func(ctx context.Context,
        )(SyncingStatus SyncingStatus, err error)

// Returns the client coinbase address.
  FnEthCoinbase func(ctx context.Context,
        )(CoinbaseAddress Address, err error)

// Returns a list of addresses owned by client.
  FnEthAccounts func(ctx context.Context,
        )(Accounts []Address, err error)

// Returns the number of most recent block.
  FnEthBlockNumber func(ctx context.Context,
        )(BlockNumber Uint, err error)

// Executes a new message call immediately without creating a transaction on the block chain.
  FnEthCall func(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        )(ReturnData Bytes, err error)

// Generates and returns an estimate of how much gas is necessary to allow the transaction to complete.
  FnEthEstimateGas func(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        )(GasUsed Uint, err error)

// Generates an access list for a transaction.
  FnEthCreateAccessList func(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        )(GasUsed struct {
            AccessList AccessList `json:"accessList"`
            Error string `json:"error"`
            GasUsed Uint `json:"gasUsed"`
            }, err error)

// Returns the current price per gas in wei.
  FnEthGasPrice func(ctx context.Context,
        )(GasPrice Uint, err error)

// Returns the current maxPriorityFeePerGas per gas in wei.
  FnEthMaxPriorityFeePerGas func(ctx context.Context,
        )(MaxPriorityFeePerGas Uint, err error)

// Transaction fee history
  FnEthFeeHistory func(ctx context.Context,
        BlockCount Uint,
        NewestBlock BlockNumberOrTag,
        RewardPercentiles []float64,
        )(FeeHistoryResult struct {
            BaseFeePerGas []Uint `json:"baseFeePerGas"`
            OldestBlock Uint `json:"oldestBlock"`
            Reward [][]Uint `json:"reward"`
            }, err error)

// Creates a filter object, based on filter options, to notify when the state changes (logs).
  FnEthNewFilter func(ctx context.Context,
        Filter *Filter,
        )(FilterIdentifier Uint, err error)

// Creates a filter in the node, to notify when a new block arrives.
  FnEthNewBlockFilter func(ctx context.Context,
        )(FilterIdentifier Uint, err error)

// Creates a filter in the node, to notify when new pending transactions arrive.
  FnEthNewPendingTransactionFilter func(ctx context.Context,
        )(FilterIdentifier Uint, err error)

// Uninstalls a filter with given id.
  FnEthUninstallFilter func(ctx context.Context,
        FilterIdentifier *Uint,
        )(Success bool, err error)

// Polling method for a filter, which returns an array of logs which occurred since last poll.
  FnEthGetFilterChanges func(ctx context.Context,
        FilterIdentifier *Uint,
        )(LogObjects FilterResults, err error)

// Returns an array of all logs matching filter with given id.
  FnEthGetFilterLogs func(ctx context.Context,
        FilterIdentifier *Uint,
        )(LogObjects FilterResults, err ernsactionCountByHash func(ctx context.Context,
        BlockHash *Hash32,
        )(TransactionCount Uint, err error)

// Returns the number of transactions in a block matching the given block number.
  FnEthGetBlockTransactionCountByNumber func(ctx context.Context,
        Block *BlockNumberOrTag,
        )(TransactionCount Uint, err error)

// Returns the number of uncles in a block from a block matching the given block hash.
  FnEthGetUncleCountByBlockHash func(ctx context.Context,
        BlockHash *Hash32,
        )(UncleCount Uint, err error)

// Returns the number of transactions in a block matching the given block number.
  FnEthGetUncleCountByBlockNumber func(ctx context.Context,
        Block *BlockNumberOrTag,
        )(UncleCount Uint, err error)

// Returns the chain ID of the current network.
  FnEthChainId func(ctx context.Context,
        )(ChainID Uint, err error)

// Returns an object with data about the sync status or false.
  FnEthSyncing func(ctx context.Context,
        )(SyncingStatus SyncingStatus, err error)

// Returns the client coinbase address.
  FnEthCoinbase func(ctx context.Context,
        )(CoinbaseAddress Address, err error)

// Returns a list of addresses owned by client.
  FnEthAccounts func(ctx context.Context,
        )(Accounts []Address, err error)

// Returns the number of most recent block.
  FnEthBlockNumber func(ctx context.Context,
        )(BlockNumber Uint, err error)

// Executes a new message call immediately without creating a transaction on the block chain.
  FnEthCall func(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        )(ReturnData Bytes, err error)

// Generates and returns an estimate of how much gas is necessary to allow the transaction to complete.
  FnEthEstimateGas func(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        )(GasUsed Uint, err error)

// Generates an access list for a transaction.
  FnEthCreateAccessList func(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        )(GasUsed struct {
            AccessList AccessList `json:"accessList"`
            Error string `json:"error"`
            GasUsed Uint `json:"gasUsed"`
            }, err error)

// Returns the current price per gas in wei.
  FnEthGasPrice func(ctx context.Context,
        )(GasPrice Uint, err error)

// Returns the current maxPriorityFeePerGas per gas in wei.
  FnEthMaxPriorityFeePerGas func(ctx context.Context,
        )(MaxPriorityFeePerGas Uint, err error)

// Transaction fee history
  FnEthFeeHistory func(ctx context.Context,
        BlockCount Uint,
        NewestBlock BlockNumberOrTag,
        RewardPercentiles []float64,
        )(FeeHistoryResult struct {
            BaseFeePerGas []Uint `json:"baseFeePerGas"`
            OldestBlock Uint `json:"oldestBlock"`
            Reward [][]Uint `json:"reward"`
            }, err error)

// Creates a filter object, based on filter options, to notify when the state changes (logs).
  FnEthNewFilter func(ctx context.Context,
        Filter *Filter,
        )(FilterIdentifier Uint, err error)

// Creates a filter in the node, to notify when a new block arrives.
  FnEthNewBlockFilter func(ctx context.Context,
        )(FilterIdentifier Uint, err error)

// Creates a filter in the node, to notify when new pending transactions arrive.
  FnEthNewPendingTransactionFilter func(ctx context.Context,
        )(FilterIdentifier Uint, err error)

// Uninstalls a filter with given id.
  FnEthUninstallFilter func(ctx context.Context,
        FilterIdentifier *Uint,
        )(Success bool, err error)

// Polling method for a filter, which returns an array of logs which occurred since last poll.
  FnEthGetFilterChanges func(ctx context.Context,
        FilterIdentifier *Uint,
        )(LogObjects FilterResults, err error)

// Returns an array of all logs matching filter with given id.
  FnEthGetFilterLogs func(ctx context.Context,
        FilterIdentifier *Uint,
        )(LogObjects FilterResults, err error)

// Returns an array of all logs matching filter with given id.
  FnEthGetLogs func(ctx context.Context,
        Filter *Filter,
        )(LogObjects FilterResults, err error)

// Returns whether the client is actively mining new blocks.
  FnEthMining func(ctx context.Context,
        )(MiningStatus bool, err error)

// Returns the number of hashes per second that the node is mining with.
  FnEthHashrate func(ctx context.Context,
        )(MiningStatus Uint, err error)

// Returns the hash of the current block, the seedHash, and the boundary condition to be met (“target”).
  FnEthGetWork func(ctx context.Context,
        )(CurrentWork []Bytes32, err error)

// Used for submitting a proof-of-work solution.
  FnEthSubmitWork func(ctx context.Context,
        Nonce Bytes8,
        Hash Bytes32,
        Digest Bytes32,
        )(Success bool, err error)

// Used for submitting mining hashrate.
  FnEthSubmitHashrate func(ctx context.Context,
        Hashrate Bytes32,
        Id Bytes32,
        )(Success bool, err error)

// Returns an EIP-191 signature over the provided data.
  FnEthSign func(ctx context.Context,
        Address Address,
        Message Bytes,
        )(Signature Bytes65, err error)

// Returns an RLP encoded transaction signed by the specified account.
  FnEthSignTransaction func(ctx context.Context,
        Transaction GenericTransaction,
        )(EncodedTransaction Bytes, err error)

// Returns the balance of the account of given address.
  FnEthGetBalance func(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        )(Balance Uint, err error)

// Returns the value from a storage position at a given address.
  FnEthGetStorageAt func(ctx context.Context,
        Address Address,
        StorageSlot Uint256,
        Block *BlockNumberOrTag,
        )(Value Bytes, err error)

// Returns the number of transactions sent from an address.
  FnEthGetTransactionCount func(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        )(TransactionCount Uint, err error)

// Returns code at a given address.
  FnEthGetCode func(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        )(Bytecode Bytes, err error)

// Returns the merkle proof for a given account and optionally some storage keys.
  FnEthGetProof func(ctx context.Context,
        Address Address,
        StorageKeys []Hash32,
        Block BlockNumberOrTag,
        )(Account AccountProof, err error)

// Signs and submits a transaction.
  FnEthSendTransaction func(ctx context.Context,
        Transaction GenericTransaction,
        )(TransactionHash Hash32, err error)

// Submits a raw transaction.
  FnEthSendRawTransaction func(ctx context.Context,
        Transaction Bytes,
        )(TransactionHash Hash32, err error)

// Returns the information about a transaction requested by transaction hash.
  FnEthGetTransactionByHash func(ctx context.Context,
        TransactionHash Hash32,
        )(TransactionInformation TransactionInfo, err error)

// Returns information about a transaction by block hash and transaction index position.
  FnEthGetTransactionByBlockHashAndIndex func(ctx context.Context,
        BlockHash Hash32,
        TransactionIndex Uint,
        )(TransactionInformation TransactionInfo, err error)

// Returns information about a transaction by block number and transaction index position.
  FnEthGetTransactionByBlockNumberAndIndex func(ctx context.Context,
        Block BlockNumberOrTag,
        TransactionIndex Uint,
        )(TransactionInformation TransactionInfo, err error)

// Returns the receipt of a transaction by transaction hash.
  FnEthGetTransactionReceipt func(ctx context.Context,
        TransactionHash *Hash32,
        )(ReceiptInformation ReceiptInfo, err error)

}

// Returns an RLP-encoded header.
    func(h *RpcHandler) DebugGetRawHeader(ctx context.Context,
        Block BlockNumberOrTag,
        ) (HeaderRLP Bytes, err error) {
        handler := h.FnDebugGetRawHeader
        return handler(ctx context.Context,
        Block Bror)

// Returns an array of all logs matching filter with given id.
  FnEthGetLogs func(ctx context.Context,
        Filter *Filter,
        )(LogObjects FilterResults, err error)

// Returns whether the client is actively mining new blocks.
  FnEthMining func(ctx context.Context,
        )(MiningStatus bool, err error)

// Returns the number of hashes per second that the node is mining with.
  FnEthHashrate func(ctx context.Context,
        )(MiningStatus Uint, err error)

// Returns the hash of the current block, the seedHash, and the boundary condition to be met (“target”).
  FnEthGetWork func(ctx context.Context,
        )(CurrentWork []Bytes32, err error)

// Used for submitting a proof-of-work solution.
  FnEthSubmitWork func(ctx context.Context,
        Nonce Bytes8,
        Hash Bytes32,
        Digest Bytes32,
        )(Success bool, err error)

// Used for submitting mining hashrate.
  FnEthSubmitHashrate func(ctx context.Context,
        Hashrate Bytes32,
        Id Bytes32,
        )(Success bool, err error)

// Returns an EIP-191 signature over the provided data.
  FnEthSign func(ctx context.Context,
        Address Address,
        Message Bytes,
        )(Signature Bytes65, err error)

// Returns an RLP encoded transaction signed by the specified account.
  FnEthSignTransaction func(ctx context.Context,
        Transaction GenericTransaction,
        )(EncodedTransaction Bytes, err error)

// Returns the balance of the account of given address.
  FnEthGetBalance func(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        )(Balance Uint, err error)

// Returns the value from a storage position at a given address.
  FnEthGetStorageAt func(ctx context.Context,
        Address Address,
        StorageSlot Uint256,
        Block *BlockNumberOrTag,
        )(Value Bytes, err error)

// Returns the number of transactions sent from an address.
  FnEthGetTransactionCount func(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        )(TransactionCount Uint, err error)

// Returns code at a given address.
  FnEthGetCode func(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        )(Bytecode Bytes, err error)

// Returns the merkle proof for a given account and optionally some storage keys.
  FnEthGetProof func(ctx context.Context,
        Address Address,
        StorageKeys []Hash32,
        Block BlockNumberOrTag,
        )(Account AccountProof, err error)

// Signs and submits a transaction.
  FnEthSendTransaction func(ctx context.Context,
        Transaction GenericTransaction,
        )(TransactionHash Hash32, err error)

// Submits a raw transaction.
  FnEthSendRawTransaction func(ctx context.Context,
        Transaction Bytes,
        )(TransactionHash Hash32, err error)

// Returns the information about a transaction requested by transaction hash.
  FnEthGetTransactionByHash func(ctx context.Context,
        TransactionHash Hash32,
        )(TransactionInformation TransactionInfo, err error)

// Returns information about a transaction by block hash and transaction index position.
  FnEthGetTransactionByBlockHashAndIndex func(ctx context.Context,
        BlockHash Hash32,
        TransactionIndex Uint,
        )(TransactionInformation TransactionInfo, err error)

// Returns information about a transaction by block number and transaction index position.
  FnEthGetTransactionByBlockNumberAndIndex func(ctx context.Context,
        Block BlockNumberOrTag,
        TransactionIndex Uint,
        )(TransactionInformation TransactionInfo, err error)

// Returns the receipt of a transaction by transaction hash.
  FnEthGetTransactionReceipt func(ctx context.Context,
        TransactionHash *Hash32,
        )(ReceiptInformation ReceiptInfo, err error)

}

// Returns an RLP-encoded header.
    func(h *RpcHandler) DebugGetRawHeader(ctx context.Context,
        Block BlockNumberOrTag,
        ) (HeaderRLP Bytes, err error) {
        handler := h.FnDebugGetRawHeader
        return handler(ctx context.Context,
        Block BlockNumberOrTag,
        )
    }
// Returns an RLP-encoded block.
    func(h *RpcHandler) DebugGetRawBlock(ctx context.Context,
        Block BlockNumberOrTag,
        ) (BlockRLP Bytes, err error) {
        handler := h.FnDebugGetRawBlock
        return handler(ctx context.Context,
        Block BlockNumberOrTag,
        )
    }
// Returns an array of EIP-2718 binary-encoded transactions.
    func(h *RpcHandler) DebugGetRawTransaction(ctx context.Context,
        TransactionHash Hash32,
        ) (EIP2718BinaryEncodedTransaction Bytes, err error) {
        handler := h.FnDebugGetRawTransaction
        return handler(ctx context.Context,
        TransactionHash Hash32,
        )
    }
// Returns an array of EIP-2718 binary-encoded receipts.
    func(h *RpcHandler) DebugGetRawReceipts(ctx context.Context,
        Block BlockNumberOrTag,
        ) (Receipts []Bytes, err error) {
        handler := h.FnDebugGetRawReceipts
        return handler(ctx context.Context,
        Block BlockNumberOrTag,
        )
    }
// Returns an array of recent bad blocks that the client has seen on the network.
    func(h *RpcHandler) DebugGetBadBlocks(ctx context.Context,
        ) (Blocks []BadBlock, err error) {
        handler := h.FnDebugGetBadBlocks
        return handler(ctx context.Context,
        )
    }
// Returns information about a block by hash.
    func(h *RpcHandler) EthGetBlockByHash(ctx context.Context,
        BlockHash Hash32,
        HydratedTransactions bool,
        ) (BlockInformation Block, err error) {
        handler := h.FnEthGetBlockByHash
        return handler(ctx context.Context,
        BlockHash Hash32,
        HydratedTransactions bool,
        )
    }
// Returns information about a block by number.
    func(h *RpcHandler) EthGetBlockByNumber(ctx context.Context,
        Block BlockNumberOrTag,
        HydratedTransactions bool,
        ) (BlockInformation Block, err error) {
        handler := h.FnEthGetBlockByNumber
        return handler(ctx context.Context,
        Block BlockNumberOrTag,
        HydratedTransactions bool,
        )
    }
// Returns the number of transactions in a block from a block matching the given block hash.
    func(h *RpcHandler) EthGetBlockTransactionCountByHash(ctx context.Context,
        BlockHash *Hash32,
        ) (TransactionCount Uint, err error) {
        handler := h.FnEthGetBlockTransactionCountByHash
        return handler(ctx context.Context,
        BlockHash *Hash32,
        )
    }
// Returns the number of transactions in a block matching the given block number.
    func(h *RpcHandler) EthGetBlockTransactionCountByNumber(ctx context.Context,
        Block *BlockNumberOrTag,
        ) (TransactionCount Uint, err error) {
        handler := h.FnEthGetBlockTransactionCountByNumber
        return handler(ctx context.Context,
        Block *BlockNumberOrTag,
        )
    }
// Returns the number of uncles in a block from a block matching the given block hash.
    func(h *RpcHandler) EthGetUncleCountByBlockHash(ctx context.Context,
        BlockHash *Hash32,
        ) (UncleCount Uint, err error) {
        handler := h.FnEthGetUncleCountByBlockHash
        return handler(ctx context.Context,
        BlockHash *Hash32,
        )
    }
// Returns the number of transactions in a block matching the given block number.
    func(h *RpcHandler) EthGetUncleCountByBlockNumber(ctx context.Context,
        Block *BlockNumberOrTag,
        ) (UncleCount Uint, err error) {
        handler := h.FnEthGetUncleCountByBlockNumber
        return handler(ctx context.Context,
        Block *BlockNumberOrTag,
        )
    }
// Returns the chain ID of the current network.
    func(h *RpcHandler) EthChainId(ctx context.Context,
        ) (ChainID Uint, err error) {
        handler := h.FnEthChainId
        return handler(ctx context.Context,
        )
    }
// Returns an object with data about the sync status or false.
    func(h *RpcHandler) EthSyncing(ctx context.Context,
        ) (SyncingStatus SyncingStatus, err error) {
        handler := h.FnEthSyncing
        return handler(ctx conlockNumberOrTag,
        )
    }
// Returns an RLP-encoded block.
    func(h *RpcHandler) DebugGetRawBlock(ctx context.Context,
        Block BlockNumberOrTag,
        ) (BlockRLP Bytes, err error) {
        handler := h.FnDebugGetRawBlock
        return handler(ctx context.Context,
        Block BlockNumberOrTag,
        )
    }
// Returns an array of EIP-2718 binary-encoded transactions.
    func(h *RpcHandler) DebugGetRawTransaction(ctx context.Context,
        TransactionHash Hash32,
        ) (EIP2718BinaryEncodedTransaction Bytes, err error) {
        handler := h.FnDebugGetRawTransaction
        return handler(ctx context.Context,
        TransactionHash Hash32,
        )
    }
// Returns an array of EIP-2718 binary-encoded receipts.
    func(h *RpcHandler) DebugGetRawReceipts(ctx context.Context,
        Block BlockNumberOrTag,
        ) (Receipts []Bytes, err error) {
        handler := h.FnDebugGetRawReceipts
        return handler(ctx context.Context,
        Block BlockNumberOrTag,
        )
    }
// Returns an array of recent bad blocks that the client has seen on the network.
    func(h *RpcHandler) DebugGetBadBlocks(ctx context.Context,
        ) (Blocks []BadBlock, err error) {
        handler := h.FnDebugGetBadBlocks
        return handler(ctx context.Context,
        )
    }
// Returns information about a block by hash.
    func(h *RpcHandler) EthGetBlockByHash(ctx context.Context,
        BlockHash Hash32,
        HydratedTransactions bool,
        ) (BlockInformation Block, err error) {
        handler := h.FnEthGetBlockByHash
        return handler(ctx context.Context,
        BlockHash Hash32,
        HydratedTransactions bool,
        )
    }
// Returns information about a block by number.
    func(h *RpcHandler) EthGetBlockByNumber(ctx context.Context,
        Block BlockNumberOrTag,
        HydratedTransactions bool,
        ) (BlockInformation Block, err error) {
        handler := h.FnEthGetBlockByNumber
        return handler(ctx context.Context,
        Block BlockNumberOrTag,
        HydratedTransactions bool,
        )
    }
// Returns the number of transactions in a block from a block matching the given block hash.
    func(h *RpcHandler) EthGetBlockTransactionCountByHash(ctx context.Context,
        BlockHash *Hash32,
        ) (TransactionCount Uint, err error) {
        handler := h.FnEthGetBlockTransactionCountByHash
        return handler(ctx context.Context,
        BlockHash *Hash32,
        )
    }
// Returns the number of transactions in a block matching the given block number.
    func(h *RpcHandler) EthGetBlockTransactionCountByNumber(ctx context.Context,
        Block *BlockNumberOrTag,
        ) (TransactionCount Uint, err error) {
        handler := h.FnEthGetBlockTransactionCountByNumber
        return handler(ctx context.Context,
        Block *BlockNumberOrTag,
        )
    }
// Returns the number of uncles in a block from a block matching the given block hash.
    func(h *RpcHandler) EthGetUncleCountByBlockHash(ctx context.Context,
        BlockHash *Hash32,
        ) (UncleCount Uint, err error) {
        handler := h.FnEthGetUncleCountByBlockHash
        return handler(ctx context.Context,
        BlockHash *Hash32,
        )
    }
// Returns the number of transactions in a block matching the given block number.
    func(h *RpcHandler) EthGetUncleCountByBlockNumber(ctx context.Context,
        Block *BlockNumberOrTag,
        ) (UncleCount Uint, err error) {
        handler := h.FnEthGetUncleCountByBlockNumber
        return handler(ctx context.Context,
        Block *BlockNumberOrTag,
        )
    }
// Returns the chain ID of the current network.
    func(h *RpcHandler) EthChainId(ctx context.Context,
        ) (ChainID Uint, err error) {
        handler := h.FnEthChainId
        return handler(ctx context.Context,
        )
    }
// Returns an object with data about the sync status or false.
    func(h *RpcHandler) EthSyncing(ctx context.Context,
        ) (SyncingStatus SyncingStatus, err error) {
        handler := h.FnEthSyncing
        return handler(ctx context.Context,
        )
    }
// Returns the client coinbase address.
    func(h *RpcHandler) EthCoinbase(ctx context.Context,
        ) (CoinbaseAddress Address, err error) {
        handler := h.FnEthCoinbase
        return handler(ctx context.Context,
        )
    }
// Returns a list of addresses owned by client.
    func(h *RpcHandler) EthAccounts(ctx context.Context,
        ) (Accounts []Address, err error) {
        handler := h.FnEthAccounts
        return handler(ctx context.Context,
        )
    }
// Returns the number of most recent block.
    func(h *RpcHandler) EthBlockNumber(ctx context.Context,
        ) (BlockNumber Uint, err error) {
        handler := h.FnEthBlockNumber
        return handler(ctx context.Context,
        )
    }
// Executes a new message call immediately without creating a transaction on the block chain.
    func(h *RpcHandler) EthCall(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        ) (ReturnData Bytes, err error) {
        handler := h.FnEthCall
        return handler(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        )
    }
// Generates and returns an estimate of how much gas is necessary to allow the transaction to complete.
    func(h *RpcHandler) EthEstimateGas(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        ) (GasUsed Uint, err error) {
        handler := h.FnEthEstimateGas
        return handler(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        )
    }
// Generates an access list for a transaction.
    func(h *RpcHandler) EthCreateAccessList(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        ) (GasUsed struct {
            AccessList AccessList `json:"accessList"`
            Error string `json:"error"`
            GasUsed Uint `json:"gasUsed"`
            }, err error) {
        handler := h.FnEthCreateAccessList
        return handler(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        )
    }
// Returns the current price per gas in wei.
    func(h *RpcHandler) EthGasPrice(ctx context.Context,
        ) (GasPrice Uint, err error) {
        handler := h.FnEthGasPrice
        return handler(ctx context.Context,
        )
    }
// Returns the current maxPriorityFeePerGas per gas in wei.
    func(h *RpcHandler) EthMaxPriorityFeePerGas(ctx context.Context,
        ) (MaxPriorityFeePerGas Uint, err error) {
        handler := h.FnEthMaxPriorityFeePerGas
        return handler(ctx context.Context,
        )
    }
// Transaction fee history
    func(h *RpcHandler) EthFeeHistory(ctx context.Context,
        BlockCount Uint,
        NewestBlock BlockNumberOrTag,
        RewardPercentiles []float64,
        ) (FeeHistoryResult struct {
            BaseFeePerGas []Uint `json:"baseFeePerGas"`
            OldestBlock Uint `json:"oldestBlock"`
            Reward [][]Uint `json:"reward"`
            }, err error) {
        handler := h.FnEthFeeHistory
        return handler(ctx context.Context,
        BlockCount Uint,
        NewestBlock BlockNumberOrTag,
        RewardPercentiles []float64,
        )
    }
// Creates a filter object, based on filter options, to notify when the state changes (logs).
    func(h *RpcHandler) EthNewFilter(ctx context.Context,
        Filter *Filter,
        ) (FilterIdentifier Uint, err error) {
        handler := h.FnEthNewFilter
        return handler(ctx context.Context,
        Filter *Filter,
        )
    }
// Creates a filter in the node, to notify when a new block arrives.
    func(h *RpcHandler) EthNewBlockFilter(ctx context.Context,
        ) (FilterIdentifier Uint, err error) {
        handler := h.FnEthNewBlockFilter
        return handler(ctx context.Context,
        )
    }
// Creates a filter in the node, to notify when new pending transactions arrive.
    func(h *RpcHandler) EthNewPendingTransactionFilter(ctx context.Context,
       text.Context,
        )
    }
// Returns the client coinbase address.
    func(h *RpcHandler) EthCoinbase(ctx context.Context,
        ) (CoinbaseAddress Address, err error) {
        handler := h.FnEthCoinbase
        return handler(ctx context.Context,
        )
    }
// Returns a list of addresses owned by client.
    func(h *RpcHandler) EthAccounts(ctx context.Context,
        ) (Accounts []Address, err error) {
        handler := h.FnEthAccounts
        return handler(ctx context.Context,
        )
    }
// Returns the number of most recent block.
    func(h *RpcHandler) EthBlockNumber(ctx context.Context,
        ) (BlockNumber Uint, err error) {
        handler := h.FnEthBlockNumber
        return handler(ctx context.Context,
        )
    }
// Executes a new message call immediately without creating a transaction on the block chain.
    func(h *RpcHandler) EthCall(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        ) (ReturnData Bytes, err error) {
        handler := h.FnEthCall
        return handler(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        )
    }
// Generates and returns an estimate of how much gas is necessary to allow the transaction to complete.
    func(h *RpcHandler) EthEstimateGas(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        ) (GasUsed Uint, err error) {
        handler := h.FnEthEstimateGas
        return handler(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        )
    }
// Generates an access list for a transaction.
    func(h *RpcHandler) EthCreateAccessList(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        ) (GasUsed struct {
            AccessList AccessList `json:"accessList"`
            Error string `json:"error"`
            GasUsed Uint `json:"gasUsed"`
            }, err error) {
        handler := h.FnEthCreateAccessList
        return handler(ctx context.Context,
        Transaction GenericTransaction,
        Block *BlockNumberOrTag,
        )
    }
// Returns the current price per gas in wei.
    func(h *RpcHandler) EthGasPrice(ctx context.Context,
        ) (GasPrice Uint, err error) {
        handler := h.FnEthGasPrice
        return handler(ctx context.Context,
        )
    }
// Returns the current maxPriorityFeePerGas per gas in wei.
    func(h *RpcHandler) EthMaxPriorityFeePerGas(ctx context.Context,
        ) (MaxPriorityFeePerGas Uint, err error) {
        handler := h.FnEthMaxPriorityFeePerGas
        return handler(ctx context.Context,
        )
    }
// Transaction fee history
    func(h *RpcHandler) EthFeeHistory(ctx context.Context,
        BlockCount Uint,
        NewestBlock BlockNumberOrTag,
        RewardPercentiles []float64,
        ) (FeeHistoryResult struct {
            BaseFeePerGas []Uint `json:"baseFeePerGas"`
            OldestBlock Uint `json:"oldestBlock"`
            Reward [][]Uint `json:"reward"`
            }, err error) {
        handler := h.FnEthFeeHistory
        return handler(ctx context.Context,
        BlockCount Uint,
        NewestBlock BlockNumberOrTag,
        RewardPercentiles []float64,
        )
    }
// Creates a filter object, based on filter options, to notify when the state changes (logs).
    func(h *RpcHandler) EthNewFilter(ctx context.Context,
        Filter *Filter,
        ) (FilterIdentifier Uint, err error) {
        handler := h.FnEthNewFilter
        return handler(ctx context.Context,
        Filter *Filter,
        )
    }
// Creates a filter in the node, to notify when a new block arrives.
    func(h *RpcHandler) EthNewBlockFilter(ctx context.Context,
        ) (FilterIdentifier Uint, err error) {
        handler := h.FnEthNewBlockFilter
        return handler(ctx context.Context,
        )
    }
// Creates a filter in the node, to notify when new pending transactions arrive.
    func(h *RpcHandler) EthNewPendingTransactionFilter(ctx context.Context,
        ) (FilterIdentifier Uint, err error) {
        handler := h.FnEthNewPendingTransactionFilter
        return handler(ctx context.Context,
        )
    }
// Uninstalls a filter with given id.
    func(h *RpcHandler) EthUninstallFilter(ctx context.Context,
        FilterIdentifier *Uint,
        ) (Success bool, err error) {
        handler := h.FnEthUninstallFilter
        return handler(ctx context.Context,
        FilterIdentifier *Uint,
        )
    }
// Polling method for a filter, which returns an array of logs which occurred since last poll.
    func(h *RpcHandler) EthGetFilterChanges(ctx context.Context,
        FilterIdentifier *Uint,
        ) (LogObjects FilterResults, err error) {
        handler := h.FnEthGetFilterChanges
        return handler(ctx context.Context,
        FilterIdentifier *Uint,
        )
    }
// Returns an array of all logs matching filter with given id.
    func(h *RpcHandler) EthGetFilterLogs(ctx context.Context,
        FilterIdentifier *Uint,
        ) (LogObjects FilterResults, err error) {
        handler := h.FnEthGetFilterLogs
        return handler(ctx context.Context,
        FilterIdentifier *Uint,
        )
    }
// Returns an array of all logs matching filter with given id.
    func(h *RpcHandler) EthGetLogs(ctx context.Context,
        Filter *Filter,
        ) (LogObjects FilterResults, err error) {
        handler := h.FnEthGetLogs
        return handler(ctx context.Context,
        Filter *Filter,
        )
    }
// Returns whether the client is actively mining new blocks.
    func(h *RpcHandler) EthMining(ctx context.Context,
        ) (MiningStatus bool, err error) {
        handler := h.FnEthMining
        return handler(ctx context.Context,
        )
    }
// Returns the number of hashes per second that the node is mining with.
    func(h *RpcHandler) EthHashrate(ctx context.Context,
        ) (MiningStatus Uint, err error) {
        handler := h.FnEthHashrate
        return handler(ctx context.Context,
        )
    }
// Returns the hash of the current block, the seedHash, and the boundary condition to be met (“target”).
    func(h *RpcHandler) EthGetWork(ctx context.Context,
        ) (CurrentWork []Bytes32, err error) {
        handler := h.FnEthGetWork
        return handler(ctx context.Context,
        )
    }
// Used for submitting a proof-of-work solution.
    func(h *RpcHandler) EthSubmitWork(ctx context.Context,
        Nonce Bytes8,
        Hash Bytes32,
        Digest Bytes32,
        ) (Success bool, err error) {
        handler := h.FnEthSubmitWork
        return handler(ctx context.Context,
        Nonce Bytes8,
        Hash Bytes32,
        Digest Bytes32,
        )
    }
// Used for submitting mining hashrate.
    func(h *RpcHandler) EthSubmitHashrate(ctx context.Context,
        Hashrate Bytes32,
        Id Bytes32,
        ) (Success bool, err error) {
        handler := h.FnEthSubmitHashrate
        return handler(ctx context.Context,
        Hashrate Bytes32,
        Id Bytes32,
        )
    }
// Returns an EIP-191 signature over the provided data.
    func(h *RpcHandler) EthSign(ctx context.Context,
        Address Address,
        Message Bytes,
        ) (Signature Bytes65, err error) {
        handler := h.FnEthSign
        return handler(ctx context.Context,
        Address Address,
        Message Bytes,
        )
    }
// Returns an RLP encoded transaction signed by the specified account.
    func(h *RpcHandler) EthSignTransaction(ctx context.Context,
        Transaction GenericTransaction,
        ) (EncodedTransaction Bytes, err error) {
        handler := h.FnEthSignTransaction
        return handler(ctx context.Context,
        Transaction GenericTransaction,
        )
    }
// Returns the balance of the account of given address.
    func(h *RpcHandler) EthGetBalance(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        ) (Balance Uint, err error) {
        handler := h.FnEthGetBalance
        return handler(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        )
 ) (FilterIdentifier Uint, err error) {
        handler := h.FnEthNewPendingTransactionFilter
        return handler(ctx context.Context,
        )
    }
// Uninstalls a filter with given id.
    func(h *RpcHandler) EthUninstallFilter(ctx context.Context,
        FilterIdentifier *Uint,
        ) (Success bool, err error) {
        handler := h.FnEthUninstallFilter
        return handler(ctx context.Context,
        FilterIdentifier *Uint,
        )
    }
// Polling method for a filter, which returns an array of logs which occurred since last poll.
    func(h *RpcHandler) EthGetFilterChanges(ctx context.Context,
        FilterIdentifier *Uint,
        ) (LogObjects FilterResults, err error) {
        handler := h.FnEthGetFilterChanges
        return handler(ctx context.Context,
        FilterIdentifier *Uint,
        )
    }
// Returns an array of all logs matching filter with given id.
    func(h *RpcHandler) EthGetFilterLogs(ctx context.Context,
        FilterIdentifier *Uint,
        ) (LogObjects FilterResults, err error) {
        handler := h.FnEthGetFilterLogs
        return handler(ctx context.Context,
        FilterIdentifier *Uint,
        )
    }
// Returns an array of all logs matching filter with given id.
    func(h *RpcHandler) EthGetLogs(ctx context.Context,
        Filter *Filter,
        ) (LogObjects FilterResults, err error) {
        handler := h.FnEthGetLogs
        return handler(ctx context.Context,
        Filter *Filter,
        )
    }
// Returns whether the client is actively mining new blocks.
    func(h *RpcHandler) EthMining(ctx context.Context,
        ) (MiningStatus bool, err error) {
        handler := h.FnEthMining
        return handler(ctx context.Context,
        )
    }
// Returns the number of hashes per second that the node is mining with.
    func(h *RpcHandler) EthHashrate(ctx context.Context,
        ) (MiningStatus Uint, err error) {
        handler := h.FnEthHashrate
        return handler(ctx context.Context,
        )
    }
// Returns the hash of the current block, the seedHash, and the boundary condition to be met (“target”).
    func(h *RpcHandler) EthGetWork(ctx context.Context,
        ) (CurrentWork []Bytes32, err error) {
        handler := h.FnEthGetWork
        return handler(ctx context.Context,
        )
    }
// Used for submitting a proof-of-work solution.
    func(h *RpcHandler) EthSubmitWork(ctx context.Context,
        Nonce Bytes8,
        Hash Bytes32,
        Digest Bytes32,
        ) (Success bool, err error) {
        handler := h.FnEthSubmitWork
        return handler(ctx context.Context,
        Nonce Bytes8,
        Hash Bytes32,
        Digest Bytes32,
        )
    }
// Used for submitting mining hashrate.
    func(h *RpcHandler) EthSubmitHashrate(ctx context.Context,
        Hashrate Bytes32,
        Id Bytes32,
        ) (Success bool, err error) {
        handler := h.FnEthSubmitHashrate
        return handler(ctx context.Context,
        Hashrate Bytes32,
        Id Bytes32,
        )
    }
// Returns an EIP-191 signature over the provided data.
    func(h *RpcHandler) EthSign(ctx context.Context,
        Address Address,
        Message Bytes,
        ) (Signature Bytes65, err error) {
        handler := h.FnEthSign
        return handler(ctx context.Context,
        Address Address,
        Message Bytes,
        )
    }
// Returns an RLP encoded transaction signed by the specified account.
    func(h *RpcHandler) EthSignTransaction(ctx context.Context,
        Transaction GenericTransaction,
        ) (EncodedTransaction Bytes, err error) {
        handler := h.FnEthSignTransaction
        return handler(ctx context.Context,
        Transaction GenericTransaction,
        )
    }
// Returns the balance of the account of given address.
    func(h *RpcHandler) EthGetBalance(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        ) (Balance Uint, err error) {
        handler := h.FnEthGetBalance
        return handler(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        )
    }
// Returns the value from a storage position at a given address.
    func(h *RpcHandler) EthGetStorageAt(ctx context.Context,
        Address Address,
        StorageSlot Uint256,
        Block *BlockNumberOrTag,
        ) (Value Bytes, err error) {
        handler := h.FnEthGetStorageAt
        return handler(ctx context.Context,
        Address Address,
        StorageSlot Uint256,
        Block *BlockNumberOrTag,
        )
    }
// Returns the number of transactions sent from an address.
    func(h *RpcHandler) EthGetTransactionCount(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        ) (TransactionCount Uint, err error) {
        handler := h.FnEthGetTransactionCount
        return handler(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        )
    }
// Returns code at a given address.
    func(h *RpcHandler) EthGetCode(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        ) (Bytecode Bytes, err error) {
        handler := h.FnEthGetCode
        return handler(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        )
    }
// Returns the merkle proof for a given account and optionally some storage keys.
    func(h *RpcHandler) EthGetProof(ctx context.Context,
        Address Address,
        StorageKeys []Hash32,
        Block BlockNumberOrTag,
        ) (Account AccountProof, err error) {
        handler := h.FnEthGetProof
        return handler(ctx context.Context,
        Address Address,
        StorageKeys []Hash32,
        Block BlockNumberOrTag,
        )
    }
// Signs and submits a transaction.
    func(h *RpcHandler) EthSendTransaction(ctx context.Context,
        Transaction GenericTransaction,
        ) (TransactionHash Hash32, err error) {
        handler := h.FnEthSendTransaction
        return handler(ctx context.Context,
        Transaction GenericTransaction,
        )
    }
// Submits a raw transaction.
    func(h *RpcHandler) EthSendRawTransaction(ctx context.Context,
        Transaction Bytes,
        ) (TransactionHash Hash32, err error) {
        handler := h.FnEthSendRawTransaction
        return handler(ctx context.Context,
        Transaction Bytes,
        )
    }
// Returns the information about a transaction requested by transaction hash.
    func(h *RpcHandler) EthGetTransactionByHash(ctx context.Context,
        TransactionHash Hash32,
        ) (TransactionInformation TransactionInfo, err error) {
        handler := h.FnEthGetTransactionByHash
        return handler(ctx context.Context,
        TransactionHash Hash32,
        )
    }
// Returns information about a transaction by block hash and transaction index position.
    func(h *RpcHandler) EthGetTransactionByBlockHashAndIndex(ctx context.Context,
        BlockHash Hash32,
        TransactionIndex Uint,
        ) (TransactionInformation TransactionInfo, err error) {
        handler := h.FnEthGetTransactionByBlockHashAndIndex
        return handler(ctx context.Context,
        BlockHash Hash32,
        TransactionIndex Uint,
        )
    }
// Returns information about a transaction by block number and transaction index position.
    func(h *RpcHandler) EthGetTransactionByBlockNumberAndIndex(ctx context.Context,
        Block BlockNumberOrTag,
        TransactionIndex Uint,
        ) (TransactionInformation TransactionInfo, err error) {
        handler := h.FnEthGetTransactionByBlockNumberAndIndex
        return handler(ctx context.Context,
        Block BlockNumberOrTag,
        TransactionIndex Uint,
        )
    }
// Returns the receipt of a transaction by transaction hash.
    func(h *RpcHandler) EthGetTransactionReceipt(ctx context.Context,
        TransactionHash *Hash32,
        ) (ReceiptInformation ReceiptInfo, err error) {
        handler := h.FnEthGetTransactionReceipt
        return handler(ctx context.Context,
        TransactionHash *Hash32,
        )
    }
type GoOpenRPCService interface {
    // Returns an RLP-encoded header.
        DebugGetRawHeader(ctx context.Context,
         }
// Returns the value from a storage position at a given address.
    func(h *RpcHandler) EthGetStorageAt(ctx context.Context,
        Address Address,
        StorageSlot Uint256,
        Block *BlockNumberOrTag,
        ) (Value Bytes, err error) {
        handler := h.FnEthGetStorageAt
        return handler(ctx context.Context,
        Address Address,
        StorageSlot Uint256,
        Block *BlockNumberOrTag,
        )
    }
// Returns the number of transactions sent from an address.
    func(h *RpcHandler) EthGetTransactionCount(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        ) (TransactionCount Uint, err error) {
        handler := h.FnEthGetTransactionCount
        return handler(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        )
    }
// Returns code at a given address.
    func(h *RpcHandler) EthGetCode(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        ) (Bytecode Bytes, err error) {
        handler := h.FnEthGetCode
        return handler(ctx context.Context,
        Address Address,
        Block *BlockNumberOrTag,
        )
    }
// Returns the merkle proof for a given account and optionally some storage keys.
    func(h *RpcHandler) EthGetProof(ctx context.Context,
        Address Address,
        StorageKeys []Hash32,
        Block BlockNumberOrTag,
        ) (Account AccountProof, err error) {
        handler := h.FnEthGetProof
        return handler(ctx context.Context,
        Address Address,
        StorageKeys []Hash32,
        Block BlockNumberOrTag,
        )
    }
// Signs and submits a transaction.
    func(h *RpcHandler) EthSendTransaction(ctx context.Context,
        Transaction GenericTransaction,
        ) (TransactionHash Hash32, err error) {
        handler := h.FnEthSendTransaction
        return handler(ctx context.Context,
        Transaction GenericTransaction,
        )
    }
// Submits a raw transaction.
    func(h *RpcHandler) EthSendRawTransaction(ctx context.Context,
        Transaction Bytes,
        ) (TransactionHash Hash32, err error) {
        handler := h.FnEthSendRawTransaction
        return handler(ctx context.Context,
        Transaction Bytes,
        )
    }
// Returns the information about a transaction requested by transaction hash.
    func(h *RpcHandler) EthGetTransactionByHash(ctx context.Context,
        TransactionHash Hash32,
        ) (TransactionInformation TransactionInfo, err error) {
        handler := h.FnEthGetTransactionByHash
        return handler(ctx context.Context,
        TransactionHash Hash32,
        )
    }
// Returns information about a transaction by block hash and transaction index position.
    func(h *RpcHandler) EthGetTransactionByBlockHashAndIndex(ctx context.Context,
        BlockHash Hash32,
        TransactionIndex Uint,
        ) (TransactionInformation TransactionInfo, err error) {
        handler := h.FnEthGetTransactionByBlockHashAndIndex
        return handler(ctx context.Context,
        BlockHash Hash32,
        TransactionIndex Uint,
        )
    }
// Returns information about a transaction by block number and transaction index position.
    func(h *RpcHandler) EthGetTransactionByBlockNumberAndIndex(ctx context.Context,
        Block BlockNumberOrTag,
        TransactionIndex Uint,
        ) (TransactionInformation TransactionInfo, err error) {
        handler := h.FnEthGetTransactionByBlockNumberAndIndex
        return handler(ctx context.Context,
        Block BlockNumberOrTag,
        TransactionIndex Uint,
        )
    }
// Returns the receipt of a transaction by transaction hash.
    func(h *RpcHandler) EthGetTransactionReceipt(ctx context.Context,
        TransactionHash *Hash32,
        ) (ReceiptInformation ReceiptInfo, err error) {
        handler := h.FnEthGetTransactionReceipt
        return handler(ctx context.Context,
        TransactionHash *Hash32,
        )
    }
type GoOpenRPCService interface {
    // Returns an RLP-encoded header.
        DebugGetRawHeader(ctx context.Context,
            Block BlockNumberOrTag,
            ) (HeaderRLP Bytes, err error)
    // Returns an RLP-encoded block.
        DebugGetRawBlock(ctx context.Context,
            Block BlockNumberOrTag,
            ) (BlockRLP Bytes, err error)
    // Returns an array of EIP-2718 binary-encoded transactions.
        DebugGetRawTransaction(ctx context.Context,
            TransactionHash Hash32,
            ) (EIP2718BinaryEncodedTransaction Bytes, err error)
    // Returns an array of EIP-2718 binary-encoded receipts.
        DebugGetRawReceipts(ctx context.Context,
            Block BlockNumberOrTag,
            ) (Receipts []Bytes, err error)
    // Returns an array of recent bad blocks that the client has seen on the network.
        DebugGetBadBlocks(ctx context.Context,
            ) (Blocks []BadBlock, err error)
    // Returns information about a block by hash.
        EthGetBlockByHash(ctx context.Context,
            BlockHash Hash32,
            HydratedTransactions bool,
            ) (BlockInformation Block, err error)
    // Returns information about a block by number.
        EthGetBlockByNumber(ctx context.Context,
            Block BlockNumberOrTag,
            HydratedTransactions bool,
            ) (BlockInformation Block, err error)
    // Returns the number of transactions in a block from a block matching the given block hash.
        EthGetBlockTransactionCountByHash(ctx context.Context,
            BlockHash *Hash32,
            ) (TransactionCount Uint, err error)
    // Returns the number of transactions in a block matching the given block number.
        EthGetBlockTransactionCountByNumber(ctx context.Context,
            Block *BlockNumberOrTag,
            ) (TransactionCount Uint, err error)
    // Returns the number of uncles in a block from a block matching the given block hash.
        EthGetUncleCountByBlockHash(ctx context.Context,
            BlockHash *Hash32,
            ) (UncleCount Uint, err error)
    // Returns the number of transactions in a block matching the given block number.
        EthGetUncleCountByBlockNumber(ctx context.Context,
            Block *BlockNumberOrTag,
            ) (UncleCount Uint, err error)
    // Returns the chain ID of the current network.
        EthChainId(ctx context.Context,
            ) (ChainID Uint, err error)
    // Returns an object with data about the sync status or false.
        EthSyncing(ctx context.Context,
            ) (SyncingStatus SyncingStatus, err error)
    // Returns the client coinbase address.
        EthCoinbase(ctx context.Context,
            ) (CoinbaseAddress Address, err error)
    // Returns a list of addresses owned by client.
        EthAccounts(ctx context.Context,
            ) (Accounts []Address, err error)
    // Returns the number of most recent block.
        EthBlockNumber(ctx context.Context,
            ) (BlockNumber Uint, err error)
    // Executes a new message call immediately without creating a transaction on the block chain.
        EthCall(ctx context.Context,
            Transaction GenericTransaction,
            Block *BlockNumberOrTag,
            ) (ReturnData Bytes, err error)
    // Generates and returns an estimate of how much gas is necessary to allow the transaction to complete.
        EthEstimateGas(ctx context.Context,
            Transaction GenericTransaction,
            Block *BlockNumberOrTag,
            ) (GasUsed Uint, err error)
    // Generates an access list for a transaction.
        EthCreateAccessList(ctx context.Context,
            Transaction GenericTransaction,
            Block *BlockNumberOrTag,
            ) (GasUsed struct {
            AccessList AccessList `json:"accessList"`
            Error string `json:"error"`
            GasUsed Uint `json:"gasUsed"`
            }, err error)
    // Returns the current price per gas in wei.
        EthGasPrice(ctx context.Context,
            ) (GasPrice Uint, err error)
    // Returns the current maxPriorityFeePerGas per gas in wei.
        EthMaxPriorityFeePerGas(ctx context.Context,
            ) (MaxPriorityFeePerGas Uint, er       Block BlockNumberOrTag,
            ) (HeaderRLP Bytes, err error)
    // Returns an RLP-encoded block.
        DebugGetRawBlock(ctx context.Context,
            Block BlockNumberOrTag,
            ) (BlockRLP Bytes, err error)
    // Returns an array of EIP-2718 binary-encoded transactions.
        DebugGetRawTransaction(ctx context.Context,
            TransactionHash Hash32,
            ) (EIP2718BinaryEncodedTransaction Bytes, err error)
    // Returns an array of EIP-2718 binary-encoded receipts.
        DebugGetRawReceipts(ctx context.Context,
            Block BlockNumberOrTag,
            ) (Receipts []Bytes, err error)
    // Returns an array of recent bad blocks that the client has seen on the network.
        DebugGetBadBlocks(ctx context.Context,
            ) (Blocks []BadBlock, err error)
    // Returns information about a block by hash.
        EthGetBlockByHash(ctx context.Context,
            BlockHash Hash32,
            HydratedTransactions bool,
            ) (BlockInformation Block, err error)
    // Returns information about a block by number.
        EthGetBlockByNumber(ctx context.Context,
            Block BlockNumberOrTag,
            HydratedTransactions bool,
            ) (BlockInformation Block, err error)
    // Returns the number of transactions in a block from a block matching the given block hash.
        EthGetBlockTransactionCountByHash(ctx context.Context,
            BlockHash *Hash32,
            ) (TransactionCount Uint, err error)
    // Returns the number of transactions in a block matching the given block number.
        EthGetBlockTransactionCountByNumber(ctx context.Context,
            Block *BlockNumberOrTag,
            ) (TransactionCount Uint, err error)
    // Returns the number of uncles in a block from a block matching the given block hash.
        EthGetUncleCountByBlockHash(ctx context.Context,
            BlockHash *Hash32,
            ) (UncleCount Uint, err error)
    // Returns the number of transactions in a block matching the given block number.
        EthGetUncleCountByBlockNumber(ctx context.Context,
            Block *BlockNumberOrTag,
            ) (UncleCount Uint, err error)
    // Returns the chain ID of the current network.
        EthChainId(ctx context.Context,
            ) (ChainID Uint, err error)
    // Returns an object with data about the sync status or false.
        EthSyncing(ctx context.Context,
            ) (SyncingStatus SyncingStatus, err error)
    // Returns the client coinbase address.
        EthCoinbase(ctx context.Context,
            ) (CoinbaseAddress Address, err error)
    // Returns a list of addresses owned by client.
        EthAccounts(ctx context.Context,
            ) (Accounts []Address, err error)
    // Returns the number of most recent block.
        EthBlockNumber(ctx context.Context,
            ) (BlockNumber Uint, err error)
    // Executes a new message call immediately without creating a transaction on the block chain.
        EthCall(ctx context.Context,
            Transaction GenericTransaction,
            Block *BlockNumberOrTag,
            ) (ReturnData Bytes, err error)
    // Generates and returns an estimate of how much gas is necessary to allow the transaction to complete.
        EthEstimateGas(ctx context.Context,
            Transaction GenericTransaction,
            Block *BlockNumberOrTag,
            ) (GasUsed Uint, err error)
    // Generates an access list for a transaction.
        EthCreateAccessList(ctx context.Context,
            Transaction GenericTransaction,
            Block *BlockNumberOrTag,
            ) (GasUsed struct {
            AccessList AccessList `json:"accessList"`
            Error string `json:"error"`
            GasUsed Uint `json:"gasUsed"`
            }, err error)
    // Returns the current price per gas in wei.
        EthGasPrice(ctx context.Context,
            ) (GasPrice Uint, err error)
    // Returns the current maxPriorityFeePerGas per gas in wei.
        EthMaxPriorityFeePerGas(ctx context.Context,
            ) (MaxPriorityFeePerGas Uint, err error)
    // Transaction fee history
        EthFeeHistory(ctx context.Context,
            BlockCount Uint,
            NewestBlock BlockNumberOrTag,
            RewardPercentiles []float64,
            ) (FeeHistoryResult struct {
            BaseFeePerGas []Uint `json:"baseFeePerGas"`
            OldestBlock Uint `json:"oldestBlock"`
            Reward [][]Uint `json:"reward"`
            }, err error)
    // Creates a filter object, based on filter options, to notify when the state changes (logs).
        EthNewFilter(ctx context.Context,
            Filter *Filter,
            ) (FilterIdentifier Uint, err error)
    // Creates a filter in the node, to notify when a new block arrives.
        EthNewBlockFilter(ctx context.Context,
            ) (FilterIdentifier Uint, err error)
    // Creates a filter in the node, to notify when new pending transactions arrive.
        EthNewPendingTransactionFilter(ctx context.Context,
            ) (FilterIdentifier Uint, err error)
    // Uninstalls a filter with given id.
        EthUninstallFilter(ctx context.Context,
            FilterIdentifier *Uint,
            ) (Success bool, err error)
    // Polling method for a filter, which returns an array of logs which occurred since last poll.
        EthGetFilterChanges(ctx context.Context,
            FilterIdentifier *Uint,
            ) (LogObjects FilterResults, err error)
    // Returns an array of all logs matching filter with given id.
        EthGetFilterLogs(ctx context.Context,
            FilterIdentifier *Uint,
            ) (LogObjects FilterResults, err error)
    // Returns an array of all logs matching filter with given id.
        EthGetLogs(ctx context.Context,
            Filter *Filter,
            ) (LogObjects FilterResults, err error)
    // Returns whether the client is actively mining new blocks.
        EthMining(ctx context.Context,
            ) (MiningStatus bool, err error)
    // Returns the number of hashes per second that the node is mining with.
        EthHashrate(ctx context.Context,
            ) (MiningStatus Uint, err error)
    // Returns the hash of the current block, the seedHash, and the boundary condition to be met (“target”).
        EthGetWork(ctx context.Context,
            ) (CurrentWork []Bytes32, err error)
    // Used for submitting a proof-of-work solution.
        EthSubmitWork(ctx context.Context,
            Nonce Bytes8,
            Hash Bytes32,
            Digest Bytes32,
            ) (Success bool, err error)
    // Used for submitting mining hashrate.
        EthSubmitHashrate(ctx context.Context,
            Hashrate Bytes32,
            Id Bytes32,
            ) (Success bool, err error)
    // Returns an EIP-191 signature over the provided data.
        EthSign(ctx context.Context,
            Address Address,
            Message Bytes,
            ) (Signature Bytes65, err error)
    // Returns an RLP encoded transaction signed by the specified account.
        EthSignTransaction(ctx context.Context,
            Transaction GenericTransaction,
            ) (EncodedTransaction Bytes, err error)
    // Returns the balance of the account of given address.
        EthGetBalance(ctx context.Context,
            Address Address,
            Block *BlockNumberOrTag,
            ) (Balance Uint, err error)
    // Returns the value from a storage position at a given address.
        EthGetStorageAt(ctx context.Context,
            Address Address,
            StorageSlot Uint256,
            Block *BlockNumberOrTag,
            ) (Value Bytes, err error)
    // Returns the number of transactions sent from an address.
        EthGetTransactionCount(ctx context.Context,
            Address Address,
            Block *BlockNumberOrTag,
            ) (TransactionCount Uint, err error)
    // Returns code at a given address.
        EthGetCode(ctx context.Context,
            Address Address,
            Block *BlockNumberOrTag,
            ) (Bytecode Bytes, err error)
    // Returns the merkle proof for a given account and optionally some storage keys.
        EthGr error)
    // Transaction fee history
        EthFeeHistory(ctx context.Context,
            BlockCount Uint,
            NewestBlock BlockNumberOrTag,
            RewardPercentiles []float64,
            ) (FeeHistoryResult struct {
            BaseFeePerGas []Uint `json:"baseFeePerGas"`
            OldestBlock Uint `json:"oldestBlock"`
            Reward [][]Uint `json:"reward"`
            }, err error)
    // Creates a filter object, based on filter options, to notify when the state changes (logs).
        EthNewFilter(ctx context.Context,
            Filter *Filter,
            ) (FilterIdentifier Uint, err error)
    // Creates a filter in the node, to notify when a new block arrives.
        EthNewBlockFilter(ctx context.Context,
            ) (FilterIdentifier Uint, err error)
    // Creates a filter in the node, to notify when new pending transactions arrive.
        EthNewPendingTransactionFilter(ctx context.Context,
            ) (FilterIdentifier Uint, err error)
    // Uninstalls a filter with given id.
        EthUninstallFilter(ctx context.Context,
            FilterIdentifier *Uint,
            ) (Success bool, err error)
    // Polling method for a filter, which returns an array of logs which occurred since last poll.
        EthGetFilterChanges(ctx context.Context,
            FilterIdentifier *Uint,
            ) (LogObjects FilterResults, err error)
    // Returns an array of all logs matching filter with given id.
        EthGetFilterLogs(ctx context.Context,
            FilterIdentifier *Uint,
            ) (LogObjects FilterResults, err error)
    // Returns an array of all logs matching filter with given id.
        EthGetLogs(ctx context.Context,
            Filter *Filter,
            ) (LogObjects FilterResults, err error)
    // Returns whether the client is actively mining new blocks.
        EthMining(ctx context.Context,
            ) (MiningStatus bool, err error)
    // Returns the number of hashes per second that the node is mining with.
        EthHashrate(ctx context.Context,
            ) (MiningStatus Uint, err error)
    // Returns the hash of the current block, the seedHash, and the boundary condition to be met (“target”).
        EthGetWork(ctx context.Context,
            ) (CurrentWork []Bytes32, err error)
    // Used for submitting a proof-of-work solution.
        EthSubmitWork(ctx context.Context,
            Nonce Bytes8,
            Hash Bytes32,
            Digest Bytes32,
            ) (Success bool, err error)
    // Used for submitting mining hashrate.
        EthSubmitHashrate(ctx context.Context,
            Hashrate Bytes32,
            Id Bytes32,
            ) (Success bool, err error)
    // Returns an EIP-191 signature over the provided data.
        EthSign(ctx context.Context,
            Address Address,
            Message Bytes,
            ) (Signature Bytes65, err error)
    // Returns an RLP encoded transaction signed by the specified account.
        EthSignTransaction(ctx context.Context,
            Transaction GenericTransaction,
            ) (EncodedTransaction Bytes, err error)
    // Returns the balance of the account of given address.
        EthGetBalance(ctx context.Context,
            Address Address,
            Block *BlockNumberOrTag,
            ) (Balance Uint, err error)
    // Returns the value from a storage position at a given address.
        EthGetStorageAt(ctx context.Context,
            Address Address,
            StorageSlot Uint256,
            Block *BlockNumberOrTag,
            ) (Value Bytes, err error)
    // Returns the number of transactions sent from an address.
        EthGetTransactionCount(ctx context.Context,
            Address Address,
            Block *BlockNumberOrTag,
            ) (TransactionCount Uint, err error)
    // Returns code at a given address.
        EthGetCode(ctx context.Context,
            Address Address,
            Block *BlockNumberOrTag,
            ) (Bytecode Bytes, err error)
    // Returns the merkle proof for a given account and optionally some storage keys.
        EthGetProof(ctx context.Context,
            Address Address,
            StorageKeys []Hash32,
            Block BlockNumberOrTag,
            ) (Account AccountProof, err error)
    // Signs and submits a transaction.
        EthSendTransaction(ctx context.Context,
            Transaction GenericTransaction,
            ) (TransactionHash Hash32, err error)
    // Submits a raw transaction.
        EthSendRawTransaction(ctx context.Context,
            Transaction Bytes,
            ) (TransactionHash Hash32, err error)
    // Returns the information about a transaction requested by transaction hash.
        EthGetTransactionByHash(ctx context.Context,
            TransactionHash Hash32,
            ) (TransactionInformation TransactionInfo, err error)
    // Returns information about a transaction by block hash and transaction index position.
        EthGetTransactionByBlockHashAndIndex(ctx context.Context,
            BlockHash Hash32,
            TransactionIndex Uint,
            ) (TransactionInformation TransactionInfo, err error)
    // Returns information about a transaction by block number and transaction index position.
        EthGetTransactionByBlockNumberAndIndex(ctx context.Context,
            Block BlockNumberOrTag,
            TransactionIndex Uint,
            ) (TransactionInformation TransactionInfo, err error)
    // Returns the receipt of a transaction by transaction hash.
        EthGetTransactionReceipt(ctx context.Context,
            TransactionHash *Hash32,
            ) (ReceiptInformation ReceiptInfo, err error)
    }
    type AccessList []AccessListEntry
    type AccessListEntry struct {
            Address Address `json:"address"`
            StorageKeys []Hash32 `json:"storageKeys"`
            }
    type AccountProof struct {
            AccountProof []Bytes `json:"accountProof"`
            Address Address `json:"address"`
            Balance Uint256 `json:"balance"`
            CodeHash Hash32 `json:"codeHash"`
            Nonce Uint64 `json:"nonce"`
            StorageHash Hash32 `json:"storageHash"`
            StorageProof []StorageProof `json:"storageProof"`
            }
    type BadBlock struct {
            Block Bytes `json:"block"`
            Hash Hash32 `json:"hash"`
            Rlp Bytes `json:"rlp"`
            }
    type Block struct {
            BaseFeePerGas Uint `json:"baseFeePerGas"`
            Difficulty Bytes `json:"difficulty"`
            ExtraData Bytes `json:"extraData"`
            GasLimit Uint `json:"gasLimit"`
            GasUsed Uint `json:"gasUsed"`
            LogsBloom Bytes256 `json:"logsBloom"`
            Miner Address `json:"miner"`
            MixHash Hash32 `json:"mixHash"`
            Nonce Bytes8 `json:"nonce"`
            Number Uint `json:"number"`
            ParentHash Hash32 `json:"parentHash"`
            ReceiptsRoot Hash32 `json:"receiptsRoot"`
            Sha3Uncles Hash32 `json:"sha3Uncles"`
            Size Uint `json:"size"`
            StateRoot Hash32 `json:"stateRoot"`
            Timestamp Uint `json:"timestamp"`
            TotalDifficulty Uint `json:"totalDifficulty"`
            Transactions struct {
            Option0 []Hash32
            Option1 []TransactionSigned
            } `json:"transactions"`
            TransactionsRoot Hash32 `json:"transactionsRoot"`
            Uncles []Hash32 `json:"uncles"`
            }
    type BlockNumberOrTag struct {
            Option0 Uint
            Option1 BlockTag
            }
    type BlockTag string
    type Filter struct {
            Address struct {
            Option0 Address
            Option1 Addresses
            } `json:"address"`
            FromBlock Uint `json:"fromBlock"`
            ToBlock Uint `json:"toBlock"`
            Topics FilterTopics `json:"topics"`
            }
    type FilterResults struct {
            Option0 []Hash32
            Option1 []Hash32
            Option2 []Log
            }
    type FilterTopic struct {
            Option0 struct{}
            Option1 Bytes32
            Option2 []Bytes32
            }
    type FilterTopics []FilterTopicetProof(ctx context.Context,
            Address Address,
            StorageKeys []Hash32,
            Block BlockNumberOrTag,
            ) (Account AccountProof, err error)
    // Signs and submits a transaction.
        EthSendTransaction(ctx context.Context,
            Transaction GenericTransaction,
            ) (TransactionHash Hash32, err error)
    // Submits a raw transaction.
        EthSendRawTransaction(ctx context.Context,
            Transaction Bytes,
            ) (TransactionHash Hash32, err error)
    // Returns the information about a transaction requested by transaction hash.
        EthGetTransactionByHash(ctx context.Context,
            TransactionHash Hash32,
            ) (TransactionInformation TransactionInfo, err error)
    // Returns information about a transaction by block hash and transaction index position.
        EthGetTransactionByBlockHashAndIndex(ctx context.Context,
            BlockHash Hash32,
            TransactionIndex Uint,
            ) (TransactionInformation TransactionInfo, err error)
    // Returns information about a transaction by block number and transaction index position.
        EthGetTransactionByBlockNumberAndIndex(ctx context.Context,
            Block BlockNumberOrTag,
            TransactionIndex Uint,
            ) (TransactionInformation TransactionInfo, err error)
    // Returns the receipt of a transaction by transaction hash.
        EthGetTransactionReceipt(ctx context.Context,
            TransactionHash *Hash32,
            ) (ReceiptInformation ReceiptInfo, err error)
    }
    type AccessList []AccessListEntry
    type AccessListEntry struct {
            Address Address `json:"address"`
            StorageKeys []Hash32 `json:"storageKeys"`
            }
    type AccountProof struct {
            AccountProof []Bytes `json:"accountProof"`
            Address Address `json:"address"`
            Balance Uint256 `json:"balance"`
            CodeHash Hash32 `json:"codeHash"`
            Nonce Uint64 `json:"nonce"`
            StorageHash Hash32 `json:"storageHash"`
            StorageProof []StorageProof `json:"storageProof"`
            }
    type BadBlock struct {
            Block Bytes `json:"block"`
            Hash Hash32 `json:"hash"`
            Rlp Bytes `json:"rlp"`
            }
    type Block struct {
            BaseFeePerGas Uint `json:"baseFeePerGas"`
            Difficulty Bytes `json:"difficulty"`
            ExtraData Bytes `json:"extraData"`
            GasLimit Uint `json:"gasLimit"`
            GasUsed Uint `json:"gasUsed"`
            LogsBloom Bytes256 `json:"logsBloom"`
            Miner Address `json:"miner"`
            MixHash Hash32 `json:"mixHash"`
            Nonce Bytes8 `json:"nonce"`
            Number Uint `json:"number"`
            ParentHash Hash32 `json:"parentHash"`
            ReceiptsRoot Hash32 `json:"receiptsRoot"`
            Sha3Uncles Hash32 `json:"sha3Uncles"`
            Size Uint `json:"size"`
            StateRoot Hash32 `json:"stateRoot"`
            Timestamp Uint `json:"timestamp"`
            TotalDifficulty Uint `json:"totalDifficulty"`
            Transactions struct {
            Option0 []Hash32
            Option1 []TransactionSigned
            } `json:"transactions"`
            TransactionsRoot Hash32 `json:"transactionsRoot"`
            Uncles []Hash32 `json:"uncles"`
            }
    type BlockNumberOrTag struct {
            Option0 Uint
            Option1 BlockTag
            }
    type BlockTag string
    type Filter struct {
            Address struct {
            Option0 Address
            Option1 Addresses
            } `json:"address"`
            FromBlock Uint `json:"fromBlock"`
            ToBlock Uint `json:"toBlock"`
            Topics FilterTopics `json:"topics"`
            }
    type FilterResults struct {
            Option0 []Hash32
            Option1 []Hash32
            Option2 []Log
            }
    type FilterTopic struct {
            Option0 struct{}
            Option1 Bytes32
            Option2 []Bytes32
            }
    type FilterTopics []FilterTopic
    type GenericTransaction struct {
            AccessList AccessList `json:"accessList"`
            ChainId Uint `json:"chainId"`
            From Address `json:"from"`
            Gas Uint `json:"gas"`
            GasPrice Uint `json:"gasPrice"`
            Input Bytes `json:"input"`
            MaxFeePerGas Uint `json:"maxFeePerGas"`
            MaxPriorityFeePerGas Uint `json:"maxPriorityFeePerGas"`
            Nonce Uint `json:"nonce"`
            To Address `json:"to"`
            Type Byte `json:"type"`
            Value Uint `json:"value"`
            }
    type Log struct {
            Address Address `json:"address"`
            BlockHash Hash32 `json:"blockHash"`
            BlockNumber Uint `json:"blockNumber"`
            Data Bytes `json:"data"`
            LogIndex Uint `json:"logIndex"`
            Removed bool `json:"removed"`
            Topics []Bytes32 `json:"topics"`
            TransactionHash Hash32 `json:"transactionHash"`
            TransactionIndex Uint `json:"transactionIndex"`
            }
    type ReceiptInfo struct {
            BlockHash Hash32 `json:"blockHash"`
            BlockNumber Uint `json:"blockNumber"`
            ContractAddress struct {
            Option0 Address
            Option1 struct{}
            } `json:"contractAddress"`
            CumulativeGasUsed Uint `json:"cumulativeGasUsed"`
            EffectiveGasPrice Uint `json:"effectiveGasPrice"`
            From Address `json:"from"`
            GasUsed Uint `json:"gasUsed"`
            Logs []Log `json:"logs"`
            LogsBloom Bytes256 `json:"logsBloom"`
            Root Bytes32 `json:"root"`
            Status Uint `json:"status"`
            To Address `json:"to"`
            TransactionHash Hash32 `json:"transactionHash"`
            TransactionIndex Uint `json:"transactionIndex"`
            }
    type StorageProof struct {
            Key Hash32 `json:"key"`
            Proof []Bytes `json:"proof"`
            Value Uint256 `json:"value"`
            }
    type SyncingStatus struct {
            Option0 struct {
            CurrentBlock Uint `json:"currentBlock"`
            HighestBlock Uint `json:"highestBlock"`
            StartingBlock Uint `json:"startingBlock"`
            }
            Option1 bool
            }
    type Transaction1559Signed struct {
            Field0 Transaction1559Unsigned
            Field1 struct {
            R Uint `json:"r"`
            S Uint `json:"s"`
            YParity Uint `json:"yParity"`
            }
            }
    type Transaction1559Unsigned struct {
            AccessList AccessList `json:"accessList"`
            ChainId Uint `json:"chainId"`
            Gas Uint `json:"gas"`
            Input Bytes `json:"input"`
            MaxFeePerGas Uint `json:"maxFeePerGas"`
            MaxPriorityFeePerGas Uint `json:"maxPriorityFeePerGas"`
            Nonce Uint `json:"nonce"`
            To Address `json:"to"`
            Type Byte `json:"type"`
            Value Uint `json:"value"`
            }
    type Transaction2930Signed struct {
            Field0 Transaction2930Unsigned
            Field1 struct {
            R Uint `json:"r"`
            S Uint `json:"s"`
            YParity Uint `json:"yParity"`
            }
            }
    type Transaction2930Unsigned struct {
            AccessList AccessList `json:"accessList"`
            ChainId Uint `json:"chainId"`
            Gas Uint `json:"gas"`
            GasPrice Uint `json:"gasPrice"`
            Input Bytes `json:"input"`
            Nonce Uint `json:"nonce"`
            To Address `json:"to"`
            Type Byte `json:"type"`
            Value Uint `json:"value"`
            }
    type TransactionInfo struct {
            Field0 struct {
            BlockHash Hash32 `json:"blockHash"`
            BlockNumber Uint `json:"blockNumber"`
            From Address `json:"from"`
            Hash Hash32 `json:"hash"`
            TransactionIndex Uint `json:"transactionIndex"`
            }
            Field1 TransactionSigned
            }
    type TransactionLegacySigned struct {
            Field0 Trans
    type GenericTransaction struct {
            AccessList AccessList `json:"accessList"`
            ChainId Uint `json:"chainId"`
            From Address `json:"from"`
            Gas Uint `json:"gas"`
            GasPrice Uint `json:"gasPrice"`
            Input Bytes `json:"input"`
            MaxFeePerGas Uint `json:"maxFeePerGas"`
            MaxPriorityFeePerGas Uint `json:"maxPriorityFeePerGas"`
            Nonce Uint `json:"nonce"`
            To Address `json:"to"`
            Type Byte `json:"type"`
            Value Uint `json:"value"`
            }
    type Log struct {
            Address Address `json:"address"`
            BlockHash Hash32 `json:"blockHash"`
            BlockNumber Uint `json:"blockNumber"`
            Data Bytes `json:"data"`
            LogIndex Uint `json:"logIndex"`
            Removed bool `json:"removed"`
            Topics []Bytes32 `json:"topics"`
            TransactionHash Hash32 `json:"transactionHash"`
            TransactionIndex Uint `json:"transactionIndex"`
            }
    type ReceiptInfo struct {
            BlockHash Hash32 `json:"blockHash"`
            BlockNumber Uint `json:"blockNumber"`
            ContractAddress struct {
            Option0 Address
            Option1 struct{}
            } `json:"contractAddress"`
            CumulativeGasUsed Uint `json:"cumulativeGasUsed"`
            EffectiveGasPrice Uint `json:"effectiveGasPrice"`
            From Address `json:"from"`
            GasUsed Uint `json:"gasUsed"`
            Logs []Log `json:"logs"`
            LogsBloom Bytes256 `json:"logsBloom"`
            Root Bytes32 `json:"root"`
            Status Uint `json:"status"`
            To Address `json:"to"`
            TransactionHash Hash32 `json:"transactionHash"`
            TransactionIndex Uint `json:"transactionIndex"`
            }
    type StorageProof struct {
            Key Hash32 `json:"key"`
            Proof []Bytes `json:"proof"`
            Value Uint256 `json:"value"`
            }
    type SyncingStatus struct {
            Option0 struct {
            CurrentBlock Uint `json:"currentBlock"`
            HighestBlock Uint `json:"highestBlock"`
            StartingBlock Uint `json:"startingBlock"`
            }
            Option1 bool
            }
    type Transaction1559Signed struct {
            Field0 Transaction1559Unsigned
            Field1 struct {
            R Uint `json:"r"`
            S Uint `json:"s"`
            YParity Uint `json:"yParity"`
            }
            }
    type Transaction1559Unsigned struct {
            AccessList AccessList `json:"accessList"`
            ChainId Uint `json:"chainId"`
            Gas Uint `json:"gas"`
            Input Bytes `json:"input"`
            MaxFeePerGas Uint `json:"maxFeePerGas"`
            MaxPriorityFeePerGas Uint `json:"maxPriorityFeePerGas"`
            Nonce Uint `json:"nonce"`
            To Address `json:"to"`
            Type Byte `json:"type"`
            Value Uint `json:"value"`
            }
    type Transaction2930Signed struct {
            Field0 Transaction2930Unsigned
            Field1 struct {
            R Uint `json:"r"`
            S Uint `json:"s"`
            YParity Uint `json:"yParity"`
            }
            }
    type Transaction2930Unsigned struct {
            AccessList AccessList `json:"accessList"`
            ChainId Uint `json:"chainId"`
            Gas Uint `json:"gas"`
            GasPrice Uint `json:"gasPrice"`
            Input Bytes `json:"input"`
            Nonce Uint `json:"nonce"`
            To Address `json:"to"`
            Type Byte `json:"type"`
            Value Uint `json:"value"`
            }
    type TransactionInfo struct {
            Field0 struct {
            BlockHash Hash32 `json:"blockHash"`
            BlockNumber Uint `json:"blockNumber"`
            From Address `json:"from"`
            Hash Hash32 `json:"hash"`
            TransactionIndex Uint `json:"transactionIndex"`
            }
            Field1 TransactionSigned
            }
    type TransactionLegacySigned struct {
            Field0 TransactionLegacyUnsigned
            Field1 struct {
            R Uint `json:"r"`
            S Uint `json:"s"`
            V Uint `json:"v"`
            }
            }
    type TransactionLegacyUnsigned struct {
            ChainId Uint `json:"chainId"`
            Gas Uint `json:"gas"`
            GasPrice Uint `json:"gasPrice"`
            Input Bytes `json:"input"`
            Nonce Uint `json:"nonce"`
            To Address `json:"to"`
            Type Byte `json:"type"`
            Value Uint `json:"value"`
            }
    type TransactionSigned struct {
            Option0 Transaction1559Signed
            Option1 Transaction2930Signed
            Option2 TransactionLegacySigned
            }
    type TransactionUnsigned struct {
            Option0 Transaction1559Unsigned
            Option1 Transaction2930Unsigned
            Option2 TransactionLegacyUnsigned
            }
    type Address string
    type Addresses []Address
    type Byte string
    type Bytes string
    type Bytes256 string
    type Bytes32 string
    type Bytes65 string
    type Bytes8 string
    type Hash32 string
    type Uint string
    type Uint256 string
    type Uint64 string


actionLegacyUnsigned
            Field1 struct {
            R Uint `json:"r"`
            S Uint `json:"s"`
            V Uint `json:"v"`
            }
            }
    type TransactionLegacyUnsigned struct {
            ChainId Uint `json:"chainId"`
            Gas Uint `json:"gas"`
            GasPrice Uint `json:"gasPrice"`
            Input Bytes `json:"input"`
            Nonce Uint `json:"nonce"`
            To Address `json:"to"`
            Type Byte `json:"type"`
            Value Uint `json:"value"`
            }
    type TransactionSigned struct {
            Option0 Transaction1559Signed
            Option1 Transaction2930Signed
            Option2 TransactionLegacySigned
            }
    type TransactionUnsigned struct {
            Option0 Transaction1559Unsigned
            Option1 Transaction2930Unsigned
            Option2 TransactionLegacyUnsigned
            }
    type Address string
    type Addresses []Address
    type Byte string
    type Bytes string
    type Bytes256 string
    type Bytes32 string
    type Bytes65 string
    type Bytes8 string
    type Hash32 string
    type Uint string
    type Uint256 string
    type Uint64 string


error: 559:28: missing ',' in argument list (and 10 more errors)
error: 559:28: missing ',' in argument list (and 10 more errors)
exit status 1
exit status 1
gogenerate.go:4: running "go": exit status 1
gogenerate.go:4: running "go": exit status 1
