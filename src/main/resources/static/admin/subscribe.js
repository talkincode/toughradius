if (!window.toughradius.admin.subscribe)
    toughradius.admin.subscribe={};


toughradius.admin.subscribe.dataViewID = "toughradius.admin.subscribe.dataViewID";
toughradius.admin.subscribe.detailFormID = "toughradius.admin.subscribe.detailFormID";
toughradius.admin.subscribe.loadPage = function(session,keyword){
    toughradius.admin.methods.setToolbar("user-o","客户业务受理","subscribe");
    var tableid = webix.uid();
    var treeid = webix.uid();
    var sideid = webix.uid();
    var queryid = webix.uid();
    toughradius.admin.subscribe.reloadData = function(){
        $$(toughradius.admin.subscribe.detailFormID).hide();
        $$(toughradius.admin.subscribe.dataViewID).show();
        $$(tableid).define("url", $$(tableid));
        $$(tableid).refresh();
        $$(tableid).clearAll();
        var params = $$(queryid).getValues();
        var args = [];
        for(var k in params){
            args.push(k+"="+params[k]);
        }
        $$(tableid).load('/admin/subscribe/query?'+args.join("&"));
    }
    var exportData = function(){
        var params = $$(queryid).getValues();
        params.export = 1;
        webix.ajax().get('/admin/subscribe/export', params).then(function (result) {
            var resp = result.json();
            webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
        });
    }
    var backupExportData = function(){
        var params = $$(queryid).getValues();
        params.export = 1;
        webix.ajax().get('/admin/subscribe/backup_export', params).then(function (result) {
            var resp = result.json();
            webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
        });
    }
    var releaseData = function(){
        var params = $$(queryid).getValues();
        toughradius.admin.subscribe.subscribeReleaseByquery(params,function(){
            reloadData();
        })
    }
    var reloadData = toughradius.admin.subscribe.reloadData;
    toughradius.admin.initUploadApi("toughradius.admin_subscribe_upload", "/admin/subscribe/import", function () {
        toughradius.admin.methods.showBusyBar(tableid,1000,reloadData);
    });
    webix.ui({
        id: toughradius.admin.panelId,
        css:"main-panel",padding:2,
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
                                view: "button", type: "form", width: 45, icon: "plus", label: "报装", click: function () {
                                    toughradius.admin.methods.requirejs("subscribe_open",session,function(){
                                        toughradius.admin.subscribe.OpenSubscribeForm(session);
                                    })
                                }
                            },
                            {
                                view: "button", type: "base", width: 45, icon: "upload", label: "导入",
                                hidden: !hasPerms(session, ['subscribe_import']), click: function () {
                                    $$("toughradius.admin_subscribe_upload").fileDialog({});
                                }
                            },
                            {
                                view: "button", type: "form", width: 45, icon: "tag", label: "工单", hidden: !hasPerms(session, ['issues_add']), click: function () {
                                    var item = $$(tableid).getSelectedItem();
                                    if (item) {
                                        toughradius.admin.subscribe.issuesAdd(session, item, function () {
                                            reloadData();
                                        });
                                    } else {
                                        webix.message({ type: 'error', text: "请选择一项", expire: 1500 });
                                    }
                                }
                            },
                            {
                                view: "button", type: "form", width: 55, icon: "key", label: "改密码", hidden: !hasPerms(session, ['subscribe_uppwd']), click: function () {
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
                                view: "button", type: "form", width: 45, icon: "gift", label: "变更", hidden: !hasPerms(session, ['subscribe_change']), click: function () {
                                    var item = $$(tableid).getSelectedItem();
                                    if (item) {
                                        toughradius.admin.subscribe.subscribeChange(session, item, function () {
                                            reloadData();
                                        });
                                    } else {
                                        webix.message({ type: 'error', text: "请选择一项", expire: 1500 });
                                    }
                                }
                            },
                            {
                                view: "button", type: "form", width: 45, icon: "step-forward", label: "续费", hidden: !hasPerms(session, ['subscribe_renew']), click: function () {
                                    var item = $$(tableid).getSelectedItem();
                                    if (item) {
                                        toughradius.admin.subscribe.subscribeRenew(session, item, function () {
                                            reloadData();
                                        });
                                    } else {
                                        webix.message({ type: 'error', text: "请选择一项", expire: 1500 });
                                    }
                                }
                            },
                            {
                                view: "button", type: "form", width: 45, icon: "stop-circle", label: "停机", hidden: !hasPerms(session, ['subscribe_pause']), click: function () {
                                    var rows = [];
                                    $$(tableid).eachRow(
                                        function (row) {
                                            var item = $$(tableid).getItem(row);
                                            if (item && item.state === 1) {
                                                if(item.status!="enabled"){
                                                    webix.message({ type: 'error', text: "用户"+item.subscriber+"当前状态不支持停机,被忽略", expire: 1500 });
                                                    return;
                                                }else{
                                                    rows.push(item.id)
                                                }

                                            }
                                        }
                                    );
                                    if (rows.length === 0) {
                                        webix.message({ type: 'error', text: "请至少勾选一项", expire: 1500 });
                                    } else {
                                        toughradius.admin.subscribe.subscribePause(rows.join(","), function () {
                                            reloadData();
                                        });
                                    }
                                }
                            },
                            {
                                view: "button", type: "form", width:45, icon: "retweet", label: "复机", hidden: !hasPerms(session, ['subscribe_resume']), click: function () {
                                    var item = $$(tableid).getSelectedItem();
                                    if (item) {
                                        if(item.status!="pause"){
                                            webix.message({ type: 'error', text: "当前状态不支持复机", expire: 1500 });
                                            return;
                                        }
                                        toughradius.admin.subscribe.subscribeResume(session, item, function () {
                                            reloadData();
                                        });
                                    } else {
                                        webix.message({ type: 'error', text: "请选择一项", expire: 1500 });
                                    }
                                }
                            },
                            {
                                view: "button", type: "form", width: 45, icon: "arrow-circle-right", label: "移机", hidden: !hasPerms(session, ['subscribe_move']), click: function () {
                                    var item = $$(tableid).getSelectedItem();
                                    if (item) {
                                        if(item.status!="enabled"){
                                            webix.message({ type: 'error', text: "当前状态不支持移机", expire: 1500 });
                                            return;
                                        }
                                        toughradius.admin.subscribe.subscribeMove(session, item, function () {
                                            reloadData();
                                        });
                                    } else {
                                        webix.message({ type: 'error', text: "请选择一项", expire: 1500 });
                                    }
                                }
                            },
                            {
                                view: "button", type: "form", width: 45, icon: "arrow-circle-right", label: "缴费", hidden: !hasPerms(session, ['subscribe_pay']), click: function () {
                                    var item = $$(tableid).getSelectedItem();
                                    if (item) {
                                        if(item.status=="cancel"){
                                            webix.message({ type: 'error', text: "当前状态不支持缴费", expire: 1500 });
                                            return;
                                        }
                                        toughradius.admin.subscribe.subscribePay(session, item, function () {
                                            reloadData();
                                        });
                                    } else {
                                        webix.message({ type: 'error', text: "请选择一项", expire: 1500 });
                                    }
                                }
                            },
                            {
                                view: "button", type: "danger", width: 45, icon: "arrow-circle-right", label: "退费", hidden: !hasPerms(session, ['subscribe_refund']), click: function () {
                                    var item = $$(tableid).getSelectedItem();
                                    if (item) {
                                        if(item.status=="cancel"){
                                            webix.message({ type: 'error', text: "当前状态不支持退费", expire: 1500 });
                                            return;
                                        }
                                        toughradius.admin.subscribe.subscribeRefund(session, item, function () {
                                            reloadData();
                                        });
                                    } else {
                                        webix.message({ type: 'error', text: "请选择一项", expire: 1500 });
                                    }
                                }
                            },
                            {
                                view: "button", type: "danger", width: 45, icon: "ban", label: "销户", hidden: !hasPerms(session, ['subscribe_cancel']), click: function () {
                                    var item = $$(tableid).getSelectedItem();
                                    if (item) {
                                        if(item.status!="enabled"){
                                            webix.message({ type: 'error', text: "当前状态不支持销户", expire: 1500 });
                                            return;
                                        }
                                        toughradius.admin.subscribe.subscribeCancel(session, item, function () {
                                            reloadData();
                                        });
                                    } else {
                                        webix.message({ type: 'error', text: "请选择一项", expire: 1500 });
                                    }
                                }
                            },
                            {
                                view: "button", type: "danger", width: 45, icon: "times", label: "删除", hidden: !hasPerms(session, ['subscribe_delete']), click: function () {
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
                                view: "menu",
                                css: { 'background': 'none' },
                                hidden: !hasPerms(session, ['subscribe_release']),
                                width: 75,
                                borderless: true,
                                on:{
                                    onMenuItemClick:function(id){
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
                                            toughradius.admin.subscribe.subscribeRelease(rows.join(","), id,function () {
                                                reloadData();
                                            });
                                        }
                                    }
                                },
                                data: [
                                    {
                                        id: "subscribe_release", value: "解绑", icon: "unlock", submenu: [
                                            { id: "subscribe_release_mac", value: "释放MAC绑定" },
                                            { id: "subscribe_release_invlan", value: "释放内层VLAN" },
                                            { id: "subscribe_release_outvlan", value: "释放外层VLAN" },
                                        ]
                                    }
                                ]
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
                                maxWidth: 2000,
                                elementsConfig: {labelWidth:60},
                                elements: [
                                    {
                                       type:"space", id:"a1", paddingY:0, rows:[{
                                         type:"space", padding:0,responsive:"a1", cols:[
                                            {
                                                cols:[
                                                    {view: "text",hidden:true, name: "node_id", value:session.node_id||""},
                                                    {view: "text", name: "node_name", readonly:true, label: "组织节点", width:180,value:session.node_name||""},
                                                    {view: "button", label: "选择", type: "icon", icon: "angle-down", borderless: true, width: 66,click:function(){
                                                        toughradius.admin.methods.openNodeTree(session.node_id,this.$view,function(item){
                                                            $$(queryid).elements['node_id'].setValue(item.id);
                                                            $$(queryid).elements['node_name'].setValue(item.value);
                                                            $$(queryid).elements['area_id'].setValue("");
                                                            $$(queryid).elements['zone_id'].setValue("");
                                                            var list = $$(queryid).elements['area_id'].getPopup().getList();
                                                            list.clearAll();
                                                            list.load("/admin/area/options?node_id=" + item.id);
                                                            var list2 = $$(queryid).elements['product_id'].getPopup().getList();
                                                            list2.clearAll();
                                                            list2.load("/admin/product/options?node_id=" + item.id);
                                                        });
                                                    }},
                                                    { view: "combo", name: "area_id", label: "区域", labelWidth:50, icon: "caret-down", width:180,maxWidth:180, on:{
                                                            onChange:function(newv, oldv){
                                                                var list = $$(queryid).elements['zone_id'].getPopup().getList();
                                                                list.clearAll();
                                                                list.load("/admin/zone/options?area_id=" + newv);
                                                            }
                                                        }, options: {
                                                        view:"suggest",url:"/admin/area/options?node_id=" + session.node_id
                                                    }},
                                                    { view: "combo", name: "zone_id", label: "小区", labelWidth:50, icon: "caret-down", width:180, options: {
                                                        view:"suggest",data:[]
                                                    }},
                                                    { view: "combo", name: "product_id", labelWidth:50, label: "商品",width:180, icon: "caret-down", options: {
                                                        iew:"suggest",url:"/admin/product/options?node_id=" + session.node_id
                                                    }},
                                                ]
                                            },
                                            {
                                                cols:[
                                                    { view: "datepicker", name: "start_date", label: "创建时间", stringResult: true,timepicker: true, format: "%Y-%m-%d" ,width:180},
                                                    { view: "datepicker", name: "end_date", label: "至", labelWidth:27, stringResult: true,timepicker: true, format: "%Y-%m-%d" ,width:160},
                                                    { view: "datepicker", name: "expire_start_date", label: "到期时间", stringResult: true, format: "%Y-%m-%d" ,width:180},
                                                    { view: "datepicker", name: "expire_end_date", labelWidth:27, label: "至",stringResult: true, format: "%Y-%m-%d" ,width:160},
                                                ]
                                            },
                                            {
                                                cols:[
                                                     {
                                                        view: "richselect", css:"nborder-input2", name: "status", label: "状态", icon: "caret-down", width:140,  labelWidth:50,
                                                        options: [
                                                            { id: 'enabled', value: "正常" },
                                                            { id: 'pause', value: "停用" },
                                                            { id: 'expire', value: "已到期" },
                                                            { id: 'padding', value: "待完成" },
                                                            { id: 'cancel', value: "已销户" },
                                                            { id: 'arrear', value: "欠费" },
                                                            { id: 'freeze', value: "冻结" },
                                                            ]
                                                     },
                                                    {view: "text", css:"nborder-input2",  name: "keyword", label: "关键字",  value: keyword || "", placeholder: "姓名/帐号/手机/邮箱/地址...", width:240},
                                                    { view: "text", name: "export_name", labelWidth:70, label: "导出文件名", value: "用户数据导出", placeholder: "导出文件名" , width:180},
                                                ]
                                            },
                                            {
                                                cols:[
                                                    {view: "button", label: "查询", type: "icon", icon: "search", borderless: true, width: 64, click: function () {
                                                        reloadData();
                                                    }},
                                                    {
                                                        view: "button", label: "重置", type: "icon", icon: "refresh", borderless: true, width: 64, click: function () {
                                                            $$(queryid).setValues({
                                                                start_date: "",
                                                                end_date: "",
                                                                keyword: "",
                                                                status: ""
                                                            });
                                                        }
                                                    },
                                                    {view: "button", label: "解绑", type: "icon", icon: "unlock", borderless: true, width: 64,
                                                        hidden:!hasPerms(session,["subscribe_release_batch"]),click: function () {
                                                        releaseData();
                                                    }},
                                                    {
                                                        view: "button", label: "导出", type: "icon", icon: "download", borderless: true, width: 64,
                                                        hidden:!hasPerms(session,['subscribe_export']),
                                                        click: function () {
                                                            this.disable();
                                                            var msg = "数据导出任务在后台进行，完成后可在[数据下载]页面下载"
                                                            webix.confirm({
                                                                title: "操作确认",
                                                                ok: "是", cancel: "否",
                                                                text: msg,
                                                                width:270,
                                                                callback: function (ev) {
                                                                    if (ev) {
                                                                        exportData();
                                                                    }
                                                                }
                                                            });
                                                            this.enable();
                                                        }
                                                    },
                                                    {
                                                        view: "button", label: "备份", type: "icon", icon: "download", borderless: true, width: 64,
                                                        hidden:session.level!=='super',
                                                        click: function () {
                                                            this.disable();
                                                            var msg = "数据备份导出的数据库可以用作导入恢复, 导出任务在后台进行，完成后可在[数据下载]页面下载"
                                                            webix.confirm({
                                                                title: "操作确认",
                                                                ok: "是", cancel: "否",
                                                                text: msg,
                                                                width:270,
                                                                callback: function (ev) {
                                                                    if (ev) {
                                                                        backupExportData();
                                                                    }
                                                                }
                                                            });
                                                            this.enable();
                                                        }
                                                    }
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
                                leftSplit: 1,
                                rightSplit: 2,
                                columns: [
                                    { id: "state", header: { content: "masterCheckbox", css: "center" }, width: 50, css: "center", template: "{common.checkbox()}" },
                                    { id: "id", header: ["ID"], width: 65, sort: "server" },
                                    { id: "subscriber", header: ["帐号"], sort: "server", width: 120 },
                                    { id: "node_name", header: ["组织"],  width: 100 },
                                    { id: "area_name", header: ["区域"],  width: 100 },
                                    { id: "realname", header: ["姓名"],  width: 150 },
                                    { id: "idcard", header: ["姓名"],  width: 150,hidden:true},
                                    {
                                        id: "status", header: ["状态"], sort: "server", template: function (obj) {
                                            if (obj.status === 'enabled' && new Date(obj.expire_time) < new Date()) {
                                                return "<span style='color:orange;'>过期</span>";
                                            } else if (obj.status === 'enabled') {
                                                return "<span style='color:green;'>正常</span>";
                                            } else if (obj.status === 'disabled') {
                                                return "<span style='color:red;'>禁用</span>";
                                            } else if (obj.status === 'pause') {
                                                return "<span style='color:red;'>暂停</span>";
                                            } else if (obj.status === 'padding') {
                                                return "<span style='color:orange;'>未生效</span>";
                                            } else if (obj.status === 'cancel') {
                                                return "<span style='color:darkred;'>已销户</span>";
                                            } else if (obj.status === 'arrear') {
                                                return "<span style='color:darkred;'>欠费</span>";
                                            } else if (obj.status === 'freeze') {
                                                return "<span style='color:darkgray;'>冻结</span>";
                                            }
                                        }
                                    },
                                    {
                                        id: "auto_renew",hidden:true, header: ["自动续费"], template: function (obj) {
                                            if (obj.auto_renew === 'enabled') {
                                                return "<span style='color:green;'>是</span>";
                                            } else if (obj.auto_renew === 'disabled') {
                                                return "<span style='color:red;'>否</span>";
                                            }
                                        }
                                    },
                                    { id: "product_name", header: ["资费"], width: 160},
                                    { id: "expire_time", header: ["过期时间"], width: 160, sort: "server" },
                                    { id: "addr_pool", header: ["地址池"], width: 160 , hidden:true},
                                    { id: "active_num", header: ["最大在线"], width: 160, hidden:true },
                                    { id: "ip_addr", header: ["ip 地址"], width: 160, hidden:true },
                                    { id: "mac_addr", header: ["MAC 地址"], width: 160, hidden:true },
                                    { id: "in_vlan", header: ["内层VLAN"], width: 160 , hidden:true},
                                    { id: "out_vlan", header: ["外层VLAN"], width: 160 , hidden:true},
                                    { id: "remark", header: ["备注"], fillspace: true },
                                    { id: "opt", header: '操作', template: function(obj){
                                           var actions = [];
                                           actions.push("<span title='测试' class='table-btn do_tester'><i class='fa fa-tty'></i></span> ");
                                           actions.push("<span title='详情' class='table-btn do_detail'><i class='fa fa-eye'></i></span> ");
                                           if(hasPerms(session, ['subscribe_update'])){
                                              actions.push("<span title='修改账号' class='table-btn do_update'><i class='fa fa-edit'></i></span> ");
                                           }
                                           if(hasPerms(session, ['subscribe_delete'])){
                                              actions.push("<span title='删除账号' class='table-btn do_delete'><i class='fa fa-times'></i></span> ");
                                           }
                                           return actions.join(" ");
                                    },width:180},
                                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                                ],
                                select: true,
                                tooltip:true,
                                hover:"tab-hover",
                                autoConfig:true,
                                clipboard:true,
                                maxWidth: 2000,
                                maxHeight: 2000,
                                resizeColumn: true,
                                autoWidth: true,
                                autoHeight: true,
                                url: "/admin/subscribe/query",
                                pager: "dataPager",
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
                                        toughradius.admin.subscribe.subscribeDetail(session, item.id, function () {
                                            reloadData();
                                        });
                                    }
                                },
                                onClick: {
                                    do_detail: function (e, id) {
                                        toughradius.admin.subscribe.subscribeDetail(session, this.getItem(id).id, function () {
                                            reloadData();
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
                                        toughradius.admin.subscribe.subscribeRadiusTest(session,this.getItem(id).id);
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
                                                $$("dataPager").define("size",parseInt(newv));
                                                $$(tableid).refresh();
                                                reloadData();
                                            }
                                        }
                                    },
                                    {
                                        id: "dataPager", view: 'pager', master: false, size: 20, group: 5,
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
    },$$(toughradius.admin.pageId),$$(toughradius.admin.panelId));
    webix.extend($$(tableid), webix.ProgressBar);
};

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
    var order_tabid = webix.uid();
    var issues_tabid = webix.uid();
    var online_tabid = webix.uid();
    var idcard_img1 = webix.uid();
    var idcard_img2 = webix.uid();
    var idcard_id1 = webix.uid();
    var idcard_id2 = webix.uid();
    webix.ajax().get('/admin/subscribe/detail', {id:itemid}).then(function (result) {
        var subs = result.json();
        toughradius.admin.initUploadApi(idcard_id1, "/admin/customer/idcard/upload/"+subs.id+"/1", function () {
            $$(idcard_img1).define('template', "<a href='javascript:void(0);' class='imgopen'><img src='/admin/customer/idcardimg/1/160/"+subs.id+"?rand="+new Date().getTime()+"'/></a>")
            $$(idcard_img1).refresh()
        });
        toughradius.admin.initUploadApi(idcard_id2, "/admin/customer/idcard/upload/"+subs.id+"/2", function () {
            $$(idcard_img2).define('template', "<a href='javascript:void(0);' class='imgopen'><img src='/admin/customer/idcardimg/2/160/"+subs.id+"?rand="+new Date().getTime()+"'/></a>")
            $$(idcard_img2).refresh()
        });
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
                            header: "基本信息",
                            body: {
                                id: formid,
                                view: "form",
                                scroll: "auto",
                                minHeight:360,
                                maxWidth: 2000,
                                maxHeight: 2000,
                                elementsConfig: { labelWidth: 100 },
                                elements: [
                                    {
                                        cols:[
                                            { view: "fieldset", label: "基本信息", body: {
                                                rows:[
                                                    {
                                                        cols: [
                                                            { view: "text", name: "realname",labelWidth:60, label: "用户姓名", css: "nborder-input", value: subs.customer.realname, readonly: true },
                                                            { view: "text", name: "idcard", labelWidth:60,label: "证件号", css: "nborder-input", value: subs.customer.idcard, readonly: true }
                                                        ]
                                                    },
                                                    {
                                                        cols: [
                                                            {view:"radio", name:"gender", labelWidth:60,disabled:true,label: "性别", value:subs.customer.gender,options:[{id:'male',value:"男"}, {id:'female',value:"女"}]},
                                                            { view: "text", name: "mobile",labelWidth:60, label: "手机号码", css: "nborder-input", value: subs.customer.mobile, readonly: true },
                                                        ]
                                                    },
                                                    {
                                                        cols: [
                                                            { view: "text", name: "email",labelWidth:60, label: "电子邮箱", css: "nborder-input", value: subs.customer.email, readonly: true },
                                                            { view: "text", name: "idcard", labelWidth:60,label: "身份证", css: "nborder-input",value: subs.customer.idcard, readonly: true },
                                                        ]
                                                    },
                                                    {
                                                        rows: [
                                                            { view: "text", name: "address",labelWidth:60, label: "地址", css: "nborder-input", value: subs.customer.address, readonly: true },
                                                            { view: "textarea", name: "remark",labelWidth:60, label: "备注", value: subs.remark, readonly: true, height: 60 }
                                                        ]
                                                    }
                                                ]
                                            }},
                                            {width:20},
                                            { view: "fieldset", label: "证件图片", body: {
                                                cols:[
                                                    {
                                                        width:180,
                                                        rows:[
                                                            {view: "button", label: "上传身份证正面", type: "icon", icon: "upload", borderless: true, width: 120, click: function () {
                                                                $$(idcard_id1).fileDialog({});
                                                            }},
                                                            {view:"template", id:idcard_img1, borderless:true,
                                                                template:"<a href='javascript:void(0);' class='imgopen'><img src='/admin/customer/idcardimg/1/160/"+subs.id+"?rand="+new Date().getTime()+"'/></a>", onClick:{
                                                                "imgopen" : function(){
                                                                    toughradius.admin.openImage("customer_idcard_img1",'/admin/customer/idcardimg/1/640/'+subs.id+'?rand='+new Date().getTime(),640,480);
                                                                }
                                                            }}
                                                        ]
                                                    },
                                                    {width:20},
                                                    {
                                                        width:180,
                                                        rows:[
                                                            {view: "button", label: "上传身份证背面", type: "icon", icon: "upload", borderless: true, width: 120, click: function () {
                                                                $$(idcard_id2).fileDialog({});
                                                            }},
                                                            {view:"template", id:idcard_img2, borderless:true,
                                                                template:"<a href='javascript:void(0);' class='imgopen'><img src='/admin/customer/idcardimg/2/160/"+subs.id+"?rand="+new Date().getTime()+"'/></a>", onClick:{
                                                                "imgopen" : function(){
                                                                    toughradius.admin.openImage("customer_idcard_img1",'/admin/customer/idcardimg/2/640/'+subs.id+'?rand='+new Date().getTime(),640,480);
                                                                }
                                                            }}                                                        ]
                                                    }
                                                ]
                                            }}
                                        ]
                                    },
                                    { view: "fieldset", label: "授权信息",  body: {
                                        rows:[
                                            {
                                                cols: [
                                                    { view: "text", name: "subscriber", label: "订阅帐号", css: "nborder-input", readonly: true, value: subs.subscriber },
                                                    { view: "text", name: "password", label: "认证密码", css: "nborder-input", readonly: true, value: subs.password },
                                                    { view: "text", name: "expire_time", label: "过期时间", css: "nborder-input", readonly: true, value: subs.expire_time }
                                                ]
                                            },
                                            {
                                                cols: [
                                                    { view: "text", name: "last_renew", label: "最后续费", css: "nborder-input", readonly: true, value: subs.last_renew },
                                                    { view: "text", name: "last_pause", label: "最后停机", css: "nborder-input", readonly: true, value: subs.last_pause },
                                                    { view: "text", name: "last_resume", label: "最后复机", css: "nborder-input", readonly: true, value: subs.last_resume }
                                                ]
                                            },
                                            {
                                                cols: [
                                                    {view: "text", name: "product_name", label: "商品", css: "nborder-input", value:subs.product.name,readonly:true},
                                                    { view: "text", name: "active_num", label: "最大在线", css: "nborder-input", value: subs.active_num,readonly:true},
                                                    { view: "text", name: "flow_amount", label: "剩余流量", css: "nborder-input", value: bytesToSize(subs.flow_amount),readonly:true}
                                                ]
                                            },
                                            {
                                                cols:[
                                                    { view: "text", name: "addr_pool", label: "地址池", css: "nborder-input",  value: subs.addr_pool,readonly:true },
                                                    { view: "text", name: "mac_addr", label: "MAc地址", css: "nborder-input", value: subs.mac_addr ,readonly:true},
                                                    { view: "text", name: "ip_addr", label: "固定IP地址", css: "nborder-input", value: subs.ip_addr ,readonly:true}
                                                ]
                                            },
                                            {
                                                cols:[
                                                    { view: "text", name: "in_vlan", label: "内层VLAN", css: "nborder-input", value: subs.in_vlan ,readonly:true},
                                                    { view: "text", name: "out_vlan", label: "外层VLAN", css: "nborder-input", value: subs.out_vlan ,readonly:true},
                                                    { view: "radio", name: "bind_vlan", label: "绑定VLAN", disabled:true, value: subs.bind_vlan?'1':'0', options: [{ id: '1', value: "是" }, { id: '0', value: "否" }] },

                                                ]
                                            },
                                            {
                                                cols:[
                                                    { view: "text", name: "up_rate", label: "上行速率(Mbps)",  value: subs.up_rate,css: "nborder-input", readonly:true},
                                                    { view: "text", name: "down_rate", label: "下行速率(Mbps)",  value: subs.down_rate,css: "nborder-input",readonly:true},
                                                    { view: "radio", name: "bind_mac", label: "绑定MAC", disabled:true,value: subs.bind_mac?'1':'0', options: [{ id: '1', value: "是" }, { id: '0', value: "否" }] },

                                                ]
                                            },
                                            {
                                                cols:[
                                                    { view: "text", name: "up_peak_rate", label: "突发上行速率(Mbps)",  value: subs.up_peak_rate,css: "nborder-input", readonly:true},
                                                    { view: "text", name: "down_peak_rate", label: "突发下行速率(Mbps)",  value: subs.down_peak_rate,css: "nborder-input",readonly:true},
                                                    { },

                                                ]
                                            },
                                            {
                                                cols:[
                                                    { view: "text", name: "up_rate_code", label: "上行速率策略",  value: subs.up_rate_code,css: "nborder-input",readonly:true},
                                                    { view: "text", name: "down_rate_code", label: "下行速率策略",  value: subs.down_rate_code,css: "nborder-input",readonly:true},
                                                    { view: "text", name: "domain", label: "认证域", value: subs.domain,css: "nborder-input",readonly:true},
                                                ]
                                            },
                                            {
                                                cols:[
                                                    { view: "radio", name: "free_auth", label: "到期免授权", disabled:true, value: subs.free_auth?'1':'0' , options: [{ id: '1', value: "是" }, { id: '0', value: "否" }] },
                                                    { view: "text", name: "free_auth_uprate", label: "免授权上行速率",  value: subs.free_auth_uprate, css: "nborder-input",readonly:true},
                                                    { view: "text", name: "free_auth_downrate", label: "免授权下行速率", value: subs.free_auth_downrate, css: "nborder-input",readonly:true}

                                                ]
                                            },
                                            {
                                                cols:[
                                                    { view: "text", name: "proxy_user", label: "代理帐号",  value: subs.proxy_user ,css: "nborder-input",readonly:true},
                                                    { view: "text", name: "proxy_pwd", label: "代理帐号密码",  value: subs.proxy_pwd ,css: "nborder-input",readonly:true},
                                                    { view: "text", name: "proxy_vlan", label: "代理帐号VLAN",  value: subs.proxy_vlan ,css: "nborder-input",readonly:true}
                                                ]
                                            },
                                            {
                                                rows: [
                                                    {
                                                        cols:[
                                                            { view: "radio", name: "proxy_enabled", label: "启用代理拨号", disabled:true, value: subs.proxy_enabled?'1':'0', options: [{ id: '1', value: "是" }, { id: '0', value: "否" }]},
                                                            { view: "radio", name: "auto_renew", label: "自动续费",disabled:true, value: subs.auto_renew, options: [{ id: 'enabled', value: "是" }, { id: 'disabled', value: "否" }] },
                                                            { view: "text", name: "policy", label: "自定义策略", value:subs.policy,css: "nborder-input",readonly:true},
                                                        ]
                                                    }
                                                ]
                                            }
                                        ]
                                    }},
                                ]
                            }
                        },
                        {
                            header:"变更历史",
                            body:{
                                view: "datatable",
                                leftSplit: 1,
                                rightSplit: 1,
                                columns: [
                                    { id: "subscriber", header: ["帐号"], sort: "string", width: 120 },
                                    { id: "status", header: ["状态"], sort: "string",width:60},
                                    { id: "auto_renew", header: ["自动续费"], sort: "string"},
                                    { id: "active_num", header: ["在线数限制"], width: 100, sort: "int" },
                                    { id: "addr_pool", header: ["地址池"], width: 100, sort: "string" },
                                    { id: "domain", header: ["认证域"], width: 100, sort: "string" },
                                    { id: "up_rate", header: ["上行速率"], width: 100, sort: "string" },
                                    { id: "down_rate", header: ["下行速率"], width: 100, sort: "string" },
                                    { id: "up_peak_rate", header: ["突发上行速率"], width: 100, sort: "string" },
                                    { id: "down_peak_rate", header: ["突发下行速率"], width: 100, sort: "string" },
                                    { id: "up_rate_code", header: ["上行速率策略"], width: 100, sort: "string" },
                                    { id: "down_rate_code", header: ["下行速率策略"], width: 100, sort: "string" },
                                    { id: "expire_time", header: ["过期时间"], width: 150, sort: "string" },
                                    { id: "remark", header: ["备注"], sort: "string", fillspace: true , hidden:true},
                                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                                ],
                                select: true,
                                maxWidth: 2000,
                                maxHeight: 2000,
                                resizeColumn: true,
                                autoWidth: true,
                                autoHeight: true,
                                url: "/admin/subscribe/historys?id="+subs.id
                            }

                        },
                        {
                            header: "交易记录",
                            body: {
                                id:order_tabid,
                                view: "datatable",
                                columns: [
                                    { id: "subscriber", header: ["订阅帐号"], sort: "string", width:100},
                                    { id: "actual_fee", header: ["金额"], sort: "string", footer: { content: "summColumn" } },
                                    {
                                        id: "order_type", header: ["类型"], sort: "string", template: function (obj) {
                                            if (obj.order_type === 'subscribe') {
                                                return "<span style='color:green;'>订阅</span>";
                                            }else if (obj.order_type === 'subscribe_pre') {
                                                return "<span style='color:blue;'>预订</span>";
                                            } else if (obj.order_type === 'renew') {
                                                return "<span style='color:blue;'>续费</span>";
                                            }else if (obj.order_type === 'subscribe_pre') {
                                                return "<span style='color:blue;'>预订</span>";
                                            } else if (obj.order_type === 'charge') {
                                                return "<span style='color:darkgreen;'>充值</span>";
                                            } else if (obj.order_type === 'move') {
                                                return "<span style='color:blue;'>移机</span>";
                                            } else if (obj.order_type === 'resume') {
                                                return "<span style='color:blue;'>复机</span>";
                                            } else if (obj.order_type === 'change') {
                                                return "<span style='color:blue;'>变更</span>";
                                            } else if (obj.order_type === 'fees') {
                                                return "<span style='color:blue;'>收费项目</span>";
                                            } else if (obj.order_type === 'device') {
                                                return "<span style='color:blue;'>设备租购</span>";
                                            } else if (obj.order_type === 'cancel') {
                                                return "<span style='color:red;'>销户</span>";
                                            } else if (obj.order_type === 'other') {
                                                return "<span style='color:blue;'>其他</span>";
                                            }else if (obj.order_type === 'refund') {
                                                return "<span style='color:orange;'>退费</span>";
                                            }
                                        }
                                    },
                                    {
                                        id: "status", header: ["状态"], sort: "string", template: function (obj) {
                                            if (obj.status === 'done') {
                                                return "<span style='color:green;'>交易成功</span>";
                                            } else if (obj.status === 'checked') {
                                                return "<span style='color:darkgreen;'>已对帐</span>";
                                            } else if (obj.status === 'padding') {
                                                return "<span style='color:red;'>未完成</span>";
                                            }
                                        }
                                    },
                                    { id: "remark", header: ["备注"], sort: "string",width:200,hidden:true},
                                    { id: "before_expire_time", header: ["交易前过期"], width: 150},
                                    { id: "after_expire_time", header: ["交易后过期"], width: 150},
                                    { id: "create_time", header: ["交易时间"], width: 150, sort: "string" },
                                    { id: "check_time", header: ["对账时间"], width: 150, sort: "string" },
                                    { id: "opt", header: '对帐操作', template: function(obj){
                                        if(obj.status==='done'){
                                            return "<span class='table-btn do_check'><i class='fa fa-eye'></i> 对帐</span> "
                                        }else{
                                            return ""
                                        }
                                    }, width: 100 },
                                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                                ],
                                on: {
                                    onItemDblClick: function (id, e, node) {
                                        var itemid = this.getSelectedItem().id;
                                        webix.require("admin/order.js?rand="+new Date().getTime(), function () {
                                            toughradius.admin.order.orderDetail(session, itemid, function () {
                                                $$(order_tabid).load("/admin/subscribe/orders?id=" + itemid);
                                            });
                                        })
                                    }
                                },
                                onClick:{
                                    do_check: function (e, id) {
                                        var itemid = this.getItem(id).id;
                                        webix.require("admin/order.js?rand="+new Date().getTime(), function () {
                                            toughradius.admin.order.checkOrderForm(session, itemid, function () {
                                                $$(order_tabid).load("/admin/subscribe/orders?id=" + itemid);
                                            });
                                        });
                                    }
                                },
                                select: true,
                                url: "/admin/subscribe/orders?id=" + itemid
                            }
                        },
                        {
                            header: "预订商品",
                            body: {
                                view: "datatable",
                                columns: [
                                    { id: "product_name", header: ["商品"], sort: "string" },
                                    { id: "opr_name", header: ["操作员"], sort: "string" },
                                    {
                                        id: "pay_status", header: ["支付状态"], sort: "string", template: function (obj) {
                                            if (obj.pay_status === 'done') {
                                                return "<span style='color:green;'>已支付</span>";
                                            } else if (obj.pay_status === 'padding') {
                                                return "<span style='color:red;'>未完成</span>";
                                            }
                                        }
                                    },
                                    {
                                        id: "status", header: ["使用状态"], sort: "string", template: function (obj) {
                                            if (obj.status === 'used') {
                                                return "<span style='color:green;'>已生效</span>";
                                            } else if (obj.status === 'padding') {
                                                return "<span style='color:red;'>未生效</span>";
                                            }else if (obj.status === 'cancel') {
                                                return "<span style='color:gray;'>已取消</span>";
                                            }
                                        }
                                    },
                                    { id: "order_num", header: ["订购数量"], sort: "int",width:70},
                                    { id: "remark", header: ["备注"], sort: "string",fillspace:true},
                                    { id: "create_time", header: ["订阅时间"], width: 150, sort: "string" },
                                ],
                                select: true,
                                url: "/admin/subscribe/pres?id=" + itemid
                            }
                        },
                        {
                            header: "工单记录",
                            body: {
                                id:issues_tabid,
                                view: "datatable",
                                leftSplit: 1,
                                rightSplit: 1,
                                columns: [
                                    { id: "id", header: ["ID"], sort: "int" },
                                    { id: "realname", header: ["姓名"], sort: "string",width:150, },
                                    { id: "subscriber", header: ["帐号"], sort: "string", width:130 },
                                    { id: "assign_opr_name", header: ["委派操作员"], sort: "string" },
                                    {
                                        id: "status", header: ["状态"], sort: "string", template: function (obj) {
                                            if (obj.status === 'init') {
                                                return "<span style='color:blue;'>首次提交</span>";
                                            } else if (obj.status === 'padding') {
                                                return "<span style='color:blue;'>处理中</span>";
                                            } else if (obj.status === 'done') {
                                                return "<span style='color:green;'>已完成</span>";
                                            } else if (obj.status === 'cancel') {
                                                return "<span style='color:darkgray;'>已取消</span>";
                                            } else if (obj.status === 'hang') {
                                                return "<span style='color:orange;'>挂起</span>";
                                            }
                                        }
                                    },
                                    {
                                        id: "type", header: ["类型"], sort: "string", width: 80, template: function (obj) {
                                            if (obj.type === 'install') {
                                                return "<span style='color:blue;'>报装</span>";
                                            } else if (obj.type === 'fault') {
                                                return "<span style='color:green;'>故障</span>";
                                            } else if (obj.type === 'complain') {
                                                return "<span style='color:orange;'>投诉</span>";
                                            } else if (obj.type === 'maintain') {
                                                return "<span style='color:orangered;'>维护</span>";
                                            }else if (obj.type === 'other') {
                                                return "<span style='color:black;'>其他</span>";
                                            }
                                        }
                                    },
                                    { id: "remindeds", header: ["催单次数"], sort: "int", width: 90 },
                                    { id: "create_time", header: ["创建时间"], sort: "string", width: 150 },
                                    { id: "opt", header: '操作', template: "<span class='table-btn do_issues_detail'><i class='fa fa-eye'></i> 详情</span> ", width: 100 },
                                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                                ],
                                select: true,
                                maxWidth: 2000,
                                maxHeight: 2000,
                                resizeColumn: true,
                                autoWidth: true,
                                autoHeight: true,
                                url: "/admin/issues/query?subscribe_id=" + subs.id,
                                on:{
                                    onItemDblClick: function(id, e, node){
                                        var itemid = this.getItem(id).id;
                                        webix.require("admin/issues.js?rand="+new Date().getTime(), function () {
                                            toughradius.admin.issues.issuesTabForm(itemid,function(){
                                               $$(issues_tabid).load("/admin/issues/query?subscribe_id=" + subs.id);
                                               $$(issues_tabid).refreash();
                                            });
                                        });
                                    }
                                },
                                onClick:{
                                    do_issues_detail: function (e, id) {
                                        var itemid = this.getItem(id).id;
                                        webix.require("admin/issues.js?rand="+new Date().getTime(), function () {
                                            toughradius.admin.issues.issuesTabForm(itemid,function(){
                                               $$(issues_tabid).load("/admin/issues/query?subscribe_id=" + subs.id);
                                               $$(issues_tabid).refreash();
                                            });
                                        });
                                    }
                                }
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
                                    { id: "username", header: ["用户名"], sort: "string" },
                                    { id: "acctSessionId", header: ["会话ID"], width: 150, sort: "string", hidden: true },
                                    { id: "nasId", header: ["BRAS 标识"], width: 120, sort: "string" },
                                    { id: "acctStartTime", header: ["上线时间"], width: 150, sort: "string" },
                                    { id: "nasAddr", header: ["BRAS IP"], width: 120, sort: "string" },
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
                                    { id: "opt", header: '操作', template: "<span class='table-btn do_clean'><i class='fa fa-unlock'></i> 清理</span> ", width: 100 },
                                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                                ],
                                select: true,
                                maxWidth: 2000,
                                maxHeight: 2000,
                                resizeColumn: true,
                                autoWidth: true,
                                autoHeight: true,
                                url: "/admin/online/query?keyword=" + subs.subscriber + "&areaId="+subs.customer.area_id,
                                onClick:{
                                    do_clean: function (e, id) {
                                        var sessionid = this.getItem(id).acctSessionId;
                                        webix.require("admin/online.js?rand="+new Date().getTime(), function () {
                                            toughradius.admin.online.onlineUnlock(sessionid,function(){
                                               $$(online_tabid).load("/admin/online/query?keyword=" + subs.subscriber+ "&areaId="+subs.customer.area_id);
                                               $$(online_tabid).refreash();
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
                                maxWidth: 2000,
                                maxHeight: 2000,
                                resizeColumn: true,
                                autoWidth: true,
                                autoHeight: true,
                                url: "/admin/syslog/query?start=0&count=200&type=radiusd&username="+ subs.subscriber
                            }
                        },
                        {
                            header: "上网日志",
                            body: {
                                view: "datatable",
                                leftSplit: 1,
                                rightSplit: 1,
                                columns: [
                                    { id: "username", header: ["用户名"], sort: "string" },
                                    { id: "acctSessionId", header: ["会话ID"], width: 150, sort: "string",  },
                                    { id: "nasId", header: ["BRAS 标识"], width: 120, sort: "string",  },
                                    { id: "nasAddr", header: ["BRAS IP"], width: 120, sort: "string" },
                                    { id: "framedIpaddr", header: ["用户 IP"], width: 120, sort: "string" },
                                    { id: "macAddr", header: ["用户 Mac"], width: 130, sort: "string" },
                                    { id: "nasPortId", header: ["端口信息"], width: 120, sort: "string", hidden: true },
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
                                    { id: "acctInputPackets", header: ["上行数据包"], width: 100, sort: "string", hidden: true },
                                    { id: "acctOutputPackets", header: ["下行数据包"], width: 100, sort: "string", hidden: true },
                                    { id: "acctStartTime", header: ["上线时间"], width: 150, sort: "string" },
                                    { id: "acctStopTime", header: ["下线时间"], width: 150, sort: "string" },
                                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                                ],
                                select: true,
                                maxWidth: 2000,
                                maxHeight: 2000,
                                resizeColumn: true,
                                autoWidth: true,
                                autoHeight: true,
                                url: "/admin/ticket/query?username=" + subs.subscriber+ "&area_id="+subs.customer.area_id
                            }
                        },
                        {
                            header: "流量统计",
                            body: {
                                view: "datatable",
                                columns: [
                                    { id: "id", header: ["id"],  hidden:true},
                                    { id: "username", header: ["用户名"], sort: "string" ,width:180},
                                    {
                                        id: "inputTotal", header: ["上传"], width: 150, sort: "int", template: function (obj) {
                                            return bytesToSize(obj.inputTotal);
                                        }
                                    },
                                    {
                                        id: "outputTotal", header: ["下载"], width: 150, sort: "int", template: function (obj) {
                                            return bytesToSize(obj.outputTotal);
                                        }
                                    },
                                    { id: "statdate", header: ["统计日期"], sort: "string" ,width:180},
                                    { id: "xxx", header: [""], fillspace:true},
                                ],
                                select: true,
                                maxWidth: 2000,
                                maxHeight: 2000,
                                resizeColumn: true,
                                autoWidth: true,
                                autoHeight: true,
                                url: "/admin/flowstat/query?username=" + subs.subscriber
                            }
                        },
                        {
                            header: "设备信息",
                            body: {
                                view: "datatable",
                                leftSplit: 1,
                                rightSplit: 1,
                                columns: [
                                    { id: "sn", header: ["序列号"], width:120, sort: "string" },
                                    { id: "model", header: ["型号"], width: 150, sort: "string"},
                                    { id: "type", header: ["分配形式"], width: 150, sort: "string", template:function(obj){
                                        if(obj.type === 'deposit'){
                                            return "租用";
                                        }else if(obj.type === 'purchase'){
                                            return "购买";
                                        }
                                    }},
                                    { id: "status", header: ["状态"], width: 150, sort: "string", template:function(obj){
                                        if(obj.status === 'using'){
                                            return "使用中";
                                        }else if(obj.status === 'return'){
                                            return "归还";
                                        }else if(obj.status === 'lost'){
                                            return "遗失";
                                        }else if(obj.status === 'repair'){
                                            return "报修";
                                        }else if(obj.status === 'expire'){
                                            return "过期";
                                        }
                                    }},
                                    { id: "fee", header: ["费用"], width: 120, sort: "string"},
                                    { id: "create_time", header: ["发放时间"], width: 150, sort: "string", fillspace:true,},
                                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                                ],
                                select: true,
                                maxWidth: 2000,
                                maxHeight: 2000,
                                resizeColumn: true,
                                autoWidth: true,
                                autoHeight: true,
                                url: "/admin/subscribe/devices?subs_id=" + subs.id
                            }
                        },
                        {
                            header: "FR 同步日志",
                            body: {
                                view: "datatable",
                                leftSplit: 1,
                                tooltip:{
                                  template:"#remark#"
                                },
                                columns: [
                                    { id: "username", header: ["用户名"], width:100, sort: "string" },
                                    { id: "password", header: ["密码"], width: 100, sort: "string"},
                                    { id: "status", header: ["状态"], width: 60, sort: "string", template:function(obj){
                                        if(obj.status === 'done'){
                                            return "<span style='color: green;'>完成</span>";
                                        }else if(obj.status === 'padding'){
                                            return "<span style='color: blue;'>等待</span>";
                                        }
                                    }},
                                    { id: "uprate", header: ["上行速率(Mbps)"], width: 80, sort: "string"},
                                    { id: "downrate", header: ["下行速率(Mbps)"], width: 80, sort: "string"},
                                    { id: "uprate_code", header: ["上行速率策略"], width: 100, sort: "string"},
                                    { id: "downrate_code", header: ["下行速率策略"], width: 100, sort: "string"},
                                    { id: "sync_type", header: ["同步类型"], width: 80, sort: "string", },
                                    { id: "sync_time", header: ["最后同步时间"], width: 150, sort: "string"},
                                    { id: "remark", header: ["同步信息"], sort: "string", fillspace:true,},
                                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                                ],
                                select: true,
                                maxWidth: 2000,
                                maxHeight: 2000,
                                resizeColumn: true,
                                autoWidth: true,
                                autoHeight: true,
                                url: "/admin/freeradius/synclog?username=" + subs.subscriber
                            }
                        }
                    ]
                }
            ]

        },$$(toughradius.admin.subscribe.detailFormID));
    })
};

toughradius.admin.subscribe.subscribeUpdate = function(session,item,callback){
    var updateWinid = "toughradius.admin.subscribe.subscribeUpdate";
    if($$(updateWinid))
        return;
    var formid = updateWinid+"_form";
    webix.ajax().get('/admin/subscribe/detail', {id:item.id}).then(function (result) {
        var subs = result.json();
        webix.ui({
            id:updateWinid,
            view: "window",
            css:"win-body",
            move:true,
            resize:true,
            width:800,
            height:600,
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
                    scroll: "auto",
                    minHeight:360,
                    maxWidth: 2000,
                    maxHeight: 2000,
                    elementsConfig: { labelWidth: 120 },
                    elements: [
                        {
                            view: "fieldset", label: "授权信息", paddingX: 20, body: {
                            paddingX: 20,
                            rows: [
                                { view: "label", css: "form-desc", label: "注意： 修改用户帐号的授权数据可能会造成用户续费，变更，销户时无法自动准确计算费用，请自行进行调整" },
                                {
                                    cols: [
                                        { view: "text", name: "subscriber", label: "订阅帐号", css: "nborder-input", readonly: true, value: subs.subscriber },
                                        { view: "text", name: "product_name", label: "当前订阅商品", css: "nborder-input", readonly: true, value: subs.product.name }
                                    ]
                                },
                                {
                                    cols: [
                                        { view: "text", name: "password", label: "认证密码", css: "nborder-input", readonly: true, value: subs.password },
                                        { view: "text", name: "expire_time", label: "过期时间", css: "nborder-input", readonly: true, value: subs.expire_time }
                                    ]
                                }
                            ]
                        }
                        },
                        {
                            view: "fieldset", label: "修改信息", body: {
                            paddingX: 20,
                            rows: [
                                {
                                    cols: [
                                        {
                                            view: "richselect", name: "new_product_id", label: "修改商品", icon: "caret-down",disabled:!hasPerms(session,['subscribe_product_modify']),
                                            value:subs.product_id,options: "/admin/product/options?node_id=" + subs.customer.node_id, validate: webix.rules.isNotEmpty,gravity:2
                                        },
                                        {
                                            view: "datepicker", name: "new_expire_time", timepicker: true, value:subs.expire_time, disabled:!hasPerms(session,['subscribe_expire_modify']),
                                            label: "修改过期时间", stringResult: true, format: session.system_config.SYSTEM_USER_EXPORT_FORMAT, validate: webix.rules.isNotEmpty
                                        }
                                    ]
                                },
                                {
                                    cols:[
                                        { view: "text", name: "addr_pool", label: "地址池",  value: subs.addr_pool },
                                        { view: "text", name: "mac_addr", label: "MAc地址",  value: subs.mac_addr },
                                        { view: "text", name: "ip_addr", label: "固定IP地址",  value: subs.ip_addr }
                                    ]
                                },
                                {
                                    cols:[
                                        { view: "text", name: "in_vlan", label: "内层VLAN",  value: subs.in_vlan },
                                        { view: "text", name: "out_vlan", label: "外层VLAN",  value: subs.out_vlan },
                                        { view: "radio", name: "bind_vlan", label: "绑定VLAN", value: subs.bind_vlan?'1':'0', options: [{ id: '1', value: "是" }, { id: '0', value: "否" }] },

                                    ]
                                },
                                {
                                    cols:[
                                        { view: "text", name: "up_rate", label: "上行速率(Mbps)",  value: subs.up_rate,disabled:!hasPerms(session,['subscribe_rate_modify'])},
                                        { view: "text", name: "down_rate", label: "下行速率(Mbps)",  value: subs.down_rate,disabled:!hasPerms(session,['subscribe_rate_modify'])},
                                        { view: "radio", name: "bind_mac", label: "绑定MAC", value: subs.bind_mac?'1':'0', options: [{ id: '1', value: "是" }, { id: '0', value: "否" }] },

                                    ]
                                },
                                {
                                    cols:[
                                        { view: "text", name: "up_peak_rate", label: "突发上行速率(Mbps)",  value: subs.up_peak_rate,disabled:!hasPerms(session,['subscribe_rate_modify'])},
                                        { view: "text", name: "down_peak_rate", label: "突发下行速率(Mbps)",  value: subs.down_peak_rate,disabled:!hasPerms(session,['subscribe_rate_modify'])},
                                        { view: "counter", name: "active_num", label: "最大在线", placeholder: "最大在线", value: subs.active_num, min: 1, max: 99999,disabled:!hasPerms(session,['subscribe_limit_modify'])}

                                    ]
                                },
                                {
                                    cols:[
                                        { view: "text", name: "up_rate_code", label: "上行速率策略",  value: subs.up_rate_code,disabled:!hasPerms(session,['subscribe_rate_modify'])},
                                        { view: "text", name: "down_rate_code", label: "下行速率策略",  value: subs.down_rate_code,disabled:!hasPerms(session,['subscribe_rate_modify'])},
                                        { view: "text", name: "domain", label: "认证域", value: subs.domain},
                                    ]
                                },
                                {
                                    cols:[
                                        { view: "radio", name: "free_auth", label: "到期免授权", disabled:!hasPerms(session,['subscribe_free_auth_modify']), value: subs.free_auth?'1':'0' , options: [{ id: '1', value: "是" }, { id: '0', value: "否" }] },
                                        { view: "text", name: "free_auth_uprate", label: "免授权上行速率",  value: subs.free_auth_uprate, disabled:!hasPerms(session,['subscribe_rate_modify'])},
                                        { view: "text", name: "free_auth_downrate", label: "免授权下行速率", value: subs.free_auth_downrate, disabled:!hasPerms(session,['subscribe_rate_modify'])}

                                    ]
                                },
                                {
                                    cols:[
                                        { view: "text", name: "proxy_user", label: "代理帐号",  value: subs.proxy_user },
                                        { view: "text", name: "proxy_pwd", label: "代理帐号密码",  value: subs.proxy_pwd },
                                        { view: "text", name: "proxy_vlan", label: "代理帐号VLAN",  value: subs.proxy_vlan }
                                    ]
                                },
                                {
                                    rows: [
                                        {
                                            cols:[
                                                { view: "radio", name: "proxy_enabled", label: "启用代理拨号", value: subs.proxy_enabled?'1':'0', options: [{ id: '1', value: "是" }, { id: '0', value: "否" }]},
                                                { view: "radio", name: "auto_renew", label: "自动续费", value: subs.auto_renew, options: [{ id: 'enabled', value: "是" }, { id: 'disabled', value: "否" }] },
                                                { view: "text", name: "policy", label: "自定义策略", value:subs.policy},
                                            ]
                                        },
                                        {
                                            cols:[
                                                { view: "textarea", name: "remark", label: "备注",value: subs.remark, height: 80 }
                                            ]
                                        }
                                    ]
                                }
                            ]
                        }
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
                                        toughradius.admin.subscribe.reloadData();
                                         $$(updateWinid).close();
                                    }
                                });
                            }
                        },
                        {view: "button", type: "base", width: 70, icon: "check-circle", label: "取消", click: function(){$$(winid).close()}}

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
        var subs = result.json();
        webix.ui({
            id:winid,
            view: "window",
            css:"win-body",
            move:true,
            width:680,
            height:500,
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
                    maxWidth: 2000,
                    maxHeight: 2000,
                    elementsConfig: { labelWidth: 120 },
                    elements: [
                        {
                            view: "fieldset", label: "授权信息", paddingX: 20, body: {
                            paddingX: 20,
                            rows: [
                                {
                                    cols: [
                                        { view: "text", name: "subscriber", label: "订阅帐号", css: "nborder-input", readonly: true, value: subs.subscriber },
                                        { view: "text", name: "product_name", label: "订阅商品", css: "nborder-input", readonly: true, value: subs.product.name }
                                    ]
                                },
                                {
                                    cols: [
                                        { view: "text", name: "password", label: "当前密码", css: "nborder-input", readonly: true, value: subs.password },
                                        { view: "text", name: "expire_time", label: "过期时间", css: "nborder-input", readonly: true, value: subs.expire_time }
                                    ]
                                }
                            ]
                        }
                        },
                        {
                            view: "fieldset", label: "修改密码", paddingX: 20, body: {
                            paddingX: 20,
                            cols: [
                                { view: "text", name: "new_password", type: "password", label: "新密码(*)", placeholder: "新密码", validate: webix.rules.isNotEmpty },
                                { view: "text", name: "new_cpassword", type: "password", label: "确认新密码(*)", placeholder: "确认新密码", validate: webix.rules.isNotEmpty }
                            ]
                        }
                        }
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
                                params.subs_id = item.id;
                                webix.ajax().post('/admin/subscribe/uppwd', params).then(function (result) {
                                    btn.enable();
                                    var resp = result.json();
                                    webix.message({ type: resp.msgtype, text: resp.msg, expire: 3000 });
                                    if (resp.code === 0) {
                                        toughradius.admin.subscribe.reloadData();
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



toughradius.admin.subscribe.issuesAdd = function (session, item, callback) {
    var winid = "toughradius.admin.subscribe.issuesAdd";
    if($$(winid))
        return;
    var formid = winid+"_form";
    webix.ajax().get('/admin/subscribe/detail', { id: item.id }).then(function (result) {
        var subs = result.json();
        webix.ui({
            id:winid,
            view: "window",
            css:"win-body",
            move:true,
            width:680,
            height:500,
            position: "center",
            head: {
                view: "toolbar",
                css:"win-toolbar",

                cols: [
                    {view: "icon", icon: "laptop", css: "alter"},
                    {view: "label", label: "创建工单"},
                    {view: "icon", icon: "times-circle", css: "alter", click: function(){
                        $$(winid).close();
                    }}
                ]
            },
            body: {
                borderless: true,
                padding: 5,
                rows: [
                    {
                        id: formid,
                        view: "form",
                        scroll: "auto",
                        maxWidth: 2000,
                        maxHeight: 2000,
                        elementsConfig: {labelWidth: 120},
                        elements: [
                            {
                                view: "fieldset", label: "授权信息", paddingX: 20, body: {
                                paddingX: 20,
                                rows: [
                                    {
                                        cols: [
                                            {
                                                view: "text",
                                                name: "subscriber",
                                                label: "订阅帐号",
                                                css: "nborder-input",
                                                readonly: true,
                                                value: subs.subscriber
                                            },
                                            {
                                                view: "text",
                                                name: "product_name",
                                                label: "订阅商品",
                                                css: "nborder-input",
                                                readonly: true,
                                                value: subs.product.name
                                            }
                                        ]
                                    },
                                    {
                                        cols: [
                                            {
                                                view: "text",
                                                name: "password",
                                                label: "当前密码",
                                                css: "nborder-input",
                                                readonly: true,
                                                value: subs.password
                                            },
                                            {
                                                view: "text",
                                                name: "expire_time",
                                                label: "过期时间",
                                                css: "nborder-input",
                                                readonly: true,
                                                value: subs.expire_time
                                            }
                                        ]
                                    }
                                ]
                            }
                            },
                            {
                                view: "fieldset", label: "工单信息", body: {
                                paddingX: 20,
                                rows: [
                                    {
                                        cols: [
                                            {
                                                view: "combo",
                                                name: "issues_opr",
                                                label: "委派操作员",
                                                icon: "caret-down",
                                                options: {
                                                    view:"suggest", url:"/admin/opr/options?node_id" + subs.customer.node_id
                                                },
                                                on: {
                                                    onChange: function (newv, oldv) {
                                                        console.log(this.getValue());
                                                        updateFeeValue();
                                                    }
                                                }
                                            },
                                            {
                                                view: "combo", name: "issues_type", label: "工单类型", value: 'install',
                                                options: [
                                                    {id: 'install', value: "新装"},
                                                    {id: 'maintain', value: "维护"},
                                                    {id: 'fault', value: "故障"},
                                                    {id: 'complain', value: "投诉"},
                                                    {id: 'other', value: "其他"},
                                                ]
                                            }
                                        ]
                                    },
                                    {
                                        view: "combo",
                                        name: "fault_id",
                                        label: "工单报障模板",
                                        icon: "caret-down",
                                        options: {
                                            view:"suggest", url:"/admin/issues/fault/options"
                                        }
                                    },
                                    {
                                        view: "textarea",
                                        name: "issues_remark",
                                        label: "工单内容",
                                        placeholder: "工单内容",
                                        height: 120,
                                    }
                                ]
                            }
                            }
                        ]
                    },
                    {
                        height: 36,
                        css: "panel-toolbar",
                        cols: [{},
                            {
                                view: "button",
                                type: "form",
                                width: 70,
                                icon: "check-circle",
                                label: "提交",
                                click: function () {
                                    if (!$$(formid).validate()) {
                                        webix.message({type: "error", text: "请正确填写资料", expire: 1000});
                                        return false;
                                    }
                                    var btn = this;
                                    btn.disable();
                                    var params = $$(formid).getValues();
                                    params.subs_id = item.id;
                                    webix.ajax().post('/admin/subscribe/issues/add', params).then(function (result) {
                                        btn.enable();
                                        var resp = result.json();
                                        webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
                                        if (resp.code === 0) {
                                            toughradius.admin.subscribe.reloadData();
                                             $$(winid).close();
                                        }
                                    });
                                }
                            },
                            {
                                view: "button",
                                type: "base",
                                icon: "times-circle",
                                width: 70,
                                css: "alter",
                                label: "取消",
                                click: function () {
                                    $$(winid).close();
                                }
                            },
                        ]
                    }
                ]
            }
        }).show()
    })
};

/**
 * 用户续费
 * @param session
 * @param item
 * @param callback
 */
toughradius.admin.subscribe.subscribeRenew = function(session,item,callback){
    var winid = "toughradius.admin.subscribe.subscribeRenew";
    if($$(winid))
        return;
    var formid = winid+"_form";
    var updateFeeValue = function(){
        var _params = {
            subs_id:item.id,
            order_num:$$(formid).elements['order_num'].getValue(),
            fee_ids:$$(formid).elements['fee_ids'].getValue(),
            product_id:$$(formid).elements['new_product_id'].getValue(),
            renew_type:$$(formid).elements['renew_type'].getValue()
        };
        webix.ajax().get('/admin/subscribe/renew/calc', _params).then(function (result) {
            var resp = result.json();
            var data = resp.data;
            console.log(data);
            $$(formid).elements['fees_value'].setValue(data.fees_value);
            $$(formid).elements['product_fee'].setValue(data.product_fee);
            $$(formid).elements['total_fee'].setValue(data.total_fee);
            $$(formid).elements['expire_time'].setValue(data.expire_time);
            $$(formid).refresh();
            if(resp.code>0){
                webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
            }
        })
    };
    webix.ajax().get('/admin/subscribe/detail', {id:item.id}).then(function (result) {
        var subs = result.json();
        webix.ui({
            id:winid,
            view: "window",
            css:"win-body",
            move:true,
            width:680,
            height:500,
            position: "center",
            head: {
                view: "toolbar",
                css:"win-toolbar",

                cols: [
                    {view: "icon", icon: "laptop", css: "alter"},
                    {view: "label", label: "帐号续费"},
                    {view: "icon", icon: "times-circle", css: "alter", click: function(){
                        $$(winid).close();
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
                    scroll: "auto",
                    maxWidth: 2000,
                    maxHeight: 2000,
                    elementsConfig: { labelWidth: 120 },
                    elements: [
                        {
                            view: "fieldset", label: "授权信息", paddingX: 20, body: {
                                paddingX: 5,
                                paddingy: 5,
                                rows:[
                                    {
                                        cols: [
                                            { view: "text", name: "subscriber", label: "订阅帐号", css: "nborder-input", readonly: true, value: subs.subscriber },
                                            { view: "text", name: "product_name", label: "当前订阅商品", css: "nborder-input", readonly: true, value: subs.product.name }
                                        ]
                                    },
                                    {
                                        cols: [
                                            { view: "text", name: "password", label: "认证密码", css: "nborder-input", readonly: true, value: subs.password },
                                            { view: "text", name: "expire_time", label: "过期时间", css: "nborder-input", readonly: true, value: subs.expire_time }
                                        ]
                                    },
                                    {
                                        cols: [
                                            {
                                                view: "richselect", name: "new_product_id", label: "续订资费为", icon: "caret-down",
                                                options: "/admin/product/options?node_id=" + subs.customer.node_id, validate: webix.rules.isNotEmpty, on: {
                                                    onChange: function (newv, oldv) {
                                                        updateFeeValue();
                                                    }
                                                }
                                            },
                                            {view: "counter", name: "order_num", label: "购买数量",  value:1, min:1, max:1000,on:{
                                                onChange:function(newv, oldv){
                                                    updateFeeValue()
                                                }
                                            }}

                                        ]
                                    },
                                    {
                                        cols:[
                                            {
                                                view: "datepicker", name: "expire_time", timepicker: true, readonly: !hasPerms(session, ['subscribe_expire_modify']),
                                                label: "续费后到期", stringResult: true, format: session.system_config.SYSTEM_USER_EXPORT_FORMAT, validate: webix.rules.isNotEmpty
                                            },
                                            {
                                                view: "radio", name: "renew_type", label: "(变更)生效方式", value: 'now',
                                                options: [{ id: 'now', value: "立即生效" }, { id: 'expire', value: "到期生效" }],on:{
                                                    onChange:function(newv, oldv){
                                                        updateFeeValue()
                                                    }
                                                }
                                            },
                                        ]
                                    },
                                ]
                        }},
                        {
                            view: "fieldset", label: "缴费信息", paddingX: 20, body: {
                                paddingX: 5,
                                paddingy: 5,
                                rows:[
                                    {
                                        cols: [
                                            {
                                                view: "multiselect", name: "fee_ids", label: "收费项", icon: "caret-down",
                                                options: "/admin/fees/options?node_id=" + subs.customer.node_id, on: {
                                                    onChange: function (newv, oldv) {
                                                        console.log(this.getValue());
                                                        updateFeeValue();
                                                    }
                                                }
                                            },
                                            { view: "text", name: "fees_value", label: "收费项合计", readonly: true, value: "0.00" }
                                        ]
                                    },
                                    {
                                        cols: [
                                            { view: "text", name: "product_fee", label: "资费金额(*)", placeholder: "资费金额", readonly: true, validate: webix.rules.isNotEmpty },
                                            {
                                                cols: [
                                                    { view: "text", name: "total_fee", label: "合计", readonly: !hasPerms(session, ['subscribe_fee_modify']), validate: webix.rules.isNotEmpty },
                                                    {
                                                        view: "button", type: "icon", width: 60, icon: "calculator", label: "计算", click: function () {
                                                            updateFeeValue();
                                                        }
                                                    }
                                                ]
                                            }
                                        ]
                                    },
                                    {
                                        rows: [
                                            {
                                                view: "radio", name: "pay_type", label: "缴费方式", value: 'cash',
                                                options: [{ id: 'cash', value: "现金" }, { id: 'bank', value: "银行卡" }, { id: 'alipay', value: "支付宝" }, { id: 'wxpay', value: "微信支付" }]
                                            },
                                            { view: "radio", name: "pay_status", label: "支付状态",hidden:true, value: 'done', options: [{ id: 'done', value: "已缴费" }, { id: 'padding', value: "未缴费" }] }
                                        ]
                                    },
                                    {
                                        rows: [
                                            { view: "textarea", name: "remark", label: "备注", height: 100 ,value:subs.remark}
                                        ]
                                    }
                                ]
                        }},

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
                                webix.ajax().post('/admin/subscribe/renew', params).then(function (result) {
                                    btn.enable();
                                    var resp = result.json();
                                    webix.message({ type: resp.msgtype, text: resp.msg, expire: 3000 });
                                    if (resp.code === 0) {
                                        toughradius.admin.subscribe.reloadData();
                                         $$(winid).close();
                                    }
                                });
                            }
                        },
                        {
                            view: "button", type: "base", icon: "times-circle", width: 70, css: "alter", label: "取消", click: function () {
                                $$(winid).close();
                            }
                        }
                    ]
                }
            ]}
        }).show();
    })
};


toughradius.admin.subscribe.subscribeChange = function(session,item,callback){
    var winid = "toughradius.admin.subscribe.subscribeChange";
    if($$(winid))
        return;
    var formid = winid+"_form";
    var updateFeeValue = function(){
        var _params = {
            subs_id:item.id,
            fee_ids:$$(formid).elements['fee_ids'].getValue(),
            product_id:$$(formid).elements['new_product_id'].getValue()
        };
        webix.ajax().get('/admin/subscribe/change/calc', _params).then(function (result) {
            var resp = result.json();
            var data = resp.data;
            console.log(data);
            $$(formid).elements['fees_value'].setValue(data.fees_value);
            $$(formid).elements['old_product_fee'].setValue(data.old_product_fee);
            $$(formid).elements['new_product_fee'].setValue(data.new_product_fee);
            $$(formid).elements['total_fee'].setValue(data.total_fee);
            $$(formid).elements['expire_time'].setValue(data.expire_time);
            $$(formid).refresh();
            if(resp.code>0){
                webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
            }
        })
    };
    webix.ajax().get('/admin/subscribe/detail', {id:item.id}).then(function (result) {
        var subs = result.json();
        webix.ui({
            id:winid,
            view: "window",
            css:"win-body",
            move:true,
            width:680,
            height:520,
            position: "center",
            head: {
                view: "toolbar",
                css:"win-toolbar",

                cols: [
                    {view: "icon", icon: "laptop", css: "alter"},
                    {view: "label", label: "帐号变更"},
                    {view: "icon", icon: "times-circle", css: "alter", click: function(){
                        $$(winid).close();
                    }}
                ]
            },
            body: {
            borderless: true,
            padding: 5,
            rows:[
                {
                    id: formid,
                    view: "form",
                    scroll: "auto",
                    maxWidth: 2000,
                    maxHeight: 2000,
                    elementsConfig: { labelWidth: 120 },
                    elements: [
                        {
                            view: "fieldset", label: "授权信息", paddingX: 20, body: {
                                paddingX: 5,
                                paddingy: 5,
                                rows:[
                                    {
                                        cols: [
                                            { view: "text", name: "subscriber", label: "订阅帐号", css: "nborder-input", readonly: true, value: subs.subscriber },
                                            { view: "text", name: "product_name", label: "当前订阅商品", css: "nborder-input", readonly: true, value: subs.product.name }
                                        ]
                                    },
                                    {
                                        cols: [
                                            { view: "text", name: "password", label: "认证密码", css: "nborder-input", readonly: true, value: subs.password },
                                            { view: "text", name: "expire_time", label: "过期时间", css: "nborder-input", readonly: true, value: subs.expire_time }
                                        ]
                                    },
                                    {
                                        cols: [
                                            {
                                                view: "richselect", name: "new_product_id", label: "变更资费为", icon: "caret-down",
                                                options: "/admin/product/options?node_id=" + subs.customer.node_id + "&exclude_id=" + subs.product_id, validate: webix.rules.isNotEmpty, on: {
                                                    onChange: function (newv, oldv) {
                                                        updateFeeValue();
                                                    }
                                                }
                                            },
                                            {
                                                view: "datepicker", name: "expire_time", timepicker: true, readonly: !hasPerms(session, ['subscribe_expire_modify']),
                                                label: "变更后到期", stringResult: true, format: session.system_config.SYSTEM_USER_EXPORT_FORMAT, validate: webix.rules.isNotEmpty
                                            }
                                        ]
                                    },
                                ]
                        }},
                        {
                            view: "fieldset", label: "缴费信息", paddingX: 20, body: {
                                paddingX: 5,
                                paddingy: 5,
                                rows:[
                                    {
                                        cols: [
                                            {
                                                view: "multiselect", name: "fee_ids", label: "收费项", icon: "caret-down",
                                                options: "/admin/fees/options?node_id=" + subs.customer.node_id, on: {
                                                    onChange: function (newv, oldv) {
                                                        console.log(this.getValue());
                                                        updateFeeValue();
                                                    }
                                                }
                                            },
                                            { view: "text", name: "fees_value", label: "收费项合计", readonly: true, value: "0.00" }
                                        ]
                                    },
                                    {
                                        rows: [
                                            { view: "label", css: "form-desc", label: "变更前资费剩余费用 = 老资费单价 × 剩余天数;变更新资费总费用 = 新资费单价 × 剩余天数" },
                                        ]
                                    },
                                    {
                                        cols: [
                                            { view: "text", name: "old_product_fee", label: "变更前资费退款(*)", placeholder: "变更前资费退款金额", readonly: true, validate: webix.rules.isNotEmpty },
                                            { view: "text", name: "new_product_fee", label: "变更新资费费用(*)", placeholder: "变更新资费费用", readonly: true, validate: webix.rules.isNotEmpty }
                                        ]
                                    },
                                    {
                                        cols: [
                                            {
                                                cols: [
                                                    { view: "text", name: "total_fee", label: "合计", readonly: !hasPerms(session, ['subscribe_fee_modify']), validate: webix.rules.isNotEmpty },
                                                    {
                                                        view: "button", type: "icon", icon: "calculator", label: "计算", click: function () {
                                                            updateFeeValue();
                                                        }
                                                    }
                                                ]
                                            }
                                        ]
                                    },
                                    {
                                        rows: [
                                            { view: "label", css: "form-desc", label: "缴费总金额 = 收费项合计 + 变更资费总费用 - 变更前资费剩余费用" },
                                        ]
                                    },
                                    {
                                        rows: [
                                            {
                                                view: "radio", name: "pay_type", label: "缴费方式", value: 'cash',
                                                options: [{ id: 'cash', value: "现金" }, { id: 'bank', value: "银行卡" }, { id: 'alipay', value: "支付宝" }, { id: 'wxpay', value: "微信支付" }]
                                            },
                                            { view: "radio", name: "pay_status", label: "支付状态",hidden:true, value: 'done', options: [{ id: 'done', value: "已缴费" }, { id: 'padding', value: "未缴费" }] }
                                        ]
                                    },
                                    {
                                        rows: [
                                            { view: "textarea", name: "remark", label: "备注", height: 80 , value:subs.remark}
                                        ]
                                    }
                                ]
                        }},

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
                                webix.ajax().post('/admin/subscribe/change', params).then(function (result) {
                                    btn.enable();
                                    var resp = result.json();
                                    webix.message({ type: resp.msgtype, text: resp.msg, expire: 3000 });
                                    if (resp.code === 0) {
                                        toughradius.admin.subscribe.reloadData();
                                         $$(winid).close();
                                    }
                                });
                            }
                        },
                        {
                            view: "button", type: "base", icon: "times-circle", width: 70, css: "alter", label: "取消", click: function () {
                                $$(winid).close();
                            }
                        }
                    ]
                }
            ]}
        }).show();
    })
};


toughradius.admin.subscribe.subscribeResume = function(session,item,callback){
    var winid = "toughradius.admin.subscribe.subscribeResume";
    if($$(winid))
        return;
    var formid = winid+"_form";
    var updateFeeValue = function(){
        var _params = {
            subs_id:item.id,
            fee_ids:$$(formid).elements['fee_ids'].getValue(),
            product_id:''
        };
        webix.ajax().get('/admin/subscribe/resume/calc', _params).then(function (result) {
            var resp = result.json();
            var data = resp.data;
            console.log(data);
            $$(formid).elements['fees_value'].setValue(data.fees_value);
            $$(formid).elements['total_fee'].setValue(data.total_fee);
            $$(formid).elements['expire_time'].setValue(data.expire_time);
            $$(formid).refresh();
            if(resp.code>0){
                webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
            }
        })

    };
    webix.ajax().get('/admin/subscribe/detail', {id:item.id}).then(function (result) {
        var subs = result.json();
        webix.ui({
            id:winid,
            view: "window",
            css:"win-body",
            move:true,
            width:680,
            height:500,
            position: "center",
            head: {
                view: "toolbar",
                css:"win-toolbar",

                cols: [
                    {view: "icon", icon: "laptop", css: "alter"},
                    {view: "label", label: "帐号复机"},
                    {view: "icon", icon: "times-circle", css: "alter", click: function(){
                        $$(winid).close();
                    }}
                ]
            },
            body: {
            borderless: true,
            padding: 5,
            rows: [
                {
                    id: formid,
                    view: "form",
                    scroll: "auto",
                    maxWidth: 2000,
                    maxHeight: 2000,
                    elementsConfig: { labelWidth: 120 },
                    elements: [
                        {
                            view: "fieldset", label: "授权信息", paddingX: 20, body: {
                                paddingX: 5,
                                paddingy: 5,
                                rows:[
                                    {
                                        paddingX: 20,
                                        cols: [
                                            { view: "text", name: "subscriber", label: "订阅帐号", css: "nborder-input", readonly: true, value: subs.subscriber },
                                            { view: "text", name: "product_name", label: "当前订阅商品", css: "nborder-input", readonly: true, value: subs.product.name }
                                        ]
                                    },
                                    {
                                        paddingX: 20,
                                        cols: [
                                            { view: "text", name: "password", label: "认证密码", css: "nborder-input", readonly: true, value: subs.password },
                                            { view: "text", name: "expire_time", label: "过期时间", css: "nborder-input", readonly: true, value: subs.expire_time }
                                        ]
                                    },
                                    {
                                        paddingX: 20,
                                        cols: [
                                            {
                                                view: "datepicker", name: "expire_time", timepicker: true, readonly: !hasPerms(session, ['subscribe_expire_modify']),
                                                label: "复机后到期", stringResult: true, format: session.system_config.SYSTEM_USER_EXPORT_FORMAT, validate: webix.rules.isNotEmpty
                                            },
                                            {
                                                view: "button", type: "icon", icon: "calculator", label: "计算到期", click: function () {
                                                    updateFeeValue();
                                                }
                                            }
                                        ]
                                    },
                                ]
                        }},
                        {
                            view: "fieldset", label: "缴费信息", paddingX: 20, body: {
                                paddingX: 5,
                                paddingy: 5,
                                rows:[
                                    {
                                        paddingX: 20,
                                        cols: [
                                            {
                                                view: "multiselect", name: "fee_ids", label: "收费项", icon: "caret-down",
                                                options: "/admin/fees/options?node_id=" + subs.customer.node_id, on: {
                                                    onChange: function (newv, oldv) {
                                                        console.log(this.getValue());
                                                        updateFeeValue();
                                                    }
                                                }
                                            },
                                            { view: "text", name: "fees_value", label: "收费项合计", readonly: true, value: "0.00" }
                                        ]
                                    },
                                    {
                                        paddingX: 20,
                                        cols: [
                                            {
                                                cols: [
                                                    { view: "text", name: "total_fee", label: "合计", readonly: !hasPerms(session, ['subscribe_fee_modify']), validate: webix.rules.isNotEmpty },
                                                    {
                                                        view: "button", type: "icon", icon: "calculator", label: "计算", click: function () {
                                                            updateFeeValue();
                                                        }
                                                    }
                                                ]
                                            }
                                        ]
                                    },
                                    {
                                        paddingX: 20,
                                        rows: [
                                            {
                                                view: "radio", name: "pay_type", label: "缴费方式", value: 'cash',
                                                options: [{ id: 'cash', value: "现金" }, { id: 'bank', value: "银行卡" }, { id: 'alipay', value: "支付宝" }, { id: 'wxpay', value: "微信支付" }]
                                            },
                                            { view: "radio", name: "pay_status", label: "支付状态",hidden:true, value: 'done', options: [{ id: 'done', value: "已缴费" }, { id: 'padding', value: "未缴费" }] }
                                        ]
                                    },
                                    {
                                        paddingX: 20,
                                        rows: [
                                            { view: "textarea", name: "remark", label: "备注", height: 100 , value:subs.remark}
                                        ]
                                    }
                                ]
                        }}

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
                                webix.ajax().post('/admin/subscribe/resume', params).then(function (result) {
                                    btn.enable();
                                    var resp = result.json();
                                    webix.message({ type: resp.msgtype, text: resp.msg, expire: 3000 });
                                    if (resp.code === 0) {
                                        toughradius.admin.subscribe.reloadData();
                                        $$(winid).close();
                                    }
                                });
                            }
                        },
                        {
                            view: "button", type: "base", icon: "times-circle", width: 70, css: "alter", label: "取消", click: function () {
                                $$(winid).close();
                            }
                        }
                    ]
                }
            ]}
        }).show();
    })
};


toughradius.admin.subscribe.subscribeMove = function(session,item,callback){
    var winid = "toughradius.admin.subscribe.subscribeMove";
    if($$(winid))
        return;
    var formid = winid+"_form";
    var updateFeeValue = function(){
        var _params = {
            subs_id:item.id,
            fee_ids:$$(formid).elements['fee_ids'].getValue(),
            product_id:''
        };
        webix.ajax().get('/admin/subscribe/move/calc', _params).then(function (result) {
            var resp = result.json();
            var data = resp.data;
            console.log(data);
            $$(formid).elements['fees_value'].setValue(data.fees_value);
            $$(formid).elements['total_fee'].setValue(data.total_fee);
            $$(formid).refresh();
            if(resp.code>0){
                webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
            }
        })

    };
    webix.ajax().get('/admin/subscribe/detail', {id:item.id}).then(function (result) {
        var subs = result.json();
        webix.ui({
            id:winid,
            view: "window",
            css:"win-body",
            move:true,
            width:680,
            height:500,
            position: "center",
            head: {
                view: "toolbar",
                css:"win-toolbar",

                cols: [
                    {view: "icon", icon: "laptop", css: "alter"},
                    {view: "label", label: "帐号迁移"},
                    {view: "icon", icon: "times-circle", css: "alter", click: function(){
                        $$(winid).close();
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
                    scroll: "auto",
                    maxWidth: 2000,
                    maxHeight: 2000,
                    elementsConfig: { labelWidth: 90 },
                    elements: [
                        { view: "fieldset", label: "授权信息", paddingX: 20, body: {
                            paddingX: 5,
                            paddingy: 5,
                            rows:[
                                {
                                    paddingX: 20,
                                    cols: [
                                        { view: "text", name: "subscriber", label: "订阅帐号", css: "nborder-input", readonly: true, value: subs.subscriber },
                                        { view: "text", name: "password", label: "认证密码", css: "nborder-input", readonly: true, value: subs.password },
                                        { view: "text", name: "expire_time", label: "过期时间", css: "nborder-input", readonly: true, value: subs.expire_time }

                                    ]
                                },
                                {
                                    paddingX:20,
                                    cols:[
                                        { view: "text", name: "product_name", label: "当前订阅商品", css: "nborder-input", readonly: true, value: subs.product.name },
                                        {view: "text", name: "address",  gravity:2, label: "客户地址", css: "nborder-input2", value: subs.customer.address,  validate:webix.rules.isNotEmpty},
                                    ]
                                },
                                {
                                    paddingX:20,
                                    cols:[
                                        {view: "richselect", name:"area_id", label: "区域(*)", value:subs.customer.area_id, icon: "caret-down",on:{
                                            onChange:function(newv, oldv){
                                                var list = $$(formid).elements['zone_id'].getPopup().getList();
                                                list.clearAll();
                                                list.load("/admin/zone/options?area_id=" + newv);
                                            }
                                        },options:"/admin/area/options?node_id="+subs.customer.node_id,validate:webix.rules.isNotEmpty},
                                        {view: "richselect", name:"zone_id", label: "小区", value:subs.customer.zone_id, icon: "caret-down",options:"/admin/zone/options?area_id="+subs.customer.area_id},
                                    ]
                                }
                            ]
                        }},

                        { view: "fieldset", label: "缴费信息", paddingX: 20, body: {
                            paddingX: 5,
                            paddingy: 5,
                            rows:[
                                {
                                    paddingX: 20,
                                    cols: [
                                        {
                                            view: "multiselect", name: "fee_ids", label: "收费项", icon: "caret-down",
                                            options: "/admin/fees/options?node_id=" + subs.customer.node_id, on: {
                                                onChange: function (newv, oldv) {
                                                    console.log(this.getValue());
                                                    updateFeeValue();
                                                }
                                            }
                                        },
                                        { view: "text", name: "fees_value", label: "收费项合计", readonly: true, value: "0.00" }
                                    ]
                                },
                                {
                                    paddingX: 20,
                                    cols: [
                                        {
                                            cols: [
                                                { view: "text", name: "total_fee", label: "合计", readonly: !hasPerms(session, ['subscribe_fee_modify']), validate: webix.rules.isNotEmpty },
                                                {
                                                    view: "button", type: "icon", icon: "calculator", label: "计算", click: function () {
                                                        updateFeeValue();
                                                    }
                                                }
                                            ]
                                        }
                                    ]
                                },
                                {
                                    paddingX: 20,
                                    cols: [
                                        {
                                            view: "radio", name: "pay_type", label: "缴费方式", value: 'cash',
                                            options: [{ id: 'cash', value: "现金" }, { id: 'bank', value: "银行卡" }, { id: 'alipay', value: "支付宝" }, { id: 'wxpay', value: "微信支付" }]
                                        },
                                        { view: "radio", name: "pay_status", label: "支付状态",hidden:true, value: 'done', options: [{ id: 'done', value: "已缴费" }, { id: 'padding', value: "未缴费" }] }
                                    ]
                                }
                            ]
                        }},
                        { view: "fieldset", label: "工单信息", paddingX: 20, body: {
                            paddingX: 5,
                            paddingy: 5,
                            rows:[
                                {
                                    paddingX:20,
                                    cols:[
                                        {view: "richselect", name:"issues_opr", label: "委派操作员", icon: "caret-down",
                                            options:"/admin/opr/options?node_id"+subs.customer.node_id, on:{
                                            onChange:function(newv, oldv){
                                                console.log(this.getValue());
                                                updateFeeValue();
                                            }

                                        }},
                                        {view:"radio", name:"issues_type", label: "工单类型", value:'maintain',hidden:true, options:[{id:'maintain',value:"维护"}]},
                                        {view:"radio", name:"issues_status", label: "状态", value:'padding', options:[{id:'padding',value:"未完成"},{id:'done',value:"已完成"}]}
                                    ]
                                },
                                {
                                    paddingX:20,
                                    cols:[
                                        {view: "textarea", name: "issues_remark", label: "备注", placeholder: "备注", height:80}
                                    ]
                                }
                            ]
                        }}
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
                                webix.ajax().post('/admin/subscribe/move', params).then(function (result) {
                                    btn.enable();
                                    var resp = result.json();
                                    webix.message({ type: resp.msgtype, text: resp.msg, expire: 3000 });
                                    if (resp.code === 0) {
                                        toughradius.admin.subscribe.reloadData();
                                        $$(winid).close()

                                    }
                                });
                            }
                        },
                        {
                            view: "button", type: "base", icon: "times-circle", width: 70, css: "alter", label: "取消", click: function () {
                                $$(winid).close()
                            }
                        }
                    ]
                }
            ]}
        }).show();
    })
};


/**
 * 用户销户
 * @param item
 * @param callback
 */

toughradius.admin.subscribe.subscribeCancel = function(session,item,callback){
    var winid = "toughradius.admin.subscribe.subscribeCancel";
    if($$(winid))
        return;
    var formid = winid+"_form";
    var updateFeeValue = function(){
        var _params = {
            subs_id:item.id,
            fee_ids:$$(formid).elements['fee_ids'].getValue(),
            product_id:item.product_id
        };
        webix.ajax().get('/admin/subscribe/cancel/calc', _params).then(function (result) {
            var resp = result.json();
            var data = resp.data;
            console.log(data);
            $$(formid).elements['fees_value'].setValue(data.fees_value);
            $$(formid).elements['product_fee'].setValue(data.product_fee);
            $$(formid).elements['total_fee'].setValue(data.total_fee);
            $$(formid).refresh();
            if(resp.code>0){
                webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
            }
        })
    };
    webix.ajax().get('/admin/subscribe/detail', {id:item.id}).then(function (result) {
        var subs = result.json();
        webix.ui({
            id:winid,
            view: "window",
            css:"win-body",
            move:true,
            width:680,
            height:500,
            position: "center",
            head: {
                view: "toolbar",
                css:"win-toolbar",

                cols: [
                    {view: "icon", icon: "laptop", css: "alter"},
                    {view: "label", label: "帐号销户"},
                    {view: "icon", icon: "times-circle", css: "alter", click: function(){
                        $$(winid).close();
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
                    scroll: "auto",
                    maxWidth: 2000,
                    maxHeight: 2000,
                    elementsConfig: { labelWidth: 120 },
                    elements: [
                        {
                            view: "fieldset", label: "授权信息", paddingX: 20, body: {
                                paddingX: 5,
                                paddingy: 5,
                                rows:[
                                    {
                                        paddingX: 20,
                                        cols: [
                                            { view: "text", name: "subscriber", label: "订阅帐号", css: "nborder-input", readonly: true, value: subs.subscriber },
                                            { view: "text", name: "product_name", label: "当前订阅商品", css: "nborder-input", readonly: true, value: subs.product.name }
                                        ]
                                    },
                                    {
                                        paddingX: 20,
                                        cols: [
                                            { view: "text", name: "password", label: "认证密码", css: "nborder-input", readonly: true, value: subs.password },
                                            { view: "text", name: "expire_time", label: "过期时间", css: "nborder-input", readonly: true, value: subs.expire_time }
                                        ]
                                    },
                                ]
                        }},
                        {
                            view: "fieldset", label: "缴费信息", paddingX: 20, body: {
                                paddingX: 5,
                                paddingy: 5,
                                rows:[
                                    {
                                        paddingX: 20,
                                        cols: [
                                            {
                                                view: "multiselect", name: "fee_ids", label: "收费项", icon: "caret-down",
                                                options: "/admin/fees/options?node_id=" + subs.customer.node_id, on: {
                                                    onChange: function (newv, oldv) {
                                                        console.log(this.getValue());
                                                        updateFeeValue();
                                                    }
                                                }
                                            },
                                            { view: "text", name: "fees_value", label: "收费项合计", readonly: true, value: "0.00" }
                                        ]
                                    },
                                    {
                                        paddingX: 20,
                                        cols: [
                                            {
                                                view: "multiselect", name: "subdev_ids", label: "租用终端退还", icon: "caret-down",
                                                options: "/admin/subscribe/device/options?subs_id=" + subs.id, on: {
                                                    onChange: function (newv, oldv) {
                                                        console.log(this.getValue());
                                                        updateFeeValue();
                                                    }
                                                }
                                            },
                                            { view: "text", name: "device_value", label: "租用终端退还费用", readonly: true, value: "0.00" }
                                        ]
                                    },
                                    {
                                        paddingX: 20,
                                        cols: [
                                            { view: "text", name: "product_fee", label: "资费退款金额(*)", placeholder: "资费退款金额", readonly: true, validate: webix.rules.isNotEmpty },
                                            {
                                                cols: [
                                                    { view: "text", name: "total_fee", label: "合计", readonly: !hasPerms(session, ['subscribe_fee_modify']), validate: webix.rules.isNotEmpty },
                                                    {
                                                        view: "button", type: "icon", width: 60, icon: "calculator", label: "计算", click: function () {
                                                            updateFeeValue();
                                                        }
                                                    }
                                                ]
                                            }
                                        ]
                                    },
                                    {
                                        paddingX: 20,
                                        rows: [
                                            {
                                                view: "radio", name: "pay_type", label: "缴费方式", value: 'cash',
                                                options: [{ id: 'cash', value: "现金" }, { id: 'bank', value: "银行卡" }, { id: 'alipay', value: "支付宝" }, { id: 'wxpay', value: "微信支付" }]
                                            },
                                            { view: "radio", name: "pay_status", label: "支付状态",hidden:true, value: 'done', options: [{ id: 'done', value: "已缴费" }, { id: 'padding', value: "未缴费" }] }
                                        ]
                                    },
                                    {
                                        paddingX: 20,
                                        rows: [
                                            { view: "textarea", name: "remark", label: "销户备注", height: 80 }
                                        ]
                                    }
                                ]
                        }}
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
                                params.subs_id = subs.id
                                webix.ajax().post('/admin/subscribe/cancel', params).then(function (result) {
                                    btn.enable();
                                    var resp = result.json();
                                    webix.message({ type: resp.msgtype, text: resp.msg, expire: 3000 });
                                    if (resp.code === 0) {
                                        if(callback){
                                            callback();
                                        }
                                        $$(winid).close();
                                    }
                                });
                            }
                        },
                        {
                            view: "button", type: "base", icon: "times-circle", width: 70, css: "alter", label: "取消", click: function () {
                                $$(winid).close();
                            }
                        }
                    ]
                }
            ]}
        }).show();
    })
};



toughradius.admin.subscribe.subscribePay = function(session,item,callback){
    var winid = "toughradius.admin.subscribe.subscribePay";
    if($$(winid))
        return;
    var formid = winid+"_form";
    webix.ajax().get('/admin/subscribe/detail', {id:item.id}).then(function (result) {
        var subs = result.json();
        webix.ui({
            id:winid,
            view: "window",
            css:"win-body",
            move:true,
            width:680,
            height:500,
            position: "center",
            head: {
                view: "toolbar",
                css:"win-toolbar",

                cols: [
                    {view: "icon", icon: "laptop", css: "alter"},
                    {view: "label", label: "帐号缴费"},
                    {view: "icon", icon: "times-circle", css: "alter", click: function(){
                        $$(winid).close();
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
                    scroll: "auto",
                    maxWidth: 2000,
                    maxHeight: 2000,
                    elementsConfig: { labelWidth: 120 },
                    elements: [
                        {
                            view: "fieldset", label: "授权信息", paddingX: 20, body: {
                                paddingX: 5,
                                paddingy: 5,
                                rows:[
                                    {
                                        cols: [
                                            { view: "text", name: "subscriber", label: "订阅帐号", css: "nborder-input", readonly: true, value: subs.subscriber },
                                            { view: "text", name: "product_name", label: "当前订阅商品", css: "nborder-input", readonly: true, value: subs.product.name }
                                        ]
                                    },
                                    {
                                        cols: [
                                            { view: "text", name: "password", label: "认证密码", css: "nborder-input", readonly: true, value: subs.password },
                                            { view: "text", name: "expire_time", label: "过期时间", css: "nborder-input", readonly: true, value: subs.expire_time }
                                        ]
                                    },
                                ]
                        }},
                        {
                            view: "fieldset", label: "缴费信息", paddingX: 20, body: {
                                paddingX: 5,
                                paddingy: 5,
                                rows:[
                                    { view: "text", name: "total_fee", label: "缴费金额",  validate: webix.rules.isNotEmpty },
                                    {
                                        view: "radio", name: "pay_type", label: "缴费方式", value: 'cash',
                                        options: [{ id: 'cash', value: "现金" }, { id: 'bank', value: "银行卡" }, { id: 'alipay', value: "支付宝" }, { id: 'wxpay', value: "微信支付" }]
                                    },
                                    { view: "textarea", name: "remark", label: "备注", height: 80 }
                                ]
                        }}
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
                                params.subs_id = subs.id
                                webix.ajax().post('/admin/subscribe/pay', params).then(function (result) {
                                    btn.enable();
                                    var resp = result.json();
                                    webix.message({ type: resp.msgtype, text: resp.msg, expire: 3000 });
                                    if (resp.code === 0) {
                                        $$(winid).close();
                                    }
                                });
                            }
                        },
                        {
                            view: "button", type: "base", icon: "times-circle", width: 70, css: "alter", label: "取消", click: function () {
                                $$(winid).close();
                            }
                        }
                    ]
                }
            ]}
        }).show();
    })
};


toughradius.admin.subscribe.subscribeRefund = function(session,item,callback){
    var winid = "toughradius.admin.subscribe.subscribeRefund";
    if($$(winid))
        return;
    var formid = winid+"_form";
    webix.ajax().get('/admin/subscribe/detail', {id:item.id}).then(function (result) {
        var subs = result.json();
        webix.ui({
            id:winid,
            view: "window",
            css:"win-body",
            move:true,
            width:680,
            height:500,
            position: "center",
            head: {
                view: "toolbar",
                css:"win-toolbar",

                cols: [
                    {view: "icon", icon: "laptop", css: "alter"},
                    {view: "label", label: "帐号退费"},
                    {view: "icon", icon: "times-circle", css: "alter", click: function(){
                        $$(winid).close();
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
                    scroll: "auto",
                    maxWidth: 2000,
                    maxHeight: 2000,
                    elementsConfig: { labelWidth: 120 },
                    elements: [
                        {
                            view: "fieldset", label: "授权信息", paddingX: 20, body: {
                                paddingX: 5,
                                paddingy: 5,
                                rows:[
                                    {
                                        cols: [
                                            { view: "text", name: "subscriber", label: "订阅帐号", css: "nborder-input", readonly: true, value: subs.subscriber },
                                            { view: "text", name: "product_name", label: "当前订阅商品", css: "nborder-input", readonly: true, value: subs.product.name }
                                        ]
                                    },
                                    {
                                        cols: [
                                            { view: "text", name: "password", label: "认证密码", css: "nborder-input", readonly: true, value: subs.password },
                                            { view: "text", name: "expire_time", label: "过期时间", css: "nborder-input", readonly: true, value: subs.expire_time }
                                        ]
                                    },
                                ]
                        }},
                        {
                            view: "fieldset", label: "退费信息", paddingX: 20, body: {
                                paddingX: 5,
                                paddingy: 5,
                                rows:[
                                    { view: "text", name: "total_fee", label: "退费金额",  validate: webix.rules.isNotEmpty },
                                    {
                                        view: "radio", name: "pay_type", label: "退费方式", value: 'cash',
                                        options: [{ id: 'cash', value: "现金" }, { id: 'bank', value: "银行卡" }, { id: 'alipay', value: "支付宝" }, { id: 'wxpay', value: "微信支付" }]
                                    },
                                    { view: "textarea", name: "remark", label: "备注", height: 80 }
                                ]
                        }}
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
                                params.subs_id = subs.id
                                webix.ajax().post('/admin/subscribe/refund', params).then(function (result) {
                                    btn.enable();
                                    var resp = result.json();
                                    webix.message({ type: resp.msgtype, text: resp.msg, expire: 3000 });
                                    if (resp.code === 0) {
                                        $$(winid).close();
                                    }
                                });
                            }
                        },
                        {
                            view: "button", type: "base", icon: "times-circle", width: 70, css: "alter", label: "取消", click: function () {
                                $$(winid).close();
                            }
                        }
                    ]
                }
            ]}
        }).show();
    })
};



toughradius.admin.subscribe.subscribePause = function (ids,callback) {
    webix.confirm({
        title: "操作确认",
        ok: "是", cancel: "否",
        text: "停机后用户认证功能暂时关闭，确认要停机吗",
        width:360,
        callback: function (ev) {
            if (ev) {
                webix.ajax().get('/admin/subscribe/pause', {ids:ids}).then(function (result) {
                    var resp = result.json();
                    webix.message({type: resp.msgtype, text: resp.msg, expire: 1500});
                    if(callback)
                        callback()
                })
            }
        }
    });
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


toughradius.admin.subscribe.subscribeRelease = function (ids,rtype,callback) {
    console.log(rtype);
    if(['subscribe_release_mac','subscribe_release_invlan','subscribe_release_outvlan'].indexOf(rtype)==-1){
        return;
    }
    webix.confirm({
        title: "操作确认",
        ok: "是", cancel: "否",
        text: "确认要释放绑定吗？",
        width:270,
        callback: function (ev) {
            if (ev) {
                webix.ajax().get('/admin/subscribe/release', {ids:ids,rtype:rtype}).then(function (result) {
                    var resp = result.json();
                    webix.message({type: resp.msgtype, text: resp.msg, expire: 1500});
                    if(callback)
                        callback()
                });
            }
        }
    });
};


toughradius.admin.subscribe.subscribeRadiusTest = function(session,itemid){
    var winid = "toughradius.admin.subscribe.subscribeRadiusTest";
    var logvid = webix.uid();
    if($$(winid))
        return;
    var formid = winid+"_form";
    var updateLog = function(iresult){
        var rst = iresult.json();
        console.log(rst);
        $$(logvid).define("template",rst.msg.replace("\n","<br>"))
        $$(logvid).refresh();
    }
    webix.ajax().get('/admin/subscribe/detail', {id:itemid}).then(function (result) {
        var subs = result.json();
        webix.ui({
            id:winid,
            view: "window",
            css:"win-body",
            move:true,
            width:680,
            height:500,
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
                    maxWidth: 2000,
                    maxHeight: 2000,
                    elementsConfig: { labelWidth: 120 },
                    elements: [
                        {
                            view: "fieldset", label: "测试帐号", paddingX: 20, body: {
                                paddingX: 20,
                                rows: [
                                    {
                                        cols: [
                                            { view: "text", name: "subscriber", label: "订阅帐号", css: "nborder-input", readonly: true, value: subs.subscriber },
                                            { view: "text", name: "product_name", label: "订阅商品", css: "nborder-input", readonly: true, value: subs.product.name }
                                        ]
                                    },
                                    {
                                        cols: [
                                            { view: "text", name: "password", label: "当前密码", css: "nborder-input", readonly: true, value: subs.password },
                                            { view: "text", name: "expire_time", label: "过期时间", css: "nborder-input", readonly: true, value: subs.expire_time }
                                        ]
                                    }
                                ]
                            }
                        },
                        {
                            id: logvid,
                            maxHeight: 2000,
                            view:"template",
                            css:"web-console",
                            borderless: true,
                            scroll:"y",
                            template:""
                        }
                    ]
                },
                {
                    height:36,
                    cols: [{},
                        {
                            view: "button", type: "form", width: 80, icon: "check-circle", label: "PAP 认证", click: function () {
                                var btn = this;
                                btn.disable();
                                var params = {username:subs.subscriber,papchap:"pap"}
                                webix.ajax().get('/admin/radius/auth/test', params).then(function (iresult) {
                                    btn.enable();
                                    updateLog(iresult);
                                });
                            }
                        },
                        {
                            view: "button", type: "form", width: 80, icon: "check-circle", label: "CHAP 认证", click: function () {
                                var btn = this;
                                btn.disable();
                                var params = {username:subs.subscriber,papchap:"pap"}
                                webix.ajax().get('/admin/radius/auth/test', params).then(function (iresult) {
                                    btn.enable();
                                    updateLog(iresult);
                                });
                            }
                        },
                        {
                            view: "button", type: "form", width: 80, icon: "check-circle", label: "上线", click: function () {
                                var btn = this;
                                btn.disable();
                                var params = {username:subs.subscriber,type:"1"}
                                webix.ajax().get('/admin/radius/acct/test', params).then(function (iresult) {
                                    btn.enable();
                                    updateLog(iresult);
                                });
                            }
                        },
                        {
                            view: "button", type: "form", width: 80, icon: "check-circle", label: "更新", click: function () {
                                var btn = this;
                                btn.disable();
                                var params = {username:subs.subscriber,type:"3"}
                                webix.ajax().get('/admin/radius/acct/test', params).then(function (iresult) {
                                    btn.enable();
                                    updateLog(iresult);
                                });
                            }
                        },
                        {
                            view: "button", type: "form", width: 80, icon: "check-circle", label: "下线", click: function () {
                                var btn = this;
                                btn.disable();
                                var params = {username:subs.subscriber,type:"2"}
                                webix.ajax().get('/admin/radius/acct/test', params).then(function (iresult) {
                                    btn.enable();
                                    updateLog(iresult);
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