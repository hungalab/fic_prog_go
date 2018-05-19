FiC FPGA configurator (golang 版)
----

nyacom (C) 2018.05 <kzh@nyacom.net>

* yakuza様のFiC progをgolangで焼き直したバージョン
* おそらく，こっちのほうが中身何やってるかわかりやすいハズ...?
* このプログラマはsudoしなくてもユーザ権限で動く

使い方
----

FiCのRPi上で以下のようにする

    $ go run ficprog.go ledtest.bin

ビルドの仕方
----

golang製なので，ビルドもできる。

    $ go build ficprog.go

すると，ficprog というバイナリがビルドされているはず。


