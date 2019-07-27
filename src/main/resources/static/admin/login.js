webix.ready(function() {
    var logviewId = webix.uid();
    var reload_vcode = function(){
        var imgurl = '/admin/verify-img.jpg?v='+new Date().getTime();
        $$("login-verify-img").define("template","<img src="+imgurl+"  width='130' height='36'/>");
        $$("login-verify-img").refresh();
    };
    var doLogin = function (formValues){
        webix.ajax().post('/admin/login',formValues).then(function (result) {
            var resp = result.json();
            if (resp.code===0){
                window.location.href = "/admin";
            }else{
                webix.message({type: resp.msgtype, text:resp.msg,expire:2000});
            }
        }).fail(function (xhr) {
            webix.message({type: 'error', text: "登录失败:"+xhr.statusText,expire:2000});
        });
    };


    webix.ui({
        css:"login-page",
        rows:[
            {gravity: 1},
            {
                align: "center,middle",
                body:{
                    css:"login-win",
                    width: 360,
                    rows:[
                        {
                            id:logviewId,
                            view: "form",
                            scroll: false,
                            autoHeight: true,
                            paddingX: 30,
                            paddingY: 30,
                            elements: [
                                {
                                    height:90,
                                    cols:[
                                        {css:"login-logo"}
                                    ]
                                },
                                {view: "text",name:"username", value: '', placeholder: "帐 号", height:35},
                                {view: "text", name:"password",type: 'password', value: '', placeholder: "密 码",height:35},
                                {
                                    cols:[
                                        {view: "text",name:"verifyCode", value: '', placeholder: "请输入验证码", height:39},
                                        {view: "template", id:"login-verify-img",css: "verify-img", maxWidth:130, template: "<img" +
                                            " src='/admin/verify-img.jpg' width='130' height='36' />", height:36 , onClick:{
                                            "verify-img":function(){
                                                reload_vcode();
                                            }
                                        }},
                                    ]
                                },
                                {
                                    margin: 0, cols: [
                                        {
                                            id:"login_btn",
                                            view: "button",
                                            label: "登 录",
                                            value: "Login",
                                            type: "form",
                                            height:39,
                                            click:function(){
                                                doLogin($$(logviewId).getValues())
                                            },
                                            hotkey: "enter",
                                            on:{
                                                onKeyPress:function(){
                                                    doLogin($$(logviewId).getValues())
                                                }
                                            }
                                        }
                                    ]
                                },
                            ]
                        }
                    ]
                }
            },
            {gravity: 2}
        ]
    });
});
