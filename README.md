# Transit

Transit is a package for TCP forwarding

It's different from general TCP proxy forwarding tools, such as
Nginx reverse proxy. Transit's behaviors are more like a TCP listener.
Unlike Nginx which can only perform one-direction TCP forwarding, Transit
can perform two-directions TCP forwarding.

Transit opens *a single* TCP port listens both for downstream and upstream.

When a TCP connection comes from the downstream, transit will forward it to
upstream. At the same time, it can also forward to a 3rd-party host.

    DownStream -> Transit -> UpStream
                    |-> 3rd-party host

When a TCP connection comes from the upstream, transit will forward it to
downstream. However, unlike above, Transit will not forward to 3rd-party host.

    DownStream <- Transit <- Upstream

Transit also supports replacement of the forwarded content(experimental). It uses the
Google re2 syntax (https://github.com/google/re2/wiki/Syntax).

## Usage

    Usage of transit.exe:
        -f string   configure file name (default "/usr/local/etc/transit.json")

## Configure file

key | value
---- | -----
IPArray | DownStream([0]) and UpStream([1]) IP
ThirdPartyAddr | 3rd-party IP in form "xx.xx.xx.xx:xxx"
IP | Bind-IP
Port | Bind-Port
Pattern | Pattern of replacement (Google re2)
Replace | Replace

Example:

    {
      "IPArray": [
        "11.11.11.104",
        "11.11.11.106"
      ],
      "ThirdPartyAddr": "11.11.11.104:8001",
      "IP": "11.11.11.109",
      "Port": 7001
      "Pattern": "(serverip=)'\\d+\\.\\d+\\.\\d+\\.\\d+'",
      "Replace": "$1'11.11.11.109'"
    }
