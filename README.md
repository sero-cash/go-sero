## Go Sero

 Anonymous cryptocurrency based on zero-knowledge proof technology and refactored ethereum protocol by Golang.


* ### Please refer to the official wiki for the tutorial
   <https://wiki.sero.cash/en/index.html?file=Start/from-the-sourcecode-base-on-centos7>


* ### Self-apply for BetaNet mining license
   <https://sero.cash/license/records.html>

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
| **gero** | Our main Gero CLI client. It is the entry point into the Sero network (alpha or dev net), capable of running as a full node (default). It can be used by other processes as a gateway into the Sero network via JSON RPC endpoints exposed on top of HTTP, WebSocket and/or IPC transports. `gero --help` and the [CLI Wiki page](https://github.com/sero-cash/go-Sero/wiki/Command-Line-Options) for command line options. |
| `bootnode` | Stripped down version of our Sero client implementation that only takes part in the network node discovery protocol, but does not run any of the higher level application protocols. It can be used as a lightweight bootstrap node to aid in finding peers in private networks. |


## Running gero

Going through all the possible command line flags is out of scope here (please consult our
[CLI Wiki page](https://github.com/sero-cash/go-sero/wiki/Command-Line-Options)), but we've
enumerated a few common parameter combos to get you up to speed quickly on how you can run your
own sero instance.

## Sero networks

sero have 2 networks: **dev**, **main**(mainnet)

For example 

```gero --dev ... ``` will connect to a private network for development

```gero ...```  will connect to Sero's main network, it is for public testing. License is needed for mining

### Go into console with the Sero network options

By far the most common scenario is people wanting to simply interact with the Sero network:
create accounts; transfer funds; deploy and interact with contracts. 
 **Mining in sero beta network need to be licensed because it is for public testing, More detailed information can be referred to:**
 
<https://wiki.sero.cash/en/index.html?file=Start/from-the-binary-package>

To do so:

```
$ gero --${NETWORK_OPTIONS} console
```

**You are connecting sero beta network if startup gero __without__ ${NETWORK_OPTIONS}, main network is not online yet, 
 it will be online soon**

This command will:


 * Start up sero's built-in interactive [JavaScript console](https://github.com/sero-cash/console/blob/master/README.md), (via the trailing `console` subcommand) 
   through which you can invoke all official(here just reference ethereum web3 style) [`web3` methods](https://github.com/ethereum/wiki/wiki/JavaScript-API) .
   This too is optional and if you leave it out you can always attach to an already running sero instance with 
   `gero --datadir=${DATADIR} attach`.
   
 

### Go into console on the Sero **alpha** network

Transitioning towards developers, if you'd like to play around with creating Sero contracts, you
almost certainly would like to do that **without any real money involved** until you get the hang of the
entire system. In other words, instead of attaching to the main network, you want to join the **alpha**
network with your node, which is fully equivalent to the main network, but with play-Sero only.

```
$ gero --dev console
```

The `console` subcommand have the exact same meaning as above. Please see above for their explanations if you've 
skipped to here.

Specifying the `--dev` flag however will reconfigure your sero instance a bit:

 * using the default data directory (`~/.sero` on Linux for example). Note, on OSX
   and Linux this also means that attaching to a running alpha network node requires the use of a custom
   endpoint since `gero attach` will try to attach to a production node endpoint by default. E.g.
   `gero attach <datadir>/alpha/gero.ipc`.
   
*Note: Although there are some internal protective measures to prevent transactions from crossing
over between the main(beta) network and alpha network, you should make sure to always use separate accounts
for play-money and real-money. Unless you manually move accounts, sero will by default correctly
separate the two networks and will not make any accounts available between them.*

### Go into console on the Sero dev network

```
$ gero --dev console
```
With dev option, developer should config bootnode in local private network and develop new functions without affect 

outside Sero networks

#### Operating a dev network

Maintaining your own private dev network is more involved as a lot of configurations taken for granted in
the official networks need to be manually set up.

### Configuration

As an alternative to passing the numerous flags to the `gero` binary, you can also pass a configuration file via:

```
$ gero --config /path/to/your_config.toml
```

To get an idea how the file should look like you can use the `dumpconfig` subcommand to export your existing configuration:

```
$ gero --your-favourite-flags dumpconfig
```




### Programatically interfacing sero nodes

As a developer, sooner rather than later you'll want to start interacting with sero and the Sero
network via your own programs and not manually through the console. To aid this, sero has built-in
support for a JSON-RPC based APIs ([standard APIs](https://github.com/ethereum/wiki/wiki/JSON-RPC) and
[sero specific APIs](https://github.com/sero-cash/console/blob/master/README.md)). These can be
exposed via HTTP, WebSockets and IPC (unix sockets on unix based platforms, and named pipes on Windows).

The IPC interface is enabled by default and exposes all the APIs supported by sero, whereas the HTTP
and WS interfaces need to manually be enabled and only expose a subset of APIs due to security reasons.
These can be turned on/off and configured as you'd expect.

HTTP based JSON-RPC API options:

  * `--rpc` Enable the HTTP-RPC server
  * `--rpcaddr` HTTP-RPC server listening interface (default: "localhost")
  * `--rpcport` HTTP-RPC server listening port (default: 8545)
  * `--rpcapi` API's offered over the HTTP-RPC interface (default: "sero,net,web3")
  * `--rpccorsdomain` Comma separated list of domains from which to accept cross origin requests (browser enforced)
  * `--ws` Enable the WS-RPC server
  * `--wsaddr` WS-RPC server listening interface (default: "localhost")
  * `--wsport` WS-RPC server listening port (default: 8546)
  * `--wsapi` API's offered over the WS-RPC interface (default: "sero,net,web3")
  * `--wsorigins` Origins from which to accept websockets requests
  * `--ipcdisable` Disable the IPC-RPC server
  * `--ipcapi` API's offered over the IPC-RPC interface (default: "admin,debug,sero,miner,net,personal,shh,txpool,web3")
  * `--ipcpath` Filename for IPC socket/pipe within the datadir (explicit paths escape it)

You'll need to use your own programming environments' capabilities (libraries, tools, etc) to connect
via HTTP, WS or IPC to a sero node configured with the above flags and you'll need to speak [JSON-RPC](http://www.jsonrpc.org/specification)
on all transports. You can reuse the same connection for multiple requests!

**Note: Please understand the security implications of opening up an HTTP/WS based transport before
doing so! Hackers on the internet are actively trying to subvert Sero nodes with exposed APIs!
Further, all browser tabs can access locally running webservers, so malicious webpages could try to
subvert locally available APIs!**



#### Creating the communction center point with bootnode

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

*Note: there is bootnodes already available in sero alpha network and sero main network, setup developer's own 
bootnode is supposed to be used for dev network.*

#### Starting up your member nodes

With the bootnode operational and externally reachable (you can try `telnet <ip> <port>` to ensure
it's indeed reachable), start every subsequent sero node pointed to the bootnode for peer discovery
via the `--bootnodes` flag. It will probably also be desirable to keep the data directory of your
private network separated, so do also specify a custom `--datadir` flag.

```
$ gero --datadir=path/to/custom/data/folder --bootnodes=<bootnode-snode-url-from-above>
```

*Note: Since your network will be completely cut off from the sero main and sero alpha networks, you'll also
need to configure a miner to process transactions and create new blocks for you.*

*Note: Mining on the public Sero network need apply license before hand(license@sero.vip). It is will earn sero coins 
in miner's account.*

#### Running a dev network miner


In a dev network setting however, a single CPU miner instance is more than enough for practical
purposes as it can produce a stable stream of blocks at the correct intervals without needing heavy
resources (consider running on a single thread, no need for multiple ones either). To start a sero
instance for mining, run it with all your usual flags, extended by:

```
$ gero <usual-flags> --mine --minerthreads=1 --serobase=2S4kr7ZHFmgue2kLLngtWnAuHMQgV6jyv34SedvHifm1h3oomx59MEqfEmtnw3mCLnSA2FDojgjTA1WWydxHkUUt
```
Which will start mining blocks and transactions on a single CPU thread, crediting all proceedings to
the account specified by `--serobase`. You can further tune the mining by changing the default gas
limit blocks converge to (`--targetgaslimit`) and the price transactions are accepted at (`--gasprice`).

If beginner want to do mining directlly , There is a script help beginner to create account and start mining. 
[`setup account and start mine` ](https://github.com/sero-cash/go-sero/wiki/start-mine)

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
   * E.g. "sero, rpc: make trace configs optional"

Please see the [Developers' Guide](https://github.com/sero-cash/go-sero/wiki/Developers'-Guide)
for more details on configuring your environment, managing project dependencies and testing procedures.

## Community resources

**Wechat:**  SERO9413

**Discord:**  <https://discord.gg/n5HVxE>

**Twitter:**  <https://twitter.com/SEROdotCASH>

**Telegram:**  <https://t.me/SeroOfficial>

**Gitter:**  <https://gitter.im/sero-cash/Lobby?utm_source=share-link&utm_medium=link&utm_campaign=share-link>


## Other resources

**Official Website:** <https://sero.cash>

**White Paper:** <http://sero-media.s3-website-ap-southeast-1.amazonaws.com/Sero_ENG_V1.06.pdf>

**WIKI:** <https://wiki.sero.cash/zh/index.html?file=home-Home>

**Block Explorer:** <https://explorer.web.sero.cash/blocks.html>

**Introduction Video:** <https://v.qq.com/x/page/s0792e921ok.html>


## License

The go-sero library (i.e. all code outside of the `cmd` directory) is licensed under the
[GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html), also
included in our repository in the `COPYING.LESSER` file.

The go-sero binaries (i.e. all code inside of the `cmd` directory) is licensed under the
[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also included
in our repository in the `COPYING` file.

*Note: Go Sero inherit with licenses of ethereum.*
