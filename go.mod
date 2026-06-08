module lina-plugins

go 1.25.0

require (
	lina-plugin-cms v0.0.0
	lina-plugin-linapro-ai-core v0.0.0
	lina-plugin-linapro-content-notice v0.0.0
	lina-plugin-linapro-demo-source v0.0.0
	lina-plugin-linapro-monitor-loginlog v0.0.0
	lina-plugin-linapro-monitor-online v0.0.0
	lina-plugin-linapro-monitor-operlog v0.0.0
	lina-plugin-linapro-monitor-server v0.0.0
	lina-plugin-linapro-ops-demo-guard v0.0.0
	lina-plugin-linapro-org-core v0.0.0
	lina-plugin-linapro-tenant-core v0.0.0
	lina-plugin-linapro-uidentity-cas v0.0.0
	lina-plugin-media v0.0.0
	lina-plugin-sicau-niu v0.0.0
	lina-plugin-water v0.0.0
)

require (
	github.com/Azure/go-ntlmssp v0.1.0 // indirect
	github.com/BurntSushi/toml v1.5.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/clbanning/mxj/v2 v2.7.0 // indirect
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/emirpasic/gods/v2 v2.0.0-alpha // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.8-0.20250403174932-29230038a667 // indirect
	github.com/go-ldap/ldap/v3 v3.4.13 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gogf/gf/contrib/drivers/oracle/v2 v2.10.1 // indirect
	github.com/gogf/gf/v2 v2.10.1 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grokify/html-strip-tags-go v0.1.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mojocn/base64Captcha v1.3.5 // indirect
	github.com/mssola/useragent v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/olekukonko/errors v1.1.0 // indirect
	github.com/olekukonko/ll v0.0.9 // indirect
	github.com/olekukonko/tablewriter v1.1.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/richardlehane/mscfb v1.0.6 // indirect
	github.com/richardlehane/msoleps v1.0.6 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/shirou/gopsutil/v4 v4.26.5 // indirect
	github.com/sijms/go-ora/v2 v2.7.10 // indirect
	github.com/tiendc/go-deepcopy v1.7.2 // indirect
	github.com/tklauser/go-sysconf v0.3.16 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	github.com/xuri/efp v0.0.1 // indirect
	github.com/xuri/excelize/v2 v2.10.1 // indirect
	github.com/xuri/nfp v0.0.2-0.20250530014748-2ddeb826f9a9 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/image v0.25.0 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lina-core v0.0.0 // indirect
)

replace (
	lina-core => ../lina-core
	lina-plugin-cms => ./cms
	lina-plugin-linapro-ai-core => ./linapro-ai-core
	lina-plugin-linapro-content-notice => ./linapro-content-notice
	lina-plugin-linapro-demo-source => ./linapro-demo-source
	lina-plugin-linapro-monitor-loginlog => ./linapro-monitor-loginlog
	lina-plugin-linapro-monitor-online => ./linapro-monitor-online
	lina-plugin-linapro-monitor-operlog => ./linapro-monitor-operlog
	lina-plugin-linapro-monitor-server => ./linapro-monitor-server
	lina-plugin-linapro-ops-demo-guard => ./linapro-ops-demo-guard
	lina-plugin-linapro-org-core => ./linapro-org-core
	lina-plugin-linapro-tenant-core => ./linapro-tenant-core
	lina-plugin-linapro-uidentity-cas => ./linapro-uidentity-cas
	lina-plugin-media => ./media
	lina-plugin-sicau-niu => ./sicau-niu
	lina-plugin-water => ./water
)
