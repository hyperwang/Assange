Assange: a blockchain explore based on bitcoind.
=======

Environment
-------

go 1.4+

APIs
-------

* /api/v1/block
* /api/v1/tx
* /api/v1/address
* /api/v1/balance

Options
-------

* --reindex

Regenrate all the data(including blocks index,transactions index, address balance)from bitcoind RPC. This option will cost very long time, please be wareness for it.  
