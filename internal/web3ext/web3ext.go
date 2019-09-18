// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// package web3ext contains gero specific web3.js extensions.
package web3ext

var Modules = map[string]string{
	"admin":      Admin_JS,
	"chequebook": Chequebook_JS,
	"clique":     Clique_JS,
	"debug":      Debug_JS,
	"sero":       SER_JS,
	"miner":      Miner_JS,
	"net":        Net_JS,
	"personal":   Personal_JS,
	"rpc":        RPC_JS,
	"shh":        Shh_JS,
	"swarmfs":    SWARMFS_JS,
	"txpool":     TxPool_JS,
	"ssi":        SSI_JS,
	"exchange":   Exchange_JS,
	"light":      LightNode_JS,
	"stake":      Stake_JS,
	"flight":     Flight_JS,
	"local":      Local_JS,
}

const Chequebook_JS = `
web3._extend({
	property: 'chequebook',
	methods: [
		new web3._extend.Method({
			name: 'deposit',
			call: 'chequebook_deposit',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Property({
			name: 'balance',
			getter: 'chequebook_balance',
			outputFormatter: web3._extend.utils.toDecimal
		}),
		new web3._extend.Method({
			name: 'cash',
			call: 'chequebook_cash',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'issue',
			call: 'chequebook_issue',
			params: 2,
			inputFormatter: [null, null]
		}),
	]
});
`

const Clique_JS = `
web3._extend({
	property: 'clique',
	methods: [
		new web3._extend.Method({
			name: 'getSnapshot',
			call: 'clique_getSnapshot',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getSnapshotAtHash',
			call: 'clique_getSnapshotAtHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getSigners',
			call: 'clique_getSigners',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getSignersAtHash',
			call: 'clique_getSignersAtHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'propose',
			call: 'clique_propose',
			params: 2
		}),
		new web3._extend.Method({
			name: 'discard',
			call: 'clique_discard',
			params: 1
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'proposals',
			getter: 'clique_proposals'
		}),
	]
});
`

const Admin_JS = `
web3._extend({
	property: 'admin',
	methods: [
		new web3._extend.Method({
			name: 'addPeer',
			call: 'admin_addPeer',
			params: 1
		}),
		new web3._extend.Method({
			name: 'removePeer',
			call: 'admin_removePeer',
			params: 1
		}),
		new web3._extend.Method({
			name: 'exportChain',
			call: 'admin_exportChain',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'importChain',
			call: 'admin_importChain',
			params: 1
		}),
		new web3._extend.Method({
			name: 'sleepBlocks',
			call: 'admin_sleepBlocks',
			params: 2
		}),
		new web3._extend.Method({
			name: 'startRPC',
			call: 'admin_startRPC',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new web3._extend.Method({
			name: 'stopRPC',
			call: 'admin_stopRPC'
		}),
		new web3._extend.Method({
			name: 'startWS',
			call: 'admin_startWS',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new web3._extend.Method({
			name: 'stopWS',
			call: 'admin_stopWS'
		}),
		new web3._extend.Method({
			name: 'close',
			call: 'admin_close',
			params: 0
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'nodeInfo',
			getter: 'admin_nodeInfo'
		}),
		new web3._extend.Property({
			name: 'peers',
			getter: 'admin_peers'
		}),
		new web3._extend.Property({
			name: 'datadir',
			getter: 'admin_datadir'
		}),
	]
});
`

const Debug_JS = `
web3._extend({
	property: 'debug',
	methods: [
		new web3._extend.Method({
			name: 'printBlock',
			call: 'debug_printBlock',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getBlockRlp',
			call: 'debug_getBlockRlp',
			params: 1
		}),
		new web3._extend.Method({
			name: 'setHead',
			call: 'debug_setHead',
			params: 1
		}),
		new web3._extend.Method({
			name: 'seedHash',
			call: 'debug_seedHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'dumpBlock',
			call: 'debug_dumpBlock',
			params: 1
		}),
		new web3._extend.Method({
			name: 'chaindbProperty',
			call: 'debug_chaindbProperty',
			params: 1,
			outputFormatter: console.log
		}),
		new web3._extend.Method({
			name: 'chaindbCompact',
			call: 'debug_chaindbCompact',
		}),
		new web3._extend.Method({
			name: 'metrics',
			call: 'debug_metrics',
			params: 1
		}),
		new web3._extend.Method({
			name: 'verbosity',
			call: 'debug_verbosity',
			params: 1
		}),
		new web3._extend.Method({
			name: 'vmodule',
			call: 'debug_vmodule',
			params: 1
		}),
		new web3._extend.Method({
			name: 'backtraceAt',
			call: 'debug_backtraceAt',
			params: 1,
		}),
		new web3._extend.Method({
			name: 'stacks',
			call: 'debug_stacks',
			params: 0,
			outputFormatter: console.log
		}),
		new web3._extend.Method({
			name: 'freeOSMemory',
			call: 'debug_freeOSMemory',
			params: 0,
		}),
		new web3._extend.Method({
			name: 'setGCPercent',
			call: 'debug_setGCPercent',
			params: 1,
		}),
		new web3._extend.Method({
			name: 'memStats',
			call: 'debug_memStats',
			params: 0,
		}),
		new web3._extend.Method({
			name: 'gcStats',
			call: 'debug_gcStats',
			params: 0,
		}),
		new web3._extend.Method({
			name: 'cpuProfile',
			call: 'debug_cpuProfile',
			params: 2
		}),
		new web3._extend.Method({
			name: 'startCPUProfile',
			call: 'debug_startCPUProfile',
			params: 1
		}),
		new web3._extend.Method({
			name: 'stopCPUProfile',
			call: 'debug_stopCPUProfile',
			params: 0
		}),
		new web3._extend.Method({
			name: 'goTrace',
			call: 'debug_goTrace',
			params: 2
		}),
		new web3._extend.Method({
			name: 'startGoTrace',
			call: 'debug_startGoTrace',
			params: 1
		}),
		new web3._extend.Method({
			name: 'stopGoTrace',
			call: 'debug_stopGoTrace',
			params: 0
		}),
		new web3._extend.Method({
			name: 'blockProfile',
			call: 'debug_blockProfile',
			params: 2
		}),
		new web3._extend.Method({
			name: 'setBlockProfileRate',
			call: 'debug_setBlockProfileRate',
			params: 1
		}),
		new web3._extend.Method({
			name: 'writeBlockProfile',
			call: 'debug_writeBlockProfile',
			params: 1
		}),
		new web3._extend.Method({
			name: 'mutexProfile',
			call: 'debug_mutexProfile',
			params: 2
		}),
		new web3._extend.Method({
			name: 'setMutexProfileFraction',
			call: 'debug_setMutexProfileFraction',
			params: 1
		}),
		new web3._extend.Method({
			name: 'writeMutexProfile',
			call: 'debug_writeMutexProfile',
			params: 1
		}),
		new web3._extend.Method({
			name: 'writeMemProfile',
			call: 'debug_writeMemProfile',
			params: 1
		}),
		new web3._extend.Method({
			name: 'traceBlock',
			call: 'debug_traceBlock',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Method({
			name: 'traceBlockFromFile',
			call: 'debug_traceBlockFromFile',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Method({
			name: 'traceBlockByNumber',
			call: 'debug_traceBlockByNumber',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Method({
			name: 'traceBlockByHash',
			call: 'debug_traceBlockByHash',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Method({
			name: 'traceTransaction',
			call: 'debug_traceTransaction',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Method({
			name: 'preimage',
			call: 'debug_preimage',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getBadBlocks',
			call: 'debug_getBadBlocks',
			params: 0,
		}),
		new web3._extend.Method({
			name: 'storageRangeAt',
			call: 'debug_storageRangeAt',
			params: 5,
		}),
		new web3._extend.Method({
			name: 'getModifiedAccountsByNumber',
			call: 'debug_getModifiedAccountsByNumber',
			params: 2,
			inputFormatter: [null, null],
		}),
		new web3._extend.Method({
			name: 'getModifiedAccountsByHash',
			call: 'debug_getModifiedAccountsByHash',
			params: 2,
			inputFormatter:[null, null],
		}),
	],
	properties: []
});
`

const SER_JS = `
web3._extend({
	property: 'sero',
	methods: [
		new web3._extend.Method({
			name: 'sign',
			call: 'sero_sign',
			params: 2,
			inputFormatter: [web3._extend.formatters.inputAddressFormatter, null]
		}),
		new web3._extend.Method({
			name: 'resend',
			call: 'sero_resend',
			params: 3,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter, web3._extend.utils.fromDecimal, web3._extend.utils.fromDecimal]
		}),
		new web3._extend.Method({
			name: 'signTransaction',
			call: 'sero_signTransaction',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		}),
		new web3._extend.Method({
			name: 'submitTransaction',
			call: 'sero_submitTransaction',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		}),
        new web3._extend.Method({
			name: 'getBlockTotalRewardByNumber',
			call: 'sero_getBlockTotalRewardByNumber',
			params: 1,
			inputFormatter: [web3._extend.utils.toHex],
			outputFormatter: web3._extend.utils.toDecimal
		}),
       new web3._extend.Method({
			name: 'getBlockPosRewardByNumber',
			call: 'sero_getBlockPosRewardByNumber',
			params: 1,
            inputFormatter: [web3._extend.utils.toHex],
			outputFormatter: web3._extend.utils.toDecimal
		}),
		new web3._extend.Method({
			name: 'getRawTransaction',
			call: 'sero_getRawTransactionByHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getRawTransactionFromBlock',
			call: function(args) {
				return (web3._extend.utils.isString(args[0]) && args[0].indexOf('0x') === 0) ? 'sero_getRawTransactionByBlockHashAndIndex' : 'sero_getRawTransactionByBlockNumberAndIndex';
			},
			params: 2,
			inputFormatter: [web3._extend.formatters.inputBlockNumberFormatter, web3._extend.utils.toHex]
		}),
        new web3._extend.Method({
			name: 'packMethod',
            call: 'sero_packMethod',	
			params:4
		}),
       new web3._extend.Method({
			name: 'packConstruct',
            call: 'sero_packConstruct',	
			params: 3
		}),
       new web3._extend.Method({
			name: 'unPack',
            call: 'sero_unPack',	
			params: 3
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'pendingTransactions',
			getter: 'sero_pendingTransactions',
			outputFormatter: function(txs) {
				var formatted = [];
				for (var i = 0; i < txs.length; i++) {
					formatted.push(web3._extend.formatters.outputTransactionFormatter(txs[i]));
					formatted[i].blockHash = null;
				}
				return formatted;
			}
		}),
	]
});
`

const Miner_JS = `
web3._extend({
	property: 'miner',
	methods: [
		new web3._extend.Method({
			name: 'start',
			call: 'miner_start',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'stop',
			call: 'miner_stop'
		}),
		new web3._extend.Method({
			name: 'setSerobase',
			call: 'miner_setSerobase',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputAddressFormatter]
		}),
		new web3._extend.Method({
			name: 'setExtra',
			call: 'miner_setExtra',
			params: 1
		}),
		new web3._extend.Method({
			name: 'setGasPrice',
			call: 'miner_setGasPrice',
			params: 1,
			inputFormatter: [web3._extend.utils.fromDecimal]
		}),
		new web3._extend.Method({
			name: 'getHashrate',
			call: 'miner_getHashrate'
		}),
	],
	properties: []
});
`

const Net_JS = `
web3._extend({
	property: 'net',
	methods: [],
	properties: [
		new web3._extend.Property({
			name: 'version',
			getter: 'net_version'
		}),
	]
});
`

const Personal_JS = `
web3._extend({
	property: 'personal',
	methods: [
		new web3._extend.Method({
			name: 'importRawKey',
			call: 'personal_importRawKey',
			params: 2
		}),
        new web3._extend.Method({
			name: 'exportRawKey',
			call: 'personal_exportRawKey',
			params: 2
		}),
        new web3._extend.Method({
			name: 'genSeed',
			call: 'personal_genSeed',
			params: 0
		}),
		new web3._extend.Method({
			name: 'sign',
			call: 'personal_sign',
			params: 3,
			inputFormatter: [null, web3._extend.formatters.inputAddressFormatter, null]
		}),
        new web3._extend.Method({
			name: 'exportMnemonic',
			call: 'personal_exportMnemonic',
			params: 2,
			inputFormatter: [web3._extend.formatters.inputAddressFormatter, null]
		}),
		new web3._extend.Method({
			name: 'ecRecover',
			call: 'personal_ecRecover',
			params: 2
		}),
		new web3._extend.Method({
			name: 'deriveAccount',
			call: 'personal_deriveAccount',
			params: 3
		}),
		new web3._extend.Method({
			name: 'signTransaction',
			call: 'personal_signTransaction',
			params: 2,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter, null]
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'listWallets',
			getter: 'personal_listWallets'
		}),
	]
})
`

const RPC_JS = `
web3._extend({
	property: 'rpc',
	methods: [],
	properties: [
		new web3._extend.Property({
			name: 'modules',
			getter: 'rpc_modules'
		}),
	]
});
`

const Shh_JS = `
web3._extend({
	property: 'shh',
	methods: [
	],
	properties:
	[
		new web3._extend.Property({
			name: 'version',
			getter: 'shh_version',
			outputFormatter: web3._extend.utils.toDecimal
		}),
		new web3._extend.Property({
			name: 'info',
			getter: 'shh_info'
		}),
	]
});
`

const SWARMFS_JS = `
web3._extend({
	property: 'swarmfs',
	methods:
	[
		new web3._extend.Method({
			name: 'mount',
			call: 'swarmfs_mount',
			params: 2
		}),
		new web3._extend.Method({
			name: 'unmount',
			call: 'swarmfs_unmount',
			params: 1
		}),
		new web3._extend.Method({
			name: 'listmounts',
			call: 'swarmfs_listmounts',
			params: 0
		}),
	]
});
`

const TxPool_JS = `
web3._extend({
	property: 'txpool',
	methods: [],
	properties:
	[
		new web3._extend.Property({
			name: 'content',
			getter: 'txpool_content'
		}),
		new web3._extend.Property({
			name: 'status',
			getter: 'txpool_status',
			outputFormatter: function(status) {
				status.pending = web3._extend.utils.toDecimal(status.pending);
				status.queued = web3._extend.utils.toDecimal(status.queued);
				return status;
			}
		}),
	]
});
`

const SSI_JS = `
web3._extend({
	property: 'ssi',
	methods: [
		new web3._extend.Method({
			name: 'createkr',
			call: 'ssi_createKr',
			params: 0
		}),
		new web3._extend.Method({
			name: 'getblocksinfo',
			call: 'ssi_getBlocksInfo',
			params: 2,
			inputFormatter: [
				web3._extend.utils.toHex,
				web3._extend.utils.toHex
			]
		}),
		new web3._extend.Method({
			name: 'detail',
			call: 'ssi_detail',
			params: 2
		}),
		new web3._extend.Method({
			name: 'gentx',
			call: 'ssi_genTx',
			params: 1
		}),
		new web3._extend.Method({
			name: 'gettx',
			call: 'ssi_getTx',
			params: 1
		}),
		new web3._extend.Method({
			name: 'committx',
			call: 'ssi_commitTx',
			params: 1
		})
	]
});
`

const Exchange_JS = `
web3._extend({
	property: 'exchange',
	methods: [
		new web3._extend.Method({
			name: 'getLockedBalances',
			call: 'exchange_getLockedBalances',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getMaxAvailable',
			call: 'exchange_getMaxAvailable',
			params: 2
		}),
       new web3._extend.Method({
			name: 'clearUsedFlag',
			call: 'exchange_clearUsedFlag',
			params: 1
		}),
       new web3._extend.Method({
			name: 'clearUsedFlagForRoot',
			call: 'exchange_clearUsedFlagForRoot',
			params: 1
		}),
       new web3._extend.Method({
			name: 'genMergeTx',
			call: 'exchange_genMergeTx',
			params: 1
		}),
       new web3._extend.Method({
			name: 'getTx',
			call: 'exchange_getTx',
			params: 1
		}),
        new web3._extend.Method({
			name: 'getBlocksInfo',
			call: 'exchange_getBlocksInfo',
			params: 2
		}),
       new web3._extend.Method({
			name: 'getPkByPkr',
			call: 'exchange_getPkByPkr',
			params: 1
		}),
        new web3._extend.Method({
			name: 'getBlockByNumber',
			call: 'exchange_getBlockByNumber',
            params: 1
		}),
        new web3._extend.Method({
			name: 'seed2Sk',
			call: 'exchange_seed2Sk',
            params: 1
		}),
        new web3._extend.Method({
			name: 'sk2Tk',
			call: 'exchange_sk2Tk',
            params: 1
		}),
        new web3._extend.Method({
			name: 'tk2Pk',
			call: 'exchange_tk2Pk',
            params: 1
		}),
        new web3._extend.Method({
			name: 'pk2Pkr',
			call: 'exchange_pk2Pkr',
            params: 2
		}),
        new web3._extend.Method({
			name: 'signTxWithSk',
			call: 'exchange_signTxWithSk',
            params: 2
		})
	]
});
`
const Stake_JS = `
web3._extend({
	property: 'stake',
	methods: [
		new web3._extend.Method({
			name: 'buyShare',
			call: 'stake_buyShare',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		}),
        new web3._extend.Method({
			name: 'estimateShares',
			call: 'stake_estimateShares',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		}),
		new web3._extend.Method({
			name: 'registStakePool',
			call: 'stake_registStakePool',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputRegistPoolFormatter]
		}),
        new web3._extend.Method({
			name: 'poolState',
			call: 'stake_poolState',
			params: 1,
			outputFormatter: web3._extend.formatters.outputStakePoolFormatter
		}),
        new web3._extend.Method({
			name: 'myShare',
			call: 'stake_myShare',
			params:1,
            outputFormatter: web3._extend.formatters.outputStakeShareFormatter
		}),
        new web3._extend.Method({
			name: 'getShareByPkr',
			call: 'stake_getShareByPkr',
			params:1,
            outputFormatter: web3._extend.formatters.outputStakeShareFormatter
		}),
        new web3._extend.Method({
			name: 'getShare',
			call: 'stake_getShare',
			params:1,
            outputFormatter: web3._extend.formatters.outputStakeShareFormatter
		}),
       new web3._extend.Method({
			name: 'stakePools',
			call: 'stake_stakePools',
			params:0,
            outputFormatter: web3._extend.formatters.outputStakePoolFormatter
		}),
       new web3._extend.Method({
			name: 'closeStakePool',
			call: 'stake_closeStakePool',
			params:1 
		}),
        new web3._extend.Method({
			name: 'modifyStakePoolFee',
			call: 'stake_modifyStakePoolFee',
			params:2,
			inputFormatter: [null, web3._extend.utils.toHex]
		}),
        new web3._extend.Method({
			name: 'modifyStakePoolVote',
			call: 'stake_modifyStakePoolVote',
			params:2 
		}),
        new web3._extend.Method({
			name: 'getStakeInfo',
			call: 'stake_getStakeInfo',
			params:3,
            inputFormatter: [null,web3._extend.utils.toHex,web3._extend.utils.toHex],
            outputFormatter: web3._extend.formatters.outputStakeInfoFormatter
		})

	],
    properties: [
       new web3._extend.Property({
			name: 'sharePoolSize',
			getter: 'stake_sharePoolSize',
            outputFormatter: web3._extend.utils.toDecimal
		}),
       new web3._extend.Property({
			name: 'sharePrice',
			getter: 'stake_sharePrice',
            outputFormatter: web3._extend.utils.toDecimal
		})
	]
});
`

const LightNode_JS = `
web3._extend({
	property: 'light',
	methods: [
		new web3._extend.Method({
			name: 'getOutsByPKr',
			call: 'light_getOutsByPKr',
			params: 3
		}),
		new web3._extend.Method({
			name: 'checkNil',
			call: 'light_checkNil',
			params: 1
		}),
	]
});
`

const Flight_JS = `
web3._extend({
	property: 'flight',
	methods: [
		new web3._extend.Method({
			name: 'getBlocksInfo',
			call: 'flight_getBlocksInfo',
			params: 2
		}),
		new web3._extend.Method({
			name: 'getBlockByNumber',
			call: 'flight_getBlockByNumber',
			params: 1
		}),
		new web3._extend.Method({
			name: 'genTxParam',
			call: 'flight_genTxParam',
			params: 2
		}),
		new web3._extend.Method({
			name: 'commitTx',
			call: 'flight_commitTx',
			params: 1
		}),
		new web3._extend.Method({
			name: 'Trace2Root',
			call: 'flight_trace2Root',
			params: 3
		}),
		new web3._extend.Method({
			name: 'getOut',
			call: 'flight_getOut',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getTx',
			call: 'flight_getTx',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getTxReceipt',
			call: 'flight_getTxReceipt',
			params: 1
		})
	]
});
`

const Local_JS = `
web3._extend({
	property: 'local',
	methods: [
        new web3._extend.Method({
			name: 'currencyToId',
			call: 'local_currencyToId',
            params: 1
		}),
        new web3._extend.Method({
			name: 'idToCurrency',
			call: 'local_idToCurrency',
            params: 1
		}),
        new web3._extend.Method({
			name: 'genSeed',
			call: 'local_genSeed',
            params: 0
		}),
        new web3._extend.Method({
			name: 'seed2Sk',
			call: 'local_seed2Sk',
            params: 1
		}),
        new web3._extend.Method({
			name: 'sk2Tk',
			call: 'local_sk2Tk',
            params: 1
		}),
        new web3._extend.Method({
			name: 'tk2Pk',
			call: 'local_tk2Pk',
            params: 1
		}),
        new web3._extend.Method({
			name: 'pk2Pkr',
			call: 'local_pk2Pkr',
            params: 2
		}),
        new web3._extend.Method({
			name: 'signTxWithSk',
			call: 'local_signTxWithSk',
            params: 2
		}),
		new web3._extend.Method({
			name: 'decOut',
			call: 'local_decOut',
			params: 2
		}),
		new web3._extend.Method({
			name: 'confirmOutZ',
			call: 'local_confirmOutZ',
			params: 2
		}),
		new web3._extend.Method({
			name: 'isPkrValid',
			call: 'local_isPkrValid',
			params: 1
		}),
		new web3._extend.Method({
			name: 'isPkValid',
			call: 'local_isPkValid',
			params: 1
		})
	]
});
`
