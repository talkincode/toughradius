if (!window.toughradius.admin.bras)
    toughradius.admin.bras={};

toughradius.admin.bras.loadPage = function(session){
    var tableid = webix.uid();
    var reloadData = function(){
        $$(tableid).refresh();
        $$(tableid).clearAll();
        $$(tableid).load('/admin/bras/query');
    };
    var cview = {
        id:"toughradius.admin.bras",
        css:"main-panel",padding:10,
        rows:[
            {
                view:"toolbar",
                css:"page-toolbar",
                cols:[
                    { view:"button", type:"form", width:70, icon:"plus", label:"添加",  click:function(){
                        toughradius.admin.bras.brasAdd(session,function(){
                            reloadData();
                        });
                    }},
                    { view:"button", type:"form",  width:70,icon:"edit", label:"修改", click:function(){
                        var item = $$(tableid).getSelectedItem();
                        if(item){
                            toughradius.admin.bras.brasUpdate(session, item,function(){
                                reloadData();
                            });
                        }else{
                            webix.message({type: 'error', text: "请选择一项", expire: 1500});
                        }
                    }},
                    { view:"button",  type:"danger",  width:70, icon:"times",label:"删除",  click:function(){
                        var item = $$(tableid).getSelectedItem();
                        if(item){
                            toughradius.admin.bras.brasDelete(item,function(){
                                reloadData();
                            });
                        }else{
                            webix.message({type: 'error', text: "请选择一项", expire: 1500});
                        }
                    }},{},
                    { view:"button", type:"icon", width:55, icon:"refresh", label:"刷新", click:function(){
                        reloadData();
                    }}
                ]
            },
            {
                id:tableid,
                rightSplit: 1,
                view:"datatable",
                columns:[
                    { id: "id", header: ["ID"],  hidden:true},
                    { id:"name",header:["名称"],  adjust:true},
                    { id:"identifier",header:["标识"], adjust:true},
                    { id:"ipaddr",header:["IP"], adjust:true},
                    { id:"authLimit",header:["认证并发"],adjust:true},
                    { id:"acctLimit",header:["记账并发"], adjust:true},
                    { id:"vendorId",header:["厂商"], adjust:true,template:function(obj){
                        if(obj.vendorId==="0"){
                            return "标准";
                        }else if(obj.vendorId==="2352"){
                            return "爱立信";
                        }else if(obj.vendorId==="18168"){
                            return "ToughProxy";
                        }else if(obj.vendorId==="3902"){
                            return "中兴";
                        }else if(obj.vendorId==="9"){
                            return "思科";
                        }else if(obj.vendorId==="25506"){
                            return "H3C";
                        }else if(obj.vendorId==="2011"){
                            return "华为";
                        }else if(obj.vendorId==="2636"){
                            return "juniper";
                        }else if(obj.vendorId==="14988"){
                            return "Mikrotik";
                        }else if(obj.vendorId==="10055"){
                            return "爱快";
                        }
                    }},
                    { id:"coaPort",header:["COA端口"], adjust:true},
                    { id:"acPort",header:["AC端口"], adjust:true},
                    { id:"portalVendor",header:["Portal协议"], adjust:true},
                    { id:"status",header:["状态"], adjust:true, template:function(obj){
                        if(obj.status === 'enabled'){
                            return "<span style='color:green;'>正常</span>";
                        }else if(obj.status === 'disabled'){
                            return "<span style='color:red;'>停用</span>";
                        }
                    }},
                    { id:"remark",header:["备注"], sort:"string", fillspace:true},
                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                ],
                select:true,
                // maxWidth:4096,
                // maxHeight:1000,
                resizeColumn:true,
                autoWidth:true,
                autoHeight:true,
                url:"/admin/bras/query",
                on:{
                    onItemDblClick: function(id, e, node){
                        console.log(this.getSelectedItem());
                        toughradius.admin.bras.brasUpdate(session,this.getSelectedItem(),function(){
                            reloadData();
                        });
                    }
                },
            }
        ]
    };
    toughradius.admin.methods.addTabView("toughradius.admin.bras","laptop","设备管理", cview, true);
};

toughradius.admin.bras.brasAdd = function(session,callback){
    var winid = "brasAdd";
    if($$(winid))
        return;
    var formid = winid+"_form";
    webix.ui({
        id:winid,
        view: "window",css:"win-body",
        move:true,
        width:360,
        height:480,
        position: "center",
        head: {
            view: "toolbar",
            css:"win-toolbar",
            cols: [
                {view: "icon", icon: "laptop", css: "alter"},
                {view: "label", label: "创建 BRAS 设备"},
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
                    scroll: "y",
                    elementsConfig: {},
                    elements: [
                        {view:"richselect", name:"vendorId", label: "设备厂商",value:"", options:BRAS_VENDOR_OPTIONS},
                        {view: "text", name: "name", label: "名称", placeholder: "名称", validate:webix.rules.isNotEmpty},
                        {view: "text", name: "identifier", label: "标识", placeholder: "标识", validate:webix.rules.isNotEmpty},
                        {view: "text", name: "secret", label: "共享密钥", placeholder: "共享密钥",validate:webix.rules.isNotEmpty},
                        {view: "text", name: "ipaddr", label: "ip地址",value:"0.0.0.0", placeholder: "ip地址",validate:webix.rules.isNotEmpty},
                        {view: "text", name: "coaPort", label: "COA 端口", value: "3799", validate:webix.rules.isNotEmpty},
                        {view:"richselect", name:"portalVendor", label: "Portal协议",value:"", options:PORTAL_VENDOR_OPTIONS},
                        {view: "text", name: "acPort", label: "AC 端口", value: "2000"},
                        {view: "counter", name: "authLimit", label: "认证并发(*)",  value:1000, min:1, max:10000},
                        {view: "counter", name: "acctLimit", label: "记帐并发(*)",  value:1000, min:1, max:10000},
                        {view: "textarea", name: "remark", label: "备注", placeholder: "备注", height:90}
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
                                webix.ajax().post('/admin/bras/create', $$(formid).getValues()).then(function (result) {
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



toughradius.admin.bras.brasUpdate = function(session,item,callback){
    var winid = "brasUpdate";
    if($$(winid))
        return;
    var formid = winid+"_form";
    webix.ui({
        id:winid,
        view: "window",css:"win-body",
        move:true,
        width:360,
        height:480,
        position: "center",
        head: {
            view: "toolbar",
            css:"win-toolbar",

            cols: [
                {view: "icon", icon: "laptop", css: "alter"},
                {view: "label", label: "修改 BRAS 设备"},
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
                    scroll: "y",
                    elementsConfig: {},
                    elements: [
                        {view:"richselect", name:"vendorId", label: "设备厂商",value:item.vendorId, options:BRAS_VENDOR_OPTIONS},
                        {view: "text", name: "name", label: "名称", value: item.name, validate:webix.rules.isNotEmpty},
                        {view: "text", name: "identifier", label: "标识", value: item.identifier, validate:webix.rules.isNotEmpty},
                        {view: "text", name: "secret", label: "共享密钥", value: item.secret,validate:webix.rules.isNotEmpty},
                        {view: "text", name: "ipaddr", label: "ip地址", value: item.ipaddr,validate:webix.rules.isNotEmpty},
                        {view: "text", name: "coa_port", label: "COA端口", value: item.coaPort, validate:webix.rules.isNotEmpty},
                        {view:"richselect", name:"portalVendor", label: "设备厂商",value:item.portalVendor, options:PORTAL_VENDOR_OPTIONS},
                        {view: "text", name: "acPort", label: "AC 端口", value: item.acPort},
                        {view: "counter", name: "authLimit", label: "认证并发(*)",  value: item.authLimit, min:1, max:10000},
                        {view: "counter", name: "acctLimit", label: "记帐并发(*)",  value: item.acctLimit, min:1, max:10000},
                        {view:"radio", name:"status", label: "状态", value:item.status, options:[{id:'enabled',value:"启用"}, {id:'disabled',value:"停用"}]},
                        {view: "textarea", name: "remark", label: "备注", value: item.remark, height:90}

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
                                var param = $$(formid).getValues();
                                param.id = item.id;
                                webix.ajax().post('/admin/bras/update',param).then(function (result) {
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


toughradius.admin.bras.brasDelete = function (item,callback) {
    webix.confirm({
        title: "操作确认",
        ok: "是", cancel: "否",
        text: "确认要删除吗，此操作不可逆。",
        callback: function (ev) {
            if (ev) {
                webix.ajax().get('/admin/bras/delete', {id: item.id}).then(function (result) {
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