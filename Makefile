static	:= "$(shell PWD)/static"
ruleset	:= "$(static)/ruleset"

# 如果文件夹不存在，则创建文件夹
create_static_folder:
	@echo "Creating $(static) folder"
	@mkdir -p $(static)
	@mkdir -p $(ruleset)

download_ruleset:
	@echo "Download Clash RuleSet"
	@curl -sSL https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/reject.txt > static/ruleset/reject.yaml
    @curl -sSL https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/icloud.txt > static/ruleset/icloud.yaml
    @curl -sSL https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/apple.txt > static/ruleset/apple.yaml
    @curl -sSL https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/google.txt > static/ruleset/google.yaml
    @curl -sSL https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/proxy.txt > static/ruleset/proxy.yaml
    @curl -sSL https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/direct.txt > static/ruleset/direct.yaml
    @curl -sSL https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/private.txt > static/ruleset/private.yaml
    @curl -sSL https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/gfw.txt > static/ruleset/gfw.yaml
    @curl -sSL https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/tld-not-cn.txt > static/ruleset/tld-not-cn.yaml
    @curl -sSL https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/telegramcidr.txt > static/ruleset/telegramcidr.yaml
    @curl -sSL https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/cncidr.txt > static/ruleset/cncidr.yaml
    @curl -sSL https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/lancidr.txt > static/ruleset/lancidr.yaml
    @curl -sSL https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/applications.txt > static/ruleset/applications.yaml

download_GeoIP2-CN_MMDB:
	@echo "Download GeoIP2-CN MMDB"
	@curl -sSL https://cdn.jsdelivr.net/gh/Hackl0us/GeoIP2-CN@release/Country.mmdb > static/Country.mmdb

build-tpclash-premium:
	@echo "Build TPClash With Clash Premium"
	GOOS={{.GOOS}} GOARCH={{.GOARCH}} GOARM={{.GOARM}} GOAMD64={{.GOAMD64}} GOMIPS={{.GOMIPS}} \
            go build -trimpath -o build/tpclash-premium-{{.GOOS}}-{{.GOARCH}}{{if .GOAMD64}}-{{.GOAMD64}}{{end}} \
              -ldflags "{{if not .DEBUG}}-w -s{{end}} \
              -X 'main.build={{.BUILD_DATE}}' \
              -X 'main.commit={{.COMMIT_SHA}}' \
              -X 'main.version={{.VERSION}}' \
              -X 'main.clash={{.PREMIUM_VERSION}}' \
              -X 'main.branch=premium' \
              -X 'main.binName=tpclash-premium-{{.GOOS}}-{{.GOARCH}}{{if .GOAMD64}}-{{.GOAMD64}}{{end}}'" \
              {{if .DEBUG}}-gcflags "all=-N -l"{{end}}
# 默认目标
.PHONY: default
default: build-tpclash-premium