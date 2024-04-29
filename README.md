# brc20-balance-monitor

## http server

启动http服务, 默认端口是`8090`, 也可以通过`-port=8090`来指定
```shell
cargo build --release
go run api/main.go [-port=8090]
```

## scan block server
启动区块索引服务, 默认是处理测试网，可以通过`-network=testnet(mainnet)`来指定
```shell
go run cmd/server.go [-network=testnet]
```

## robot server
启动机器人服务
```shell
go run robot/robot.go
```

## 设置环境变量

环境变量的添加参考`.env.example`文件
```shell
DBUSER= #PostgreSQL 用户名
PASSWORD= #PostgreSQL 密码
HOST= # PostgreSQL host
PORT= # PostgreSQL 端口
DBName= # PostgreSQL 库名
StartBlock= # 开始同步的区块号 例如(1213513)
Interval= # 扫描区块的时间间隔，单位秒
GIN_MODE= # gin框架的模式，测试环境使用debug, 正式可以使用release或者debug均可
ENDPOINT= # 请求平台api的地址 例如https://prod-testnet.prod.findora.org
PLATINNERPORT= # 平台内部接口端口8668
PLATAPIPORT= # 平台api端口 26657
UNISATDOMAIN= # unisat 请求域名 正式: https://open-api.unisat.io, 测试: https://open-api-testnet.unisat.io
UNISATAPIKEY= # unisat 平台apikey
AIRDROPUSER= # 发送空投的账户
AIRDROPMNEMONIC= # 发送空投的账户助记词
AIRDROPAMOUNT= # 空投数量(1000)
REDISURL= # redis信息,格式为[redis://user:password@localhost:6789/3?dial_timeout=3&db=1&read_timeout=6s&max_retries=2]  例如redis://localhost:6379/0
RATELIMITSECOND= # 请求unisat每秒限制, 例如(5)
RATELIMITDAY= # 请求unisat每天限制, 例如(10000)
ROBOTTICK= # 机器人操作的tick, 测试网，test
```
