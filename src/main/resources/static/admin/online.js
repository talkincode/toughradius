if (!window.xspeedercloud.customer.auth_online)
    xspeedercloud.customer.auth_online={};


xspeedercloud.customer.auth_online.loadPage = function(session){
    xspeedercloud.customer.methods.setToolbar("cube","在线用户管理","auth_online");
    var tableid = webix.uid();
    var queryid = webix.uid();
    var reloadData = function(){
        $$(tableid).define("url", $$(tableid));
        $$(tableid).refresh();
        $$(tableid).clearAll();
        var params = $$(queryid).getValues();
        var args = [];
        for(var k in params){
            args.push(k+"="+params[k]);
        }
        $$(tableid).load('/customer/auth/online/query?'+args.join("&"));
    };
    webix.ui({
        id:xspeedercloud.customer.panelId,
        css:"main-panel",padding:2,
        rows:[
            {
                css:"page-toolbar",
                id: queryid,
                height:50,
                view: "form",
                hidden: false,
                maxWidth: 2000,
                borderless:true,
                elements: [
                    {
                        margin:10,
                        cols:[
                            { view: "text", name: "nas_addr", label: "设备 IP 地址", width:200, placeholder: "设备 IP 地址"},
                            { view: "text", name: "keyword", label: "关键字", labelWidth:50,  placeholder: "账号/IP/MAC", width:240},
                            {view: "button", label: "查询", type: "icon", icon: "search",hotkey:"enter", borderless: true, width: 55,click:function(){
                                reloadData();
                            }},
                            {view: "button", label: "重置", type: "icon", icon: "refresh", borderless: true, width: 55,click:function(){
                                $$(queryid).setValues({
                                    nas_addr: "",
                                    keyword: ""
                                });
                            }}
                        ]
                    }
                ]
            },
            {
                id: tableid,
                view: "datatable",
                leftSplit: 1,
                rightSplit: 1,
                columns: [
                    { id: "username", header: ["用户名"], sort: "string" },
                    { id: "acctSessionId", header: ["会话ID"], width: 150, sort: "string", hidden: true },
                    { id: "nasId", header: ["BRAS 标识"], width: 120, sort: "string" },
                    { id: "acctStartTime", header: ["上线时间"], width: 150, sort: "string" },
                    { id: "nasPaddr", header: ["BRAS IP"], width: 120, sort: "string" },
                    { id: "framedIpaddr", header: ["用户 IP"], width: 120, sort: "string" },
                    { id: "macAddr", header: ["用户 Mac"], width: 140, sort: "string" },
                    { id: "nasPortId", header: ["端口信息"], width: 120, sort: "string" },
                    {
                        id: "acctInputTotal", header: ["上传"], width: 80, sort: "nt", template: function (obj) {
                            return bytesToSize(obj.acctInputTotal);
                        }
                    },
                    {
                        id: "acctOutputTotal", header: ["下载"], width: 80, sort: "int", template: function (obj) {
                            return bytesToSize(obj.acctOutputTotal);
                        }
                    },
                    { id: "acctInputPackets", header: ["上行数据包"], width: 140, sort: "string" },
                    { id: "acctOutputPackets", header: ["下行数据包"], width: 140, sort: "string"},
                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                ],
                select: true,
                maxWidth: 2000,
                maxHeight: 2000,
                resizeColumn: true,
                autoWidth: true,
                autoHeight: true,
                url: "/customer/auth/online/query",
                pager: "dataPager",
                datafetch: 40,
                loadahead: 15,
                on: {}
            },
            {
                paddingY: 3,
                cols: [
                    {
                        id: "dataPager", view: 'pager', master: false, size: 20, group: 5,
                        template: '{common.first()} {common.prev()} {common.pages()} {common.next()} {common.last()} total:#count#'
                    }
                ]
            }
        ]
    },$$(xspeedercloud.customer.pageId),$$(xspeedercloud.customer.panelId));
};


