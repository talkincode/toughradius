if (!window.toughsms.admin.syslog)
    toughsms.admin.syslog={};


toughsms.admin.syslog.loadPage = function(session){
    toughsms.admin.methods.setToolbar("hdd-o","系统日志","syslog");
    var tableid = webix.uid();
    var queryid = webix.uid();
    var reloadData = function(){
        $$(tableid).define("url", $$(tableid));
        $$(tableid).refresh();
        $$(tableid).clearAll();
        var params = $$(queryid).getValues();
        var args = [];
        for(var k in params){
            args.push(k+"="+params[k]);
        }
        $$(tableid).load('/admin/syslog/query?'+args.join("&"));
    };
    webix.ui({
        id:toughsms.admin.panelId,
        css:"main-panel",padding:2,
        rows:[
            {
                view:"toolbar",
                css:"page-toolbar",
                cols:[
                    {
                        id: queryid,
                        css:"query-form",
                        // height:40,
                        paddingY:5,
                        view: "form",
                        hidden: false,
                        maxWidth: 2000,
                        borderless:true,
                        elements: [
                            {
                                cols:[
                                    {view: "richselect", name: "type", label: "类型", value: "radiusd",width:140,labelWidth:40,
                                        options: [{ id: "radiusd", value: "认证" },{ id: "system", value: "系统" },
                                            { id: "api", value: "接口" },{ id: "bras", value: "设备" },{ id: "all", value: "所有" }]},
                                    {view: "datepicker", timepicker:true, name:"startDate",label:"", labelWidth:0, width:170, stringResult:true, format: "%Y-%m-%d %H:%i"},
                                    {view: "datepicker", timepicker:true, name:"endDate",label:"至", labelWidth:27, width:190, stringResult:true,format: "%Y-%m-%d %H:%i"},
                                    { view: "text", name: "username", label: "", labelWidth: 0, placeholder: "用户帐号", width:120},
                                    { view: "text", name: "keyword", label: "", labelWidth: 0, placeholder: "内容关键字", width:150},
                                    {view: "button", label: "查询", type: "icon", icon: "search", borderless: true, width: 55,click:function(){
                                        reloadData();
                                    }},
                                    {view: "button", label: "重置", type: "icon", icon: "refresh", borderless: true, width: 55,click:function(){
                                        $$(queryid).setValues({
                                            start_date: "",
                                            end_date: "",
                                            level: "",
                                            keyword: ""
                                        });
                                    }}
                                ]
                            }
                        ]
                    }
                ]
            },
            {
                id:tableid,
                view:"treetable",
                scroll:"y",
                leftSplit:1,
                subview:{
                    borderless:true,
                    view:"template",
                    height:240,
                    template:"<div style='padding: 5px;'>#msg#</div>"
                },
                on:{
                    onSubViewCreate:function(view, item){
                        item.msg = item.msg.replace("\n","<br>");
                        view.setValues(item);
                    }
                },
                columns:[
                    { id:"name", header:["帐号"], width:120,  template:"{common.subrow()} #name#"},
                    { id:"time",header:["时间"], width:180},
                    { id:"type",header:["类型"], width:70},
                    { id:"msg",header:["内容"], fillspace:true},
                ],
                select:true,
                maxWidth:2000,
                maxHeight:1000,
                resizeColumn:true,
                autoWidth:true,
                autoHeight:true,
                url:"/admin/syslog/query",
                pager: "dataPager",
                datafetch:40,
				loadahead:15,
                ready:function () {
                    reloadData();
                }
            },
            {
                paddingY: 3,
                cols:[
                    {
                        view: "richselect", name: "page_num", label: "每页显示", value: 20,width:130,labelWidth:60,
                        options: [{ id: 20, value: "20" },
                            { id: 50, value: "50" },
                            { id: 100, value: "100" },
                            { id: 500, value: "500" },
                            { id: 1000, value: "1000" }],on: {
                            onChange: function (newv, oldv) {
                                $$("dataPager").define("size",parseInt(newv));
                                $$(tableid).refresh();
                                reloadData();
                            }
                        }
                    },
                    {
                        id:"dataPager", view: 'pager', master:false, size: 20, group: 5,
                        template: '{common.first()} {common.prev()} {common.pages()} {common.next()} {common.last()} total:#count#'
                    },{}
                ]
            }
        ]
    },$$(toughsms.admin.pageId),$$(toughsms.admin.panelId));
    webix.extend($$(tableid), webix.ProgressBar);
};




