if (!window.toughradius.admin.online)
    toughradius.admin.online={};


toughradius.admin.online.loadPage = function(session){
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
        $$(tableid).load('/admin/online/query?'+args.join("&"));
    };
    var clearData = function(){
        $$(tableid).define("url", $$(tableid));
        $$(tableid).refresh();
        $$(tableid).clearAll();
        var params = $$(queryid).getValues();
        var rparam = {
            nodeId: params.node_id,
            areaId: params.area_id,
            beginTime: params.start_time,
            endTime: params.end_time,
            keyword: params.keyword
        };
        webix.ajax().get('/admin/online/clear', rparam).then(function (result) {
            var resp = result.json();
            webix.message({type: resp.msgtype, text: resp.msg, expire: 500});
            reloadData();
        })
    };
    var cview = {
        id:"toughradius.admin.online",
        css:"main-panel",padding:10,
        rows: [
            {
                view: "toolbar",
                css: "page-toolbar",
                cols: [
                    {
                        view: "button", type: "danger", width: 100, icon: "times", label: "选中批量清除",  click: function () {
                            var rows = [];
                            $$(tableid).eachRow(
                                function (row) {
                                    var item = $$(tableid).getItem(row);
                                    if (item && item.state === 1) {
                                        rows.push(item.acctSessionId)
                                    }
                                }
                            );
                            if (rows.length === 0) {
                                webix.message({ type: 'error', text: "请至少勾选一项", expire: 1500 });
                            } else {
                                toughradius.admin.online.onlineDelete(rows.join(","), function () {
                                    reloadData();
                                });
                            }
                        }
                    },
                    {
                        view: "button", type: "danger", width: 100, icon: "times", label: "选中批量踢线", click: function () {
                            var rows = [];
                            $$(tableid).eachRow(
                                function (row) {
                                    var item = $$(tableid).getItem(row);
                                    if (item && item.state === 1) {
                                        rows.push(item.acctSessionId)
                                    }
                                }
                            );
                            if (rows.length === 0) {
                                webix.message({ type: 'error', text: "请至少勾选一项", expire: 1500 });
                            } else {
                                toughradius.admin.online.onlineUnlock(rows.join(","), function () {
                                    reloadData();
                                });
                            }
                        }
                    },
                    { gravity: 4 },
                    {
                        view: "button", type: "icon", width: 70, icon: "refresh", label: "刷新", click: function () {
                            reloadData();
                        }
                    },
                ]
            },
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
                       type:"space", id:"a1", rows:[{
                         type:"space", padding:0, responsive:"a1", cols:[
                            { view: "datepicker", name: "beginTime", label: "上线时间",stringResult:true, timepicker: true, format: "%Y-%m-%d %h:%i" },
                            { view: "datepicker", name: "endTime", label: "至", labelWidth:27,stringResult:true, timepicker: true, format: "%Y-%m-%d %h:%i" },
                            {view: "text", name: "keyword", label: "关键字",  placeholder: "帐号 / IP地址 / MAC地址 .."},
                            {
                                cols:[

                                    {view: "button", label: "查询", type: "icon", icon: "search", borderless: true, width: 66,click:function(){
                                        reloadData();
                                    }},
                                    {view: "button", label: "重置", type: "icon", icon: "refresh", borderless: true, width: 66,click:function(){
                                        $$(queryid).setValues({
                                            start_time: "",
                                            end_time: "",
                                            keyword: "",
                                            area_id: "",
                                            status: ""
                                        });
                                    }},
                                    {
                                        view: "button", label: "清理", type: "icon", icon: "trash",width:66, borderless: true, click: function () {
                                              webix.confirm({
                                                title: "操作确认",
                                                ok: "是", cancel: "否",
                                                text: "确认要清除符合条件的记录吗？",
                                                callback: function (ev) {
                                                    if (ev) {
                                                        clearData()
                                                    }
                                                }
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
                rightSplit: 1,
                columns: [
                    { id: "state", header: { content: "masterCheckbox", css: "center" }, width: 50, css: "center", template: "{common.checkbox()}" },
                    { id: "realname", header: ["用户姓名"], adjust:true },
                    { id: "username", header: ["用户名"] , adjust:true ,sort: "string", },
                    { id: "acctSessionId", header: ["会话ID"], hidden:true, adjust:true  },
                    { id: "nasId", header: ["BRAS 标识"] ,sort: "string",},
                    { id: "acctStartTime", header: ["上线时间"], sort: "string", adjust:true  },
                    { id: "nasAddr", header: ["BRAS IP"]  , adjust:true },
                    { id: "framedIpaddr", header: ["用户 IP"],sort: "string", adjust:true },
                    { id: "macAddr", header: ["用户 Mac"] , sort: "string",adjust:true },
                    { id: "nasPortId", header: ["端口信息"], adjust:true },
                    {
                        id: "acctInputTotal", header: ["上传"],sort: "int", adjust:true , template: function (obj) {
                            return bytesToSize(obj.acctInputTotal);
                        }
                    },
                    {
                        id: "acctOutputTotal", header: ["下载"],  sort: "int",  adjust:true ,template: function (obj) {
                            return bytesToSize(obj.acctOutputTotal);
                        }
                    },
                    { id: "acctInputPackets", header: ["上行数据包"], sort: "int", adjust:true },
                    { id: "acctOutputPackets", header: ["下行数据包"],sort: "int", adjust:true },
                    { id: "_", header: [""],   fillspace:true},
                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                ],
                select: true,
                resizeColumn: true,
                autoWidth: true,
                autoHeight: true,
                url: "/admin/online/query",
                pager: "online_dataPager",
                datafetch: 40,
                loadahead: 15,
                on: {}
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
                                $$("online_dataPager").define("size",parseInt(newv));
                                $$(tableid).refresh();
                                reloadData();
                            }
                        }
                    },
                    {
                        id: "online_dataPager", view: 'pager', master: false, size: 20, group: 5,
                        template: '{common.first()} {common.prev()} {common.pages()} {common.next()} {common.last()} total:#count#'
                    },{}
                ]
            }
        ]
    };
    toughradius.admin.methods.addTabView("toughradius.admin.online","users","在线查询", cview, true);
    webix.extend($$(tableid), webix.ProgressBar);
};



toughradius.admin.online.onlineDelete = function (ids,callback) {
    webix.confirm({
        title: "操作确认",
        ok: "是", cancel: "否",
        text: "确认要删除吗，此操作不可逆。",
        callback: function (ev) {
            if (ev) {
                webix.ajax().get('/admin/online/delete', {ids:ids}).then(function (result) {
                    var resp = result.json();
                    webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
                    if(callback)
                        callback()
                }).fail(function (xhr) {
                    webix.message({type: 'error', text: "删除失败:" + xhr.statusText, expire: 3000});
                });
            }
        }
    });
};

toughradius.admin.online.onlineUnlock = function (ids,callback) {
    webix.confirm({
        title: "操作确认",
        ok: "是", cancel: "否",
        text: "确认要踢线吗，该操作将会向路由设备发送强制下线指令，请确认你的路由设备支持。",
        callback: function (ev) {
            if (ev) {
                webix.ajax().get('/admin/online/unlock', {ids:ids,sessionId:""}).then(function (result) {
                    var resp = result.json();
                    webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
                    if(callback)
                        callback()
                });
            }
        }
    });
};