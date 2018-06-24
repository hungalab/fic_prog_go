FiC FPGA configurator (golang 版)
----

nyacom (C) 2018.05 <kzh@nyacom.net>

* yakuza様のFiC progをgolangで焼き直したバージョン
* おそらく，こっちのほうが中身何やってるかわかりやすいハズ...?
* このプログラマはsudoしなくてもユーザ権限で動く

使い方
----

FiCのRPi上で以下のようにする

    $ go run ficprog.go ledtest.bin [-m {8, 16}]

-m は Selectmapのモードを指定するオプションで，8bit幅か16bit幅を選択する

ビルドの仕方
----

golang製なので，ビルドもできる。

    $ go build ficprog.go

すると，ficprog というバイナリがビルドされているはず。


ロックファイル
----
GPIO操作中の他のプログラムとの干渉を避けるため，/tmp/gpio.lock というファイルを監視する

* LOCKEXPIRE 秒以上前に作成されたgpio.lockファイルは無視して更新する
* TIMEOUT秒以上gpio.lockが取れないとあきらめる

