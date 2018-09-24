if (!window.xspeedercloud.customer.auth_user)
    xspeedercloud.customer.auth_user={};


xspeedercloud.customer.auth_user.loadPage = function(session){
    xspeedercloud.customer.methods.setToolbar("cube","用户管理","auth_user");
    var tableid = webix.uid();
    var queryid = webix.uid();
    var reloadData = function(){

        var params = $$(queryid).getValues();
        var args = [];
        for(var k in params){
            args.push(k+"="+params[k]);
        }

        $$(tableid).clearAll();
        $$(tableid).load("/customer/auth/user/query?"+args.join("&"),"json");
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
                maxWidth: 4000,
                borderless:true,
                hotkey:"enter",
                elements: [
                    {
                        margin:10,
                        cols:[
                            {view: "text", name: "keyword", label: "关键字",  placeholder: "名称/备注",labelWidth:50, maxWidth:300},
                            {view: "button", label: "查询", type: "icon", icon: "search", borderless: true, width: 55,click:function(){
                                reloadData();
                            }},
                            {view: "button", label: "重置", type: "icon", icon: "refresh", borderless: true, width: 55,click:function(){
                                $$(queryid).setValues({
                                    level: "all",
                                    keyword: ""
                                });
                            }},{},
                            { view:"button", type:"form", width:70, icon:"plus", label:"添加",  click:function(){
                                    xspeedercloud.customer.auth_user.addUserForm(session,function(){
                                        reloadData();
                                    });
                                }},
                            { view:"button", type:"form",  width:70,icon:"edit", label:"修改", click:function(){
                                    var item = $$(tableid).getSelectedItem();
                                    if(item){
                                        xspeedercloud.customer.auth_user.editUserForm(session, item,function(){
                                            reloadData();
                                        });
                                    }else{
                                        webix.message({type: 'error', text: "请选择一项", expire: 1500});
                                    }
                                }},
                            { view:"button",  type:"danger",  width:70, icon:"times",label:"删除", click:function(){
                                    var item = $$(tableid).getSelectedItem();
                                    if(item){
                                        xspeedercloud.customer.auth_user.deleteUser(item,function(){
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
                leftSplit: 1,
                rightSplit: 1,
                columns:[
                    { id: "id", header: ["ID"], width: 60, sort: "string" },
                    { id:"fullname",header:["全名"], width:120, sort:"string"},
                    { id:"mobile",header:["用户手机"], width:120, sort:"string"},
                    { id:"email",header:["用户邮箱"], width:150, sort:"string"},
                    { id:"username",header:["账号名"], sort:"string",width:120},
                    { id:"ipAddr",header:["IP地址"], sort:"string",width:140},
                    { id:"addrPool",header:["地址池"], sort:"string",width:120},
                    { id:"status",header:["状态"], sort:"string",width:100},
                    { id:"groupPolicy",header:["启用组策略"], sort:"string",width:120},
                    { id:"remark",header:["备注"], sort:"string",width:160},
                    { id:"",header:[""], fillspace:true},
                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                ],
                select:true,
                maxWidth:4000,
                maxHeight:4000,
                resizeColumn:true,
                autoWidth:true,
                autoHeight:true,
                url:"/customer/auth/user/query",
                on:{
                    onItemDblClick: function(id, e, node){
                        console.log(this.getSelectedItem());
                        xspeedercloud.customer.auth_user.editUserForm(session,this.getSelectedItem(),function(){
                            reloadData();
                        });
                    }
                },
                pager: "dataPager"
            },
            {
                paddingY: 3,
                cols:[
                    {
                        id:"dataPager", view: 'pager', master:false, size: 20, nas: 7,
                        template: '{common.first()} {common.prev()} {common.pages()} {common.next()} {common.last()}'
                    }
                ]
            }

        ]
    },$$(xspeedercloud.customer.pageId),$$(xspeedercloud.customer.panelId));
};



xspeedercloud.customer.auth_user.addUserForm = function(session,callback){
    var winid = "addUserForm";
    if($$(winid))
        return;
    var formid = winid+"_form";
    webix.ui({
        id:winid,
        view: "window",css:"win-body",
        move:true,
        width:640,
        height:580,
        position: "center",
        head: {
            view: "toolbar",
            css:"win-toolbar",
            margin: -4,
            cols: [
                {view: "icon", icon: "user", css: "alter"},
                {view: "label", label: "创建认证用户"},
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
                    scroll: false,
                    maxWidth: 2000,
                    maxHeight: 2000,
                    elementsConfig: {},
                    elements: [
                        {view: "text", name: "fullname", label: "全名", placeholder: "全名", validate:webix.rules.isNotEmpty},
                        {view: "text", name: "mobile", label: "用户手机", placeholder: "用户手机"},
                        {view: "text", name: "email", label: "用户邮箱", placeholder: "用户邮箱" },
                        {view: "text", name: "username", label: "账号名", placeholder: "账号名" },
                        {view: "text", name: "password", label: "密码", placeholder: "密码" },
                        {view: "text", name: "radiusAttrs", label: "策略", placeholder: "策略" },
                        {view: "text", name: "ipAddr", label: "IP地址", placeholder: "IP地址" },
                        {view: "text", name: "addrPool", label: "地址池", placeholder: "地址池" },
                        {view:"radio", name:"status", label: "状态",value:"enabled", options:[{id:'enabled',value:"启用"}, {id:'disabled',value:"停用"}]},
                        {view:"radio", name:"groupPolicy", label: "启用组策略",value:"enabled", options:[{id:'enabled',value:"启用"}, {id:'disabled',value:"停用"}]},
                        {view: "textarea", name: "remark", label: "备注", placeholder: "备注", height:80},

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
                            width: 90,
                            height:36,
                            click: function () {
                                if (!$$(formid).validate()){
                                    webix.message({type: "error", text:"请正确填写数据",expire:1000});
                                    return false;
                                }
                                var btn = this;
                                webix.ajax().post('/customer/auth/user/add', $$(formid).getValues()).then(function (result) {
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
            ]
        }
    }).show();
};



xspeedercloud.customer.auth_user.editUserForm = function(session,item,callback){
    var winid = "editUserForm";
    var pattrs = [];
    if($$(winid))
        return;
    var formid = winid+"_form";
    webix.ajax().get("/customer/auth/user/queryUser?id="+item.id).then(function (initdata) {
        iresult = initdata.json().data;
        webix.ui({
            id:winid,
            view: "window",css:"win-body",
            move:true,
            width:640,
            height:580,
            position: "center",
            head: {
                view: "toolbar",
                css:"win-toolbar",
                margin: -4,
                cols: [
                    {view: "icon", icon: "user", css: "alter"},
                    {view: "label", label: "修改认证用户"},
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
                            {view: "text", name: "fullname", label: "全名", placeholder: "全名", validate:webix.rules.isNotEmpty,value:iresult.fullname},
                            {view: "text", name: "mobile", label: "用户手机", placeholder: "用户手机",value:iresult.mobile},
                            {view: "text", name: "email", label: "用户邮箱", placeholder: "用户邮箱" ,value:iresult.email},
                            {view: "text", name: "username", label: "账号名", placeholder: "账号名" ,value:iresult.username},
                            {view: "text", name: "password", label: "密码", placeholder: "密码" ,value:iresult.password},
                            {view: "text", name: "radiusAttrs", label: "策略", placeholder: "策略",value:iresult.radiusAttrs },
                            {view: "text", name: "ipAddr", label: "IP地址", placeholder: "IP地址",value:iresult.ipAddr },
                            {view: "text", name: "addrPool", label: "地址池", placeholder: "地址池" ,value:iresult.addrPool},
                            {view:"radio", name:"status", label: "状态", value:iresult.status, options:[{id:'enabled',value:"启用"}, {id:'disabled',value:"停用"}]},
                            {view:"radio", name:"groupPolicy", label: "启用组策略", value:iresult.groupPolicy, options:[{id:'enabled',value:"启用"}, {id:'disabled',value:"停用"}]},
                            {view: "textarea", name: "remark", label: "备注", placeholder: "备注", height:80},

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
                                    webix.ajax().post('/customer/auth/user/update',param).then(function (result) {
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


xspeedercloud.customer.auth_user.deleteUser = function (item,callback) {
    webix.confirm({
        title: "操作确认",
        ok: "是", cancel: "否",
        text: "确认要删除吗，此操作不可逆。",
        callback: function (ev) {
            if (ev) {
                webix.ajax().get('/customer/auth/user/delete', {id: item.id}).then(function (result) {
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