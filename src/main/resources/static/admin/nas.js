if (!window.xspeedercloud.customer.auth_nas)
    xspeedercloud.customer.auth_nas={};

xspeedercloud.customer.auth_nas.loadOprPage = function(session){

};


xspeedercloud.customer.auth_nas.loadPage = function(session){
    xspeedercloud.customer.methods.setToolbar("cube","接入设备管理","auth_nas");
    var tableid = webix.uid();
    var queryid = webix.uid();
    var reloadData = function(){

        var params = $$(queryid).getValues();
        var args = [];
        for(var k in params){
            args.push(k+"="+params[k]);
        }

        $$(tableid).clearAll();
        $$(tableid).load("/customer/auth/nas/query?"+args.join("&"),"json");
    };
    webix.ui({
        id:xspeedercloud.customer.panelId,
        css:"main-panel",padding:2,
        rows:[
            {
                id: queryid,
                css:"page-toolbar",
                height:50,
                view: "form",
                hidden: false,
                maxWidth: 4000,
                borderless:true,
                elements: [
                    {
                        margin:10,
                        cols:[
                            {view: "text", name: "keyword", label: "关键字",  placeholder: "名称/备注",labelWidth:50, maxWidth:300},
                            {view: "button", label: "查询", type: "icon", icon: "search", hotkey:"enter",borderless: true, width: 55,click:function(){
                                    reloadData();
                                }},
                            {view: "button", label: "重置", type: "icon", icon: "refresh", borderless: true, width: 55,click:function(){
                                $$(queryid).setValues({
                                level: "all",
                                keyword: ""
                            });
                            }},{},
                            { view:"button", type:"form", width:70, icon:"plus", label:"添加",  click:function(){
                                    xspeedercloud.customer.auth_nas.addNasForm(session,function(){
                                        reloadData();
                                    });
                                }},
                            { view:"button", type:"form",  width:70,icon:"edit", label:"修改", click:function(){
                                    var item = $$(tableid).getSelectedItem();
                                    if(item){
                                        xspeedercloud.customer.auth_nas.editNasForm(session, item,function(){
                                            reloadData();
                                        });
                                    }else{
                                        webix.message({type: 'error', text: "请选择一项", expire: 1500});
                                    }
                                }},
                            { view:"button",  type:"danger",  width:70, icon:"times",label:"删除", click:function(){
                                    var item = $$(tableid).getSelectedItem();
                                    if(item){
                                        xspeedercloud.customer.auth_nas.deleteNas(item,function(){
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
                    { id:"name",header:["名称"], width:180, sort:"string"},
                    { id:"identifier",header:["设备标识"], width:180, sort:"string"},
                    { id:"ipaddr",header:["IP地址"], width:180, sort:"string"},
                    { id:"vendorId",header:["厂商标识"], sort:"string",fillspace:true},
                    { id:"secret",header:["共享密钥"], sort:"string",fillspace:true},
                    { id:"coaPort",header:["CoA 端口"], sort:"string",fillspace:true},
                    { id:"acPort",header:["AC 端口"], sort:"string",fillspace:true},
                    { id:"portalVendor", header:["portal类型"],fillspace:true },
                    { id:"remark",header:["备注"], sort:"string",fillspace:true}

                ],
                select:true,
                maxWidth:2000,
                maxHeight:1000,
                resizeColumn:true,
                autoWidth:true,
                autoHeight:true,
                url:"/customer/auth/nas/query",
                on:{
                    onItemDblClick: function(id, e, node){
                        console.log(this.getSelectedItem());
                        xspeedercloud.customer.auth_nas.editNasForm(session,this.getSelectedItem(),function(){
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
                        id:"dataPager", view: 'pager', master:false, size: 20, nas: 7,
                        template: '{common.first()} {common.prev()} {common.pages()} {common.next()} {common.last()}'
                    }
                ]
            }

        ]
    },$$(xspeedercloud.customer.pageId),$$(xspeedercloud.customer.panelId));
};



xspeedercloud.customer.auth_nas.addNasForm = function(session,callback){
    var winid = "addNasForm";
    if($$(winid))
        return;
    var formid = winid+"_form";
    webix.ui({
        id:winid,
        view: "window",css:"win-body",
        move:true,
        width:480,
        height:360,
        position: "center",
        head: {
            view: "toolbar",
            css:"win-toolbar",
            margin: -4,
            cols: [
                {view: "icon", icon: "user", css: "alter"},
                {view: "label", label: "创建接入设备"},
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
                        {view: "text", name: "name", label: "名称", placeholder: "认证组名称", validate:webix.rules.isNotEmpty},
                        {view: "text", name: "identifier", label: "设备标识", placeholder: "设备标识"},
                        {view: "text", name: "ipaddr", label: "IP地址", placeholder: "IP地址" },
                        {view: "text", name: "vendorId", label: "厂商标识", placeholder: "厂商标识" },
                        {view:"richselect", name:"vendorId", label: "portal类型", value:"14988",
                            options:[{id:'0',value:"标准"}, {id:'9',value:"思科"}, {id:'3902',value:"中兴"}, {id:'2011',value:"华为"}, {id:'14988',value:"神行者"}]},
                        {view: "text", name: "secret", label: "共享密钥", placeholder: "共享密钥" },
                        {view: "text", name: "coaPort", label: "CoA 端口", placeholder: "CoA 端口" },
                        {view: "text", name: "acPort", label: "AC 端口", placeholder: "AC 端口" },
                        {view:"radio", name:"portalVendor", label: "portal类型",  options:[{id:'cmccv1',value:"cmccv1"}, {id:'cmccv2',value:"cmccv2"}, {id:'huaweiv1',value:"huaweiv1"}, {id:'huaweiv2',value:"huaweiv2"}]},
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
                            width: 120,
                            height:36,
                            click: function () {
                                if (!$$(formid).validate()){
                                    webix.message({type: "error", text:"请正确填写数据",expire:1000});
                                    return false;
                                }
                                var btn = this;
                                webix.ajax().post('/customer/auth/nas/add', $$(formid).getValues()).then(function (result) {
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



xspeedercloud.customer.auth_nas.editNasForm = function(session,item,callback){
    var winid = "editNasForm";
    var pattrs = [];
    if($$(winid))
        return;
    var formid = winid+"_form";
    webix.ajax().get("/customer/auth/nas/queryNas?id="+item.id).then(function (initdata) {
        iresult = initdata.json().data;
        webix.ui({
            id:winid,
            view: "window",css:"win-body",
            move:true,
            width:640,
            height:510,
            position: "center",
            head: {
                view: "toolbar",
                css:"win-toolbar",
                margin: -4,
                cols: [
                    {view: "icon", icon: "user", css: "alter"},
                    {view: "label", label: "修改接入设备"},
                    {view: "icon", icon: "times-circle", css: "alter", click: function(){
                            $$(winid).close();
                        }}
                ]
            },
            body:
                {
                    rows:[
                        {
                            id: formid,
                            view: "form",
                            scroll: false,
                            maxWidth: 2000,
                            maxHeight: 2000,
                            elementsConfig: {},
                            elements: [
                                {view: "text", name: "name", label: "名称", placeholder: "认证组名称", validate:webix.rules.isNotEmpty,value:iresult.name},
                                {view: "text", name: "identifier", label: "设备标识", placeholder: "设备标识",value:iresult.identifier},
                                {view: "text", name: "ipaddr", label: "IP地址", placeholder: "IP地址" ,value:iresult.ipaddr},
                                {view:"richselect", name:"vendorId", label: "portal类型", value:iresult.vendorId,
                                    options:[{id:'0',value:"标准"}, {id:'9',value:"思科"}, {id:'3902',value:"中兴"}, {id:'2011',value:"华为"}, {id:'14988',value:"神行者"}]},
                                {view: "text", name: "secret", label: "共享密钥", placeholder: "共享密钥",value:iresult.secret },
                                {view: "text", name: "coaPort", label: "CoA 端口", placeholder: "CoA 端口" ,value:iresult.coaPort},
                                {view: "text", name: "acPort", label: "AC 端口", placeholder: "AC 端口" ,value:iresult.acPort},
                                {view:"radio", name:"portalVendor", label: "portal类型",  options:[{id:'cmccv1',value:"cmccv1"}, {id:'cmccv2',value:"cmccv2"}, {id:'huaweiv1',value:"huaweiv1"}, {id:'huaweiv2',value:"huaweiv2"}],value:iresult.portalVendor},
                                {view: "textarea", name: "remark", label: "备注", placeholder: "备注", height:80,value:iresult.remark},

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
                                        webix.ajax().post('/customer/auth/nas/update',param).then(function (result) {
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


xspeedercloud.customer.auth_nas.deleteNas = function (item,callback) {
    webix.confirm({
        title: "操作确认",
        ok: "是", cancel: "否",
        text: "确认要删除吗，此操作不可逆。",
        callback: function (ev) {
            if (ev) {
                webix.ajax().get('/customer/auth/nas/delete', {id: item.id}).then(function (result) {
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