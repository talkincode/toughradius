if (!window.toughradius.admin.group)
    toughradius.admin.group={};

toughradius.admin.group.loadPage = function(session){
    toughradius.admin.methods.setToolbar("cube","用户组管理","auth_group");
    var tableid = webix.uid();
    var queryid = webix.uid();
    var reloadData = function(){

        var params = $$(queryid).getValues();
        var args = [];
        for(var k in params){
            args.push(k+"="+params[k]);
        }

        $$(tableid).clearAll();
        $$(tableid).load("/customer/auth/group/query?"+args.join("&"),"json");
    };
    webix.ui({
        id:toughradius.admin.panelId,
        css:"main-panel",padding:2,
        rows:[
            {
                id: queryid,
                css:"page-toolbar",
                height:50,
                view: "form",
                hidden: false,
                maxWidth: 2000,
                borderless:true,
                elements: [
                    {
                        margin:10,
                        cols:[
                            {view: "text", name: "keyword", label: "关键字",  placeholder: "组名称/描述", maxWidth:300},
                            {view: "button", label: "查询", type: "icon", icon: "search", hotkey:"enter",borderless: true, width: 55,click:function(){
                                    reloadData();
                                }},
                            {view: "button", label: "重置", type: "icon", icon: "refresh", borderless: true, width: 55,click:function(){
                                $$(queryid).setValues({
                                    level: "all",
                                    keyword: ""
                                });
                            }},
                            {},
                            { view:"button", type:"form", width:70, icon:"plus", label:"添加",  click:function(){
                                    toughradius.admin.group.addGroupForm(session,function(){
                                        reloadData();
                                    });
                                }},
                            { view:"button", type:"form",  width:70,icon:"edit", label:"修改", click:function(){
                                    var item = $$(tableid).getSelectedItem();
                                    if(item){
                                        toughradius.admin.group.editGroupForm(session, item,function(){
                                            reloadData();
                                        });
                                    }else{
                                        webix.message({type: 'error', text: "请选择一项", expire: 1500});
                                    }
                                }},
                            { view:"button",  type:"danger",  width:70, icon:"times",label:"删除", click:function(){
                                    var item = $$(tableid).getSelectedItem();
                                    if(item){
                                        toughradius.admin.group.deleteGroup(item,function(){
                                            reloadData();
                                        });
                                    }else{
                                        webix.message({type: 'error', text: "请选择一项", expire: 1500});
                                    }
                                }}

                        ]
                    }
                ]
            },
            {
                id:tableid,
                view:"datatable",
                columns:[
                    { id: "id", header: ["ID"], width: 60, sort: "string" },
                    { id:"name",header:["组名称"], width:180, sort:"string"},
                    { id:"radiusAttrs",header:["radius策略"], width:180, sort:"string"},
                    { id:"onlineSum",header:["在线数"], sort:"string",fillspace:true},
                    { id:"addrPool",header:["地址池"], sort:"string",fillspace:true},
                    { id:"upRate",header:["上行速率bps"], sort:"string",fillspace:true},
                    { id:"downRate",header:["下行速率bps"], sort:"string",fillspace:true},
                    { id:"remark",header:["描述"], sort:"string",fillspace:true}

                ],
                select:true,
                maxWidth:2000,
                maxHeight:1000,
                resizeColumn:true,
                autoWidth:true,
                autoHeight:true,
                url:"/customer/auth/group/query",
                on:{
                    onItemDblClick: function(id, e, node){
                        console.log(this.getSelectedItem());
                        toughradius.admin.group.editGroupForm(session,this.getSelectedItem(),function(){
                            reloadData();
                        });
                    }
                },
                pager: "dataPager",
            },
            {
                paddingY: 3,
                cols:[
                    {
                        id:"dataPager", view: 'pager', master:false, size: 20, group: 7,
                        template: '{common.first()} {common.prev()} {common.pages()} {common.next()} {common.last()}'
                    }
                ]
            }

        ]
    },$$(toughradius.admin.panelId));
};



toughradius.admin.group.addGroupForm = function(session,callback){
    var winid = "addGroupForm";
    if($$(winid))
        return;
    var formid = winid+"_form";
    webix.ui({
        id:winid,
        view: "window",css:"win-body",
        move:true,
        width:640,
        height:480,
        position: "center",
        head: {
            view: "toolbar",
            css:"win-toolbar",
            margin: -4,
            cols: [
                {view: "icon", icon: "user", css: "alter"},
                {view: "label", label: "创建认证组"},
                {view: "icon", icon: "times-circle", css: "alter", click: function(){
                        $$(winid).close();
                    }}
            ]
        },
        body:{
            rows:[
            {
            id: formid,
            view: "form",
            scroll: false,
            maxWidth: 2000,
            maxHeight: 2000,
            elementsConfig: {},
            elements: [
                {view: "text", name: "name", label: "名称", placeholder: "认证组名称", validate:webix.rules.isNotEmpty},
                {view: "text", name: "radiusAttrs", label: "扩展策略", placeholder: "扩展策略"},
                {view: "text", name: "addrPool", label: "地址池", placeholder: "地址池" },
                {view: "text", name: "upRate", label: "上行速率", placeholder: "上行速率" },
                {view: "text", name: "downRate", label: "上行速率", placeholder: "上行速率" },
                {view: "textarea", name: "remark", label: "备注", placeholder: "备注", height:80},


            ]
            },
            {
                padding:5,
                cols: [
                    {},
                    {view: "button", name: "submit", type: "form", value: "提交数据", width: 90, height:36,
                        click: function () {
                            if (!$$(formid).validate()){
                                webix.message({type: "error", text:"请正确填写数据",expire:1000});
                                return false;
                            }
                            var btn = this;
                            webix.ajax().post('/customer/auth/group/add', $$(formid).getValues()).then(function (result) {
                                btn.enable();
                                var resp = result.json();
                                webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
                                if(resp.code===0){
                                    $$(winid).close();
                                    if(callback)
                                        callback()
                                }
                            }).fail(function (xhr) {
                                btn.enable();
                                webix.message({type: 'error', text: "操作失败:" + xhr.statusText, expire: 1500});
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
};



toughradius.admin.group.editGroupForm = function(session,item,callback){
    var winid = "editGroupForm";
    var pformid = winid+"_pattr";
    var pattrs = [];
    if($$(winid))
        return;
    var formid = winid+"_form";
    webix.ajax().get("/customer/auth/group/queryGroup?id="+item.id).then(function (initdata) {
        iresult = initdata.json().data;
        webix.ui({
            id:winid,
            view: "window",css:"win-body",
            move:true,
            width:480,
            height:320,
            position: "center",
            head: {
                view: "toolbar",
                css:"win-toolbar",
                margin: -4,
                cols: [
                    {view: "icon", icon: "user", css: "alter"},
                    {view: "label", label: "修改用户组"},
                    {view: "icon", icon: "times-circle", css: "alter", click: function(){
                            $$(winid).close();
                        }}
                ]
            },
            body:{
                rows:[
                    {
                        id: formid,
                        view: "form",
                        scroll: false,
                        maxWidth: 2000,
                        maxHeight: 2000,
                        elementsConfig: {},
                        elements: [
                            {view: "text", name: "name", label: "名称", placeholder: "组名称", value:iresult.name,validate:webix.rules.isNotEmpty},
                            {view: "text", name: "radiusAttrs", label: "扩展策略", placeholder: "扩展策略", value:iresult.radiusAttrs},
                            {view: "text", name: "addrPool", label: "地址池", placeholder: "地址池", value:iresult.addrPool },
                            {view: "text", name: "upRate", label: "上行速率", placeholder: "上行速率" , value:iresult.upRate},
                            {view: "text", name: "downRate", label: "上行速率", placeholder: "上行速率" , value:iresult.downRate},
                        ]
                    },
                    {
                        padding:5,
                        cols: [{},
                            {
                                view: "button",
                                name: "submit",
                                type: "form",
                                value: "提交数据",
                                // disabled:!hasPerm,
                                width: 120,
                                height:36,
                                click: function () {
                                    if (!$$(formid).validate()){
                                        webix.message({type: "error", text:"请正确填写数据",expire:1000});
                                        return false;
                                    }
                                    var btn = this;
                                    var param = $$(formid).getValues();
                                    param.id = item.id;
                                    webix.ajax().post('/customer/auth/group/update',param).then(function (result) {
                                        btn.enable();
                                        var resp = result.json();
                                        webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
                                        if(resp.code===0){
                                            $$(winid).close();
                                            if(callback)
                                                callback()
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
                ]
            }
        }).show();
    })
};


toughradius.admin.group.deleteGroup = function (item,callback) {
    webix.confirm({
        title: "操作确认",
        ok: "是", cancel: "否",
        text: "确认要删除吗，此操作不可逆。",
        callback: function (ev) {
            if (ev) {
                webix.ajax().get('/customer/auth/group/delete', {id: item.id}).then(function (result) {
                    var resp = result.json();
                    webix.message({type: resp.msgtype, text: resp.msg, expire: 500});
                    if(callback)
                        callback()
                }).fail(function (xhr) {
                    webix.message({type: 'error', text: "删除失败:" + xhr.statusText, expire: 500});
                });
            }
        }
    });
};