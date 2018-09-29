## Go Sero

 Anonymous cash tech based on zeroknowledge proof tech and refacted ethereum protocol using golang .



## Building the source

For prerequisites and detailed build instructions please read the
[Installation Instructions](https://github.com/sero-cash/go-sero/wiki/Building-Sero)
on the wiki.

Building sero requires both a Go (version 1.7 or later) and a C++ compiler.
You can install them using your favourite package manager.
Once the dependencies are installed, run

    make 

or, to build the full suite of utilities:

    make all

## Executables

The go-sero project comes with several wrappers/executables found in the `cmd` directory.

| Command    | Description |
|:----------:|-------------|
| **sero** | Our main Sero CLI client. It is the entry point into the Sero network (main-, test- or private net), capable of running as a full node (default), archive node (retaining all historical state) or a light node (retrieving data live). It can be used by other processes as a gateway into the Sero network via JSON RPC endpoints exposed on top of HTTP, WebSocket and/or IPC transports. `sero --help` and the [CLI Wiki page](https://github .com/Sero/go-Sero/wiki/Command-Line-Options) for command line options. |
| `abigen` | Source code generator to convert Sero contract definitions into easy to use, compile-time type-safe Go packages. It operates on plain [Sero contract ABIs](https://github.com/ethereum/wiki/wiki/Ethereum-Contract-ABI) with expanded functionality if the contract bytecode is also available. However it also accepts Solidity source files, making development much more streamlined. Please see our [Native DApps](https://github.com/sero-cash/go-sero/wiki/Native-DApps:-Go-bindings-to-Ethereum-contracts) wiki page for details. |
| `bootnode` | Stripped down version of our Sero client implementation that only takes part in the network node discovery protocol, but does not run any of the higher level application protocols. It can be used as a lightweight bootstrap node to aid in finding peers in private networks. |
| `evm` | Developer utility version of the EVM (Sero Virtual Machine) that is capable of running bytecode snippets within a configurable environment and execution mode. Its purpose is to allow isolated, fine-grained debugging of EVM opcodes (e.g. `evm --code 60ff60ff --debug`). |
| `rlpdump` | Developer utility tool to convert binary RLP ([Recursive Length Prefix](https://github.com/ethereum/wiki/wiki/RLP)) dumps (data encoding used by the Sero protocol both network as well as consensus wise) to user friendlier hierarchical representation (e.g. `rlpdump --hex CE0183FFFFFFC4C304050583616263`). |


## Running sero

Going through all the possible command line flags is out of scope here (please consult our
[CLI Wiki page](https://github.com/sero-cash/go-sero/wiki/Command-Line-Options)), but we've
enumerated a few common parameter combos to get you up to speed quickly on how you can run your
own sero instance.

### Full node on the main Sero network

By far the most common scenario is people wanting to simply interact with the Sero network:
create accounts; transfer funds; deploy and interact with contracts. For this particular use-case
the user doesn't care about years-old historical data, so we can fast-sync quickly to the current
state of the network. To do so:

```
$ gero console
```

This command will:


 * Start up sero's built-in interactive [JavaScript console](https://github.com/sero-cash/go-sero/wiki/JavaScript-Console),
   (via the trailing `console` subcommand) through which you can invoke all official [`web3` methods](https://github.com/ethereum/wiki/wiki/JavaScript-API)
   as well as sero's own [management APIs](https://github.com/sero-cash/go-sero/wiki/Management-APIs).
   This too is optional and if you leave it out you can always attach to an already running sero instance
   with `sero attach`.

### Full node on the Sero alpha network

Transitioning towards developers, if you'd like to play around with creating Sero contracts, you
almost certainly would like to do that without any real money involved until you get the hang of the
entire system. In other words, instead of attaching to the main network, you want to join the **alpha**
network with your node, which is fully equivalent to the main network, but with play-Sero only.

```
$ gero --alpha console
```

The `console` subcommand have the exact same meaning as above and they are equally useful on the
testnet too. Please see above for their explanations if you've skipped to here.

Specifying the `--alpha` flag however will reconfigure your sero instance a bit:

 * Instead of using the default data directory (`~/.datadir` on Linux for example), sero will nest
   itself one level deeper into a `alpha` subfolder (`~/.datadir/alpha` on Linux). Note, on OSX
   and Linux this also means that attaching to a running testnet node requires the use of a custom
   endpoint since `sero attach` will try to attach to a production node endpoint by default. E.g.
   `gero attach <datadir>/alpha/sero.ipc`. Windows users are not affected by this.
 * Instead of connecting the main Sero network, the client will connect to the test network,
   which uses different P2P bootnodes, different network IDs and genesis states.
   
*Note: Although there are some internal protective measures to prevent transactions from crossing
over between the main(beta) network and alpha network, you should make sure to always use separate accounts
for play-money and real-money. Unless you manually move accounts, sero will by default correctly
separate the two networks and will not make any accounts available between them.*



### Configuration

As an alternative to passing the numerous flags to the `gero` binary, you can also pass a configuration file via:

```
$ gero --config /path/to/your_config.toml
```

To get an idea how the file should look like you can use the `dumpconfig` subcommand to export your existing configuration:

```
$ gero --your-favourite-flags dumpconfig
```



```

This will start sero in fast-sync mode with a DB memory allowance of 1GB just as the above command does.  It will also create a persistent volume in your home directory for saving your blockchain as well as map the default ports. There is also an `alpine` tag available for a slim version of the image.

Do not forget `--rpcaddr 0.0.0.0`, if you want to access RPC from other containers and/or hosts. By default, `sero` binds to the local interface and RPC endpoints is not accessible from the outside.

### Programatically interfacing sero nodes

As a developer, sooner rather than later you'll want to start interacting with sero and the Sero
network via your own programs and not manually through the console. To aid this, sero has built-in
support for a JSON-RPC based APIs ([standard APIs](https://github.com/ethereum/wiki/wiki/JSON-RPC) and
[sero specific APIs](https://github.com/sero-cash/go-sero/wiki/Management-APIs)). These can be
exposed via HTTP, WebSockets and IPC (unix sockets on unix based platforms, and named pipes on Windows).

The IPC interface is enabled by default and exposes all the APIs supported by sero, whereas the HTTP
and WS interfaces need to manually be enabled and only expose a subset of APIs due to security reasons.
These can be turned on/off and configured as you'd expect.

HTTP based JSON-RPC API options:

  * `--rpc` Enable the HTTP-RPC server
  * `--rpcaddr` HTTP-RPC server listening interface (default: "localhost")
  * `--rpcport` HTTP-RPC server listening port (default: 8545)
  * `--rpcapi` API's offered over the HTTP-RPC interface (default: "eth,net,web3")
  * `--rpccorsdomain` Comma separated list of domains from which to accept cross origin requests (browser enforced)
  * `--ws` Enable the WS-RPC server
  * `--wsaddr` WS-RPC server listening interface (default: "localhost")
  * `--wsport` WS-RPC server listening port (default: 8546)
  * `--wsapi` API's offered over the WS-RPC interface (default: "eth,net,web3")
  * `--wsorigins` Origins from which to accept websockets requests
  * `--ipcdisable` Disable the IPC-RPC server
  * `--ipcapi` API's offered over the IPC-RPC interface (default: "admin,debug,eth,miner,net,personal,shh,txpool,web3")
  * `--ipcpath` Filename for IPC socket/pipe within the datadir (explicit paths escape it)

You'll need to use your own programming environments' capabilities (libraries, tools, etc) to connect
via HTTP, WS or IPC to a sero node configured with the above flags and you'll need to speak [JSON-RPC](http://www.jsonrpc.org/specification)
on all transports. You can reuse the same connection for multiple requests!

**Note: Please understand the security implications of opening up an HTTP/WS based transport before
doing so! Hackers on the internet are actively trying to subvert Sero nodes with exposed APIs!
Further, all browser tabs can access locally running webservers, so malicious webpages could try to
subvert locally available APIs!**

### Operating a private network

Maintaining your own private network is more involved as a lot of configurations taken for granted in
the official networks need to be manually set up.

```

#### Creating the rendezvous point

With all nodes that you want to run initialized to the desired genesis state, you'll need to start a
bootstrap node that others can use to find each other in your network and/or over the internet. The
clean way is to configure and run a dedicated bootnode:

```
$ bootnode --genkey=boot.key
$ bootnode --nodekey=boot.key
```

With the bootnode online, it will display an [`snode` URL](https://github.com/sero-cash/go-sero/wiki/snode-url-format)
that other nodes can use to connect to it and exchange peer information. Make sure to replace the
displayed IP address information (most probably `[::]`) with your externally accessible IP to get the
actual `snode` URL.

*Note: You could also use a full fledged sero node as a bootnode, but it's the less recommended way.*

#### Starting up your member nodes

With the bootnode operational and externally reachable (you can try `telnet <ip> <port>` to ensure
it's indeed reachable), start every subsequent sero node pointed to the bootnode for peer discovery
via the `--bootnodes` flag. It will probably also be desirable to keep the data directory of your
private network separated, so do also specify a custom `--datadir` flag.

```
$ gero --datadir=path/to/custom/data/folder --bootnodes=<bootnode-snode-url-from-above>
```

*Note: Since your network will be completely cut off from the main and test networks, you'll also
need to configure a miner to process transactions and create new blocks for you.*

#### Running a private miner

Mining on the public Sero network need apply license before hand.


In a private network setting however, a single CPU miner instance is more than enough for practical
purposes as it can produce a stable stream of blocks at the correct intervals without needing heavy
resources (consider running on a single thread, no need for multiple ones either). To start a sero
instance for mining, run it with all your usual flags, extended by:

```
$ gero <usual-flags> --mine --minerthreads=1 --serobase=2S4kr7ZHFmgue2kLLngtWnAuHMQgV6jyv34SedvHifm1h3oomx59MEqfEmtnw3mCLnSA2FDojgjTA1WWydxHkUUt
```
There is a script help beginner to create account and start mining. 
[`setup account and start mine` ](https://github.com/sero-cash/go-sero/wiki/start-mine)

Which will start mining blocks and transactions on a single CPU thread, crediting all proceedings to
the account specified by `--serobase`. You can further tune the mining by changing the default gas
limit blocks converge to (`--targetgaslimit`) and the price transactions are accepted at (`--gasprice`).

## Contribution

Thank you for considering to help out with the source code! We welcome contributions from
anyone on the internet, and are grateful for even the smallest of fixes!

If you'd like to contribute to go-sero, please fork, fix, commit and send a pull request
for the maintainers to review and merge into the main code base. 

Please make sure your contributions adhere to our coding guidelines:

 * Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting) guidelines (i.e. uses [gofmt](https://golang.org/cmd/gofmt/)).
 * Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary) guidelines.
 * Pull requests need to be based on and opened against the `master` branch.
 * Commit messages should be prefixed with the package(s) they modify.
   * E.g. "eth, rpc: make trace configs optional"

Please see the [Developers' Guide](https://github.com/sero-cash/go-sero/wiki/Developers'-Guide)
for more details on configuring your environment, managing project dependencies and testing procedures.

## License

The go-sero library (i.e. all code outside of the `cmd` directory) is licensed under the
[GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html), also
included in our repository in the `COPYING.LESSER` file.

The go-sero binaries (i.e. all code inside of the `cmd` directory) is licensed under the
[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also included
in our repository in the `COPYING` file.
