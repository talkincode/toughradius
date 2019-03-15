if (!window.toughradius.admin.config)
    toughradius.admin.config={};


toughradius.admin.config.loadPage = function(session){
    toughradius.admin.methods.setToolbar("cogs","系统配置","config");
    var tableid = webix.uid();
    var reloadData = function(){
        $$(tableid).clearAll();
        $$(tableid).load("/admin/config/query");
        $$(tableid).refresh()
    };
    webix.ui({
        id:toughradius.admin.panelId,
        view:"scrollview",
        css:"main-panel",padding:2,
        body:{
            rows:[
                {
                    borderless:true,
                    css:"config-tabs",
                    view:"tabview",
                    cells:[
                        {
                            header:"RADIUS 设置",
                            body:{
                                id: "radius_settings",
                                view: "form",
                                paddingX:30,
                                maxHeight:4096,
                                maxWidth:4096,
                                elementsConfig: {
                                    labelWidth:200,
                                    labelPosition:"left"
                                },
                                url:"/admin/config/load/radius",
                                elements: [

                                    {view: "counter", name: "RADIUS_INTERIM_INTELVAL", label: "Radius记账间隔(秒)",  value:300},
                                    {view: "counter", name: "RADIUS_MAX_SESSION_TIMEOUT", label: "Radius最大会话时长(秒)",  value:864000},
                                    {view: "counter", name: "RADIUS_TICKET_HISTORY_DAYS", label: "RADIUS 上网日志保存最大天数",  value:180},
                                    {view: "richselect", name: "RADIUS_IGNORE_PASSWORD", label: "RADIUS 免密码认证:",value:"0", options:[{id:"1",value:"否"},{id:"0",value:'是'}], width:480},
                                    {view: "text", name: "RADIUS_EXPORE_ADDR_POOL", label: "到期用户下发地址池",  width:480},
                                    {view:"radio", name:"RADIUS_SYNC_FREERADIUS", label: "freeRADIUS 同步",  options:[{id:'enabled',value:"启用"}, {id:'disabled',value:"停用"}]},
                                    {view:"radio", name:"RADIUS_ONLINE_EXPIRE_CHECK", label: "RADIUS 在线过期定时清理",  options:[{id:'enabled',value:"启用"}, {id:'disabled',value:"停用"}]},
                                    {
                                        cols: [
                                            {view: "button", name: "submit", type: "form", value: "保存配置", width: 120, height:36, click: function () {
                                                    if (!$$("radius_settings").validate()){
                                                        webix.message({type: "error", text:"请正确填写",expire:1000});
                                                        return false;
                                                    }
                                                    var param =  $$("radius_settings").getValues();
                                                    param['ctype'] = 'radius';
                                                    webix.ajax().post('/admin/config/update',param).then(function (result) {
                                                        var resp = result.json();
                                                        webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
                                                    });
                                                }
                                            },
                                            {}
                                        ]
                                    }
                                ]
                            }
                        }
                    ]

                }
            ]
        }

    },$$(toughradius.admin.panelId));
};



