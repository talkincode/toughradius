if (!window.toughradius.admin.subscribe)
    toughradius.admin.subscribe={};


toughradius.admin.subscribe.dataViewID = "toughradius.admin.subscribe.dataViewID";
toughradius.admin.subscribe.detailFormID = "toughradius.admin.subscribe.detailFormID";
toughradius.admin.subscribe.loadPage = function(session,keyword){
    var tableid = webix.uid();
    var queryid = webix.uid();
    toughradius.admin.subscribe.reloadData = function(){
        $$(toughradius.admin.subscribe.detailFormID).hide();
        $$(toughradius.admin.subscribe.dataViewID).show();
        $$(tableid).refresh();
        $$(tableid).clearAll();
        var params = $$(queryid).getValues();
        var args = [];
        for(var k in params){
            args.push(k+"="+params[k]);
        }
        $$(tableid).load('/admin/subscribe/query?'+args.join("&"));
    };

    var reloadData = toughradius.admin.subscribe.reloadData;

    var cview = {
        id: "toughradius.admin.subscribe",
        css:"main-panel",padding:10,
        rows:[
            {
                id:toughradius.admin.subscribe.dataViewID,
                rows:[
                    {
                        view: "toolbar",
                        height:40,
                        css: "page-toolbar",
                        cols: [
                            {
                                view: "button", type: "form", width: 70, icon: "plus", label: "创建用户", click: function () {
                                    toughradius.admin.subscribe.OpenSubscribeForm(session,function () {
                                        reloadData();
                                    });
                                }
                            },
                            {
                                view: "button", type: "form", width: 70, icon: "plus", label: "批量创建", click: function () {
                                    toughradius.admin.subscribe.batchOpenSubscribeForm(session,function () {
                                        reloadData();
                                    });
                                }
                            },
                            {
                                view: "button", type: "form", width: 55, icon: "key", label: "改密码", click: function () {
                                    var item = $$(tableid).getSelectedItem();
                                    if (item) {
                                        toughradius.admin.subscribe.subscribeUppwd(session, item, function () {
                                            reloadData();
                                        });
                                    } else {
                                        webix.message({ type: 'error', text: "请选择一项", expire: 1500 });
                                    }
                                }
                            },
                            {
                                view: "button", type: "danger", width: 45, icon: "times", label: "删除",  click: function () {
                                    var rows = [];
                                    $$(tableid).eachRow(
                                        function (row) {
                                            var item = $$(tableid).getItem(row);
                                            if (item && item.state === 1) {
                                                rows.push(item.id)
                                            }
                                        }
                                    );
                                    if (rows.length === 0) {
                                        webix.message({ type: 'error', text: "请至少勾选一项", expire: 1500 });
                                    } else {
                                        toughradius.admin.subscribe.subscribeDelete(rows.join(","), function () {
                                            reloadData();
                                        });
                                    }
                                }
                            },
                            {
                                view: "button", type: "danger", width: 45, icon: "times", label: "解绑", click: function () {
                                    var rows = [];
                                    $$(tableid).eachRow(
                                        function (row) {
                                            var item = $$(tableid).getItem(row);
                                            if (item && item.state === 1) {
                                                rows.push(item.id)
                                            }
                                        }
                                    );
                                    if (rows.length === 0) {
                                        webix.message({ type: 'error', text: "请至少勾选一项", expire: 1500 });
                                    } else {
                                        toughradius.admin.subscribe.subscribeRelease(rows.join(","), function () {
                                            reloadData();
                                        });
                                    }
                                }
                            },
                            {}
                        ]
                    },
                    {
                        rows: [
                            {
                                id: queryid,
                                css:"query-form",
                                view: "form",
                                hidden: false,
                                paddingX: 10,
                                paddingY: 5,
                                elementsConfig: {minWidth:180},
                                elements: [
                                    {
                                       type:"space", id:"a1", paddingY:0, rows:[{
                                         type:"space", padding:0,responsive:"a1", cols:[
                                                { view: "datepicker", name: "createTime", label: "创建时间不超过", labelWidth:100, stringResult: true,timepicker: true, format: "%Y-%m-%d" },
                                                { view: "datepicker", name: "expireTime", label: "到期时间不超过", labelWidth:100,stringResult: true, format: "%Y-%m-%d" },
                                                {
                                                    view: "richselect", css:"nborder-input2", name: "status", value:"enabled", label: "用户状态", icon: "caret-down",
                                                    options: [
                                                        { id: 'enabled', value: "正常" },
                                                        { id: 'disabled', value: "停用" },
                                                        { id: 'expire', value: "已到期" }
                                                    ]
                                                },
                                                {view: "text", css:"nborder-input2",  name: "subscriber", label: "用户名", placeholder: "帐号精确匹配", width:240},
                                                {view: "text", css:"nborder-input2",  name: "keyword", label: "",labelWidth:0,  value: keyword || "", placeholder: "帐号模糊匹配", width:180},

                                            {
                                                cols:[
                                                    {view: "button", label: "查询", type: "icon", icon: "search", borderless: true, width: 64, click: function () {
                                                        reloadData();
                                                    }},
                                                    {
                                                        view: "button", label: "重置", type: "icon", icon: "refresh", borderless: true, width: 64, click: function () {
                                                            $$(queryid).setValues({
                                                                createTime: "",
                                                                expireTime: "",
                                                                keyword: "",
                                                                status: ""
                                                            });
                                                        }
                                                    },{}
                                                ]
                                            }
                                         ]}
                                       ]
                                    }
                                ]
                            },
                            {
                                id: tableid,
                                view: "datatable",
                                rightSplit: 2,
                                columns: [
                                    { id: "state", header: { content: "masterCheckbox", css: "center" }, width: 35, css: "center", template: "{common.checkbox()}" },
                                    { id: "id", header: ["ID"], hidden:true},
                                    { id: "subscriber", header: ["帐号"],adjust:true},
                                    { id: "realname", header: ["姓名"],adjust:true},
                                    {
                                        id: "status", header: ["状态"], sort: "string",  adjust:true, template: function (obj) {
                                            if (obj.status === 'enabled' && new Date(obj.expireTime) < new Date()) {
                                                return "<span style='color:orange;'>过期</span>";
                                            } else if (obj.status === 'enabled') {
                                                return "<span style='color:green;'>正常</span>";
                                            } else if (obj.status === 'disabled') {
                                                return "<span style='color:red;'>禁用</span>";
                                            }
                                        }
                                    },
                                    { id: "expireTime", header: ["过期时间"],sort:"date",adjust:true},
                                    { id: "addrPool", header: ["地址池"] ,adjust:true},
                                    { id: "activeNum", header: ["最大在线"],adjust:true},
                                    { id: "ipAddr", header: ["ip 地址"],adjust:true},
                                    { id: "macAddr", header: ["MAC 地址"],adjust:true},
                                    { id: "inVlan", header: ["内层VLAN"],adjust:true},
                                    { id: "outVlan", header: ["外层VLAN"],adjust:true},
                                    { id: "remark", header: ["备注"],fillspace:true},
                                    { id: "opt", header: '操作', adjust:true,template: function(obj){
                                           var actions = [];
                                           actions.push("<span title='测试' class='table-btn do_tester'><i class='fa fa-tty'></i></span> ");
                                           actions.push("<span title='详情' class='table-btn do_detail'><i class='fa fa-eye'></i></span> ");
                                            actions.push("<span title='修改' class='table-btn do_update'><i class='fa fa-edit'></i></span> ");
                                            // actions.push("<span title='删除账号' class='table-btn do_delete'><i class='fa fa-times'></i></span> ");
                                           return actions.join(" ");
                                    }},
                                    { header: { content: "headerMenu" }, headermenu: false, width: 32 }
                                ],
                                select: true,
                                tooltip:true,
                                hover:"tab-hover",
                                autoConfig:true,
                                clipboard:true,
                                resizeColumn: true,
                                autoWidth: true,
                                autoHeight: true,
                                url: "/admin/subscribe/query",
                                pager: "subs_dataPager",
                                datafetch: 40,
                                loadahead: 15,
                                ready: function () {
                                    if (keyword) {
                                        reloadData();
                                    }
                                },
                                on: {
                                    onItemDblClick: function(id, e, node){
                                        var item = this.getSelectedItem();
                                        webix.require("admin/subscribeDetail.js?rand="+new Date().getTime(), function () {
                                            toughradius.admin.subscribe.subscribeDetail(session, item.id, function () {
                                                reloadData();
                                            });
                                        });

                                    }
                                },
                                onClick: {
                                    do_detail: function (e, id) {
                                        var item= this.getItem(id);
                                        webix.require("admin/subscribeDetail.js?rand="+new Date().getTime(), function () {
                                            toughradius.admin.subscribe.subscribeDetail(session, item.id, function () {
                                                reloadData();
                                            });
                                        });
                                    },
                                    do_update: function(e, id){
                                        toughradius.admin.subscribe.subscribeUpdate(session, this.getItem(id), function () {
                                            reloadData();
                                        });
                                    },
                                    do_delete: function(e, id){
                                        toughradius.admin.subscribe.subscribeDelete(this.getItem(id).id, function () {
                                            reloadData();
                                        });
                                    },
                                    do_tester: function(e, id){
                                        toughradius.admin.subscribe.subscribeRadiusTest(session,this.getItem(id));
                                    }
                                }
                            },
                            {
                                paddingY: 3,
                                cols: [
                                    {
                                        view: "richselect", name: "page_num", label: "每页显示", value: 20,width:130,labelWidth:60,
                                        options: [{ id: 20, value: "20" },
                                            { id: 50, value: "50" },
                                            { id: 100, value: "100" },
                                            { id: 500, value: "500" },
                                            { id: 1000, value: "1000" }],on: {
                                            onChange: function (newv, oldv) {
                                                $$("subs_dataPager").define("size",parseInt(newv));
                                                $$(tableid).refresh();
                                                reloadData();
                                            }
                                        }
                                    },
                                    {
                                        id: "subs_dataPager", view: 'pager', master: false, size: 20, group: 5,
                                        template: '{common.first()} {common.prev()} {common.pages()} {common.next()} {common.last()} total:#count#'
                                    },{},

                                ]
                            }
                        ]
                    },
                ]
            },
            {
                id: toughradius.admin.subscribe.detailFormID,
                hidden:true
            }
        ]
    };
    toughradius.admin.methods.addTabView("toughradius.admin.subscribe","user-o","用户管理", cview, true);
    webix.extend($$(tableid), webix.ProgressBar);
};


/**
 * 新用户报装
 * @param session
 * @constructor
 */
toughradius.admin.subscribe.OpenSubscribeForm = function(session, callback){
    var winid = "toughradius.admin.subscribe.OpenSubscribeForm";
    if($$(winid))
        return;
    var formid = winid+"_form";
    webix.ui({
        id:winid,
        view: "window",
        css:"win-body",
        move:true,
        width:340,
        height:480,
        position: "center",
        head: {
            view: "toolbar",
            css:"win-toolbar",

            cols: [
                {view: "icon", icon: "laptop", css: "alter"},
                {view: "label", label: "创建用户"},
                {view: "icon", icon: "times-circle", css: "alter", click: function(){
                        $$(winid).close();
                    }}
            ]
        },
        body: {
            rows:[
                {
                    id: formid,
                    view: "form",
                    scroll: 'y',
                    elementsConfig: { labelWidth: 110 },
                    elements: [
                        { view: "text", name: "subscriber", label: "帐号", validate:webix.rules.isNotEmpty },
                        { view: "text", name: "password", label: "认证密码", validate:webix.rules.isNotEmpty},
                        { view: "datepicker", name: "expireTime", label: "过期时间", stringResult:true, timepicker: true, format: "%Y-%m-%d %h:%i", validate:webix.rules.isNotEmpty },
                        { view: "text", name: "addrPool", label: "地址池" },
                        { view: "text", name: "ipAddr", label: "固定IP地址" , placeholder: "可选，填写后则地址池无效"},
                        { view: "counter", name: "activeNum", label: "最大在线", placeholder: "最大在线", value: 1, min: 1, max: 99999},
                        { view: "radio", name: "bindMac", label: "绑定MAC", value: '0', options: [{ id: '1', value: "是" }, { id: '0', value: "否" }] },
                        { view: "radio", name: "bindVlan", label: "绑定VLAN", value: '0', options: [{ id: '1', value: "是" }, { id: '0', value: "否" }] },
                        { view: "text", name: "upRate", label: "上行速率(Mbps)", validate:webix.rules.isNumber},
                        { view: "text", name: "downRate", label: "下行速率(Mbps)", validate:webix.rules.isNumber}
                    ]
                },
                {
                    view: "toolbar",
                    height:42,
                    css: "page-toolbar",
                    cols: [
                        {},
                        {
                            view: "button", type: "form", width: 100, icon: "check-circle", label: "提交", click: function () {
                                if (!$$(formid).validate()) {
                                    webix.message({ type: "error", text: "请正确填写资料", expire: 1000 });
                                    return false;
                                }
                                var btn = this;
                                btn.disable();
                                var params = $$(formid).getValues();
                                webix.ajax().post('/admin/subscribe/create', params).then(function (result) {
                                    btn.enable();
                                    var resp = result.json();
                                    webix.message({ type: resp.msgtype, text: resp.msg, expire: 3000 });
                                    if (resp.code === 0) {
                                        callback();
                                        $$(winid).close();
                                    }
                                });
                            }
                        },
                        {
                            view: "button", type: "base", width: 100, icon: "times-circle", label: "取消", click: function () {
                                $$(winid).close();
                            }
                        }
                    ]
                }
            ]
        }

    }).show();
};


/**
 * 批量开用户
 * @param session
 * @constructor
 */
toughradius.admin.subscribe.batchOpenSubscribeForm = function(session, callback){
    var winid = "toughradius.admin.subscribe.batchOpenSubscribeForm";
    if($$(winid))
        return;
    var formid = winid+"_form";
    webix.ui({
        id:winid,
        view: "window",
        css:"win-body",
        move:true,
        width:340,
        height:480,
        position: "center",
        head: {
            view: "toolbar",
            css:"win-toolbar",

            cols: [
                {view: "icon", icon: "laptop", css: "alter"},
                {view: "label", label: "批量创建用户"},
                {view: "icon", icon: "times-circle", css: "alter", click: function(){
                        $$(winid).close();
                    }}
            ]
        },
        body: {
            rows:[
                {
                    id: formid,
                    view: "form",
                    scroll: 'y',
                    elementsConfig: { labelWidth: 110 },
                    elements: [
                        { view: "text", name: "userPrefix", label: "帐号前缀", validate:webix.rules.isNotEmpty },
                        { view: "counter", name: "openNum", label: "数量", placeholder: "数量（最大1000）", value: 10, min: 10, max: 1000},
                        { view: "radio", name: "randPasswd", label: "密码类型 ", value: '0', options: [{ id: '1', value: "随机" }, { id: '0', value: "固定" }] },
                        { view: "text", name: "password", label: "固定密码"},
                        { view: "datepicker", name: "expireTime", label: "过期时间", stringResult:true, timepicker: true, format: "%Y-%m-%d %h:%i", validate:webix.rules.isNotEmpty },
                        { view: "text", name: "addrPool", label: "地址池" },
                        { view: "counter", name: "activeNum", label: "最大在线", placeholder: "最大在线", value: 1, min: 1, max: 99999},
                        { view: "radio", name: "bindMac", label: "绑定MAC", value: '0', options: [{ id: '1', value: "是" }, { id: '0', value: "否" }] },
                        { view: "radio", name: "bindVlan", label: "绑定VLAN", value: '0', options: [{ id: '1', value: "是" }, { id: '0', value: "否" }] },
                        { view: "text", name: "upRate", label: "上行速率(Mbps)", validate:webix.rules.isNumber},
                        { view: "text", name: "downRate", label: "下行速率(Mbps)", validate:webix.rules.isNumber}
                    ]
                },
                {
                    view: "toolbar",
                    height:42,
                    css: "page-toolbar",
                    cols: [
                        {},
                        {
                            view: "button", type: "form", width: 100, icon: "check-circle", label: "提交", click: function () {
                                if (!$$(formid).validate()) {
                                    webix.message({ type: "error", text: "请正确填写资料", expire: 1000 });
                                    return false;
                                }
                                var btn = this;
                                btn.disable();
                                var params = $$(formid).getValues();
                                webix.ajax().post('/admin/subscribe/batchcreate', params).then(function (result) {
                                    btn.enable();
                                    var resp = result.json();
                                    webix.message({ type: resp.msgtype, text: resp.msg, expire: 3000 });
                                    if (resp.code === 0) {
                                        callback();
                                        $$(winid).close();
                                    }
                                });
                            }
                        },
                        {
                            view: "button", type: "base", width: 100, icon: "times-circle", label: "取消", click: function () {
                                $$(winid).close();
                            }
                        }
                    ]
                }
            ]
        }

    }).show();
};




toughradius.admin.subscribe.subscribeUpdate = function(session,item,callback){
    var updateWinid = "toughradius.admin.subscribe.subscribeUpdate";
    if($$(updateWinid))
        return;
    var formid = updateWinid+"_form";
    webix.ajax().get('/admin/subscribe/detail', {id:item.id}).then(function (result) {
        var resp = result.json();
        if(resp.code>0){
            webix.message({ type: "error", text: resp.msg, expire: 3000 });
            return;
        }
        var subs = resp.data;
        webix.ui({
            id:updateWinid,
            view: "window",
            css:"win-body",
            move:true,
            resize:true,
            width:360,
            height:480,
            position: "center",
            head: {
                view: "toolbar",
                css:"win-toolbar",

                cols: [
                    {view: "icon", icon: "laptop", css: "alter"},
                    {view: "label", label: "帐号修改"},
                    {view: "icon", icon: "times-circle", css: "alter", click: function(){
                        $$(updateWinid).close();
                    }}
                ]
            },
            body: {
                borderless: true,
                padding:5,
                rows:[
                {
                    id: formid,
                    view: "form",
                    scroll: "y",
                    elementsConfig: { labelWidth: 120 },
                    paddingX:10,
                    elements: [
                        { view: "text", name: "id",  hidden: true, value: subs.id },
                        { view: "text", name: "subscriber", label: "帐号", css: "nborder-input", readonly: true, value: subs.subscriber , validate:webix.rules.isNotEmpty},
                        { view: "text", name: "realname", label: "姓名",value: subs.realname , validate:webix.rules.isNotEmpty},
                        { view: "radio", name: "status", label: "状态", value: subs.status, options: [{ id: 'enabled', value: "正常" }, { id: 'disabled', value: "停用" }] },
                        {
                            view: "datepicker", name: "expireTime", timepicker: true, value:subs.expireTime,
                            label: "过期时间", stringResult: true,  format: "%Y-%m-%d %h:%i", validate: webix.rules.isNotEmpty
                        },
                        { view: "text", name: "addrPool", label: "地址池",  value: subs.addrPool },
                        { view: "radio", name: "bindMac", label: "绑定MAC", value: subs.bindMac?'1':'0', options: [{ id: '1', value: "是" }, { id: '0', value: "否" }] },
                        { view: "radio", name: "bindVlan", label: "绑定VLAN", value: subs.bindVlan?'1':'0', options: [{ id: '1', value: "是" }, { id: '0', value: "否" }] },
                        { view: "text", name: "macAddr", label: "MAc地址",  value: subs.macAddr },
                        { view: "text", name: "ipAddr", label: "固定IP地址",  value: subs.ipAddr },
                        { view: "text", name: "inVlan", label: "内层VLAN",  value: subs.inVlan },
                        { view: "text", name: "outVlan", label: "外层VLAN",  value: subs.outVlan },
                        { view: "text", name: "upRate", label: "上行速率(Mbps)",  value: subs.upRate},
                        { view: "text", name: "downRate", label: "下行速率(Mbps)",  value: subs.downRate},
                        { view: "text", name: "upPeakRate", label: "突发上行速率(Mbps)",  value: subs.upPeakRate, validate:webix.rules.isNumber},
                        { view: "text", name: "downPeakRate", label: "突发下行速率(Mbps)",  value: subs.downPeakRate, validate:webix.rules.isNumber},
                        { view: "counter", name: "activeNum", label: "最大在线", placeholder: "最大在线", value: subs.activeNum, min: 1, max: 99999},
                        { view: "text", name: "upRateCode", label: "上行速率策略",  value: subs.upRateCode},
                        { view: "text", name: "downRateCode", label: "下行速率策略",  value: subs.downRateCode},
                        { view: "text", name: "domain", label: "认证域", value: subs.domain},
                        { view: "text", name: "policy", label: "自定义策略", value:subs.policy},
                        {
                            cols:[
                                { view: "textarea", name: "remark", label: "备注",value: subs.remark, height: 80 }
                            ]
                        }
                    ]
                },
                {
                    height:36,
                    css: "panel-toolbar",
                    cols: [{},
                        {
                            view: "button", type: "form", width: 70, icon: "check-circle", label: "提交", click: function () {
                                if (!$$(formid).validate()) {
                                    webix.message({ type: "error", text: "请正确填写资料", expire: 1000 });
                                    return false;
                                }
                                var btn = this;
                                btn.disable();
                                var params = $$(formid).getValues();
                                params.subs_id = item.id;
                                webix.ajax().post('/admin/subscribe/update', params).then(function (result) {
                                    btn.enable();
                                    var resp = result.json();
                                    webix.message({ type: resp.msgtype, text: resp.msg, expire: 3000 });
                                    if (resp.code === 0) {
                                         callback();
                                         $$(updateWinid).close();
                                    }
                                });
                            }
                        },
                        {view: "button", type: "base", width: 70, icon: "check-circle", label: "取消", click: function(){$$(updateWinid).close()}}

                    ]
                }
            ]}
        }).show();
    })
};

toughradius.admin.subscribe.subscribeUppwd = function(session,item,callback){
    var winid = "toughradius.admin.subscribe.subscribeUppwd";
    if($$(winid))
        return;
    var formid = winid+"_form";
    webix.ajax().get('/admin/subscribe/detail', {id:item.id}).then(function (result) {
        var resp = result.json();
        if(resp.code>0){
            webix.message({ type: resp.msgtype, text: resp.msg, expire: 3000 });
            return;
        }
        var subs = resp.data;
        webix.ui({
            id:winid,
            view: "window",
            css:"win-body",
            move:true,
            width:360,
            height:480,
            position: "center",
            head: {
                view: "toolbar",
                css:"win-toolbar",

                cols: [
                    {view: "icon", icon: "laptop", css: "alter"},
                    {view: "label", label: "帐号密码修改"},
                    {view: "icon", icon: "times-circle", css: "alter", click: function(){
                        $$(winid).close();
                    }}
                ]
            },
            body:{
                borderless: true,
                padding:5,
                rows:[
                {
                    id: formid,
                    view: "form",
                    scroll: "auto",
                    elementsConfig: { labelWidth: 120 },
                    elements: [
                        { view: "text", name: "id",  hidden: true, value: subs.id },
                        { view: "text", name: "subscriber", label: "订阅帐号", css: "nborder-input", readonly: true, value: subs.subscriber },
                        { view: "text", name: "oldpassword", label: "当前密码", css: "nborder-input", readonly: true, value: subs.password },
                        { view: "text", name: "expire_time", label: "过期时间", css: "nborder-input", readonly: true, value: subs.expireTime },
                        { view: "text", name: "password", type: "password", label: "新密码(*)", placeholder: "新密码", validate: webix.rules.isNotEmpty },
                        { view: "text", name: "cpassword", type: "password", label: "确认新密码(*)", placeholder: "确认新密码", validate: webix.rules.isNotEmpty }
                    ]
                },
                {
                    height:36,
                    cols: [{},
                        {
                            view: "button", type: "form", width: 70, icon: "check-circle", label: "提交", click: function () {
                                if (!$$(formid).validate()) {
                                    webix.message({ type: "error", text: "请正确填写资料", expire: 1000 });
                                    return false;
                                }
                                var btn = this;
                                btn.disable();
                                var params = $$(formid).getValues();
                                params.id = item.id;
                                webix.ajax().post('/admin/subscribe/uppwd', params).then(function (result) {
                                    btn.enable();
                                    var resp = result.json();
                                    webix.message({ type: resp.msgtype, text: resp.msg, expire: 3000 });
                                    if (resp.code === 0) {
                                        callback();
                                         $$(winid).close();
                                    }
                                });
                            }
                        },
                        {
                            view: "button", type: "base", icon: "times-circle", width: 70, css: "alter", label: "关闭", click: function () {
                                 $$(winid).close();
                            }
                        }
                    ]
                }
            ]
            }
        }).show(0)
    })
};



toughradius.admin.subscribe.subscribeDelete = function (ids,callback) {
    webix.confirm({
        title: "操作确认",
        ok: "是", cancel: "否",
        text: "删除帐号会同时删除相关所有数据，此操作不可逆，确认要删除吗？",
        width:360,
        callback: function (ev) {
            if (ev) {
                webix.ajax().get('/admin/subscribe/delete', {ids:ids}).then(function (result) {
                    var resp = result.json();
                    webix.message({type: resp.msgtype, text: resp.msg, expire: 1500});
                    if(callback)
                        callback()
                });
            }
        }
    });
};

toughradius.admin.subscribe.subscribeReleaseByquery = function (params,callback) {
    webix.confirm({
        title: "操作确认",
        ok: "是", cancel: "否",
        text: "将根据查询条件批量解除帐号的MAC，VLAN绑定，确认要这么做吗？",
        width:270,
        callback: function (ev) {
            if (ev) {
                webix.ajax().get('/admin/subscribe/release_by_query', params).then(function (result) {
                    var resp = result.json();
                    webix.message({type: resp.msgtype, text: resp.msg, expire: 1500});
                    if(callback)
                        callback()
                });
            }
        }
    });
};


toughradius.admin.subscribe.subscribeRelease = function (ids,callback) {
    webix.confirm({
        title: "操作确认",
        ok: "是", cancel: "否",
        text: "确认要释放绑定吗？",
        width:270,
        callback: function (ev) {
            if (ev) {
                webix.ajax().get('/admin/subscribe/release', {ids:ids}).then(function (result) {
                    var resp = result.json();
                    webix.message({type: resp.msgtype, text: resp.msg, expire: 1500});
                    if(callback)
                        callback()
                });
            }
        }
    });
};


toughradius.admin.subscribe.subscribeRadiusTest = function(session,item){
    var winid = "toughradius.admin.subscribe.subscribeRadiusTest";
    var logvid = webix.uid();
    if($$(winid))
        return;
    var formid = winid+"_form";
    var updateLog = function(iresult){
        var rst = iresult.text();
        console.log(rst);
        $$(logvid).define("template",rst);
        $$(logvid).refresh();
    };
    webix.ui({
        id:winid,
        view: "window",
        css:"win-body",
        move:true,
        width:360,
        height:480,
        position: "center",
        head: {
            view: "toolbar",
            css:"win-toolbar",

            cols: [
                {view: "icon", icon: "laptop", css: "alter"},
                {view: "label", label: "用户拨号测试"},
                {view: "icon", icon: "times-circle", css: "alter", click: function(){
                        $$(winid).close();
                    }}
            ]
        },
        body:{
            borderless: true,
            padding:5,
            rows:[
                {
                    id: formid,
                    view: "form",
                    scroll: "auto",
                    elementsConfig: { labelWidth: 40 },
                    elements: [
                        {
                            view: "fieldset", label: "测试帐号", paddingX: 5, body: {
                                rows: [
                                    {
                                        cols: [
                                            { view: "text", name: "subscriber", label: "帐号", css: "nborder-input", readonly: true, value: item.subscriber },
                                            { view: "text", name: "password", label: "密码", css: "nborder-input", readonly: true, value: item.password },

                                        ],
                                    }
                                ]
                            }
                        },
                        {
                            id: logvid,
                            width:320,
                            view:"template",
                            css:"web-console",
                            borderless: true,
                            scroll:"y",
                            template:""
                        }
                    ]
                },
                {
                    height:72,
                    rows:[
                        {
                            cols:[
                                {
                                    view: "button", type: "form",  icon: "check-circle", label: "PAP 认证", click: function () {
                                        var btn = this;
                                        btn.disable();
                                        var params = {username:item.subscriber,papchap:"pap"}
                                        webix.ajax().get('/admin/radius/auth/test', params).then(function (iresult) {
                                            btn.enable();
                                            updateLog(iresult);
                                        });
                                    }
                                },
                                {
                                    view: "button", type: "form",  icon: "check-circle", label: "CHAP 认证", click: function () {
                                        var btn = this;
                                        btn.disable();
                                        var params = {username:item.subscriber,papchap:"pap"}
                                        webix.ajax().get('/admin/radius/auth/test', params).then(function (iresult) {
                                            btn.enable();
                                            updateLog(iresult);
                                        });
                                    }
                                },
                                {
                                    view: "button", type: "form",  icon: "check-circle", label: "上线", click: function () {
                                        var btn = this;
                                        btn.disable();
                                        var params = {username:item.subscriber,type:"1"}
                                        webix.ajax().get('/admin/radius/acct/test', params).then(function (iresult) {
                                            btn.enable();
                                            updateLog(iresult);
                                        });
                                    }
                                },
                            ]
                        },
                        {
                            cols: [

                                {
                                    view: "button", type: "form",  icon: "check-circle", label: "更新", click: function () {
                                        var btn = this;
                                        btn.disable();
                                        var params = {username:item.subscriber,type:"3"}
                                        webix.ajax().get('/admin/radius/acct/test', params).then(function (iresult) {
                                            btn.enable();
                                            updateLog(iresult);
                                        });
                                    }
                                },
                                {
                                    view: "button", type: "form",  icon: "check-circle", label: "下线", click: function () {
                                        var btn = this;
                                        btn.disable();
                                        var params = {username:item.subscriber,type:"2"}
                                        webix.ajax().get('/admin/radius/acct/test', params).then(function (iresult) {
                                            btn.enable();
                                            updateLog(iresult);
                                        });
                                    }
                                },
                                {
                                    view: "button", type: "danger", icon: "times-circle",  css: "alter", label: "关闭", click: function () {
                                        $$(winid).close();
                                    }
                                }
                            ]
                        }
                    ]
                }
            ]
        }
    }).show(0)
};