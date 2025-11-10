#!/bin/bash
#
# RADIUS RFC 文档下载脚本
# 用于下载和更新 RADIUS 相关的 RFC 协议文档
#

# 设置颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 目标目录
DEST_DIR="$(cd "$(dirname "$0")" && pwd)"

echo -e "${BLUE}开始下载 RADIUS RFC 文档到: ${DEST_DIR}${NC}"
echo ""

# RFC 文档列表
declare -A rfcs=(
    # 核心协议
    ["2865"]="RADIUS - Remote Authentication Dial In User Service"
    ["2866"]="RADIUS Accounting"
    
    # 扩展协议
    ["2867"]="RADIUS Accounting Modifications for Tunnel Protocol Support"
    ["2868"]="RADIUS Attributes for Tunnel Protocol Support"
    ["2869"]="RADIUS Extensions"
    ["6929"]="RADIUS Protocol Extensions"
    ["7499"]="Support of Fragmentation of RADIUS Packets"
    
    # NAS 支持
    ["2882"]="Network Access Servers Requirements"
    ["5607"]="RADIUS Authorization for NAS Management"
    
    # IPv6 支持
    ["3162"]="RADIUS and IPv6"
    ["4818"]="RADIUS Delegated-IPv6-Prefix Attribute"
    ["5997"]="Use of Status-Server Packets in RADIUS"
    ["6519"]="RADIUS Extensions Support for Dual-Stack Lite"
    ["6911"]="RADIUS Attributes for IPv6 Access Networks"
    ["6930"]="RADIUS Attribute for IPv6 Rapid Deployment (6rd)"
    ["5969"]="IPv6 Rapid Deployment (6rd) Protocol"
    
    # EAP 核心
    ["2284"]="PPP Extensible Authentication Protocol (EAP) [Obsolete]"
    ["3748"]="Extensible Authentication Protocol (EAP)"
    ["5247"]="EAP Key Management Framework"
    
    # RADIUS-EAP 集成
    ["3579"]="RADIUS Support For EAP"
    ["3580"]="IEEE 802.1X RADIUS Usage Guidelines"
    
    # EAP 认证方法
    ["3851"]="S/MIME Version 3.1 (includes EAP-MD5)"
    ["5216"]="The EAP-TLS Authentication Protocol"
    ["5281"]="EAP-TTLSv0"
    ["4186"]="EAP-SIM"
    ["4187"]="EAP-AKA"
    ["4764"]="EAP-PSK"
    ["5448"]="EAP-IKEv2 / EAP-AKA'"
    ["7170"]="TEAP Version 1"
    ["7542"]="EAP-PWD"
    
    # EAP 扩展
    ["6124"]="EAP Authentication Method Based on EKE"
    
    # 厂商特定
    ["2548"]="Microsoft Vendor-specific RADIUS Attributes"
    ["2759"]="Microsoft MS-CHAP-V2"
    ["4679"]="DSL Forum Vendor-Specific RADIUS Attributes"
    
    # IEEE 802 网络
    ["5904"]="RADIUS Attributes for IEEE 802 Networks"
    ["7268"]="RADIUS Attributes for IEEE 802 Networks (Updated)"
    
    # 隧道/VPN
    ["2809"]="Implementation of L2TP Compulsory Tunneling via RADIUS"
    
    # 高级特性
    ["4372"]="Chargeable User Identity"
    ["3576"]="Dynamic Authorization Extensions (Early)"
    ["5176"]="Dynamic Authorization Extensions (Updated)"
    ["4675"]="RADIUS Attributes for VLAN and Priority Support"
    
    # 认证扩展
    ["4590"]="RADIUS Extension for Digest Authentication"
    ["5090"]="RADIUS Extension for Digest Authentication (Updated)"
    
    # 位置信息
    ["5580"]="Carrying Location Objects in RADIUS and Diameter"
    
    # 移动网络
    ["6572"]="RADIUS Support for Proxy Mobile IPv6"
    
    # 安全传输
    ["6613"]="RADIUS over TCP"
    ["6614"]="RADIUS over TLS (RadSec)"
    ["7585"]="Dynamic Peer Discovery for RADIUS/TLS"
    
    # 实现指南
    ["5080"]="Common RADIUS Implementation Issues"
    
    # Diameter
    ["7155"]="Diameter Network Access Server Application"
)

# 下载函数
download_rfc() {
    local rfc_num=$1
    local description=$2
    local filename="rfc${rfc_num}.txt"
    
    if [ -f "${DEST_DIR}/${filename}" ]; then
        echo -e "${GREEN}✓${NC} RFC ${rfc_num} 已存在，跳过"
    else
        echo -e "  下载 RFC ${rfc_num}: ${description}"
        if curl -s -o "${DEST_DIR}/${filename}" "https://www.rfc-editor.org/rfc/rfc${rfc_num}.txt"; then
            echo -e "${GREEN}✓${NC} RFC ${rfc_num} 下载成功"
        else
            echo -e "${RED}✗${NC} RFC ${rfc_num} 下载失败"
        fi
    fi
}

# 下载所有 RFC
count=0
total=${#rfcs[@]}

for rfc_num in $(echo "${!rfcs[@]}" | tr ' ' '\n' | sort -n); do
    count=$((count + 1))
    echo ""
    echo "[${count}/${total}] RFC ${rfc_num}"
    download_rfc "${rfc_num}" "${rfcs[$rfc_num]}"
done

echo ""
echo -e "${BLUE}======================================${NC}"
echo -e "${GREEN}下载完成！${NC}"
echo ""
echo "已下载的文件列表："
ls -lh "${DEST_DIR}"/*.txt 2>/dev/null | wc -l | xargs echo "总计 RFC 文档数量："
echo ""
echo "文件大小统计："
du -sh "${DEST_DIR}" 2>/dev/null
echo ""
