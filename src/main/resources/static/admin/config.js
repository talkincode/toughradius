if (!window.toughradius.admin.config)
    toughradius.admin.config={};


toughradius.admin.config.loadPage = function(session){
    var cview = {
        id:"toughradius.admin.config",
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
                                    labelPosition:"top"
                                },
                                url:"/admin/config/load/radius",
                                elements: [

                                    {view: "counter", name: "RADIUS_INTERIM_INTELVAL", label: "RADIUS 记账间隔(秒)",  value:300},
                                    {view: "counter", name: "RADIUS_MAX_SESSION_TIMEOUT", label: "RADIUS 最大会话时长(秒)",  value:864000},
                                    {view: "counter", name: "RADIUS_TICKET_HISTORY_DAYS", label: "RADIUS 上网日志保存最大天数",  value:180},
                                    {view: "richselect", name: "RADIUS_IGNORE_PASSWORD", label: "RADIUS 免密码认证:",value:"0", options:[{id:"1",value:"否"},{id:"0",value:'是'}], width:220},
                                    {view: "richselect", name:"RADIUS_ONLINE_EXPIRE_CHECK", label: "RADIUS 在线过期定时清理",  options:[{id:'enabled',value:"启用"}, {id:'disabled',value:"停用"}],width:220},
                                    {view: "text", name: "RADIUS_EXPORE_ADDR_POOL", label: "RADIUS 到期下发地址池",  width:220},
                                    {
                                        cols: [
                                            {view: "button", name: "submit", type: "form", value: "保存配置", width: 120, height:36, click: function () {
                                                    if (!$$("radius_settings").validate()){
                                                        webix.message({type: "error", text:"请正确填写",expire:1000});
                                                        return false;
                                                    }
                                                    var param =  $$("radius_settings").getValues();
                                                    param['ctype'] = 'radius';
                                                    webix.ajax().post('/admin/config/radius/update',param).then(function (result) {
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

    };
    toughradius.admin.methods.addTabView("toughradius.admin.config","cogs","系统配置", cview, true);
};



