module tokamak-sybil-resistance

go 1.22.5

<<<<<<< HEAD
require gorm.io/driver/postgres v1.5.9
=======
require (
	github.com/BurntSushi/toml v0.3.1
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/dghubble/sling v1.3.0
	github.com/ethereum/go-ethereum v1.10.6
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.7.2
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/gobuffalo/packr/v2 v2.8.1
	github.com/hermeznetwork/tracerr v0.3.1-0.20210120162744-5da60b576169
	github.com/iden3/go-iden3-crypto v0.0.6-0.20210308142348-8f85683b2cef
	github.com/iden3/go-merkletree v0.0.0-20210308143313-8b63ca866189
	github.com/jmoiron/sqlx v1.2.1-0.20200615141059-0794cb1f47ee
	github.com/joho/godotenv v1.3.0
	github.com/lib/pq v1.8.0
	github.com/libp2p/go-libp2p-core v0.8.6
	github.com/libp2p/go-libp2p-discovery v0.5.1
	github.com/libp2p/go-libp2p-kad-dht v0.12.2
	github.com/libp2p/go-libp2p-pubsub v0.5.0
	github.com/multiformats/go-multiaddr v0.3.3
	github.com/rubenv/sql-migrate v0.0.0-20200616145509-8d140a17f351
	github.com/russross/meddler v1.0.1
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli v1.22.1
	go.uber.org/zap v1.16.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
)
>>>>>>> b1bb9df (Merged sequencer initialisation changes into coordinator node initialisation)

require (
<<<<<<< HEAD
	github.com/DataDog/zstd v1.4.5 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
=======
	github.com/Microsoft/go-winio v0.6.2 // indirect
=======
>>>>>>> 34f8275 (Added initial setup for node in sequencer)
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/VictoriaMetrics/fastcache v1.12.2 // indirect
>>>>>>> d06ea65 (Added node initialisation for sequencer and initialisation cmd in main)
	github.com/beorn7/perks v1.0.1 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
=======
	github.com/bits-and-blooms/bitset v1.10.0 // indirect
=======
>>>>>>> 34f8275 (Added initial setup for node in sequencer)
=======
	github.com/StackExchange/wmi v0.0.0-20180116203802-5d049714c4a6 // indirect
	github.com/VictoriaMetrics/fastcache v1.6.0 // indirect
	github.com/allegro/bigcache v1.2.1 // indirect
>>>>>>> 2d9a15e (updated dependencies)
	github.com/btcsuite/btcd v0.22.0-beta // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/cockroachdb/errors v1.8.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20190617123548-eb05cc24525f // indirect
	github.com/cockroachdb/pebble v0.0.0-20201229190758-9e27ae169fdd // indirect
	github.com/cockroachdb/redact v1.0.8 // indirect
	github.com/cockroachdb/sentry-go v0.6.1-cockroachdb.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dchest/blake512 v1.0.0 // indirect
	github.com/deckarep/golang-set v0.0.0-20180603214616-504e848d77ea // indirect
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/gballet/go-libpcsclite v0.0.0-20190607065134-2772fd86a8ff // indirect
<<<<<<< HEAD
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
<<<<<<< HEAD
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gobuffalo/logger v1.0.6 // indirect
	github.com/gobuffalo/packd v1.0.1 // indirect
	github.com/gobuffalo/packr/v2 v2.8.3 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
=======
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gobuffalo/logger v1.0.6 // indirect
	github.com/gobuffalo/packd v1.0.1 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
=======
	github.com/gofrs/flock v0.8.1 // indirect
>>>>>>> 6ec8170 (Added node initialisation for sequencer and initialisation cmd in main)
=======
	github.com/golang/protobuf v1.5.4 // indirect
>>>>>>> 9ff4157 (Added initial setup for node in sequencer)
	github.com/google/go-querystring v1.1.0 // indirect
=======
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-ole/go-ole v1.2.1 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gobuffalo/logger v1.0.3 // indirect
	github.com/gobuffalo/packd v1.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
>>>>>>> 4f4a3c1 (updated dependencies)
	github.com/google/gopacket v1.1.19 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d // indirect
	github.com/holiman/bloomfilter/v2 v2.0.3 // indirect
	github.com/holiman/uint256 v1.2.0 // indirect
	github.com/huin/goupnp v1.0.1-0.20210626160114-33cdcbb30dda // indirect
	github.com/ipfs/go-cid v0.0.7 // indirect
	github.com/ipfs/go-datastore v0.4.5 // indirect
	github.com/ipfs/go-ipfs-util v0.0.2 // indirect
	github.com/ipfs/go-ipns v0.0.2 // indirect
	github.com/ipfs/go-log v1.0.5 // indirect
	github.com/ipfs/go-log/v2 v2.1.3 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/jbenet/goprocess v0.1.4 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/markbates/errx v1.1.0 // indirect
	github.com/markbates/oncer v1.0.0 // indirect
	github.com/markbates/safe v1.0.1 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.0 // indirect
	github.com/prometheus/client_model v0.2.1-0.20210607210712-147c58e9608a // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rubenv/sql-migrate v1.7.0 // indirect
	github.com/russross/meddler v1.0.1 // indirect
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/supranational/blst v0.3.11 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/term v0.22.0 // indirect
=======
=======
	github.com/karalabe/hid v1.0.1-0.20240306101548-573246063e52 // indirect
>>>>>>> 6ec8170 (Added node initialisation for sequencer and initialisation cmd in main)
=======
	github.com/karalabe/usb v0.0.2 // indirect
>>>>>>> 9ff4157 (Added initial setup for node in sequencer)
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
=======
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/karalabe/usb v0.0.0-20190919080040-51dc0efba356 // indirect
	github.com/karrick/godirwalk v1.15.8 // indirect
	github.com/klauspost/cpuid/v2 v2.0.4 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/kr/text v0.1.0 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/libp2p/go-buffer-pool v0.0.2 // indirect
>>>>>>> 4f4a3c1 (updated dependencies)
	github.com/libp2p/go-cidranger v1.1.0 // indirect
	github.com/libp2p/go-eventbus v0.2.1 // indirect
	github.com/libp2p/go-libp2p v0.14.4 // indirect
	github.com/libp2p/go-libp2p-asn-util v0.0.0-20200825225859-85005c6cf052 // indirect
	github.com/libp2p/go-libp2p-kbucket v0.4.7 // indirect
	github.com/libp2p/go-libp2p-peerstore v0.2.7 // indirect
	github.com/libp2p/go-libp2p-record v0.1.3 // indirect
	github.com/libp2p/go-libp2p-swarm v0.5.0 // indirect
	github.com/libp2p/go-msgio v0.0.6 // indirect
	github.com/libp2p/go-netroute v0.1.6 // indirect
	github.com/libp2p/go-openssl v0.0.7 // indirect
	github.com/libp2p/go-sockaddr v0.1.1 // indirect
	github.com/logrusorgru/aurora v0.0.0-20181002194514-a7b3b318ed4e // indirect
	github.com/markbates/errx v1.1.0 // indirect
	github.com/markbates/oncer v1.0.0 // indirect
	github.com/markbates/safe v1.0.1 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.0.3 // indirect
	github.com/multiformats/go-base36 v0.1.0 // indirect
	github.com/multiformats/go-multiaddr-fmt v0.1.0 // indirect
	github.com/multiformats/go-multiaddr-net v0.2.0 // indirect
	github.com/multiformats/go-multibase v0.0.3 // indirect
	github.com/multiformats/go-multihash v0.0.15 // indirect
	github.com/multiformats/go-multistream v0.2.2 // indirect
	github.com/multiformats/go-varint v0.0.6 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/peterh/liner v1.1.1-0.20190123174540-a2c9a5303de7 // indirect
<<<<<<< HEAD
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/polydawn/refmt v0.89.0 // indirect
	github.com/prometheus/client_golang v1.20.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/prometheus/tsdb v0.7.1 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rjeczalik/notify v0.9.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spacemonkeygo/spacelog v0.0.0-20180420211403-2296661a0572 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/status-im/keycard-go v0.1.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/whyrusleeping/go-keyspace v0.0.0-20160322163242-5b898ac5add1 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/otel v1.27.0 // indirect
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/trace v1.27.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/term v0.23.0 // indirect
	gonum.org/v1/gonum v0.15.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	lukechampine.com/blake3 v1.3.0 // indirect
<<<<<<< HEAD
>>>>>>> 73c16ff (Merged sequencer initialisation changes into coordinator node initialisation)
	rsc.io/tmplfunc v0.0.3 // indirect
=======
>>>>>>> 9ff4157 (Added initial setup for node in sequencer)
)

require (
>>>>>>> 7e38234 (W.I.P Basic coordinator pipeline setup)
	github.com/bytedance/sonic v1.11.9 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/cbergoon/merkletree v0.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/cockroachdb/errors v1.11.3 // indirect
	github.com/cockroachdb/fifo v0.0.0-20240606204812-0bbfbd93a7ce // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v1.1.1 // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/gabriel-vasile/mimetype v1.4.4 // indirect
	github.com/getsentry/sentry-go v0.27.0 // indirect
	github.com/gin-contrib/cors v1.7.2
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-gonic/gin v1.10.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/go-playground/validator/v10 v10.22.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
<<<<<<< HEAD
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/iden3/go-merkletree v0.1.0
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
=======
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
<<<<<<< HEAD
	github.com/iden3/go-merkletree v0.1.0
>>>>>>> 5c80401 (Cleared commented code and added basic setup for node initialisation)
=======
	github.com/iden3/go-merkletree v0.0.0-20210308143313-8b63ca866189
>>>>>>> 34f8275 (Added initial setup for node in sequencer)
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/libp2p/go-libp2p-discovery v0.5.1
	github.com/libp2p/go-libp2p-kad-dht v0.26.1
	github.com/libp2p/go-libp2p-pubsub v0.12.0
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
=======
>>>>>>> 4f4a3c1 (updated dependencies)
	github.com/pkg/errors v0.9.1 // indirect
<<<<<<< HEAD
<<<<<<< HEAD
	github.com/prometheus/client_golang v1.12.0 // indirect
	github.com/prometheus/client_model v0.2.1-0.20210607210712-147c58e9608a // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/crypto v0.25.0 // indirect
	golang.org/x/exp v0.0.0-20230626212559-97b1e661b5df // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
=======
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/exp v0.0.0-20240808152545-0cdaa3abc0fa // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
>>>>>>> 5c80401 (Cleared commented code and added basic setup for node initialisation)
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/gorm v1.25.10 // indirect
=======
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/tsdb v0.7.1 // indirect
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/spacemonkeygo/spacelog v0.0.0-20180420211403-2296661a0572 // indirect
	github.com/status-im/keycard-go v0.0.0-20190316090335-8537d3370df4 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210305035536-64b5b1c73954 // indirect
	github.com/tklauser/go-sysconf v0.3.5 // indirect
	github.com/tklauser/numcpus v0.2.2 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/ugorji/go/codec v1.1.7 // indirect
	github.com/whyrusleeping/go-keyspace v0.0.0-20160322163242-5b898ac5add1 // indirect
	github.com/whyrusleeping/timecache v0.0.0-20160911033111-cfcb2f1abfee // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
	golang.org/x/exp v0.0.0-20200513190911-00229845015e // indirect
	golang.org/x/net v0.0.0-20210423184538-5f58ad60dda6 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1 // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/gorp.v1 v1.7.2 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
>>>>>>> 2d9a15e (updated dependencies)
)
