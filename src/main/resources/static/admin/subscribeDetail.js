if (!window.toughradius.admin.subscribeDetail)
    toughradius.admin.subscribeDetail={};


/**
 * 用户订阅详情
 * @param itemid
 * @param callback
 */
toughradius.admin.subscribe.subscribeDetail = function(session,itemid,callback){
    $$(toughradius.admin.subscribe.detailFormID).show();
    $$(toughradius.admin.subscribe.dataViewID).hide();
    var detailWinid = "toughradius.admin.subscribe.subscribeDetail";
    // if($$(detailWinid))
    //     return;
    var formid = detailWinid+"_form";
    var online_tabid = webix.uid();
    webix.ajax().get('/admin/subscribe/detail', {id:itemid}).then(function (result) {
        var resp = result.json();
        if(resp.code>0){
            webix.message({ type: "error", text: resp.msg, expire: 3000 });
            return;
        }
        var subs = resp.data;
        webix.ui({
            id:toughradius.admin.subscribe.detailFormID,
            borderless:true,
            padding:5,
            rows:[
                {
                    view: "toolbar",
                    css: "page-toolbar",
                    cols: [
                        {view:"icon", icon:"user"},
                        {view: "label", label: "用户详情"},
                        { },
                        {
                            view: "button", type: "icon", width: 80, icon: "reply", label: "返回", click: function () {
                                $$(toughradius.admin.subscribe.detailFormID).hide();
                                $$(toughradius.admin.subscribe.dataViewID).show();
                            }
                        }
                    ]
                },
                {
                    view: "tabview",
                    cells: [
                        {
                            header: "用户信息",
                            body: {
                                id: formid,
                                view: "form",
                                scroll: "auto",
                                elementsConfig: { labelWidth: 110 },
                                elements: [
                                    { view: "fieldset", label: "基本信息",  body: {
                                            rows:[
                                                {
                                                    cols: [
                                                        { view: "text", name: "subscriber", label: "订阅帐号", css: "nborder-input", readonly: true, value: subs.subscriber },
                                                        { view: "text", name: "password", label: "认证密码", css: "nborder-input", readonly: true, value: subs.password },
                                                    ]
                                                },
                                                {
                                                    cols:[
                                                        { view: "text", name: "expireime", label: "过期时间", css: "nborder-input", readonly: true, value: subs.expireTime },
                                                        { view: "text", name: "addrPool", label: "地址池", css: "nborder-input",  value: subs.addrPool,readonly:true },
                                                    ]
                                                },
                                                {
                                                    cols:[
                                                        { view: "text", name: "ipAddr", label: "固定IP地址", css: "nborder-input", value: subs.ipAddr ,readonly:true},
                                                        { view: "text", name: "macAddr", label: "MAc地址", css: "nborder-input", value: subs.macAddr ,readonly:true},
                                                    ]
                                                },
                                                {
                                                    cols:[
                                                        { view: "text", name: "inVlan", label: "内层VLAN", css: "nborder-input", value: subs.in_vlan ,readonly:true},
                                                        { view: "text", name: "outVlan", label: "外层VLAN", css: "nborder-input", value: subs.out_vlan ,readonly:true},

                                                    ]
                                                }
                                            ]

                                        }},
                                    { view: "fieldset", label: "授权策略",  body: {
                                            rows:[
                                                {
                                                    cols: [
                                                        { view: "text", name: "activeNum", label: "最大在线", css: "nborder-input", value: subs.activeNum,readonly:true},{}
                                                    ]
                                                },
                                                {
                                                    cols:[
                                                        { view: "radio", name: "bindVlan", label: "绑定VLAN", disabled:true, value: subs.bind_vlan?'1':'0', options: [{ id: '1', value: "是" }, { id: '0', value: "否" }] },
                                                        { view: "radio", name: "bindMac", label: "绑定MAC", disabled:true,value: subs.bindMac?'1':'0', options: [{ id: '1', value: "是" }, { id: '0', value: "否" }] },
                                                    ]
                                                },
                                                {
                                                    cols:[
                                                        { view: "text", name: "upRate", label: "上行速率(Mbps)",  value: subs.upRate,css: "nborder-input", readonly:true},
                                                        { view: "text", name: "downRate", label: "下行速率(Mbps)",  value: subs.downRate,css: "nborder-input",readonly:true},
                                                    ]
                                                },
                                                {
                                                    cols:[
                                                        { view: "text", name: "upPeakRate", label: "突发上行(Mbps)",  value: subs.upPeakRate,css: "nborder-input", readonly:true},
                                                        { view: "text", name: "downPeakRate", label: "突发下行(Mbps)",  value: subs.downPeakRate,css: "nborder-input",readonly:true},
                                                    ]
                                                },
                                                {
                                                    cols:[
                                                        { view: "text", name: "upRateCode", label: "上行速率策略",  value: subs.upRateCode,css: "nborder-input",readonly:true},
                                                        { view: "text", name: "downRateCode", label: "下行速率策略",  value: subs.downRateCode,css: "nborder-input",readonly:true},
                                                    ]
                                                },
                                                {
                                                    cols:[
                                                        { view: "text", name: "domain", label: "认证域", css: "nborder-input", value: subs.domain,readonly:true},
                                                        { view: "text", name: "policy", label: "扩展策略", css: "nborder-input", value: subs.policy,readonly:true},
                                                    ]
                                                }

                                            ]
                                        }},
                                    {
                                        view: "treetable",
                                        scroll: "y",
                                        subview: {
                                            borderless: true,
                                            view: "template",
                                            height: 180,
                                            template: "<div style='padding: 5px;'>#msg#</div>"
                                        },
                                        on: {
                                            onSubViewCreate: function (view, item) {
                                                item.msg = item.msg.replace("\n", "<br>");
                                                view.setValues(item);
                                            }
                                        },
                                        columns: [
                                            {
                                                id: "time",
                                                header: ["时间"],
                                                adjust: true,
                                                template: "{common.subrow()} #time#"
                                            },
                                            {id: "msg", header: ["最近 20 条认证失败信息"], fillspace: true}
                                        ],
                                        select: true,
                                        resizeColumn: true,
                                        autoWidth: true,
                                        autoHeight: true,
                                        url: "/admin/syslog/query?start=0&count=20&type=error&username=" + subs.subscriber
                                    }
                                ]
                            }
                        },
                        {
                            header: "在线信息",
                            body: {
                                id:online_tabid,
                                view: "datatable",
                                leftSplit: 1,
                                rightSplit: 2,
                                columns: [
                                    { id: "username", header: ["用户名"], sort: "string", adjust:true },
                                    { id: "acctSessionId", header: ["会话ID"], sort: "string",  adjust:true },
                                    { id: "nasId", header: ["BRAS 标识"], sort: "string", adjust:true },
                                    { id: "acctStartTime", header: ["上线时间"], sort: "string", adjust:true },
                                    { id: "nasAddr", header: ["BRAS IP"], sort: "string" , adjust:true},
                                    { id: "framedIpaddr", header: ["用户 IP"],  sort: "string", adjust:true },
                                    { id: "macAddr", header: ["用户 Mac"],  sort: "string", adjust:true },
                                    { id: "nasPortId", header: ["端口信息"], sort: "string", adjust:true },
                                    {
                                        id: "acctInputTotal", header: ["上传"],  sort: "int", adjust:true, template: function (obj) {
                                            return bytesToSize(obj.acctInputTotal);
                                        }
                                    },
                                    {
                                        id: "acctOutputTotal", header: ["下载"], sort: "int", adjust:true, template: function (obj) {
                                            return bytesToSize(obj.acctOutputTotal);
                                        }
                                    },
                                    { id: "acctInputPackets", header: ["上行数据包"], sort: "string", adjust:true },
                                    { id: "acctOutputPackets", header: ["下行数据包"],  sort: "string", adjust:true},
                                    { id: "_", header: [""],   fillspace:true},
                                    { id: "opt", header: '操作', template: "<span class='table-btn do_clean'><i class='fa fa-unlock'></i> 清理</span> ", width: 100 },
                                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                                ],
                                select: true,
                                resizeColumn: true,
                                autoWidth: true,
                                autoHeight: true,
                                url: "/admin/online/query?keyword=" + subs.subscriber,
                                onClick:{
                                    do_clean: function (e, id) {
                                        var sessionid = this.getItem(id).acctSessionId;
                                        webix.require("admin/online.js?rand="+new Date().getTime(), function () {
                                            toughradius.admin.online.onlineUnlock(sessionid,function(){
                                                $$(online_tabid).load("/admin/online/query?keyword=" + subs.subscriber);
                                                $$(online_tabid).refresh();
                                            });
                                        });
                                    }
                                }
                            }
                        },
                        {
                            header: "认证日志",
                            body: {
                                view:"treetable",
                                scroll:"y",
                                subview:{
                                    borderless:true,
                                    view:"template",
                                    height:180,
                                    template:"<div style='padding: 5px;'>#msg#</div>"
                                },
                                on:{
                                    onSubViewCreate:function(view, item){
                                        item.msg = item.msg.replace("\n","<br>");
                                        view.setValues(item);
                                    }
                                },
                                columns: [
                                    { id: "time", header: ["时间"], width: 180, template:"{common.subrow()} #time#"},
                                    { id: "msg", header: ["最近200条记录"], fillspace:true  }
                                ],
                                select: true,
                                resizeColumn: true,
                                autoWidth: true,
                                autoHeight: true,
                                url: "/admin/syslog/query?username="+subs.subscriber+"&type=radiusd"
                            }
                        },
                        {
                            header: "上网日志",
                            body: {
                                view: "datatable",
                                rightSplit: 1,
                                columns: [
                                    { id: "username", header: ["用户名"], sort: "string" , adjust:true},
                                    { id: "acctSessionId", header: ["会话ID"],sort: "string", adjust:true },
                                    { id: "nasId", header: ["BRAS 标识"],  sort: "string", adjust:true },
                                    { id: "nasAddr", header: ["BRAS IP"],  sort: "string" , adjust:true},
                                    { id: "framedIpaddr", header: ["用户 IP"],  sort: "string" , adjust:true},
                                    { id: "macAddr", header: ["用户 Mac"],  sort: "string" , adjust:true},
                                    { id: "nasPortId", header: ["端口信息"],  sort: "string",  adjust:true},
                                    {
                                        id: "acctInputTotal", header: ["上传"], sort: "int", adjust:true, template: function (obj) {
                                            return bytesToSize(obj.acctInputTotal);
                                        }
                                    },
                                    {
                                        id: "acctOutputTotal", header: ["下载"],  sort: "int", adjust:true, template: function (obj) {
                                            return bytesToSize(obj.acctOutputTotal);
                                        }
                                    },
                                    { id: "acctInputPackets", header: ["上行数据包"],  sort: "string",  adjust:true },
                                    { id: "acctOutputPackets", header: ["下行数据包"],  sort: "string", adjust:true },
                                    { id: "acctStartTime", header: ["上线时间"],  sort: "string" , adjust:true},
                                    { id: "acctStopTime", header: ["下线时间"],  sort: "string" , adjust:true},
                                    { id: "_", header: [""],   fillspace:true},
                                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                                ],
                                select: true,
                                resizeColumn: true,
                                autoWidth: true,
                                autoHeight: true,
                                url: "/admin/ticket/query?username=" + subs.subscriber
                            }
                        }
                    ]
                }
            ]

        },$$(toughradius.admin.subscribe.detailFormID));
    })
};