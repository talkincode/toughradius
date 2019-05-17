if (!window.toughradius.admin.config)
    toughradius.admin.config={};


toughradius.admin.config.loadPage = function(session){
    var cview = {
        id:"toughradius.admin.config",
        view:"scrollview",
        css:"main-panel",padding:15,
        body:{
            maxWidth:8192,
            rows:[
                {
                    borderless:true,
                    css:"config-tabs",
                    view:"tabview",
                    cells:[
                        {
                            header:"RADIUS 配置",
                            body:{
                                id: "radius_settings",
                                view: "form",
                                paddingX:30,
                                elementsConfig: {
                                    // labelWidth:200,
                                    labelPosition:"top"
                                },
                                url:"/admin/config/load/radius",
                                elements: [

                                    {view: "counter", name: "RADIUS_INTERIM_INTELVAL", label: "RADIUS 记账间隔(秒)",  value:300},
                                    {view: "counter", name: "RADIUS_TICKET_HISTORY_DAYS", label: "RADIUS 上网日志保存最大天数",  value:180},
                                    {view: "richselect", name: "RADIUS_IGNORE_PASSWORD", label: "RADIUS 免密码认证:",value:"0", options:[{id:"1",value:"否"},{id:"0",value:'是'}], width:220},
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
                        },
                        {
                            header:"短信网关",
                            body:{
                                id: "sms_settings",
                                view: "form",
                                paddingX:30,
                                elementsConfig: {
                                    // labelWidth:200,
                                    labelPosition:"top"
                                },
                                url:"/admin/config/load/sms",
                                elements: [
                                    {view: "richselect", name: "SMS_GATEWAY", label: "短信网关:",value:"qcloud", options:[{id:"qcloud",value:"腾讯云短信"}], width:220},
                                    {view: "text", name: "SMS_APPID", label: "短信网关APPID",  width:220},
                                    {view: "text", type:"password", name: "SMS_APPKEY", label: "短信网关APPKEY",  width:220},
                                    {
                                        cols: [
                                            {view: "button", name: "submit", type: "form", value: "保存配置", width: 120, height:36, click: function () {
                                                    if (!$$("radius_settings").validate()){
                                                        webix.message({type: "error", text:"请正确填写",expire:1000});
                                                        return false;
                                                    }
                                                    var param =  $$("sms_settings").getValues();
                                                    param['ctype'] = 'sms';
                                                    webix.ajax().post('/admin/config/sms/update',param).then(function (result) {
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
                        },
                        {
                            header:"无线认证",
                            body:{
                                id: "wlan_settings",
                                view: "form",
                                paddingX:30,
                                elementsConfig: {
                                    // labelWidth:200,
                                    labelPosition:"top"
                                },
                                url:"/admin/config/load/sms",
                                elements: [
                                    {view: "text", name: "WLAN_WECHAT_SSID", label: "微信连 Wifi SSID",  width:220},
                                    {view: "text", name: "WLAN_WECHAT_SHOPID", label: "微信连 Wifi 门店ID",  width:220},
                                    {view: "text", name: "WLAN_WECHAT_APPID", label: "微信连 Wifi APPID",  width:220},
                                    {view: "text", name: "WLAN_WECHAT_SECRETKEY", label: "微信连 Wifi APP密钥",  width:220},
                                    {
                                        cols: [
                                            {view: "button", name: "submit", type: "form", value: "保存配置", width: 120, height:36, click: function () {
                                                    if (!$$("radius_settings").validate()){
                                                        webix.message({type: "error", text:"请正确填写",expire:1000});
                                                        return false;
                                                    }
                                                    var param =  $$("wlan_settings").getValues();
                                                    param['ctype'] = 'wlan';
                                                    webix.ajax().post('/admin/config/wlan/update',param).then(function (result) {
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



