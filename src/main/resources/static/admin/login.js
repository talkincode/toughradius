webix.ready(function() {
    var logviewId = webix.uid();
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
                    rows:[
                        {
                            maxWidth:340,
                            width:320,
                            height:90,
                            cols:[
                                {css:"login-logo"}
                            ]
                        },
                        {
                            id:logviewId,
                            view: "form",
                            scroll: false,
                            maxWidth: 340,
                            autoHeight: true,
                            paddingX: 30,
                            elements: [
                                {view: "text",name:"username", value: '', placeholder: "帐 号", height:35},
                                {view: "text", name:"password",type: 'password', value: '', placeholder: "密 码",height:35},
                                {
                                    margin: 3, cols: [
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
                                {height:20}
                            ]
                        }
                    ]
                }
            },
            {gravity: 2}
        ]
    });
});
