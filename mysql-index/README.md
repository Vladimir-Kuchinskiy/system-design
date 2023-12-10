## Performance of selection

### 1) Without index

![Explain-Without-Index.png](docs%2Fimages%2FExplain-Withou-Index.png)
![Select-Without-Index.png](docs%2Fimages%2FSelect-Without-Index.png)

### 2) With BTree Index

![Explain-With-Index.png](docs%2Fimages%2FExplain-With-Index.png)
![Select-BTree-Index.png](docs%2Fimages%2FSelect-BTree-Index.png)

### 3) With Hash Index
![Select-Hash-Index.png](docs%2Fimages%2FSelect-Hash-Index.png)


## Performance of insertion for batches of users (2000) for 40mil total users

### 1) Avg latency using innodb_flush_log_at_trx_commit=1
![flush-log-1.png](docs%2Fimages%2Fflush-log-1.png)

### 2) Avg latency using innodb_flush_log_at_trx_commit=2
![flush-log-2.png](docs%2Fimages%2Fflush-log-2.png)

### 3) Avg latency using innodb_flush_log_at_trx_commit=20
![flush-log-20.png](docs%2Fimages%2Fflush-log-20.png)

