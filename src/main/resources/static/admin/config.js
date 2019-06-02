if (!window.toughradius.admin.config)
    toughradius.admin.config={};


toughradius.admin.config.loadPage = function(session){
    var cview = {
        id:"toughradius.admin.config",
        css:"main-panel",
        padding:10,
        // borderless:true,
        view:"tabview",
        tabbar:{
            optionWidth:160,
        },
        cells:[
            {
                header:"RADIUS 配置",
                body:{
                    id: "radius_settings",
                    view: "form",
                    paddingX:10,
                    elementsConfig: {
                        labelWidth:160,
                        // labelPosition:"top"
                    },
                    url:"/admin/config/load/radius",
                    elements: [
                        { view: "fieldset", label: "基本设置",  body: {
                            rows:[
                                {view: "counter", name: "radiusInterimIntelval", label: "记账间隔(秒)",  value:300},
                                {view: "counter", name: "radiusTicketHistoryDays", label: "上网日志保存最大天数",  value:180},
                                {view: "radio", name: "radiusIgnorePassword", label: "免密码认证:",value:"disabled", options:[{id:"disabled",value:"关闭"},{id:"enabled",value:'开启'}]},
                                {view: "text", name: "radiusExpireAddrPool", label: "到期下发地址池"},
                            ]
                        }},
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
                        },{}
                    ]
                }
            },
            {
                header:"短信配置",
                body:{
                    id: "sms_settings",
                    view: "form",
                    paddingX:10,
                    elementsConfig: {
                        labelWidth:160,
                        // labelPosition:"top"
                    },
                    url:"/admin/config/load/sms",
                    elements: [
                        { view: "fieldset", label: "短信网关",  body: {
                            rows:[
                                {view: "richselect", name: "smsGateway", label: "短信网关:",value:"qcloud", options:[{id:"qcloud",value:"腾讯云短信"}]},
                                {view: "text", name: "smsAppid", label: "短信网关APPID"},
                                {view: "text", type:"password", name: "smsAppkey", label: "短信网关APPKEY"},
                                {view: "text", name: "smsVcodeTemplate", label: "短信验证码模板"}
                            ]
                        }},
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
                        },{}
                    ]
                }
            },
            {
                header:"无线认证",
                body:{
                    id: "wlan_settings",
                    view: "form",
                    paddingX:10,
                    elementsConfig: {
                        labelWidth:160,
                        // labelPosition:"top"
                    },
                    url:"/admin/config/load/wlan",
                    elements: [
                        { view: "fieldset", label: "基本设置",  body: {
                            rows:[
                                {view: "richselect", name: "wlanTemplate", label: "认证模板:",value:"default", options:[{id:"default",value:"默认"}]},
                                {view: "text", name: "wlanJoinUrl", label: "自助服务网站"},
                                {view: "text", name: "wlanResultUrl", label: "认证成功跳转"},
                                {view:"radio", name:"wlanUserauthEnabled", label: "用户密码认证",  options:[{id:'enabled',value:"开启"}, {id:'disabled',value:"关闭"}]},
                                {view:"radio", name:"wlanPwdauthEnabled", label: "固定密码认证",  options:[{id:'enabled',value:"开启"}, {id:'disabled',value:"关闭"}]},
                                {view:"radio", name:"wlanSmsauthEnabled", label: "手机短信认证",  options:[{id:'enabled',value:"开启"}, {id:'disabled',value:"关闭"}]},
                                {view:"radio", name:"wlanWxauthEnabled", label: "微信连WiFi认证",  options:[{id:'enabled',value:"开启"}, {id:'disabled',value:"关闭"}]},
                            ]
                        }},
                        { view: "fieldset", label: "微信连WiFi设置",  body: {
                            rows:[
                                {view: "text", name: "wlanWechatSsid", label: "SSID"},
                                {view: "text", name: "wlanWechatShopid", label: "门店ID"},
                                {view: "text", name: "wlanWechatAppid", label: "APPID"},
                                {view: "text", name: "wlanWechatSecretkey", label: "APP密钥"}
                            ]
                        }},
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
                                {view: "button", name: "preview", type: "base", value: "预览界面", width: 120, height:36, click: function () {
                                        toughradius.admin.config.openWlanTemplate();
                                    }
                                },
                                {}
                            ]
                        },{}
                    ]
                }
            },
            {
                header:"API 设置",
                body:{
                    id: "api_settings",
                    view: "form",
                    paddingX:10,
                    elementsConfig: {
                        labelWidth:160,
                        // labelPosition:"top"
                    },
                    url:"/admin/config/load/api",
                    elements: [
                        { view: "fieldset", label: "API",  body: {
                                rows:[
                                    {view: "richselect", name: "apiType", label: "API 类型:",value:"basic", options:[{id:"basic",value:"Basic"}]},
                                    {view: "text", name: "apiUsername", label: "Basic 用户"},
                                    {view: "text", type:"password", name: "apiPasswd", label: "Basic 密码"},
                                    {view: "text", name: "apiAllowIplist", label: "IP 白名单"},
                                    {view: "text", name: "apiBlackIplist", label: "IP 黑名单"}
                                ]
                            }},
                        {
                            cols: [
                                {view: "button", name: "submit", type: "form", value: "保存配置", width: 120, height:36, click: function () {
                                        if (!$$("radius_settings").validate()){
                                            webix.message({type: "error", text:"请正确填写",expire:1000});
                                            return false;
                                        }
                                        var param =  $$("api_settings").getValues();
                                        param['ctype'] = 'sms';
                                        webix.ajax().post('/admin/config/api/update',param).then(function (result) {
                                            var resp = result.json();
                                            webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
                                        });
                                    }
                                },
                                {}
                            ]
                        },{}
                    ]
                }
            },
        ]


    };
    toughradius.admin.methods.addTabView("toughradius.admin.config","cogs","系统配置", cview, true);
};


toughradius.admin.config.openWlanTemplate = function(){
    var winid = "toughradius.admin.config.openWlanTemplate";
    if($$(winid))
        return;
    webix.ui({
        id:winid,
        view: "window",css:"win-body",
        move:true,
        width:360,
        height:640,
        position: "center",
        head: {
            view: "toolbar",
            css:"win-toolbar",
            cols: [
                {view: "icon", icon: "laptop", css: "alter"},
                {view: "label", label: "预览模板"},
                {view: "icon", icon: "times-circle", css: "alter", click: function(){
                        $$(winid).close();
                    }}
            ]
        },
        body:{
            view:"iframe",
            height:640,
            id:"frame-body",
            src:"/wlandemo"
        }
    }).show();
};


