#!/bin/bash

#########################################################
# ToughRADIUS RadSec 服务器证书生成脚本
#
# 此脚本用于生成 RadSec (RADIUS over TLS) 服务器所需的证书
# 包括：CA 证书、服务器证书和密钥
#
# 使用方法:
#   ./scripts/gen-radsec-certs.sh [选项]
#
# 选项:
#   -d DIR     输出目录 (默认: ./rundata/private)
#   -h HOST    服务器主机名 (默认: radius.example.com)
#   -i IPS     服务器 IP 地址，逗号分隔 (默认: 127.0.0.1)
#   -y DAYS    证书有效期天数 (默认: 3650)
#   -o ORG     组织名称 (默认: ToughRADIUS)
#   -h         显示帮助信息
#
# 示例:
#   # 使用默认配置生成证书
#   ./scripts/gen-radsec-certs.sh
#
#   # 指定主机名和 IP
#   ./scripts/gen-radsec-certs.sh -h radius.mycompany.com -i "192.168.1.100,10.0.0.1"
#
#   # 自定义所有参数
#   ./scripts/gen-radsec-certs.sh -d ./certs -h radius.local -i 127.0.0.1 -y 1825 -o "My Company"
#
#########################################################

set -e

# 默认配置
OUTPUT_DIR="./rundata/private"
SERVER_HOST="radius.example.com"
SERVER_IPS="127.0.0.1"
VALID_DAYS=3650
ORGANIZATION="ToughRADIUS"
COUNTRY="CN"
PROVINCE="Shanghai"
LOCALITY="Shanghai"
ORG_UNIT="IT"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 显示帮助信息
show_help() {
    cat << EOF
ToughRADIUS RadSec 服务器证书生成脚本

用法:
    $0 [选项]

选项:
    -d DIR      输出目录 (默认: ./rundata/private)
    -n HOST     服务器主机名 (默认: radius.example.com)
    -i IPS      服务器 IP 地址，逗号分隔 (默认: 127.0.0.1)
    -y DAYS     证书有效期天数 (默认: 3650)
    -o ORG      组织名称 (默认: ToughRADIUS)
    -c COUNTRY  国家代码 (默认: CN)
    -p PROVINCE 省份 (默认: Shanghai)
    -l LOCALITY 城市 (默认: Shanghai)
    -u UNIT     组织单元 (默认: IT)
    -h          显示此帮助信息

示例:
    # 使用默认配置
    $0

    # 指定主机名和 IP
    $0 -n radius.mycompany.com -i "192.168.1.100,10.0.0.1"

    # 完整配置
    $0 -d ./certs -n radius.local -i 127.0.0.1 -y 1825 -o "My Company"

生成的文件:
    ca.crt              - CA 根证书（公钥）
    ca.key              - CA 私钥
    radsec.tls.crt      - RadSec 服务器证书（公钥）
    radsec.tls.key      - RadSec 服务器私钥

EOF
}

# 解析命令行参数
while getopts "d:n:i:y:o:c:p:l:u:h" opt; do
    case $opt in
        d) OUTPUT_DIR="$OPTARG" ;;
        n) SERVER_HOST="$OPTARG" ;;
        i) SERVER_IPS="$OPTARG" ;;
        y) VALID_DAYS="$OPTARG" ;;
        o) ORGANIZATION="$OPTARG" ;;
        c) COUNTRY="$OPTARG" ;;
        p) PROVINCE="$OPTARG" ;;
        l) LOCALITY="$OPTARG" ;;
        u) ORG_UNIT="$OPTARG" ;;
        h)
            show_help
            exit 0
        ;;
        \?)
            echo -e "${RED}错误: 无效选项 -$OPTARG${NC}" >&2
            show_help
            exit 1
        ;;
    esac
done

# 打印配置信息
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  ToughRADIUS RadSec 证书生成${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}配置信息:${NC}"
echo -e "  输出目录:     ${YELLOW}$OUTPUT_DIR${NC}"
echo -e "  服务器主机名: ${YELLOW}$SERVER_HOST${NC}"
echo -e "  服务器 IP:    ${YELLOW}$SERVER_IPS${NC}"
echo -e "  有效期:       ${YELLOW}$VALID_DAYS 天${NC}"
echo -e "  组织名称:     ${YELLOW}$ORGANIZATION${NC}"
echo -e "  国家:         ${YELLOW}$COUNTRY${NC}"
echo -e "  省份:         ${YELLOW}$PROVINCE${NC}"
echo -e "  城市:         ${YELLOW}$LOCALITY${NC}"
echo -e "  组织单元:     ${YELLOW}$ORG_UNIT${NC}"
echo ""

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 检查 certgen 工具是否存在
if ! command -v go &> /dev/null; then
    echo -e "${RED}错误: 未找到 Go 工具链${NC}"
    echo "请先安装 Go: https://golang.org/dl/"
    exit 1
fi

# 构建 certgen 工具（如果需要）
CERTGEN_BIN="./certgen"
if [ ! -f "$CERTGEN_BIN" ]; then
    echo -e "${YELLOW}正在构建 certgen 工具...${NC}"
    go build -o "$CERTGEN_BIN" ./cmd/certgen
    echo -e "${GREEN}✓ certgen 工具构建完成${NC}"
    echo ""
fi

# 准备 DNS 名称列表（包含主机名及其通配符）
DNS_NAMES="${SERVER_HOST},*.${SERVER_HOST},localhost"

# 生成证书
echo -e "${YELLOW}正在生成证书...${NC}"
echo ""

"$CERTGEN_BIN" \
-type all \
-output "$OUTPUT_DIR" \
-days "$VALID_DAYS" \
-org "$ORGANIZATION" \
-ou "$ORG_UNIT" \
-country "$COUNTRY" \
-province "$PROVINCE" \
-locality "$LOCALITY" \
-ca-cn "$ORGANIZATION RadSec CA" \
-server-cn "$SERVER_HOST" \
-server-dns "$DNS_NAMES" \
-server-ips "$SERVER_IPS" \
-client-cn "radius-client"

# 重命名服务器证书文件为 RadSec 标准命名
if [ -f "$OUTPUT_DIR/server.crt" ]; then
    mv "$OUTPUT_DIR/server.crt" "$OUTPUT_DIR/radsec.tls.crt"
    echo -e "${GREEN}✓ 服务器证书已重命名为 radsec.tls.crt${NC}"
fi

if [ -f "$OUTPUT_DIR/server.key" ]; then
    mv "$OUTPUT_DIR/server.key" "$OUTPUT_DIR/radsec.tls.key"
    echo -e "${GREEN}✓ 服务器私钥已重命名为 radsec.tls.key${NC}"
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✓ RadSec 服务器证书生成完成!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}生成的文件:${NC}"
echo -e "  CA 证书:          ${YELLOW}$OUTPUT_DIR/ca.crt${NC}"
echo -e "  CA 私钥:          ${YELLOW}$OUTPUT_DIR/ca.key${NC}"
echo -e "  服务器证书:       ${YELLOW}$OUTPUT_DIR/radsec.tls.crt${NC}"
echo -e "  服务器私钥:       ${YELLOW}$OUTPUT_DIR/radsec.tls.key${NC}"
echo -e "  客户端证书:       ${YELLOW}$OUTPUT_DIR/client.crt${NC}"
echo -e "  客户端私钥:       ${YELLOW}$OUTPUT_DIR/client.key${NC}"
echo ""
echo -e "${YELLOW}注意事项:${NC}"
echo -e "  1. 请妥善保管私钥文件（*.key）"
echo -e "  2. CA 证书需要分发给所有 RadSec 客户端"
echo -e "  3. 在 toughradius.yml 中配置证书路径:"
echo -e "     ${BLUE}radiusd:${NC}"
echo -e "       ${BLUE}radsec_ca_cert: $OUTPUT_DIR/ca.crt${NC}"
echo -e "       ${BLUE}radsec_cert: $OUTPUT_DIR/radsec.tls.crt${NC}"
echo -e "       ${BLUE}radsec_key: $OUTPUT_DIR/radsec.tls.key${NC}"
echo ""
echo -e "${GREEN}证书有效期至: $(date -v+${VALID_DAYS}d '+%Y-%m-%d' 2>/dev/null || date -d "+${VALID_DAYS} days" '+%Y-%m-%d' 2>/dev/null || echo '无法计算')${NC}"
echo ""
