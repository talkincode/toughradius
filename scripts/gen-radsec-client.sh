#!/bin/bash

#########################################################
# ToughRADIUS RadSec 客户端证书生成脚本
#
# 此脚本用于生成 RadSec 客户端证书
# 前提：必须先运行 gen-radsec-certs.sh 生成 CA 证书
#
# 使用方法:
#   ./scripts/gen-radsec-client.sh <客户端名称> [选项]
#
# 选项:
#   -d DIR     CA 证书目录 (默认: ./rundata/private)
#   -o DIR     客户端证书输出目录 (默认: ./rundata/private/clients)
#   -h HOST    客户端主机名 (默认: 使用客户端名称)
#   -i IPS     客户端 IP 地址，逗号分隔 (可选)
#   -y DAYS    证书有效期天数 (默认: 3650)
#   -g ORG     组织名称 (默认: ToughRADIUS)
#   -h         显示帮助信息
#
# 示例:
#   # 为 NAS 设备生成客户端证书
#   ./scripts/gen-radsec-client.sh nas01
#
#   # 为 NAS 指定主机名和 IP
#   ./scripts/gen-radsec-client.sh nas02 -h nas02.example.com -i "192.168.1.50"
#
#   # 指定自定义输出目录
#   ./scripts/gen-radsec-client.sh wifi-controller -o ./certs/clients
#
#########################################################

set -e

# 默认配置
CA_DIR="./rundata/private"
OUTPUT_DIR="./rundata/private/clients"
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
ToughRADIUS RadSec 客户端证书生成脚本

用法:
    $0 <客户端名称> [选项]

参数:
    客户端名称      客户端标识符（必需）

选项:
    -d DIR          CA 证书目录 (默认: ./rundata/private)
    -o DIR          客户端证书输出目录 (默认: ./rundata/private/clients)
    -n HOST         客户端主机名 (默认: 使用客户端名称)
    -i IPS          客户端 IP 地址，逗号分隔 (可选)
    -y DAYS         证书有效期天数 (默认: 3650)
    -g ORG          组织名称 (默认: ToughRADIUS)
    -c COUNTRY      国家代码 (默认: CN)
    -p PROVINCE     省份 (默认: Shanghai)
    -l LOCALITY     城市 (默认: Shanghai)
    -u UNIT         组织单元 (默认: IT)
    -h              显示此帮助信息

示例:
    # 基本用法 - 为 NAS 设备生成证书
    $0 nas01

    # 指定主机名和 IP
    $0 nas02 -n nas02.example.com -i "192.168.1.50"

    # 指定多个 IP
    $0 wifi-controller -i "192.168.1.100,10.0.0.100"

    # 自定义所有参数
    $0 branch-radius -d ./certs -o ./certs/clients -n branch.example.com -y 365

前提条件:
    必须先运行 gen-radsec-certs.sh 生成 CA 证书
    CA 目录中必须存在: ca.crt 和 ca.key

生成的文件:
    <客户端名称>.crt    - 客户端证书（公钥）
    <客户端名称>.key    - 客户端私钥

EOF
}

# 解析命令行参数
CLIENT_NAME=""
CLIENT_HOST=""
CLIENT_IPS=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -d)
            CA_DIR="$2"
            shift 2
        ;;
        -o)
            OUTPUT_DIR="$2"
            shift 2
        ;;
        -n)
            CLIENT_HOST="$2"
            shift 2
        ;;
        -i)
            CLIENT_IPS="$2"
            shift 2
        ;;
        -y)
            VALID_DAYS="$2"
            shift 2
        ;;
        -g)
            ORGANIZATION="$2"
            shift 2
        ;;
        -c)
            COUNTRY="$2"
            shift 2
        ;;
        -p)
            PROVINCE="$2"
            shift 2
        ;;
        -l)
            LOCALITY="$2"
            shift 2
        ;;
        -u)
            ORG_UNIT="$2"
            shift 2
        ;;
        -h)
            show_help
            exit 0
        ;;
        -*)
            echo -e "${RED}错误: 无效选项 $1${NC}" >&2
            show_help
            exit 1
        ;;
        *)
            if [ -z "$CLIENT_NAME" ]; then
                CLIENT_NAME="$1"
            else
                echo -e "${RED}错误: 多余的参数 '$1'${NC}" >&2
                show_help
                exit 1
            fi
            shift
        ;;
    esac
done

# 验证客户端名称
if [ -z "$CLIENT_NAME" ]; then
    echo -e "${RED}错误: 必须指定客户端名称${NC}" >&2
    echo ""
    show_help
    exit 1
fi

# 如果未指定主机名，使用客户端名称
if [ -z "$CLIENT_HOST" ]; then
    CLIENT_HOST="$CLIENT_NAME"
fi

# 验证 CA 证书是否存在
if [ ! -f "$CA_DIR/ca.crt" ] || [ ! -f "$CA_DIR/ca.key" ]; then
    echo -e "${RED}错误: CA 证书不存在${NC}"
    echo "请先运行 gen-radsec-certs.sh 生成 CA 证书"
    echo ""
    echo "CA 证书路径:"
    echo "  $CA_DIR/ca.crt"
    echo "  $CA_DIR/ca.key"
    exit 1
fi

# 打印配置信息
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  ToughRADIUS RadSec 客户端证书生成${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}配置信息:${NC}"
echo -e "  客户端名称:   ${YELLOW}$CLIENT_NAME${NC}"
echo -e "  客户端主机名: ${YELLOW}$CLIENT_HOST${NC}"
if [ -n "$CLIENT_IPS" ]; then
    echo -e "  客户端 IP:    ${YELLOW}$CLIENT_IPS${NC}"
else
    echo -e "  客户端 IP:    ${YELLOW}(未指定)${NC}"
fi
echo -e "  CA 目录:      ${YELLOW}$CA_DIR${NC}"
echo -e "  输出目录:     ${YELLOW}$OUTPUT_DIR${NC}"
echo -e "  有效期:       ${YELLOW}$VALID_DAYS 天${NC}"
echo -e "  组织名称:     ${YELLOW}$ORGANIZATION${NC}"
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

# 准备 DNS 名称列表
DNS_NAMES="${CLIENT_HOST},localhost"

# 如果输出目录与 CA 目录不同，需要先复制 CA 证书到输出目录
# 因为 certgen 工具会从输出目录读取 CA 证书
if [ "$OUTPUT_DIR" != "$CA_DIR" ]; then
    echo -e "${YELLOW}复制 CA 证书到输出目录...${NC}"
    cp "$CA_DIR/ca.crt" "$OUTPUT_DIR/"
    cp "$CA_DIR/ca.key" "$OUTPUT_DIR/"
fi

# 构建 certgen 命令
CERTGEN_ARGS=(
    -type client
    -output "$OUTPUT_DIR"
    -days "$VALID_DAYS"
    -org "$ORGANIZATION"
    -ou "$ORG_UNIT"
    -country "$COUNTRY"
    -province "$PROVINCE"
    -locality "$LOCALITY"
    -client-cn "$CLIENT_HOST"
    -client-dns "$DNS_NAMES"
)

# 如果指定了 IP 地址，添加到参数中
if [ -n "$CLIENT_IPS" ]; then
    CERTGEN_ARGS+=(-client-ips "$CLIENT_IPS")
fi

# 生成客户端证书
echo -e "${YELLOW}正在生成客户端证书...${NC}"
echo ""

"$CERTGEN_BIN" "${CERTGEN_ARGS[@]}"

# 如果之前复制了 CA 证书，现在删除它们（保持输出目录干净）
if [ "$OUTPUT_DIR" != "$CA_DIR" ]; then
    rm -f "$OUTPUT_DIR/ca.crt" "$OUTPUT_DIR/ca.key"
fi

# 重命名客户端证书文件
if [ -f "$OUTPUT_DIR/client.crt" ]; then
    mv "$OUTPUT_DIR/client.crt" "$OUTPUT_DIR/${CLIENT_NAME}.crt"
    echo -e "${GREEN}✓ 客户端证书已重命名为 ${CLIENT_NAME}.crt${NC}"
fi

if [ -f "$OUTPUT_DIR/client.key" ]; then
    mv "$OUTPUT_DIR/client.key" "$OUTPUT_DIR/${CLIENT_NAME}.key"
    echo -e "${GREEN}✓ 客户端私钥已重命名为 ${CLIENT_NAME}.key${NC}"
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✓ RadSec 客户端证书生成完成!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}生成的文件:${NC}"
echo -e "  客户端证书: ${YELLOW}$OUTPUT_DIR/${CLIENT_NAME}.crt${NC}"
echo -e "  客户端私钥: ${YELLOW}$OUTPUT_DIR/${CLIENT_NAME}.key${NC}"
echo ""
echo -e "${YELLOW}注意事项:${NC}"
echo -e "  1. 请妥善保管客户端私钥文件"
echo -e "  2. 将证书和私钥部署到客户端设备"
echo -e "  3. 客户端还需要 CA 证书进行验证:"
echo -e "     ${BLUE}$CA_DIR/ca.crt${NC}"
echo ""
echo -e "${GREEN}证书有效期至: $(date -v+${VALID_DAYS}d '+%Y-%m-%d' 2>/dev/null || date -d "+${VALID_DAYS} days" '+%Y-%m-%d' 2>/dev/null || echo '无法计算')${NC}"
echo ""
echo -e "${YELLOW}客户端配置示例 (FreeRADIUS):${NC}"
echo -e "  ${BLUE}home_server radsec {${NC}"
echo -e "    ${BLUE}type = auth+acct${NC}"
echo -e "    ${BLUE}ipaddr = <服务器 IP>${NC}"
echo -e "    ${BLUE}port = 2083${NC}"
echo -e "    ${BLUE}proto = tcp${NC}"
echo -e "    ${BLUE}secret = radsec${NC}"
echo -e "    ${BLUE}status_check = status-server${NC}"
echo -e "  ${BLUE}}${NC}"
echo ""
echo -e "  ${BLUE}tls {${NC}"
echo -e "    ${BLUE}ca_file = $CA_DIR/ca.crt${NC}"
echo -e "    ${BLUE}certificate_file = $OUTPUT_DIR/${CLIENT_NAME}.crt${NC}"
echo -e "    ${BLUE}private_key_file = $OUTPUT_DIR/${CLIENT_NAME}.key${NC}"
echo -e "  ${BLUE}}${NC}"
echo ""
